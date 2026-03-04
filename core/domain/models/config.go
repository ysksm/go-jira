package models

import "time"

// Settings represents the application configuration.
type Settings struct {
	Jira            *JiraConfig      `json:"jira,omitempty"`
	JiraEndpoints   []JiraEndpoint   `json:"jiraEndpoints,omitempty"`
	ActiveEndpoint  string           `json:"activeEndpoint,omitempty"`
	Projects        []ProjectConfig  `json:"projects"`
	Database        DatabaseConfig   `json:"database"`
	Embeddings      *EmbeddingsConfig `json:"embeddings,omitempty"`
	Log             *LogConfig       `json:"log,omitempty"`
	Sync            *SyncSettings    `json:"sync,omitempty"`
	DebugMode       bool             `json:"debugMode,omitempty"`
}

// JiraConfig represents a single JIRA endpoint configuration (legacy).
type JiraConfig struct {
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	APIKey   string `json:"apiKey"`
}

// JiraEndpoint represents a named JIRA endpoint configuration.
type JiraEndpoint struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Endpoint    string `json:"endpoint"`
	Username    string `json:"username"`
	APIKey      string `json:"apiKey"`
}

// ProjectConfig represents per-project configuration.
type ProjectConfig struct {
	ID                  string              `json:"id"`
	Key                 string              `json:"key"`
	Name                string              `json:"name"`
	SyncEnabled         bool                `json:"syncEnabled"`
	LastSynced          *time.Time          `json:"lastSynced,omitempty"`
	Endpoint            string              `json:"endpoint,omitempty"`
	SyncCheckpoint      *SyncCheckpoint     `json:"syncCheckpoint,omitempty"`
	SnapshotCheckpoint  *SnapshotCheckpoint `json:"snapshotCheckpoint,omitempty"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Path        string `json:"path,omitempty"`
	DatabaseDir string `json:"databaseDir"`
}

// EmbeddingsConfig represents embeddings configuration.
type EmbeddingsConfig struct {
	Provider     string `json:"provider"`
	ModelName    string `json:"modelName,omitempty"`
	Endpoint     string `json:"endpoint,omitempty"`
	AutoGenerate bool   `json:"autoGenerate"`
}

// LogConfig represents logging configuration.
type LogConfig struct {
	FileEnabled bool   `json:"fileEnabled"`
	FileDir     string `json:"fileDir,omitempty"`
	Level       string `json:"level"`
	MaxFiles    int    `json:"maxFiles"`
}

// SyncSettings represents sync-specific settings.
type SyncSettings struct {
	IncrementalSyncEnabled       bool `json:"incrementalSyncEnabled"`
	IncrementalSyncMarginMinutes int  `json:"incrementalSyncMarginMinutes"`
}

// SyncCheckpoint stores resumable sync state.
type SyncCheckpoint struct {
	LastIssueUpdatedAt time.Time `json:"lastIssueUpdatedAt"`
	LastIssueKey       string    `json:"lastIssueKey"`
	ItemsProcessed     int       `json:"itemsProcessed"`
	TotalItems         int       `json:"totalItems"`
}

// SnapshotCheckpoint stores resumable snapshot generation state.
type SnapshotCheckpoint struct {
	LastIssueID        string `json:"lastIssueId"`
	LastIssueKey       string `json:"lastIssueKey"`
	IssuesProcessed    int    `json:"issuesProcessed"`
	TotalIssues        int    `json:"totalIssues"`
	SnapshotsGenerated int    `json:"snapshotsGenerated"`
}

// GetActiveEndpoint returns the active JIRA endpoint configuration.
// Falls back to the legacy JiraConfig if no endpoints are configured.
func (s *Settings) GetActiveEndpoint() *JiraEndpoint {
	if len(s.JiraEndpoints) > 0 {
		for i := range s.JiraEndpoints {
			if s.JiraEndpoints[i].Name == s.ActiveEndpoint {
				return &s.JiraEndpoints[i]
			}
		}
		return &s.JiraEndpoints[0]
	}
	if s.Jira != nil {
		return &JiraEndpoint{
			Name:     "default",
			Endpoint: s.Jira.Endpoint,
			Username: s.Jira.Username,
			APIKey:   s.Jira.APIKey,
		}
	}
	return nil
}

// GetSyncSettings returns sync settings with defaults applied.
func (s *Settings) GetSyncSettings() SyncSettings {
	if s.Sync != nil {
		return *s.Sync
	}
	return SyncSettings{
		IncrementalSyncEnabled:       true,
		IncrementalSyncMarginMinutes: 5,
	}
}
