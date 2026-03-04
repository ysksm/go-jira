package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ysksm/go-jira/core/domain/models"
)

// SnapshotRepositoryImpl implements repository.SnapshotRepository.
type SnapshotRepositoryImpl struct {
	db *sql.DB
	tx *sql.Tx
}

// NewSnapshotRepository creates a new SnapshotRepositoryImpl.
func NewSnapshotRepository(db *sql.DB) *SnapshotRepositoryImpl {
	return &SnapshotRepositoryImpl{db: db}
}

func (r *SnapshotRepositoryImpl) BatchInsert(ctx context.Context, snapshots []models.IssueSnapshot) error {
	stmt := `INSERT INTO issue_snapshots (
		issue_id, issue_key, project_id, version, valid_from, valid_to,
		summary, description, status, priority, assignee, reporter,
		issue_type, resolution, labels, components, fix_versions,
		sprint, parent_key, raw_data, updated_date, resolved_date, due_date, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	execFn := r.db.ExecContext
	if r.tx != nil {
		execFn = r.tx.ExecContext
	}

	for _, s := range snapshots {
		labelsJSON, _ := json.Marshal(s.Labels)
		componentsJSON, _ := json.Marshal(s.Components)
		fixVersionsJSON, _ := json.Marshal(s.FixVersions)

		var rawDataStr interface{}
		if s.RawData != nil {
			rawDataStr = string(*s.RawData)
		}

		_, err := execFn(ctx, stmt,
			s.IssueID, s.IssueKey, s.ProjectID, s.Version,
			s.ValidFrom, s.ValidTo,
			s.Summary, s.Description, s.Status, s.Priority,
			s.Assignee, s.Reporter, s.IssueType, s.Resolution,
			string(labelsJSON), string(componentsJSON), string(fixVersionsJSON),
			s.Sprint, s.ParentKey, rawDataStr,
			s.UpdatedDate, s.ResolvedDate, s.DueDate, s.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert snapshot %s v%d: %w", s.IssueKey, s.Version, err)
		}
	}
	return nil
}

func (r *SnapshotRepositoryImpl) DeleteByIssueID(ctx context.Context, issueID string) error {
	stmt := `DELETE FROM issue_snapshots WHERE issue_id = ?`
	if r.tx != nil {
		_, err := r.tx.ExecContext(ctx, stmt, issueID)
		return err
	}
	_, err := r.db.ExecContext(ctx, stmt, issueID)
	return err
}

func (r *SnapshotRepositoryImpl) BeginTransaction(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	r.tx = tx
	return nil
}

func (r *SnapshotRepositoryImpl) CommitTransaction(_ context.Context) error {
	if r.tx == nil {
		return fmt.Errorf("no active transaction")
	}
	err := r.tx.Commit()
	r.tx = nil
	return err
}

func (r *SnapshotRepositoryImpl) RollbackTransaction(_ context.Context) error {
	if r.tx == nil {
		return nil
	}
	err := r.tx.Rollback()
	r.tx = nil
	return err
}
