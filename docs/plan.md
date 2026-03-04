# Plan - go-jira アーキテクチャ設計

## ディレクトリ構成

```
go-jira/
├── cmd/
│   ├── go-jira/
│   │   └── main.go                  # CLI エントリーポイント
│   └── go-jira-server/
│       └── main.go                  # REST API サーバー エントリーポイント (v2)
├── internal/
│   ├── cli/                         # CLI層（ユーザーインターフェース）
│   │   ├── root.go                  # ルートコマンド定義
│   │   ├── config_cmd.go            # config サブコマンド群
│   │   ├── project_cmd.go           # project サブコマンド群
│   │   ├── sync_cmd.go              # sync サブコマンド
│   │   ├── issue_cmd.go             # issue サブコマンド群
│   │   ├── query_cmd.go             # query サブコマンド群
│   │   └── progress.go              # 進捗表示ヘルパー
│   └── api/                         # REST API層 (v2)
│       ├── server.go                # HTTPサーバー起動・設定
│       ├── router.go                # ルーティング定義
│       ├── middleware.go            # CORS, ログ, エラーハンドリング
│       ├── config_handler.go        # /api/config.* ハンドラー
│       ├── project_handler.go       # /api/projects.* ハンドラー
│       ├── sync_handler.go          # /api/sync.* ハンドラー + SSE進捗通知
│       ├── issue_handler.go         # /api/issues.* ハンドラー
│       ├── metadata_handler.go      # /api/metadata.* ハンドラー
│       ├── query_handler.go         # /api/sql.* ハンドラー
│       └── types.go                 # API リクエスト/レスポンス型
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
├── web/                             # Svelte フロントエンド (v3)
│   ├── src/
│   │   ├── routes/                  # SvelteKit ページ
│   │   │   ├── +page.svelte         # ダッシュボード
│   │   │   ├── projects/
│   │   │   ├── sync/
│   │   │   ├── issues/
│   │   │   ├── query/
│   │   │   ├── visualization/
│   │   │   └── settings/
│   │   ├── lib/
│   │   │   ├── api.ts               # REST API クライアント
│   │   │   ├── types.ts             # 型定義（Go の models に対応）
│   │   │   └── stores/              # Svelte stores（状態管理）
│   │   └── app.html
│   ├── package.json
│   ├── svelte.config.js
│   └── vite.config.ts
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

### Core と提供層の分離

```
┌──────────────┐  ┌──────────────┐  ┌──────────────────┐
│ CLI層         │  │ REST API層   │  │ Web フロントエンド │
│ internal/cli  │  │ internal/api │  │ web/ (Svelte)     │
│ cobra コマンド │  │ HTTP ハンドラ │  │ SvelteKit         │
└──────┬───────┘  └──────┬───────┘  └────────┬─────────┘
       │                 │                    │ HTTP
       │ Go 関数呼出     │ Go 関数呼出        │
       │                 │                    │
┌──────▼─────────────────▼────────────────────▼─────────┐
│  Core層 (core/)                                       │
│  - 全提供形態で共有されるビジネスロジック                │
│  ┌──────────────────────────────────────────────────┐ │
│  │ service/ (ユースケース)                           │ │
│  │ - SyncService, IssueService 等                   │ │
│  └───────────────────┬──────────────────────────────┘ │
│                      │                                │
│  ┌───────────────────▼──────────────────────────────┐ │
│  │ domain/ (モデル + インターフェース)                │ │
│  │ - models: Issue, Project 等                      │ │
│  │ - repository: インターフェース定義                │ │
│  └───────────────────┬──────────────────────────────┘ │
│                      │ 実装                           │
│  ┌───────────────────▼──────────────────────────────┐ │
│  │ infrastructure/ (外部依存)                        │ │
│  │ - jira/: JIRA REST API クライアント              │ │
│  │ - database/: DuckDB リポジトリ実装               │ │
│  │ - config/: 設定ファイル管理                      │ │
│  └──────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────┘
```

**ポイント:**
- `core/` は `go-jira/core` として外部importが可能
- CLI / REST API は `internal/` に配置し、外部非公開
- Web フロントエンドは REST API を経由して core にアクセス
- `domain/repository` はインターフェースのみ定義 → テスト時にモック差し替え可能
- CLI も REST API も同じ `service` 層を呼ぶだけ → ロジックの重複なし

### 依存関係の方向

```
cli ──→ service → domain ← infrastructure
api ──↗
web ──→ api (HTTP)
```

- `service` は `domain/repository` インターフェースに依存
- `infrastructure` は `domain/repository` インターフェースを実装
- `cli` と `api` は `service` のみに依存（同じインターフェース）
- `web` (Svelte) は `api` の HTTP エンドポイントに依存

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

### Go バックエンド

| 用途 | ライブラリ | 理由 |
|---|---|---|
| CLI フレームワーク | `github.com/spf13/cobra` | Go CLI のデファクト |
| DuckDB ドライバー | `github.com/marcboeker/go-duckdb` | Go用DuckDBドライバー |
| HTTP クライアント | `net/http` (標準) | 外部依存を最小化 |
| HTTP サーバー (v2) | `net/http` (標準) | Go 1.22+ のルーティング強化で十分 |
| JSON 処理 | `encoding/json` (標準) | 標準で十分 |
| ログ | `log/slog` (標準) | Go 1.21+ 標準の構造化ログ |
| テスト | `testing` (標準) + `github.com/stretchr/testify` | アサーション強化 |

### Svelte フロントエンド (v3)

| 用途 | ライブラリ | 理由 |
|---|---|---|
| フレームワーク | SvelteKit | Svelte の公式アプリフレームワーク |
| 状態管理 | Svelte stores (組み込み) | 外部ライブラリ不要 |
| HTTP通信 | fetch (標準) | SvelteKit の load 関数と統合 |
| チャート | ECharts or Chart.js | 可視化（バーンダウン、ベロシティ、CFD） |
| スタイリング | Tailwind CSS | ユーティリティファースト |

## REST API 設計 (v2)

### エンドポイント設計

既存のフロントエンド（jd）と同じ RPC スタイル API を採用。全操作 POST メソッド。

```
POST /api/config.get          → ConfigGetResponse
POST /api/config.update       → ConfigUpdateResponse
POST /api/config.initialize   → ConfigInitResponse

POST /api/projects.list       → ProjectListResponse
POST /api/projects.initialize → ProjectInitResponse
POST /api/projects.enable     → ProjectEnableResponse
POST /api/projects.disable    → ProjectDisableResponse

POST /api/sync.execute        → SyncExecuteResponse
POST /api/sync.status         → SyncStatusResponse

POST /api/issues.search       → IssueSearchResponse
POST /api/issues.get          → IssueGetResponse
POST /api/issues.history      → IssueHistoryResponse

POST /api/metadata.get        → MetadataGetResponse

POST /api/sql.execute         → SqlExecuteResponse
POST /api/sql.get-schema      → SqlGetSchemaResponse
POST /api/sql.list-queries    → SqlQueryListResponse
POST /api/sql.save-query      → SqlQuerySaveResponse
POST /api/sql.delete-query    → SqlQueryDeleteResponse
```

### 同期進捗のリアルタイム通知

```
GET /api/sync.progress → SSE (Server-Sent Events)

event: progress
data: {"projectKey":"PROJ","phase":"fetch_issues","current":50,"total":200,"message":"..."}

event: complete
data: {"projectKey":"PROJ","success":true}
```

SSE を選択する理由:
- 単方向通信で十分（サーバー→クライアント）
- HTTP/2 との相性が良い
- WebSocket より実装がシンプル
- ブラウザの EventSource API で簡単に接続可能

### API ハンドラーの実装パターン

```go
// 全ハンドラーが同じパターンに従う
func (h *SyncHandler) Execute(w http.ResponseWriter, r *http.Request) {
    var req models.SyncExecuteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, err)
        return
    }
    result, err := h.syncService.Execute(r.Context(), req)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err)
        return
    }
    writeJSON(w, result)
}
```

### サーバー起動

```
$ go-jira-server --port 8080
  Server listening on :8080
  API: http://localhost:8080/api/
```

## Svelte フロントエンド設計 (v3)

### ページ構成

```
web/src/routes/
├── +page.svelte              # ダッシュボード（同期状態, プロジェクト概要）
├── +layout.svelte            # 共通レイアウト（ナビゲーション）
├── projects/
│   ├── +page.svelte          # プロジェクト一覧・管理
│   └── [key]/
│       ├── +page.svelte      # プロジェクト詳細
│       ├── issues/+page.svelte
│       ├── query/+page.svelte
│       └── charts/+page.svelte
├── sync/+page.svelte         # 同期実行・進捗表示
├── issues/+page.svelte       # 課題検索（リスト/ボード/カレンダー）
├── query/+page.svelte        # SQLクエリエディタ
├── visualization/+page.svelte # チャート生成
└── settings/+page.svelte     # 設定管理
```

### API クライアント

```typescript
// web/src/lib/api.ts
// Go バックエンドの RPC API と1:1対応

class ApiClient {
  private baseUrl: string;

  async configGet(): Promise<ConfigGetResponse> {
    return this.post('/api/config.get', {});
  }

  async syncExecute(req: SyncExecuteRequest): Promise<SyncExecuteResponse> {
    return this.post('/api/sync.execute', req);
  }

  // SSE で同期進捗を購読
  subscribeSyncProgress(callback: (progress: SyncProgress) => void): EventSource {
    const es = new EventSource(`${this.baseUrl}/api/sync.progress`);
    es.addEventListener('progress', (e) => callback(JSON.parse(e.data)));
    return es;
  }
}
```

### 状態管理

Svelte の組み込み store を使用（外部ライブラリ不要）:

```typescript
// web/src/lib/stores/projects.ts
import { writable } from 'svelte/store';

export const projects = writable<Project[]>([]);
export const syncProgress = writable<SyncProgress | null>(null);
```

### 開発時の構成

```
[Svelte dev server :5173] → proxy /api/* → [Go API server :8080]
```

- `vite.config.ts` で `/api` プレフィックスを Go サーバーにプロキシ
- 本番時は Go サーバーが静的ファイルも配信（`web/build/` を埋め込み）

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
