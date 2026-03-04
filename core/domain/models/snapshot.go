package models

import (
	"encoding/json"
	"time"
)

// IssueSnapshot represents a point-in-time snapshot of an issue.
type IssueSnapshot struct {
	IssueID     string           `json:"issueId"`
	IssueKey    string           `json:"issueKey"`
	ProjectID   string           `json:"projectId"`
	Version     int              `json:"version"`
	ValidFrom   time.Time        `json:"validFrom"`
	ValidTo     *time.Time       `json:"validTo,omitempty"`
	Summary     string           `json:"summary"`
	Description string           `json:"description,omitempty"`
	Status      string           `json:"status,omitempty"`
	Priority    string           `json:"priority,omitempty"`
	Assignee    string           `json:"assignee,omitempty"`
	Reporter    string           `json:"reporter,omitempty"`
	IssueType   string           `json:"issueType,omitempty"`
	Resolution  string           `json:"resolution,omitempty"`
	Labels      []string         `json:"labels,omitempty"`
	Components  []string         `json:"components,omitempty"`
	FixVersions []string         `json:"fixVersions,omitempty"`
	Sprint      string           `json:"sprint,omitempty"`
	ParentKey   string           `json:"parentKey,omitempty"`
	RawData     *json.RawMessage `json:"rawData,omitempty"`
	UpdatedDate *time.Time       `json:"updatedDate,omitempty"`
	ResolvedDate *time.Time      `json:"resolvedDate,omitempty"`
	DueDate     *time.Time       `json:"dueDate,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
}
