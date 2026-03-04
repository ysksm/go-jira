# Tasks - go-jira 実装タスク

## Phase 0: プロジェクト基盤

- [ ] **T-001** `go mod init` + 基本依存追加（cobra, go-duckdb, testify）
- [ ] **T-002** ディレクトリ構成の作成（cmd, internal/cli, core/domain, core/service, core/infrastructure）
- [ ] **T-003** `cmd/go-jira/main.go` エントリーポイント作成
- [ ] **T-004** CLI ルートコマンド定義（`internal/cli/root.go`）

## Phase 1: ドメインモデル

- [ ] **T-010** `core/domain/models/config.go` — Settings, JiraEndpoint, ProjectConfig, SyncCheckpoint, SnapshotCheckpoint
- [ ] **T-011** `core/domain/models/project.go` — Project
- [ ] **T-012** `core/domain/models/issue.go` — Issue
- [ ] **T-013** `core/domain/models/change_history.go` — ChangeHistoryItem
- [ ] **T-014** `core/domain/models/snapshot.go` — IssueSnapshot
- [ ] **T-015** `core/domain/models/metadata.go` — Status, Priority, IssueType, Label, Component, FixVersion, ProjectMetadata
- [ ] **T-016** `core/domain/models/sync.go` — SyncResult, SyncProgress
- [ ] **T-017** `core/domain/models/query.go` — SavedQuery, SqlTable, SqlColumn

## Phase 2: リポジトリインターフェース

- [ ] **T-020** `core/domain/repository/issue_repo.go` — IssueRepository インターフェース
  - BatchInsert, FindByProjectCursor, CountByProject, FindByProjectPaginated, MarkDeletedNotInCurrentSync, CountByStatus
  - IssueCursor インターフェース（Next, Issue, Err, Close）— メモリ効率のためのカーソル読出
- [ ] **T-021** `core/domain/repository/change_history_repo.go` — ChangeHistoryRepository インターフェース
  - BatchInsert, DeleteByIssueID, FindByIssueKey
- [ ] **T-022** `core/domain/repository/snapshot_repo.go` — SnapshotRepository インターフェース
  - BatchInsert, DeleteByIssueID, BeginTransaction, CommitTransaction, RollbackTransaction
- [ ] **T-023** `core/domain/repository/metadata_repo.go` — MetadataRepository インターフェース
  - UpsertStatuses, UpsertPriorities, UpsertIssueTypes, UpsertLabels, UpsertComponents, UpsertFixVersions, GetByProject
- [ ] **T-024** `core/domain/repository/sync_history_repo.go` — SyncHistoryRepository インターフェース
  - Insert, UpdateCompleted, UpdateFailed

## Phase 3: インフラ層 - JIRA クライアント

- [ ] **T-030** `core/infrastructure/jira/client.go` — JIRA HTTP クライアント
  - Basic認証、リトライ（指数バックオフ、最大3回、30秒タイムアウト）
  - `context.Context` によるキャンセル対応
- [ ] **T-031** `core/infrastructure/jira/types.go` — JIRA APIレスポンス型定義
  - JiraSearchResponse, JiraIssue, JiraChangelog, JiraProject 等
- [ ] **T-032** `core/infrastructure/jira/parser.go` — レスポンスパーサー
  - JiraIssue → domain Issue 変換
  - changelog → []ChangeHistoryItem 抽出
  - カスタムフィールド（Sprint, Team）の抽出
  - 日付パーサー（RFC3339 + JIRA形式対応）
- [ ] **T-033** JIRAクライアントのメソッド実装
  - FetchIssues(jql, nextPageToken) → ([]Issue, []ChangeHistoryItem, nextToken)
  - FetchProjects() → []Project
  - FetchStatuses(projectKey) → []Status
  - FetchPriorities() → []Priority
  - FetchIssueTypes(projectID) → []IssueType
  - FetchComponents(projectKey) → []Component
  - FetchVersions(projectKey) → []FixVersion
  - FetchFields() → []JiraField

## Phase 4: インフラ層 - データベース

- [ ] **T-040** `core/infrastructure/database/connection.go` — DuckDB接続管理
  - プロジェクトごとのDB作成・接続
  - パス: `{db_dir}/{project_key}/data.duckdb`
- [ ] **T-041** `core/infrastructure/database/schema.go` — DDL定義・マイグレーション
  - issues, issue_change_history, issue_snapshots テーブル
  - statuses, priorities, issue_types, labels, components, fix_versions テーブル
  - sync_history, jira_fields テーブル
  - インデックス作成
- [ ] **T-042** `core/infrastructure/database/issue_repository.go` — IssueRepository 実装
  - BatchInsert: ON CONFLICT DO UPDATE (UPSERT) → 即DB書込でメモリ解放
  - FindByProjectCursor: カーソルベース読出（1件ずつ、メモリ効率重視）
  - MarkDeletedNotInCurrentSync: DB内サブクエリで完結（全キーをメモリに持たない）
- [ ] **T-043** `core/infrastructure/database/change_history_repository.go` — ChangeHistoryRepository 実装
- [ ] **T-044** `core/infrastructure/database/snapshot_repository.go` — SnapshotRepository 実装
  - トランザクション管理（BEGIN/COMMIT/ROLLBACK）
- [ ] **T-045** `core/infrastructure/database/metadata_repository.go` — MetadataRepository 実装
- [ ] **T-046** `core/infrastructure/database/sync_history_repository.go` — SyncHistoryRepository 実装

## Phase 5: インフラ層 - 設定ファイル

- [ ] **T-050** `core/infrastructure/config/file_config.go` — 設定の読み書き
  - パス: `~/.config/go-jira/settings.json`
  - Load / Save / Initialize / Update
  - チェックポイントの保存・読込

## Phase 6: サービス層

- [ ] **T-060** `core/service/config_service.go` — 設定管理サービス
  - Get, Initialize, Update
  - エンドポイント追加・削除・切り替え
- [ ] **T-061** `core/service/project_service.go` — プロジェクト管理サービス
  - List, FetchFromJira, Enable, Disable
- [ ] **T-062** `core/service/sync_service.go` — **同期サービス（最重要）**
  - Execute: 全フェーズのオーケストレーション
  - FetchIssues: ページネーション + チェックポイント + 進捗報告
  - SyncMetadata: 6種類のメタデータ同期
  - VerifyIntegrity: 件数比較
  - ProgressCallback 型定義とコールバック機構
- [ ] **T-063** `core/service/snapshot_service.go` — スナップショット生成サービス
  - GenerateForProject: バッチ + トランザクション + チェックポイント
  - BuildInitialState: 変更履歴の逆適用で初期状態復元
  - ApplyChangesForward: 各変更を順適用してバージョン構築
  - フィールドタイプ判定（DirectString, ObjectWithName, ArrayOfStrings等）
- [ ] **T-064** `core/service/issue_service.go` — 課題検索・取得サービス
  - Search (フィルタ + ページネーション)
  - Get (キーで取得)
  - GetHistory (変更履歴取得)
- [ ] **T-065** `core/service/query_service.go` — SQLクエリサービス
  - Execute (読み取り専用SQL実行)
  - GetSchema (テーブル・カラム情報)
  - SavedQuery の CRUD

## Phase 7: CLI コマンド

- [ ] **T-070** `internal/cli/config_cmd.go` — config init / show / set
- [ ] **T-071** `internal/cli/project_cmd.go` — project list / fetch / enable / disable
- [ ] **T-072** `internal/cli/sync_cmd.go` — sync run / status
  - 進捗表示（フェーズ, プログレスバー, 件数）
  - `--project`, `--force` フラグ
- [ ] **T-073** `internal/cli/issue_cmd.go` — issue search / get / history
  - テーブル形式の出力
  - フィルタフラグ（--project, --status, --assignee, --limit 等）
- [ ] **T-074** `internal/cli/query_cmd.go` — query exec / schema / list / save / delete
  - テーブル形式のクエリ結果出力
  - `--project`, `--all-projects` フラグ
- [ ] **T-075** `internal/cli/progress.go` — 進捗表示ユーティリティ
  - プログレスバー、フェーズ表示、サマリー表示

## Phase 8: テスト

- [ ] **T-080** JIRA クライアントテスト（httptest スタブ）
- [ ] **T-081** パーサーテスト（JIRAレスポンス → ドメインモデル変換）
- [ ] **T-082** sync_service テスト（リポジトリモック使用）
- [ ] **T-083** snapshot_service テスト（変更履歴の逆適用・順適用のロジック）
- [ ] **T-084** データベースリポジトリ統合テスト（DuckDB使用）

## Phase 9: 仕上げ

- [ ] **T-090** エラーハンドリング整備（カスタムエラー型、ユーザー向けメッセージ）
- [ ] **T-091** ログ出力整備（slog、--verbose/--quiet）
- [ ] **T-092** README.md 更新（使い方、インストール手順）

---

## 実装順序の推奨

```
Phase 0 (基盤)
    ↓
Phase 1 (モデル) → Phase 2 (インターフェース)
    ↓
Phase 3 (JIRA) + Phase 4 (DB) + Phase 5 (設定)  ← 並行可能
    ↓
Phase 6 (サービス)  ← Phase 3-5 完了後
    ↓
Phase 7 (CLI)  ← Phase 6 完了後
    ↓
Phase 8 (テスト) + Phase 9 (仕上げ)  ← 並行可能
```

## 注意事項

### 同期処理の実装ポイント
1. **チェックポイント保存はバッチ完了ごとに行う** — 中断からの再開を確実にする
2. **インクリメンタル同期のマージン** — JQLの分精度に対応するため、前回日時から5分引く
3. **トークンベースページネーション** — offset ではなく `nextPageToken` を使う
4. **changelog は raw JSON から抽出** — API レスポンスの `changelog.histories[].items[]` を解析
5. **スナップショットの初期状態復元** — 全変更を逆順に適用して初期状態を得る

### スナップショット生成のフィールドタイプ
| タイプ | 該当フィールド | 処理方法 |
|---|---|---|
| DirectString | summary, description | そのまま from/to を適用 |
| ObjectWithName | status, priority, issuetype, resolution, sprint | `name` フィールドで比較・適用 |
| ObjectWithDisplayName | assignee, reporter | `displayName` で比較・適用 |
| ArrayOfStrings | labels | カンマ区切り文字列として処理 |
| ArrayOfObjectsWithName | components, fixVersions | オブジェクト配列から `name` 抽出 |
