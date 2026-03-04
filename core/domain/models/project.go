package models

// Project represents a JIRA project.
type Project struct {
	ID           string `json:"id"`
	Key          string `json:"key"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Enabled      bool   `json:"enabled"`
	LastSyncedAt string `json:"lastSyncedAt,omitempty"`
}
