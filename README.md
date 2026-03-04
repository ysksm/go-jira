# go-jira

JIRA のデータをローカルに同期し、高速な検索・SQL クエリ・分析を行う CLI ツール。

## 特徴

- JIRA データのローカル同期（インクリメンタル対応）
- DuckDB による高速 SQL クエリ
- 課題の時系列スナップショット自動生成
- チェックポイントによる中断・再開対応
- メモリ効率重視（ディスクファースト設計）
- core/CLI 分離アーキテクチャ（ライブラリとしても利用可能）

## 前提条件

- Go 1.23 以上
- JIRA Cloud アカウント + API トークン

## インストール

```bash
go install github.com/ysksm/go-jira/cmd/go-jira@latest
```

ソースからビルドする場合:

```bash
git clone https://github.com/ysksm/go-jira.git
cd go-jira
go mod tidy
go build -o go-jira ./cmd/go-jira
```

## 使い方

### 1. 初期設定

```bash
go-jira config init \
  --endpoint https://your-domain.atlassian.net \
  --username your-email@example.com \
  --api-key your-api-token
```

設定ファイルは `~/.config/go-jira/settings.json` に保存されます。

### 2. プロジェクト取得・有効化

```bash
# JIRA からプロジェクト一覧を取得
go-jira project fetch

# プロジェクト一覧を確認
go-jira project list

# 同期したいプロジェクトを有効化
go-jira project enable PROJ
```

### 3. データ同期

```bash
# 有効な全プロジェクトを同期
go-jira sync run

# 特定プロジェクトのみ同期
go-jira sync run --project PROJ

# フル同期（インクリメンタルを無視）
go-jira sync run --force

# 同期状態を確認
go-jira sync status
```

同期は以下の 4 フェーズで実行されます:

```
Phase 1/4: Fetching issues       — JIRA API から課題を取得（バッチ 100 件ずつ）
Phase 2/4: Syncing metadata      — ステータス、優先度、課題タイプ等を同期
Phase 3/4: Generating snapshots  — 変更履歴から時系列スナップショットを生成
Phase 4/4: Verifying integrity   — JIRA とローカルの件数を照合
```

### 4. 課題検索

```bash
# プロジェクト内の課題を検索
go-jira issue search --project PROJ

# 件数制限
go-jira issue search --project PROJ --limit 20

# 課題の詳細を表示
go-jira issue get PROJ-123

# 変更履歴を表示
go-jira issue history PROJ-123
```

### 5. SQL クエリ

```bash
# SQL を直接実行
go-jira query exec "SELECT key, summary, status FROM issues LIMIT 10" --project PROJ

# ステータス別の集計
go-jira query exec "SELECT status, COUNT(*) as cnt FROM issues GROUP BY status ORDER BY cnt DESC" --project PROJ

# スキーマを確認
go-jira query schema --project PROJ
```

### 6. その他

```bash
# 設定を表示
go-jira config show

# バージョン表示
go-jira version

# 詳細ログ出力
go-jira sync run --verbose

# エラーのみ出力
go-jira sync run --quiet
```

## コマンド一覧

```
go-jira
├── config
│   ├── init       初期設定（JIRA URL, 認証情報）
│   └── show       現在の設定を表示
├── project
│   ├── list       プロジェクト一覧
│   ├── fetch      JIRA からプロジェクト取得
│   ├── enable     同期を有効化
│   └── disable    同期を無効化
├── sync
│   ├── run        同期実行（--project, --force）
│   └── status     同期状態を表示
├── issue
│   ├── search     課題検索（--project, --limit, --offset）
│   ├── get        課題詳細
│   └── history    変更履歴
├── query
│   ├── exec       SQL クエリ実行（--project, --limit）
│   └── schema     スキーマ表示
└── version        バージョン表示
```

## データベーススキーマ

プロジェクトごとに DuckDB ファイルが作成されます（`~/.local/share/go-jira/data/{PROJECT_KEY}/data.duckdb`）。

| テーブル | 説明 |
|---|---|
| `issues` | 課題の現在の状態 |
| `issue_change_history` | 全フィールドの変更履歴 |
| `issue_snapshots` | 課題の時系列スナップショット（バージョン管理） |
| `statuses` | ステータス一覧 |
| `priorities` | 優先度一覧 |
| `issue_types` | 課題タイプ一覧 |
| `labels` | ラベル一覧 |
| `components` | コンポーネント一覧 |
| `fix_versions` | バージョン一覧 |
| `sync_history` | 同期履歴 |

## 開発者向け

### プロジェクト構成

```
go-jira/
├── cmd/go-jira/           エントリーポイント
├── internal/cli/          CLI コマンド定義（外部非公開）
├── core/                  コアライブラリ（外部利用可能）
│   ├── domain/
│   │   ├── models/        ドメインモデル
│   │   └── repository/    リポジトリインターフェース
│   ├── service/           ビジネスロジック
│   └── infrastructure/
│       ├── jira/          JIRA REST API クライアント
│       ├── database/      DuckDB リポジトリ実装
│       └── config/        設定ファイル管理
└── docs/                  設計ドキュメント
```

### 依存関係の方向

```
CLI (internal/cli)
  ↓
Service (core/service)
  ↓
Domain (core/domain)    ← Infrastructure (core/infrastructure)
  models + interfaces       jira, database, config
```

- `service` は `domain/repository` インターフェースに依存
- `infrastructure` がインターフェースを実装
- CLI は `service` のみを呼び出す

### セットアップ

```bash
git clone https://github.com/ysksm/go-jira.git
cd go-jira
go mod tidy
```

### ビルド

```bash
go build -o go-jira ./cmd/go-jira
```

### テスト

```bash
go test ./...
```

### コアライブラリとしての利用

`core/` パッケージは外部から import して利用できます:

```go
import (
    "github.com/ysksm/go-jira/core/service"
    "github.com/ysksm/go-jira/core/infrastructure/config"
    "github.com/ysksm/go-jira/core/infrastructure/database"
)

// 設定を読み込み
store, _ := config.NewFileConfigStore("")
settings, _ := store.Load()

// DB 接続
connMgr := database.NewConnection(settings.Database.DatabaseDir)
defer connMgr.Close()

// 同期実行
syncSvc := service.NewSyncService(store, connMgr, nil)
results, _ := syncSvc.Execute(ctx, service.SyncOptions{})
```

### 設計ドキュメント

- [docs/requirements.md](docs/requirements.md) — 機能要件・非機能要件
- [docs/plan.md](docs/plan.md) — アーキテクチャ設計・同期シーケンス
- [docs/tasks.md](docs/tasks.md) — 実装タスク一覧

### メモリ効率設計

本プロジェクトはディスクファースト設計を採用しています:

- **課題取得**: 100 件バッチで取得 → 即 DB 書込 → メモリ解放
- **スナップショット生成**: カーソルで 1 課題ずつ DB から読出 → 処理 → DB 書戻
- **ソフトデリート判定**: DB 内サブクエリで完結（全キーをメモリに持たない）
- **インメモリ許容**: メタデータ（数十件）、1 課題の変更履歴、設定ファイルのみ

### ロードマップ

| フェーズ | 状態 | 内容 |
|---|---|---|
| v1 CLI | 実装済み | コマンドラインツール |
| v2 REST API | 計画中 | RPC スタイル HTTP API サーバー |
| v3 Web UI | 計画中 | Svelte フロントエンド |

## ライセンス

MIT License
