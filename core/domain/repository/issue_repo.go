package repository

import (
	"context"
	"time"

	"github.com/ysksm/go-jira/core/domain/models"
)

// IssuePage represents a paginated result of issues.
type IssuePage struct {
	Issues []models.Issue
	Total  int
}

// IssueCursor provides memory-efficient row-by-row iteration over issues.
type IssueCursor interface {
	// Next advances the cursor to the next issue. Returns false when done.
	Next() bool
	// Issue returns the current issue.
	Issue() *models.Issue
	// Err returns any error encountered during iteration.
	Err() error
	// Close releases the cursor resources.
	Close() error
}

// IssueRepository defines operations for issue persistence.
type IssueRepository interface {
	// BatchInsert upserts a batch of issues into the database.
	// Data is written immediately and the caller can release the slice.
	BatchInsert(ctx context.Context, issues []models.Issue) error

	// FindByProjectCursor returns a cursor for memory-efficient iteration.
	// Used by snapshot generation to process one issue at a time.
	FindByProjectCursor(ctx context.Context, projectID string) (IssueCursor, error)

	// FindByProjectPaginated returns a page of issues for search/display.
	FindByProjectPaginated(ctx context.Context, projectID string, offset, limit int) (*IssuePage, error)

	// FindByKey returns a single issue by its key.
	FindByKey(ctx context.Context, key string) (*models.Issue, error)

	// CountByProject returns the total number of active issues for a project.
	CountByProject(ctx context.Context, projectID string) (int, error)

	// CountByStatus returns issue counts grouped by status.
	CountByStatus(ctx context.Context, projectID string) (map[string]int, error)

	// MarkDeletedNotInCurrentSync soft-deletes issues not updated in the current sync.
	// Uses a DB subquery to avoid loading all keys into memory.
	MarkDeletedNotInCurrentSync(ctx context.Context, projectID string, syncedAt time.Time) (int, error)
}
