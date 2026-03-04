# Plan - go-jira アーキテクチャ設計

## ディレクトリ構成

```
go-jira/
├── cmd/
│   └── go-jira/
│       └── main.go                  # エントリーポイント
├── internal/
│   └── cli/                         # CLI層（ユーザーインターフェース）
│       ├── root.go                  # ルートコマンド定義
│       ├── config_cmd.go            # config サブコマンド群
│       ├── project_cmd.go           # project サブコマンド群
│       ├── sync_cmd.go              # sync サブコマンド
│       ├── issue_cmd.go             # issue サブコマンド群
│       ├── query_cmd.go             # query サブコマンド群
│       └── progress.go              # 進捗表示ヘルパー
├── core/                            # コアライブラリ（外部から利用可能）
│   ├── domain/
│   │   ├── models/                  # ドメインモデル
│   │   │   ├── project.go
│   │   │   ├── issue.go
│   │   │   ├── change_history.go
│   │   │   ├── snapshot.go
│   │   │   ├── metadata.go
│   │   │   └── config.go
│   │   └── repository/              # リポジトリインターフェース
│   │       ├── issue_repo.go
│   │       ├── change_history_repo.go
│   │       ├── snapshot_repo.go
│   │       ├── metadata_repo.go
│   │       └── sync_history_repo.go
│   ├── service/                     # ユースケース・ビジネスロジック
│   │   ├── config_service.go        # 設定の読み書き
│   │   ├── project_service.go       # プロジェクト管理
│   │   ├── sync_service.go          # 同期オーケストレーション
│   │   ├── issue_service.go         # 課題検索・取得
│   │   ├── query_service.go         # SQLクエリ実行
│   │   └── snapshot_service.go      # スナップショット生成
│   └── infrastructure/
│       ├── jira/                    # JIRA APIクライアント
│       │   ├── client.go            # HTTPクライアント（認証・リトライ）
│       │   ├── types.go             # JIRA APIレスポンス型
│       │   └── parser.go            # レスポンスパーサー
│       ├── database/                # DuckDBアクセス層
│       │   ├── connection.go        # DB接続管理
│       │   ├── schema.go            # DDL・マイグレーション
│       │   ├── issue_repository.go  # Issue リポジトリ実装
│       │   ├── change_history_repository.go
│       │   ├── snapshot_repository.go
│       │   ├── metadata_repository.go
│       │   └── sync_history_repository.go
│       └── config/
│           └── file_config.go       # settings.json 読み書き
├── docs/
│   ├── requirements.md
│   ├── plan.md
│   └── tasks.md
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

## 設計方針

### Core / CLI 分離

```
┌─────────────────────────────────────────────┐
│  CLI層 (internal/cli)                       │
│  - cobra によるコマンド定義                   │
│  - ユーザー入出力・進捗表示                   │
│  - core パッケージを呼び出すだけ              │
└──────────────────┬──────────────────────────┘
                   │ 呼び出し
┌──────────────────▼──────────────────────────┐
│  Core層 (core/)                             │
│  - 外部パッケージとして公開可能               │
│  ┌────────────────────────────────────────┐ │
│  │ service/ (ユースケース)                 │ │
│  │ - SyncService, IssueService 等         │ │
│  └─────────────┬──────────────────────────┘ │
│                │                            │
│  ┌─────────────▼──────────────────────────┐ │
│  │ domain/ (モデル + インターフェース)      │ │
│  │ - models: Issue, Project 等            │ │
│  │ - repository: インターフェース定義      │ │
│  └─────────────┬──────────────────────────┘ │
│                │ 実装                       │
│  ┌─────────────▼──────────────────────────┐ │
│  │ infrastructure/ (外部依存)              │ │
│  │ - jira/: JIRA REST API クライアント    │ │
│  │ - database/: DuckDB リポジトリ実装     │ │
│  │ - config/: 設定ファイル管理            │ │
│  └────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
```

**ポイント:**
- `core/` は `go-jira/core` として外部importが可能
- CLI は `internal/` に配置し、外部非公開
- `domain/repository` はインターフェースのみ定義 → テスト時にモック差し替え可能

### 依存関係の方向

```
cli → service → domain ← infrastructure
```

- `service` は `domain/repository` インターフェースに依存
- `infrastructure` は `domain/repository` インターフェースを実装
- `cli` は `service` のみに依存

## 同期シーケンス（詳細）

### ユーザー視点

```
$ go-jira sync
  ▸ Syncing project: PROJ-A
    Phase 1/4: Fetching issues...
      [=============================>     ] 850/1200 (70%)
      Checkpoint saved at issue PROJ-A-850
    Phase 2/4: Syncing metadata...
      ✓ Statuses (12), Priorities (5), IssueTypes (8)
    Phase 3/4: Generating snapshots...
      [==================>                ] 500/850 (58%)
    Phase 4/4: Verifying integrity...
      ✓ Local: 1200, JIRA: 1200 — OK
    ✓ PROJ-A completed (1200 issues, 45.2s)

  ▸ Syncing project: PROJ-B
    Phase 1/4: Fetching issues (incremental)...
      [======================================] 32/32 (100%)
    ...
    ✓ PROJ-B completed (32 issues updated, 3.1s)

  Summary:
    Projects synced: 2
    Total issues: 1232
    Duration: 48.3s
```

### 実装者視点（内部シーケンス）

```
SyncService.Execute(ctx, options)
│
├─ 1. Load settings & get enabled projects
│
├─ 2. For each enabled project:
│   │
│   ├─ 2.1 Phase: FetchIssues
│   │   │
│   │   ├─ Determine sync mode:
│   │   │   ├─ checkpoint exists? → Resume from checkpoint
│   │   │   ├─ lastSyncedAt exists? → Incremental (JQL: updated >= "date - margin")
│   │   │   └─ else → Full sync
│   │   │
│   │   ├─ Loop (paginated):
│   │   │   ├─ GET /rest/api/3/search/jql
│   │   │   │   params: jql, maxResults=100, nextPageToken, fields, expand=changelog
│   │   │   │
│   │   │   ├─ Parse response → []Issue + extract []ChangeHistoryItem from changelog
│   │   │   │
│   │   │   ├─ issueRepo.BatchInsert(issues)        // UPSERT
│   │   │   ├─ changeHistoryRepo.BatchInsert(items)  // INSERT
│   │   │   │
│   │   │   ├─ Save checkpoint → settings.json
│   │   │   │   { lastIssueUpdatedAt, lastIssueKey, itemsProcessed }
│   │   │   │
│   │   │   ├─ progressCallback(current, total)
│   │   │   │
│   │   │   └─ if nextPageToken == "" → break
│   │   │
│   │   └─ If full sync: issueRepo.MarkDeletedNotInKeys(projectID, allKeys)
│   │
│   ├─ 2.2 Phase: SyncMetadata
│   │   │
│   │   ├─ Fetch from JIRA API (6 parallel requests):
│   │   │   ├─ GET /project/{key}/statuses
│   │   │   ├─ GET /priority
│   │   │   ├─ GET /issuetype/project?projectId={id}
│   │   │   ├─ GET /project/{key}/components
│   │   │   ├─ GET /project/{key}/versions
│   │   │   └─ Extract labels from issues
│   │   │
│   │   └─ metadataRepo.Upsert*(projectID, data)
│   │
│   ├─ 2.3 Phase: GenerateSnapshots
│   │   │
│   │   ├─ snapshotRepo.BeginTransaction()
│   │   │
│   │   ├─ For each issue (batches of 100):
│   │   │   ├─ Fetch change history for issue
│   │   │   │
│   │   │   ├─ Build initial state (version=1):
│   │   │   │   └─ Current state - reverse all changes → initial state
│   │   │   │
│   │   │   ├─ Build subsequent versions (version=2,3,...):
│   │   │   │   └─ Apply each change group forward
│   │   │   │
│   │   │   ├─ Delete old snapshots → Insert new snapshots
│   │   │   │
│   │   │   └─ Save snapshot checkpoint
│   │   │
│   │   └─ snapshotRepo.CommitTransaction()
│   │       (or RollbackTransaction() on error)
│   │
│   └─ 2.4 Phase: VerifyIntegrity
│       │
│       ├─ Get JIRA total count via JQL
│       ├─ Get local count by status
│       └─ Log discrepancies (if any)
│
└─ 3. Return SyncResult[]
```

## CLI コマンド体系

```
go-jira
├── config
│   ├── init          # 初期設定（JIRA URL, 認証情報）
│   ├── show          # 現在の設定を表示
│   └── set           # 設定値を変更
├── project
│   ├── list          # プロジェクト一覧
│   ├── fetch         # JIRAからプロジェクト取得
│   ├── enable <key>  # 同期を有効化
│   └── disable <key> # 同期を無効化
├── sync
│   ├── run           # 同期実行（デフォルト: 全有効プロジェクト）
│   │   ├── --project <key>   # 特定プロジェクトのみ
│   │   └── --force           # フル同期（インクリメンタルを無視）
│   └── status        # 同期状態を表示
├── issue
│   ├── search        # 課題検索
│   │   ├── --query <text>
│   │   ├── --project <key>
│   │   ├── --status <status>
│   │   ├── --assignee <name>
│   │   └── --limit <n>
│   ├── get <key>     # 課題詳細
│   └── history <key> # 変更履歴
├── query
│   ├── exec <sql>    # SQLクエリ実行
│   │   ├── --project <key>
│   │   └── --all-projects
│   ├── schema        # スキーマ表示
│   ├── list          # 保存済みクエリ一覧
│   ├── save          # クエリ保存
│   └── delete <id>   # クエリ削除
└── version           # バージョン表示
```

## 主要ライブラリ

| 用途 | ライブラリ | 理由 |
|---|---|---|
| CLI フレームワーク | `github.com/spf13/cobra` | Go CLI のデファクト |
| DuckDB ドライバー | `github.com/marcboeker/go-duckdb` | Go用DuckDBドライバー |
| HTTP クライアント | `net/http` (標準) | 外部依存を最小化 |
| JSON 処理 | `encoding/json` (標準) | 標準で十分 |
| ログ | `log/slog` (標準) | Go 1.21+ 標準の構造化ログ |
| テスト | `testing` (標準) + `github.com/stretchr/testify` | アサーション強化 |

## エラーハンドリング方針

- カスタムエラー型を `core/domain/` に定義
- `fmt.Errorf("context: %w", err)` でラップ
- CLI層でユーザー向けメッセージに変換
- リトライ可能なエラー（ネットワーク等）とそうでないエラーを区別

## テスト戦略

- `domain/repository` インターフェースをモックしてservice層をユニットテスト
- `infrastructure/jira` は httptest でスタブサーバーを立ててテスト
- `infrastructure/database` はインメモリDuckDBでテスト
- CLI層はE2Eテスト（実際のコマンド実行）
