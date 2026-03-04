package models

import "time"

// ChangeHistoryItem represents a single field change in an issue's history.
type ChangeHistoryItem struct {
	ID                string    `json:"id,omitempty"`
	IssueID           string    `json:"issueId"`
	IssueKey          string    `json:"issueKey"`
	HistoryID         string    `json:"historyId"`
	AuthorAccountID   string    `json:"authorAccountId,omitempty"`
	AuthorDisplayName string    `json:"authorDisplayName,omitempty"`
	Field             string    `json:"field"`
	FieldType         string    `json:"fieldType,omitempty"`
	FromValue         string    `json:"fromValue,omitempty"`
	FromString        string    `json:"fromString,omitempty"`
	ToValue           string    `json:"toValue,omitempty"`
	ToString          string    `json:"toString,omitempty"`
	ChangedAt         time.Time `json:"changedAt"`
}
