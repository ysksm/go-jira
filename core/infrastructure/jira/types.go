package jira

import "encoding/json"

// JiraSearchResponse represents the response from JIRA search API.
type JiraSearchResponse struct {
	Issues        []JiraIssue `json:"issues"`
	Total         int         `json:"total"`
	MaxResults    int         `json:"maxResults"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
}

// JiraIssue represents a JIRA issue from the API.
type JiraIssue struct {
	ID        string           `json:"id"`
	Key       string           `json:"key"`
	Self      string           `json:"self"`
	Fields    json.RawMessage  `json:"fields"`
	Changelog *JiraChangelog   `json:"changelog,omitempty"`
}

// JiraIssueFields represents parsed issue fields.
type JiraIssueFields struct {
	Summary     string              `json:"summary"`
	Description interface{}         `json:"description"`
	Status      *JiraNamedField     `json:"status"`
	Priority    *JiraNamedField     `json:"priority"`
	Assignee    *JiraUserField      `json:"assignee"`
	Reporter    *JiraUserField      `json:"reporter"`
	IssueType   *JiraNamedField     `json:"issuetype"`
	Resolution  *JiraNamedField     `json:"resolution"`
	Labels      []string            `json:"labels"`
	Components  []JiraNamedField    `json:"components"`
	FixVersions []JiraNamedField    `json:"fixVersions"`
	Parent      *JiraParentField    `json:"parent"`
	DueDate     string              `json:"duedate"`
	Created     string              `json:"created"`
	Updated     string              `json:"updated"`
	Project     *JiraProjectField   `json:"project"`
}

// JiraNamedField represents a JIRA field with a name.
type JiraNamedField struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
}

// JiraUserField represents a JIRA user field.
type JiraUserField struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
}

// JiraParentField represents a JIRA parent issue reference.
type JiraParentField struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

// JiraProjectField represents a JIRA project reference in issue fields.
type JiraProjectField struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// JiraChangelog represents the changelog from a JIRA issue.
type JiraChangelog struct {
	Histories []JiraHistory `json:"histories"`
}

// JiraHistory represents a single history entry.
type JiraHistory struct {
	ID      string            `json:"id"`
	Author  *JiraUserField    `json:"author"`
	Created string            `json:"created"`
	Items   []JiraHistoryItem `json:"items"`
}

// JiraHistoryItem represents a single field change in a history entry.
type JiraHistoryItem struct {
	Field      string `json:"field"`
	FieldType  string `json:"fieldtype"`
	FromValue  string `json:"from"`
	FromString string `json:"fromString"`
	ToValue    string `json:"to"`
	ToString   string `json:"toString"`
}

// JiraProject represents a JIRA project from the API.
type JiraProject struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// JiraStatusCategory represents a status category.
type JiraStatusCategory struct {
	ID   int    `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// JiraStatusDetail represents a status with category from project statuses API.
type JiraStatusDetail struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description,omitempty"`
	StatusCategory *JiraStatusCategory `json:"statusCategory,omitempty"`
}

// JiraIssueTypeWithStatuses is returned by the project statuses endpoint.
type JiraIssueTypeWithStatuses struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Subtask  bool               `json:"subtask"`
	Statuses []JiraStatusDetail `json:"statuses"`
}

// JiraField represents a JIRA field definition.
type JiraField struct {
	ID         string           `json:"id"`
	Key        string           `json:"key"`
	Name       string           `json:"name"`
	Custom     bool             `json:"custom"`
	Searchable bool            `json:"searchable"`
	Navigable  bool            `json:"navigable"`
	Schema     *JiraFieldSchema `json:"schema,omitempty"`
}

// JiraFieldSchema represents the schema of a JIRA field.
type JiraFieldSchema struct {
	Type     string `json:"type"`
	Items    string `json:"items,omitempty"`
	System   string `json:"system,omitempty"`
	Custom   string `json:"custom,omitempty"`
	CustomID int64  `json:"customId,omitempty"`
}

// JiraVersion represents a JIRA project version.
type JiraVersion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Released    bool   `json:"released"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}

// JiraComponent represents a JIRA project component.
type JiraComponent struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Lead        *JiraUserField `json:"lead,omitempty"`
}
