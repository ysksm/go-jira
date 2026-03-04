package service

import (
	"testing"
	"time"

	"github.com/ysksm/go-jira/core/domain/models"
)

func TestGroupChangesByTime(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC)

	history := []models.ChangeHistoryItem{
		{Field: "status", ChangedAt: t1, FromString: "Open", ToString: "In Progress"},
		{Field: "assignee", ChangedAt: t1, FromString: "", ToString: "Alice"},
		{Field: "status", ChangedAt: t2, FromString: "In Progress", ToString: "Done"},
	}

	groups := groupChangesByTime(history)

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if len(groups[0].changes) != 2 {
		t.Errorf("group[0] expected 2 changes, got %d", len(groups[0].changes))
	}
	if len(groups[1].changes) != 1 {
		t.Errorf("group[1] expected 1 change, got %d", len(groups[1].changes))
	}
	if !groups[0].timestamp.Equal(t1) {
		t.Errorf("group[0].timestamp: got %v, want %v", groups[0].timestamp, t1)
	}
}

func TestReverseAndApplyChange(t *testing.T) {
	state := issueState{
		Status:   "Done",
		Assignee: "Bob",
		Priority: "High",
	}

	// Reverse: Done → In Progress (fromString = "In Progress")
	reverseChange(&state, &models.ChangeHistoryItem{
		Field: "status", FromString: "In Progress", ToString: "Done",
	})
	if state.Status != "In Progress" {
		t.Errorf("after reverse, status: got %s, want In Progress", state.Status)
	}

	// Apply forward: In Progress → Done
	applyChange(&state, &models.ChangeHistoryItem{
		Field: "status", FromString: "In Progress", ToString: "Done",
	})
	if state.Status != "Done" {
		t.Errorf("after apply, status: got %s, want Done", state.Status)
	}
}

func TestReverseMultipleChanges(t *testing.T) {
	// Current state: status=Done, assignee=Bob
	// Change 1 (t1): status Open→InProgress, assignee ""→Alice
	// Change 2 (t2): status InProgress→Done, assignee Alice→Bob

	state := issueState{
		Status:   "Done",
		Assignee: "Bob",
	}

	changes := []models.ChangeHistoryItem{
		{Field: "status", FromString: "Open", ToString: "In Progress"},
		{Field: "assignee", FromString: "", ToString: "Alice"},
		{Field: "status", FromString: "In Progress", ToString: "Done"},
		{Field: "assignee", FromString: "Alice", ToString: "Bob"},
	}

	// Reverse all changes (from last to first) to get initial state.
	for i := len(changes) - 1; i >= 0; i-- {
		reverseChange(&state, &changes[i])
	}

	if state.Status != "Open" {
		t.Errorf("initial status: got %s, want Open", state.Status)
	}
	if state.Assignee != "" {
		t.Errorf("initial assignee: got %s, want empty", state.Assignee)
	}

	// Apply all changes forward to reconstruct final state.
	for i := range changes {
		applyChange(&state, &changes[i])
	}

	if state.Status != "Done" {
		t.Errorf("final status: got %s, want Done", state.Status)
	}
	if state.Assignee != "Bob" {
		t.Errorf("final assignee: got %s, want Bob", state.Assignee)
	}
}

func TestBuildSnapshotFromIssue(t *testing.T) {
	created := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2024, 1, 20, 15, 0, 0, 0, time.UTC)

	issue := &models.Issue{
		ID:          "123",
		Key:         "PROJ-1",
		ProjectID:   "P1",
		Summary:     "Test",
		Status:      "Open",
		Priority:    "High",
		CreatedDate: &created,
		UpdatedDate: &updated,
	}

	snapshot := buildSnapshotFromIssue(issue, 1, issue.CreatedDate, nil)

	if snapshot.IssueID != "123" {
		t.Errorf("IssueID: got %s, want 123", snapshot.IssueID)
	}
	if snapshot.Version != 1 {
		t.Errorf("Version: got %d, want 1", snapshot.Version)
	}
	if !snapshot.ValidFrom.Equal(created) {
		t.Errorf("ValidFrom: got %v, want %v", snapshot.ValidFrom, created)
	}
	if snapshot.ValidTo != nil {
		t.Errorf("ValidTo: got %v, want nil", snapshot.ValidTo)
	}
	if snapshot.Status != "Open" {
		t.Errorf("Status: got %s, want Open", snapshot.Status)
	}
}
