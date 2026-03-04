package models

// Status represents a JIRA workflow status.
type Status struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category"`
}

// Priority represents a JIRA priority level.
type Priority struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
}

// IssueType represents a JIRA issue type.
type IssueType struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
	Subtask     bool   `json:"subtask"`
}

// Label represents a JIRA label.
type Label struct {
	Name string `json:"name"`
}

// Component represents a JIRA project component.
type Component struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Lead        string `json:"lead,omitempty"`
}

// FixVersion represents a JIRA fix version.
type FixVersion struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Released    bool   `json:"released"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}

// ProjectMetadata holds all metadata for a project.
type ProjectMetadata struct {
	ProjectKey  string       `json:"projectKey"`
	Statuses    []Status     `json:"statuses"`
	Priorities  []Priority   `json:"priorities"`
	IssueTypes  []IssueType  `json:"issueTypes"`
	Labels      []Label      `json:"labels"`
	Components  []Component  `json:"components"`
	FixVersions []FixVersion `json:"fixVersions"`
}
