package api

import "github.com/ysksm/go-jira/core/domain/models"

// --- Config ---

type ConfigGetRequest struct{}
type ConfigGetResponse struct {
	Settings models.Settings `json:"settings"`
}

type ConfigUpdateRequest struct {
	Jira              *models.JiraConfig      `json:"jira,omitempty"`
	Database          *models.DatabaseConfig   `json:"database,omitempty"`
	Embeddings        *models.EmbeddingsConfig `json:"embeddings,omitempty"`
	Log               *models.LogConfig        `json:"log,omitempty"`
	Sync              *models.SyncSettings     `json:"sync,omitempty"`
	AddEndpoint       *models.JiraEndpoint     `json:"addEndpoint,omitempty"`
	RemoveEndpoint    string                   `json:"removeEndpoint,omitempty"`
	SetActiveEndpoint string                   `json:"setActiveEndpoint,omitempty"`
}
type ConfigUpdateResponse struct {
	Success  bool            `json:"success"`
	Settings models.Settings `json:"settings"`
}

type ConfigInitRequest struct {
	Endpoint     string `json:"endpoint"`
	Username     string `json:"username"`
	APIKey       string `json:"apiKey"`
	DatabasePath string `json:"databasePath,omitempty"`
}
type ConfigInitResponse struct {
	Success  bool            `json:"success"`
	Settings models.Settings `json:"settings"`
}

// --- Projects ---

type ProjectListRequest struct{}
type ProjectListResponse struct {
	Projects []models.Project `json:"projects"`
}

type ProjectInitRequest struct {
	EndpointName string `json:"endpointName,omitempty"`
}
type ProjectInitResponse struct {
	Projects []models.Project `json:"projects"`
	NewCount int              `json:"newCount"`
}

type ProjectEnableRequest struct {
	Key string `json:"key"`
}
type ProjectEnableResponse struct {
	Project models.Project `json:"project"`
}

type ProjectDisableRequest struct {
	Key string `json:"key"`
}
type ProjectDisableResponse struct {
	Project models.Project `json:"project"`
}

// --- Sync ---

type SyncExecuteRequest struct {
	ProjectKey string `json:"projectKey,omitempty"`
	Force      bool   `json:"force,omitempty"`
}
type SyncExecuteResponse struct {
	Results []models.SyncResult `json:"results"`
}

type SyncStatusRequest struct{}
type SyncStatusResponse struct {
	InProgress bool                 `json:"inProgress"`
	Progress   *models.SyncProgress `json:"progress,omitempty"`
}

// --- Issues ---

type IssueSearchRequest struct {
	Query     string `json:"query,omitempty"`
	Project   string `json:"project,omitempty"`
	Status    string `json:"status,omitempty"`
	Assignee  string `json:"assignee,omitempty"`
	Priority  string `json:"priority,omitempty"`
	IssueType string `json:"issueType,omitempty"`
	Team      string `json:"team,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}
type IssueSearchResponse struct {
	Issues []models.Issue `json:"issues"`
	Total  int            `json:"total"`
}

type IssueGetRequest struct {
	Key string `json:"key"`
}
type IssueGetResponse struct {
	Issue models.Issue `json:"issue"`
}

type IssueHistoryRequest struct {
	Key   string `json:"key"`
	Field string `json:"field,omitempty"`
	Limit int    `json:"limit,omitempty"`
}
type IssueHistoryResponse struct {
	History []models.ChangeHistoryItem `json:"history"`
}

// --- Metadata ---

type MetadataGetRequest struct {
	ProjectKey string `json:"projectKey"`
}
type MetadataGetResponse struct {
	Metadata models.ProjectMetadata `json:"metadata"`
}

// --- SQL ---

type SqlExecuteRequest struct {
	ProjectKey  string `json:"projectKey,omitempty"`
	AllProjects bool   `json:"allProjects,omitempty"`
	Query       string `json:"query"`
	Limit       int    `json:"limit,omitempty"`
}
type SqlExecuteResponse struct {
	Columns         []string        `json:"columns"`
	Rows            [][]interface{} `json:"rows"`
	RowCount        int             `json:"rowCount"`
	ExecutionTimeMs int64           `json:"executionTimeMs"`
}

type SqlGetSchemaRequest struct {
	ProjectKey string `json:"projectKey,omitempty"`
}
type SqlGetSchemaResponse struct {
	Tables []models.SqlTable `json:"tables"`
}
