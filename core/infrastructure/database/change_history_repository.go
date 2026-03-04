package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ysksm/go-jira/core/domain/models"
)

// ChangeHistoryRepositoryImpl implements repository.ChangeHistoryRepository.
type ChangeHistoryRepositoryImpl struct {
	db *sql.DB
}

// NewChangeHistoryRepository creates a new ChangeHistoryRepositoryImpl.
func NewChangeHistoryRepository(db *sql.DB) *ChangeHistoryRepositoryImpl {
	return &ChangeHistoryRepositoryImpl{db: db}
}

func (r *ChangeHistoryRepositoryImpl) BatchInsert(ctx context.Context, items []models.ChangeHistoryItem) error {
	stmt := `INSERT INTO issue_change_history (
		issue_id, issue_key, history_id, author_account_id, author_display_name,
		field, field_type, from_value, from_string, to_value, to_string, changed_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, item := range items {
		_, err := r.db.ExecContext(ctx, stmt,
			item.IssueID, item.IssueKey, item.HistoryID,
			item.AuthorAccountID, item.AuthorDisplayName,
			item.Field, item.FieldType,
			item.FromValue, item.FromString,
			item.ToValue, item.ToString,
			item.ChangedAt,
		)
		if err != nil {
			return fmt.Errorf("insert change history for %s: %w", item.IssueKey, err)
		}
	}
	return nil
}

func (r *ChangeHistoryRepositoryImpl) DeleteByIssueID(ctx context.Context, issueID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM issue_change_history WHERE issue_id = ?`, issueID)
	return err
}

func (r *ChangeHistoryRepositoryImpl) FindByIssueKey(ctx context.Context, issueKey string) ([]models.ChangeHistoryItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT issue_id, issue_key, history_id, author_account_id, author_display_name,
			field, field_type, from_value, from_string, to_value, to_string, changed_at
		FROM issue_change_history
		WHERE issue_key = ?
		ORDER BY changed_at ASC`,
		issueKey,
	)
	if err != nil {
		return nil, fmt.Errorf("query change history for %s: %w", issueKey, err)
	}
	defer rows.Close()

	var items []models.ChangeHistoryItem
	for rows.Next() {
		var item models.ChangeHistoryItem
		var authorAccountID, authorDisplayName sql.NullString
		var fromValue, fromString, toValue, toString sql.NullString
		var fieldType sql.NullString

		err := rows.Scan(
			&item.IssueID, &item.IssueKey, &item.HistoryID,
			&authorAccountID, &authorDisplayName,
			&item.Field, &fieldType,
			&fromValue, &fromString,
			&toValue, &toString,
			&item.ChangedAt,
		)
		if err != nil {
			return nil, err
		}

		item.AuthorAccountID = authorAccountID.String
		item.AuthorDisplayName = authorDisplayName.String
		item.FieldType = fieldType.String
		item.FromValue = fromValue.String
		item.FromString = fromString.String
		item.ToValue = toValue.String
		item.ToString = toString.String

		items = append(items, item)
	}
	return items, rows.Err()
}
