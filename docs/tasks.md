# Tasks - go-jira 実装タスク

## Phase 0: プロジェクト基盤

- [x] **T-001** `go mod init` + 基本依存追加（cobra, go-duckdb, testify）
- [x] **T-0\1** ディレクトリ構成の作成（cmd, internal/cli, core/domain, core/service, core/infrastructure）
- [x] **T-0\1** `cmd/go-jira/main.go` エントリーポイント作成
- [x] **T-0\1** CLI ルートコマンド定義（`internal/cli/root.go`）

## Phase 1: ドメインモデル

- [x] **T-0\1** `core/domain/models/config.go` — Settings, JiraEndpoint, ProjectConfig, SyncCheckpoint, SnapshotCheckpoint
- [x] **T-0\1** `core/domain/models/project.go` — Project
- [x] **T-0\1** `core/domain/models/issue.go` — Issue
- [x] **T-0\1** `core/domain/models/change_history.go` — ChangeHistoryItem
- [x] **T-0\1** `core/domain/models/snapshot.go` — IssueSnapshot
- [x] **T-0\1** `core/domain/models/metadata.go` — Status, Priority, IssueType, Label, Component, FixVersion, ProjectMetadata
- [x] **T-0\1** `core/domain/models/sync.go` — SyncResult, SyncProgress
- [x] **T-0\1** `core/domain/models/query.go` — SavedQuery, SqlTable, SqlColumn

## Phase 2: リポジトリインターフェース

- [x] **T-0\1** `core/domain/repository/issue_repo.go` — IssueRepository インターフェース
  - BatchInsert, FindByProjectCursor, CountByProject, FindByProjectPaginated, MarkDeletedNotInCurrentSync, CountByStatus
  - IssueCursor インターフェース（Next, Issue, Err, Close）— メモリ効率のためのカーソル読出
- [x] **T-0\1** `core/domain/repository/change_history_repo.go` — ChangeHistoryRepository インターフェース
  - BatchInsert, DeleteByIssueID, FindByIssueKey
- [x] **T-0\1** `core/domain/repository/snapshot_repo.go` — SnapshotRepository インターフェース
  - BatchInsert, DeleteByIssueID, BeginTransaction, CommitTransaction, RollbackTransaction
- [x] **T-0\1** `core/domain/repository/metadata_repo.go` — MetadataRepository インターフェース
  - UpsertStatuses, UpsertPriorities, UpsertIssueTypes, UpsertLabels, UpsertComponents, UpsertFixVersions, GetByProject
- [x] **T-0\1** `core/domain/repository/sync_history_repo.go` — SyncHistoryRepository インターフェース
  - Insert, UpdateCompleted, UpdateFailed

## Phase 3: インフラ層 - JIRA クライアント

- [x] **T-0\1** `core/infrastructure/jira/client.go` — JIRA HTTP クライアント
  - Basic認証、リトライ（指数バックオフ、最大3回、30秒タイムアウト）
  - `context.Context` によるキャンセル対応
- [x] **T-0\1** `core/infrastructure/jira/types.go` — JIRA APIレスポンス型定義
  - JiraSearchResponse, JiraIssue, JiraChangelog, JiraProject 等
- [x] **T-0\1** `core/infrastructure/jira/parser.go` — レスポンスパーサー
  - JiraIssue → domain Issue 変換
  - changelog → []ChangeHistoryItem 抽出
  - カスタムフィールド（Sprint, Team）の抽出
  - 日付パーサー（RFC3339 + JIRA形式対応）
- [x] **T-0\1** JIRAクライアントのメソッド実装
  - FetchIssues(jql, nextPageToken) → ([]Issue, []ChangeHistoryItem, nextToken)
  - FetchProjects() → []Project
  - FetchStatuses(projectKey) → []Status
  - FetchPriorities() → []Priority
  - FetchIssueTypes(projectID) → []IssueType
  - FetchComponents(projectKey) → []Component
  - FetchVersions(projectKey) → []FixVersion
  - FetchFields() → []JiraField

## Phase 4: インフラ層 - データベース

- [x] **T-0\1** `core/infrastructure/database/connection.go` — DuckDB接続管理
  - プロジェクトごとのDB作成・接続
  - パス: `{db_dir}/{project_key}/data.duckdb`
- [x] **T-0\1** `core/infrastructure/database/schema.go` — DDL定義・マイグレーション
  - issues, issue_change_history, issue_snapshots テーブル
  - statuses, priorities, issue_types, labels, components, fix_versions テーブル
  - sync_history, jira_fields テーブル
  - インデックス作成
- [x] **T-0\1** `core/infrastructure/database/issue_repository.go` — IssueRepository 実装
  - BatchInsert: ON CONFLICT DO UPDATE (UPSERT) → 即DB書込でメモリ解放
  - FindByProjectCursor: カーソルベース読出（1件ずつ、メモリ効率重視）
  - MarkDeletedNotInCurrentSync: DB内サブクエリで完結（全キーをメモリに持たない）
- [x] **T-0\1** `core/infrastructure/database/change_history_repository.go` — ChangeHistoryRepository 実装
- [x] **T-0\1** `core/infrastructure/database/snapshot_repository.go` — SnapshotRepository 実装
  - トランザクション管理（BEGIN/COMMIT/ROLLBACK）
- [x] **T-0\1** `core/infrastructure/database/metadata_repository.go` — MetadataRepository 実装
- [x] **T-0\1** `core/infrastructure/database/sync_history_repository.go` — SyncHistoryRepository 実装

## Phase 5: インフラ層 - 設定ファイル

- [x] **T-0\1** `core/infrastructure/config/file_config.go` — 設定の読み書き
  - パス: `~/.config/go-jira/settings.json`
  - Load / Save / Initialize / Update
  - チェックポイントの保存・読込

## Phase 6: サービス層

- [x] **T-0\1** `core/service/config_service.go` — 設定管理サービス
  - Get, Initialize, Update
  - エンドポイント追加・削除・切り替え
- [x] **T-0\1** `core/service/project_service.go` — プロジェクト管理サービス
  - List, FetchFromJira, Enable, Disable
- [x] **T-0\1** `core/service/sync_service.go` — **同期サービス（最重要）**
  - Execute: 全フェーズのオーケストレーション
  - FetchIssues: ページネーション + チェックポイント + 進捗報告
  - SyncMetadata: 6種類のメタデータ同期
  - VerifyIntegrity: 件数比較
  - ProgressCallback 型定義とコールバック機構
- [x] **T-0\1** `core/service/snapshot_service.go` — スナップショット生成サービス
  - GenerateForProject: バッチ + トランザクション + チェックポイント
  - BuildInitialState: 変更履歴の逆適用で初期状態復元
  - ApplyChangesForward: 各変更を順適用してバージョン構築
  - フィールドタイプ判定（DirectString, ObjectWithName, ArrayOfStrings等）
- [x] **T-0\1** `core/service/issue_service.go` — 課題検索・取得サービス
  - Search (フィルタ + ページネーション)
  - Get (キーで取得)
  - GetHistory (変更履歴取得)
- [x] **T-0\1** `core/service/query_service.go` — SQLクエリサービス
  - Execute (読み取り専用SQL実行)
  - GetSchema (テーブル・カラム情報)
  - SavedQuery の CRUD

## Phase 7: CLI コマンド

- [x] **T-0\1** `internal/cli/config_cmd.go` — config init / show / set
- [x] **T-0\1** `internal/cli/project_cmd.go` — project list / fetch / enable / disable
- [x] **T-0\1** `internal/cli/sync_cmd.go` — sync run / status
  - 進捗表示（フェーズ, プログレスバー, 件数）
  - `--project`, `--force` フラグ
- [x] **T-0\1** `internal/cli/issue_cmd.go` — issue search / get / history
  - テーブル形式の出力
  - フィルタフラグ（--project, --status, --assignee, --limit 等）
- [x] **T-0\1** `internal/cli/query_cmd.go` — query exec / schema / list / save / delete
  - テーブル形式のクエリ結果出力
  - `--project`, `--all-projects` フラグ
- [x] **T-0\1** `internal/cli/progress.go` — 進捗表示ユーティリティ
  - プログレスバー、フェーズ表示、サマリー表示

## Phase 8: テスト

- [x] **T-080** JIRA クライアントテスト（httptest スタブ）
- [x] **T-081** パーサーテスト（JIRAレスポンス → ドメインモデル変換）
- [x] **T-082** sync_service テスト（リポジトリモック使用）
- [x] **T-083** snapshot_service テスト（変更履歴の逆適用・順適用のロジック）
- [x] **T-084** データベースリポジトリ統合テスト（DuckDB使用）

## Phase 9: 仕上げ

- [x] **T-090** エラーハンドリング整備（カスタムエラー型、ユーザー向けメッセージ）
- [x] **T-091** ログ出力整備（slog、--verbose/--quiet）
- [x] **T-092** README.md 更新（使い方、インストール手順）

---

# v2: REST API サーバー

## Phase 10: API 基盤

- [x] **T-100** `cmd/go-jira-server/main.go` — サーバーエントリーポイント
  - ポート指定（--port, デフォルト 8080）
  - Graceful shutdown（SIGINT/SIGTERM）
- [x] **T-101** `internal/api/server.go` — HTTP サーバー設定・起動
- [x] **T-102** `internal/api/router.go` — ルーティング定義
  - 全エンドポイントを `POST /api/{domain}.{action}` で登録
- [x] **T-103** `internal/api/middleware.go` — ミドルウェア
  - CORS（開発時用、Originの許可設定）
  - リクエストログ（slog）
  - JSON Content-Type 強制
  - エラーリカバリ（panic → 500）
- [x] **T-104** `internal/api/types.go` — API リクエスト/レスポンス型
  - domain models をそのまま利用 + API 固有のラッパー型
- [x] **T-105** `internal/api/helpers.go` — 共通ヘルパー
  - `decodeRequest`, `writeJSON`, `writeError`

## Phase 11: API ハンドラー

- [x] **T-110** `internal/api/config_handler.go` — 設定 API
  - POST /api/config.get → ConfigGetResponse
  - POST /api/config.update → ConfigUpdateResponse
  - POST /api/config.initialize → ConfigInitResponse
- [x] **T-111** `internal/api/project_handler.go` — プロジェクト API
  - POST /api/projects.list → ProjectListResponse
  - POST /api/projects.initialize → ProjectInitResponse
  - POST /api/projects.enable → ProjectEnableResponse
  - POST /api/projects.disable → ProjectDisableResponse
- [x] **T-112** `internal/api/sync_handler.go` — 同期 API
  - POST /api/sync.execute → SyncExecuteResponse
  - POST /api/sync.status → SyncStatusResponse
  - GET  /api/sync.progress → SSE（Server-Sent Events）
- [x] **T-113** `internal/api/issue_handler.go` — 課題 API
  - POST /api/issues.search → IssueSearchResponse
  - POST /api/issues.get → IssueGetResponse
  - POST /api/issues.history → IssueHistoryResponse
- [x] **T-114** `internal/api/metadata_handler.go` — メタデータ API
  - POST /api/metadata.get → MetadataGetResponse
- [x] **T-115** `internal/api/query_handler.go` — SQL クエリ API
  - POST /api/sql.execute → SqlExecuteResponse
  - POST /api/sql.get-schema → SqlGetSchemaResponse

## Phase 12: API テスト

- [ ] **T-120** ハンドラーテスト（httptest + サービスモック）
- [ ] **T-121** SSE 進捗通知テスト
- [ ] **T-122** CORS・ミドルウェアテスト

---

# v3: Svelte フロントエンド

## Phase 13: Svelte プロジェクト基盤

- [ ] **T-130** SvelteKit プロジェクト初期化（`web/`）
  - Svelte 5, TypeScript, Tailwind CSS
- [ ] **T-131** Vite プロキシ設定（`/api` → Go サーバー）
- [ ] **T-132** `web/src/lib/api.ts` — API クライアント
  - 全エンドポイント対応（POST + SSE）
- [ ] **T-133** `web/src/lib/types.ts` — 型定義（Go models に対応）
- [ ] **T-134** `web/src/lib/stores/` — Svelte stores
  - projects, syncProgress, settings

## Phase 14: Svelte ページ実装

- [ ] **T-140** 共通レイアウト + ナビゲーション
- [ ] **T-141** ダッシュボード（同期状態、プロジェクト概要）
- [ ] **T-142** プロジェクト管理ページ（一覧、有効/無効切替）
- [ ] **T-143** 同期ページ（実行ボタン、SSE リアルタイム進捗表示）
- [ ] **T-144** 課題検索ページ（フィルタ、リスト/ボード切替）
- [ ] **T-145** SQL クエリページ（エディタ、スキーマブラウザ、結果テーブル）
- [ ] **T-146** 設定ページ（JIRA接続、DB、同期設定）

## Phase 15: ビルド統合

- [ ] **T-150** Go サーバーに Svelte ビルド成果物を埋め込み（`embed`）
- [ ] **T-151** 本番ビルドスクリプト（Makefile）
- [ ] **T-152** README 更新（v2/v3 の使い方追加）

---

## 実装順序

```
v1 (完了)
    ↓
Phase 10-11 (REST API) → Phase 12 (API テスト)
    ↓
Phase 13-14 (Svelte) → Phase 15 (ビルド統合)
```
