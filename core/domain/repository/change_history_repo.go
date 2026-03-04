package repository

import (
	"context"

	"github.com/ysksm/go-jira/core/domain/models"
)

// ChangeHistoryRepository defines operations for change history persistence.
type ChangeHistoryRepository interface {
	// BatchInsert inserts change history items into the database.
	BatchInsert(ctx context.Context, items []models.ChangeHistoryItem) error

	// DeleteByIssueID deletes all change history for a specific issue.
	DeleteByIssueID(ctx context.Context, issueID string) error

	// FindByIssueKey returns all change history items for an issue.
	// Used during snapshot generation (one issue at a time).
	FindByIssueKey(ctx context.Context, issueKey string) ([]models.ChangeHistoryItem, error)
}
