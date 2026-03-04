package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SyncHistoryRepositoryImpl implements repository.SyncHistoryRepository.
type SyncHistoryRepositoryImpl struct {
	db *sql.DB
}

// NewSyncHistoryRepository creates a new SyncHistoryRepositoryImpl.
func NewSyncHistoryRepository(db *sql.DB) *SyncHistoryRepositoryImpl {
	return &SyncHistoryRepositoryImpl{db: db}
}

func (r *SyncHistoryRepositoryImpl) Insert(ctx context.Context, projectID string, syncType string, startedAt time.Time) (int64, error) {
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO sync_history (project_id, sync_type, started_at, status)
		 VALUES (?, ?, ?, 'in_progress') RETURNING id`,
		projectID, syncType, startedAt)

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("insert sync history: %w", err)
	}
	return id, nil
}

func (r *SyncHistoryRepositoryImpl) UpdateCompleted(ctx context.Context, id int64, itemsSynced int, completedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sync_history SET status = 'success', items_synced = ?, completed_at = ? WHERE id = ?`,
		itemsSynced, completedAt, id)
	return err
}

func (r *SyncHistoryRepositoryImpl) UpdateFailed(ctx context.Context, id int64, errorMessage string, completedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sync_history SET status = 'failed', error_message = ?, completed_at = ? WHERE id = ?`,
		errorMessage, completedAt, id)
	return err
}
