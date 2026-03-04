package repository

import (
	"context"
	"time"
)

// SyncHistoryRepository defines operations for sync history tracking.
type SyncHistoryRepository interface {
	// Insert records the start of a sync operation. Returns the record ID.
	Insert(ctx context.Context, projectID string, syncType string, startedAt time.Time) (int64, error)

	// UpdateCompleted marks a sync as successfully completed.
	UpdateCompleted(ctx context.Context, id int64, itemsSynced int, completedAt time.Time) error

	// UpdateFailed marks a sync as failed with an error message.
	UpdateFailed(ctx context.Context, id int64, errorMessage string, completedAt time.Time) error
}
