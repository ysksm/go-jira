# Requirements - go-jira

## 概要

JIRA のデータをローカルに同期・蓄積し、高速な検索・分析・SQLクエリを可能にするツール。
既存の Rust + Angular 実装 (jd) のバックエンドロジックを Go で再実装する。

### 提供形態（ロードマップ）

| フェーズ | 提供形態 | 説明 |
|---|---|---|
| **v1 (現在)** | CLI | コマンドラインツールとして提供 |
| **v2 (将来)** | REST API サーバー | HTTP API として core 機能を公開 |
| **v3 (将来)** | Web アプリ | Svelte フロントエンド + REST API バックエンド |

## 機能要件

### 1. 設定管理 (Config)

- JIRA エンドポイント設定（URL, ユーザー名, APIキー）
- 複数エンドポイント対応（`jira_endpoints[]` + `active_endpoint`）
- データベース保存先パス設定
- 同期設定（インクリメンタル同期の有効/無効、マージン分数）
- ログ設定（ファイル出力、レベル、ローテーション）
- 設定ファイル: `~/.config/go-jira/settings.json`

### 2. プロジェクト管理 (Projects)

- JIRA API からプロジェクト一覧を取得・登録
- プロジェクト単位で同期の有効/無効を制御
- 複数エンドポイントからのプロジェクト取得対応
- プロジェクトごとのメタデータ管理（ステータス、優先度、課題タイプ等）

### 3. データ同期 (Sync)

最も重要な機能。以下のフェーズで構成される：

#### Phase 1: 課題取得 (Fetch Issues)
- JIRA REST API v3 (`/rest/api/3/search/jql`) を使用
- トークンベースページネーション（`nextPageToken`）
- インクリメンタル同期: 前回同期日時以降の更新のみ取得
- バッチ処理（100件ずつ）
- チェックポイント保存（中断からの再開対応）
- 変更履歴（changelog）の抽出・保存
- 削除された課題のソフトデリート

#### Phase 2: メタデータ同期 (Sync Metadata)
- ステータス、優先度、課題タイプ、ラベル、コンポーネント、バージョンの同期
- UPSERT（存在すれば更新、なければ挿入）

#### Phase 3: スナップショット生成 (Generate Snapshots)
- 課題の変更履歴から時系列スナップショットを生成
- 変更履歴を逆適用して初期状態を復元
- 各変更を順適用してバージョンごとの状態を構築
- バッチ処理 + トランザクション管理
- チェックポイントによる再開対応

#### Phase 4: データ整合性検証 (Verify Integrity)
- JIRA 上の件数とローカルの件数を比較
- ステータス別の件数確認
- 差異があればログに記録

#### 進捗報告
- 各フェーズの進行状況をリアルタイム表示
- バッチ処理ごとにカウント更新

### 4. データ検索・クエリ (Issues / SQL)

- 課題の全文検索（フィルタ付き）
- フィルタ: プロジェクト、ステータス、担当者、優先度、課題タイプ、チーム
- ページネーション対応
- 課題単体の詳細取得
- 課題の変更履歴取得
- SQL クエリの直接実行（読み取り専用）
- データベーススキーマの参照
- クエリの保存・一覧・削除

### 5. メタデータ管理

- プロジェクトごとのメタデータ取得
  - ステータス一覧（カテゴリ付き）
  - 優先度一覧
  - 課題タイプ一覧
  - ラベル一覧
  - コンポーネント一覧
  - バージョン一覧

### 6. REST API サーバー（v2）

RPC スタイルの HTTP API として core 機能を公開する。

- 全操作は POST メソッド（リクエストボディでパラメータ渡し）
- エンドポイント体系: `/api/{domain}.{action}`
  - `/api/config.get`, `/api/config.update`, `/api/config.initialize`
  - `/api/projects.list`, `/api/projects.initialize`, `/api/projects.enable`, `/api/projects.disable`
  - `/api/sync.execute`, `/api/sync.status`
  - `/api/issues.search`, `/api/issues.get`, `/api/issues.history`
  - `/api/metadata.get`
  - `/api/sql.execute`, `/api/sql.get-schema`, `/api/sql.list-queries`, `/api/sql.save-query`, `/api/sql.delete-query`
- 同期進捗のリアルタイム通知（WebSocket または SSE）
- CORS 設定（フロントエンド開発用）
- JSON リクエスト/レスポンス

### 7. Web フロントエンド（v3）

Svelte によるWebアプリケーション。

- **フレームワーク:** SvelteKit
- **ページ構成:**
  - ダッシュボード — 同期状態・プロジェクト概要
  - プロジェクト管理 — プロジェクト一覧・有効/無効切り替え
  - 同期 — 同期実行・リアルタイム進捗表示
  - 課題検索 — フィルタ付き検索・複数ビュー（リスト/ボード/カレンダー）
  - SQLクエリ — マルチタブエディタ・スキーマブラウザ
  - 可視化 — チャート生成（バーンダウン、ベロシティ、CFD等）
  - 設定 — JIRA接続・DB・ログ設定の管理
- **REST API バックエンドと通信**

## 非機能要件

### パフォーマンス
- バッチ処理による効率的なデータ取得・保存
- インクリメンタル同期による差分取得
- DuckDB による高速なSQLクエリ実行

### 信頼性
- チェックポイントによる中断耐性
- トランザクション管理
- リトライ機構（指数バックオフ、最大3回、30秒タイムアウト）
- エラー時のロールバック

### 拡張性
- core パッケージとCLI/API/Webの分離（ライブラリとして他プロジェクトから利用可能）
- リポジトリパターンによるストレージ層の抽象化（インターフェース定義）
- JIRAクライアントの抽象化（テスト容易性）
- core → CLI / REST API / Web フロントエンドの3つの提供形態を同一コアで実現

### ユーザビリティ
- CLI での進捗表示（フェーズ、件数、パーセンテージ）
- わかりやすいエラーメッセージ
- `--verbose` / `--quiet` フラグによる出力制御
- サブコマンド体系による直感的な操作

## データモデル

### コアエンティティ

| エンティティ | 説明 |
|---|---|
| Project | JIRAプロジェクト（ID, Key, Name, 同期状態） |
| Issue | 課題（ID, Key, Summary, Status, Priority, Assignee等） |
| ChangeHistoryItem | 課題の変更履歴（フィールド名, From/To値, 変更日時） |
| IssueSnapshot | 課題の時系列スナップショット（バージョン, 有効期間） |
| ProjectMetadata | プロジェクトメタデータ（ステータス, 優先度, 課題タイプ等） |

### 設定エンティティ

| エンティティ | 説明 |
|---|---|
| Settings | 全体設定（JIRA, DB, 同期, ログ） |
| JiraEndpoint | JIRAエンドポイント（URL, 認証情報） |
| ProjectConfig | プロジェクト設定（有効/無効, チェックポイント） |
| SyncCheckpoint | 同期チェックポイント（再開用情報） |
| SnapshotCheckpoint | スナップショットチェックポイント |

## JIRA API エンドポイント

| 操作 | エンドポイント |
|---|---|
| 課題検索 | `GET /rest/api/3/search/jql` |
| プロジェクト一覧 | `GET /rest/api/3/project` |
| ステータス一覧 | `GET /rest/api/3/project/{key}/statuses` |
| 優先度一覧 | `GET /rest/api/3/priority` |
| 課題タイプ一覧 | `GET /rest/api/3/issuetype/project?projectId={id}` |
| コンポーネント一覧 | `GET /rest/api/3/project/{key}/components` |
| バージョン一覧 | `GET /rest/api/3/project/{key}/versions` |
| フィールド一覧 | `GET /rest/api/3/field` |

認証方式: Basic認証（`Authorization: Basic base64(username:api_key)`）

## データベース

- **エンジン:** DuckDB
- **構成:** プロジェクトごとに分離（`{db_dir}/{project_key}/data.duckdb`）
- **テーブル:** issues, issue_change_history, issue_snapshots, statuses, priorities, issue_types, labels, components, fix_versions, sync_history, jira_fields
