package repository

import (
	"context"

	"github.com/ysksm/go-jira/core/domain/models"
)

// SnapshotRepository defines operations for issue snapshot persistence.
type SnapshotRepository interface {
	// BatchInsert inserts snapshots into the database.
	BatchInsert(ctx context.Context, snapshots []models.IssueSnapshot) error

	// DeleteByIssueID deletes all snapshots for a specific issue.
	DeleteByIssueID(ctx context.Context, issueID string) error

	// BeginTransaction starts a database transaction.
	BeginTransaction(ctx context.Context) error

	// CommitTransaction commits the current transaction.
	CommitTransaction(ctx context.Context) error

	// RollbackTransaction rolls back the current transaction.
	RollbackTransaction(ctx context.Context) error
}
