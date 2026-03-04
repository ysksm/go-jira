package models

import "time"

// Issue represents a JIRA issue.
type Issue struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"projectId"`
	Key         string    `json:"key"`
	Summary     string    `json:"summary"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status,omitempty"`
	Priority    string    `json:"priority,omitempty"`
	Assignee    string    `json:"assignee,omitempty"`
	Reporter    string    `json:"reporter,omitempty"`
	IssueType   string    `json:"issueType,omitempty"`
	Resolution  string    `json:"resolution,omitempty"`
	Labels      []string  `json:"labels,omitempty"`
	Components  []string  `json:"components,omitempty"`
	FixVersions []string  `json:"fixVersions,omitempty"`
	Sprint      string    `json:"sprint,omitempty"`
	Team        string    `json:"team,omitempty"`
	ParentKey   string    `json:"parentKey,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	CreatedDate *time.Time `json:"createdDate,omitempty"`
	UpdatedDate *time.Time `json:"updatedDate,omitempty"`
	RawJSON     string    `json:"-"`
	SyncedAt    *time.Time `json:"syncedAt,omitempty"`
	IsDeleted   bool      `json:"isDeleted,omitempty"`
}
