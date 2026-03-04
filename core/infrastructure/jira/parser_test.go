package jira

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseJiraDateTime(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2024-01-15T10:30:00.000+0900", "2024-01-15T01:30:00Z"},
		{"2024-01-15T10:30:00.000Z", "2024-01-15T10:30:00Z"},
		{"2024-01-15T10:30:00Z", "2024-01-15T10:30:00Z"},
		{"2024-01-15T10:30:00+09:00", "2024-01-15T01:30:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseJiraDateTime(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Format(time.RFC3339) != tt.want {
				t.Errorf("got %s, want %s", got.Format(time.RFC3339), tt.want)
			}
		})
	}
}

func TestParseJiraDate(t *testing.T) {
	got, err := parseJiraDate("2024-03-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Format("2006-01-02") != "2024-03-15" {
		t.Errorf("got %s, want 2024-03-15", got.Format("2006-01-02"))
	}
}

func TestParseIssue(t *testing.T) {
	fields := JiraIssueFields{
		Summary: "Test issue",
		Status:  &JiraNamedField{Name: "Open"},
		Priority: &JiraNamedField{Name: "High"},
		Assignee: &JiraUserField{AccountID: "abc", DisplayName: "Alice"},
		Reporter: &JiraUserField{AccountID: "def", DisplayName: "Bob"},
		IssueType: &JiraNamedField{Name: "Bug"},
		Labels:   []string{"backend", "urgent"},
		Components: []JiraNamedField{{Name: "API"}, {Name: "Auth"}},
		FixVersions: []JiraNamedField{{Name: "1.0"}},
		Parent:   &JiraParentField{Key: "PROJ-1"},
		Created:  "2024-01-15T10:00:00.000Z",
		Updated:  "2024-01-20T15:30:00.000Z",
		Project:  &JiraProjectField{ID: "10001", Key: "PROJ", Name: "Project"},
	}

	fieldsJSON, _ := json.Marshal(fields)
	ji := &JiraIssue{
		ID:     "12345",
		Key:    "PROJ-42",
		Fields: fieldsJSON,
	}

	syncedAt := time.Date(2024, 1, 20, 16, 0, 0, 0, time.UTC)
	issue, err := ParseIssue(ji, syncedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if issue.ID != "12345" {
		t.Errorf("ID: got %s, want 12345", issue.ID)
	}
	if issue.Key != "PROJ-42" {
		t.Errorf("Key: got %s, want PROJ-42", issue.Key)
	}
	if issue.Summary != "Test issue" {
		t.Errorf("Summary: got %s, want Test issue", issue.Summary)
	}
	if issue.Status != "Open" {
		t.Errorf("Status: got %s, want Open", issue.Status)
	}
	if issue.Priority != "High" {
		t.Errorf("Priority: got %s, want High", issue.Priority)
	}
	if issue.Assignee != "Alice" {
		t.Errorf("Assignee: got %s, want Alice", issue.Assignee)
	}
	if issue.Reporter != "Bob" {
		t.Errorf("Reporter: got %s, want Bob", issue.Reporter)
	}
	if issue.IssueType != "Bug" {
		t.Errorf("IssueType: got %s, want Bug", issue.IssueType)
	}
	if issue.ProjectID != "10001" {
		t.Errorf("ProjectID: got %s, want 10001", issue.ProjectID)
	}
	if issue.ParentKey != "PROJ-1" {
		t.Errorf("ParentKey: got %s, want PROJ-1", issue.ParentKey)
	}
	if len(issue.Labels) != 2 || issue.Labels[0] != "backend" {
		t.Errorf("Labels: got %v, want [backend urgent]", issue.Labels)
	}
	if len(issue.Components) != 2 || issue.Components[0] != "API" {
		t.Errorf("Components: got %v, want [API Auth]", issue.Components)
	}
	if len(issue.FixVersions) != 1 || issue.FixVersions[0] != "1.0" {
		t.Errorf("FixVersions: got %v, want [1.0]", issue.FixVersions)
	}
}

func TestExtractChangeHistory(t *testing.T) {
	fields, _ := json.Marshal(JiraIssueFields{Summary: "test"})
	ji := &JiraIssue{
		ID:     "12345",
		Key:    "PROJ-42",
		Fields: fields,
		Changelog: &JiraChangelog{
			Histories: []JiraHistory{
				{
					ID:      "100",
					Author:  &JiraUserField{AccountID: "abc", DisplayName: "Alice"},
					Created: "2024-01-16T10:00:00.000Z",
					Items: []JiraHistoryItem{
						{
							Field:      "status",
							FieldType:  "jira",
							FromValue:  "1",
							FromString: "Open",
							ToValue:    "2",
							ToString:   "In Progress",
						},
						{
							Field:      "assignee",
							FieldType:  "jira",
							FromString: "",
							ToString:   "Bob",
						},
					},
				},
			},
		},
	}

	items := ExtractChangeHistory(ji)

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].Field != "status" {
		t.Errorf("items[0].Field: got %s, want status", items[0].Field)
	}
	if items[0].FromString != "Open" {
		t.Errorf("items[0].FromString: got %s, want Open", items[0].FromString)
	}
	if items[0].ToString != "In Progress" {
		t.Errorf("items[0].ToString: got %s, want In Progress", items[0].ToString)
	}
	if items[0].AuthorDisplayName != "Alice" {
		t.Errorf("items[0].Author: got %s, want Alice", items[0].AuthorDisplayName)
	}
	if items[0].IssueKey != "PROJ-42" {
		t.Errorf("items[0].IssueKey: got %s, want PROJ-42", items[0].IssueKey)
	}

	if items[1].Field != "assignee" {
		t.Errorf("items[1].Field: got %s, want assignee", items[1].Field)
	}
	if items[1].ToString != "Bob" {
		t.Errorf("items[1].ToString: got %s, want Bob", items[1].ToString)
	}
}

func TestExtractSprint(t *testing.T) {
	raw := map[string]interface{}{
		"summary": "test",
		"customfield_10020": []interface{}{
			map[string]interface{}{"name": "Sprint 1", "state": "closed"},
			map[string]interface{}{"name": "Sprint 2", "state": "active"},
		},
	}
	data, _ := json.Marshal(raw)

	got := extractSprint(data)
	if got != "Sprint 2" {
		t.Errorf("got %s, want Sprint 2 (active sprint)", got)
	}
}

func TestExtractSprintNoActive(t *testing.T) {
	raw := map[string]interface{}{
		"summary": "test",
		"customfield_10020": []interface{}{
			map[string]interface{}{"name": "Sprint 1", "state": "closed"},
			map[string]interface{}{"name": "Sprint 3", "state": "future"},
		},
	}
	data, _ := json.Marshal(raw)

	got := extractSprint(data)
	if got != "Sprint 3" {
		t.Errorf("got %s, want Sprint 3 (last sprint)", got)
	}
}

func TestExtractTeamObject(t *testing.T) {
	raw := map[string]interface{}{
		"summary":          "test",
		"customfield_10001": map[string]interface{}{"name": "Backend Team"},
	}
	data, _ := json.Marshal(raw)

	got := extractTeam(data)
	if got != "Backend Team" {
		t.Errorf("got %s, want Backend Team", got)
	}
}
