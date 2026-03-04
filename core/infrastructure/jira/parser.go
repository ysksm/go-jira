package jira

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ysksm/go-jira/core/domain/models"
)

// ParseIssue converts a JIRA API issue to a domain Issue.
func ParseIssue(ji *JiraIssue, syncedAt time.Time) (*models.Issue, error) {
	var fields JiraIssueFields
	if err := json.Unmarshal(ji.Fields, &fields); err != nil {
		return nil, fmt.Errorf("unmarshal fields for %s: %w", ji.Key, err)
	}

	issue := &models.Issue{
		ID:       ji.ID,
		Key:      ji.Key,
		Summary:  fields.Summary,
		SyncedAt: &syncedAt,
		RawJSON:  string(ji.Fields),
	}

	if fields.Project != nil {
		issue.ProjectID = fields.Project.ID
	}

	if fields.Description != nil {
		switch v := fields.Description.(type) {
		case string:
			issue.Description = v
		default:
			b, _ := json.Marshal(v)
			issue.Description = string(b)
		}
	}

	if fields.Status != nil {
		issue.Status = fields.Status.Name
	}
	if fields.Priority != nil {
		issue.Priority = fields.Priority.Name
	}
	if fields.Assignee != nil {
		issue.Assignee = fields.Assignee.DisplayName
	}
	if fields.Reporter != nil {
		issue.Reporter = fields.Reporter.DisplayName
	}
	if fields.IssueType != nil {
		issue.IssueType = fields.IssueType.Name
	}
	if fields.Resolution != nil {
		issue.Resolution = fields.Resolution.Name
	}
	if fields.Parent != nil {
		issue.ParentKey = fields.Parent.Key
	}

	issue.Labels = fields.Labels

	if len(fields.Components) > 0 {
		issue.Components = make([]string, len(fields.Components))
		for i, c := range fields.Components {
			issue.Components[i] = c.Name
		}
	}

	if len(fields.FixVersions) > 0 {
		issue.FixVersions = make([]string, len(fields.FixVersions))
		for i, v := range fields.FixVersions {
			issue.FixVersions[i] = v.Name
		}
	}

	if fields.DueDate != "" {
		if t, err := parseJiraDate(fields.DueDate); err == nil {
			issue.DueDate = &t
		}
	}
	if fields.Created != "" {
		if t, err := parseJiraDateTime(fields.Created); err == nil {
			issue.CreatedDate = &t
		}
	}
	if fields.Updated != "" {
		if t, err := parseJiraDateTime(fields.Updated); err == nil {
			issue.UpdatedDate = &t
		}
	}

	// Extract sprint and team from custom fields in raw JSON.
	issue.Sprint = extractSprint(ji.Fields)
	issue.Team = extractTeam(ji.Fields)

	return issue, nil
}

// ExtractChangeHistory extracts change history items from a JIRA issue.
func ExtractChangeHistory(ji *JiraIssue) []models.ChangeHistoryItem {
	if ji.Changelog == nil {
		return nil
	}

	var items []models.ChangeHistoryItem
	for _, h := range ji.Changelog.Histories {
		changedAt, err := parseJiraDateTime(h.Created)
		if err != nil {
			continue
		}

		authorAccountID := ""
		authorDisplayName := ""
		if h.Author != nil {
			authorAccountID = h.Author.AccountID
			authorDisplayName = h.Author.DisplayName
		}

		for _, item := range h.Items {
			items = append(items, models.ChangeHistoryItem{
				IssueID:           ji.ID,
				IssueKey:          ji.Key,
				HistoryID:         h.ID,
				AuthorAccountID:   authorAccountID,
				AuthorDisplayName: authorDisplayName,
				Field:             item.Field,
				FieldType:         item.FieldType,
				FromValue:         item.FromValue,
				FromString:        item.FromString,
				ToValue:           item.ToValue,
				ToString:          item.ToString,
				ChangedAt:         changedAt,
			})
		}
	}
	return items
}

// parseJiraDateTime parses a JIRA datetime string.
// Handles both RFC3339 and JIRA format (with/without timezone colon).
func parseJiraDateTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05Z",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", s)
}

// parseJiraDate parses a JIRA date-only string.
func parseJiraDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
	}
	return t, nil
}

// extractSprint looks for sprint name in custom fields.
func extractSprint(fieldsJSON json.RawMessage) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(fieldsJSON, &raw); err != nil {
		return ""
	}

	sprintFields := []string{"customfield_10020", "customfield_10010", "customfield_10016"}
	for _, field := range sprintFields {
		data, ok := raw[field]
		if !ok || string(data) == "null" {
			continue
		}

		// Sprint can be an array of objects with "name" field.
		var sprints []struct {
			Name  string `json:"name"`
			State string `json:"state"`
		}
		if err := json.Unmarshal(data, &sprints); err == nil && len(sprints) > 0 {
			// Return the active sprint, or the last one.
			for _, s := range sprints {
				if strings.EqualFold(s.State, "active") {
					return s.Name
				}
			}
			return sprints[len(sprints)-1].Name
		}
	}
	return ""
}

// extractTeam looks for team name in custom fields.
func extractTeam(fieldsJSON json.RawMessage) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(fieldsJSON, &raw); err != nil {
		return ""
	}

	teamFields := []string{"customfield_10001", "customfield_10002"}
	for _, field := range teamFields {
		data, ok := raw[field]
		if !ok || string(data) == "null" {
			continue
		}

		// Team can be an object with "name" or "value" field.
		var team struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal(data, &team); err == nil {
			if team.Name != "" {
				return team.Name
			}
			if team.Value != "" {
				return team.Value
			}
		}

		// Or a simple string.
		var teamStr string
		if err := json.Unmarshal(data, &teamStr); err == nil && teamStr != "" {
			return teamStr
		}
	}
	return ""
}
