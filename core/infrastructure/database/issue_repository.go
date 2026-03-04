package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ysksm/go-jira/core/domain/models"
	"github.com/ysksm/go-jira/core/domain/repository"
)

// IssueRepositoryImpl implements repository.IssueRepository using DuckDB.
type IssueRepositoryImpl struct {
	db *sql.DB
}

// NewIssueRepository creates a new IssueRepositoryImpl.
func NewIssueRepository(db *sql.DB) *IssueRepositoryImpl {
	return &IssueRepositoryImpl{db: db}
}

func (r *IssueRepositoryImpl) BatchInsert(ctx context.Context, issues []models.Issue) error {
	stmt := `INSERT INTO issues (
		id, project_id, key, summary, description, status, priority,
		assignee, reporter, issue_type, resolution, labels, components,
		fix_versions, sprint, team, parent_key, due_date, created_date,
		updated_date, raw_data, synced_at, is_deleted
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, false)
	ON CONFLICT (id) DO UPDATE SET
		project_id = excluded.project_id,
		key = excluded.key,
		summary = excluded.summary,
		description = excluded.description,
		status = excluded.status,
		priority = excluded.priority,
		assignee = excluded.assignee,
		reporter = excluded.reporter,
		issue_type = excluded.issue_type,
		resolution = excluded.resolution,
		labels = excluded.labels,
		components = excluded.components,
		fix_versions = excluded.fix_versions,
		sprint = excluded.sprint,
		team = excluded.team,
		parent_key = excluded.parent_key,
		due_date = excluded.due_date,
		created_date = excluded.created_date,
		updated_date = excluded.updated_date,
		raw_data = excluded.raw_data,
		synced_at = excluded.synced_at,
		is_deleted = false`

	for _, issue := range issues {
		labelsJSON, _ := json.Marshal(issue.Labels)
		componentsJSON, _ := json.Marshal(issue.Components)
		fixVersionsJSON, _ := json.Marshal(issue.FixVersions)

		_, err := r.db.ExecContext(ctx, stmt,
			issue.ID, issue.ProjectID, issue.Key, issue.Summary, issue.Description,
			issue.Status, issue.Priority, issue.Assignee, issue.Reporter,
			issue.IssueType, issue.Resolution, string(labelsJSON),
			string(componentsJSON), string(fixVersionsJSON),
			issue.Sprint, issue.Team, issue.ParentKey,
			issue.DueDate, issue.CreatedDate, issue.UpdatedDate,
			issue.RawJSON, issue.SyncedAt,
		)
		if err != nil {
			return fmt.Errorf("upsert issue %s: %w", issue.Key, err)
		}
	}
	return nil
}

func (r *IssueRepositoryImpl) FindByProjectCursor(ctx context.Context, projectID string) (repository.IssueCursor, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, key, summary, description, status, priority,
			assignee, reporter, issue_type, resolution, labels, components,
			fix_versions, sprint, team, parent_key, due_date, created_date,
			updated_date, raw_data, synced_at, is_deleted
		FROM issues
		WHERE project_id = ? AND (is_deleted IS NULL OR is_deleted = false)
		ORDER BY id`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("query issues cursor: %w", err)
	}
	return &issueCursor{rows: rows}, nil
}

func (r *IssueRepositoryImpl) FindByProjectPaginated(ctx context.Context, projectID string, offset, limit int) (*repository.IssuePage, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM issues WHERE project_id = ? AND (is_deleted IS NULL OR is_deleted = false)`,
		projectID,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count issues: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, project_id, key, summary, description, status, priority,
			assignee, reporter, issue_type, resolution, labels, components,
			fix_versions, sprint, team, parent_key, due_date, created_date,
			updated_date, synced_at, is_deleted
		FROM issues
		WHERE project_id = ? AND (is_deleted IS NULL OR is_deleted = false)
		ORDER BY key
		LIMIT ? OFFSET ?`,
		projectID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query issues paginated: %w", err)
	}
	defer rows.Close()

	var issues []models.Issue
	for rows.Next() {
		issue, err := scanIssueSummary(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, *issue)
	}
	return &repository.IssuePage{Issues: issues, Total: total}, nil
}

func (r *IssueRepositoryImpl) FindByKey(ctx context.Context, key string) (*models.Issue, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, key, summary, description, status, priority,
			assignee, reporter, issue_type, resolution, labels, components,
			fix_versions, sprint, team, parent_key, due_date, created_date,
			updated_date, raw_data, synced_at, is_deleted
		FROM issues WHERE key = ?`, key)

	issue, err := scanIssueFromRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find issue by key %s: %w", key, err)
	}
	return issue, nil
}

func (r *IssueRepositoryImpl) CountByProject(ctx context.Context, projectID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM issues WHERE project_id = ? AND (is_deleted IS NULL OR is_deleted = false)`,
		projectID,
	).Scan(&count)
	return count, err
}

func (r *IssueRepositoryImpl) CountByStatus(ctx context.Context, projectID string) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT status, COUNT(*) FROM issues
		 WHERE project_id = ? AND (is_deleted IS NULL OR is_deleted = false)
		 GROUP BY status`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		result[status] = count
	}
	return result, nil
}

func (r *IssueRepositoryImpl) MarkDeletedNotInCurrentSync(ctx context.Context, projectID string, syncedAt time.Time) (int, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE issues SET is_deleted = true
		 WHERE project_id = ? AND (synced_at IS NULL OR synced_at < ?) AND (is_deleted IS NULL OR is_deleted = false)`,
		projectID, syncedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("mark deleted: %w", err)
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// issueCursor implements repository.IssueCursor.
type issueCursor struct {
	rows    *sql.Rows
	current *models.Issue
	err     error
}

func (c *issueCursor) Next() bool {
	if !c.rows.Next() {
		c.err = c.rows.Err()
		return false
	}

	var issue models.Issue
	var labelsJSON, componentsJSON, fixVersionsJSON sql.NullString
	var rawData sql.NullString
	var description, status, priority, assignee, reporter sql.NullString
	var issueType, resolution, sprint, team, parentKey sql.NullString
	var dueDate, createdDate, updatedDate, syncedAt sql.NullTime
	var isDeleted sql.NullBool

	c.err = c.rows.Scan(
		&issue.ID, &issue.ProjectID, &issue.Key, &issue.Summary,
		&description, &status, &priority,
		&assignee, &reporter, &issueType, &resolution,
		&labelsJSON, &componentsJSON, &fixVersionsJSON,
		&sprint, &team, &parentKey,
		&dueDate, &createdDate, &updatedDate,
		&rawData, &syncedAt, &isDeleted,
	)
	if c.err != nil {
		return false
	}

	issue.Description = description.String
	issue.Status = status.String
	issue.Priority = priority.String
	issue.Assignee = assignee.String
	issue.Reporter = reporter.String
	issue.IssueType = issueType.String
	issue.Resolution = resolution.String
	issue.Sprint = sprint.String
	issue.Team = team.String
	issue.ParentKey = parentKey.String
	issue.RawJSON = rawData.String
	if isDeleted.Valid {
		issue.IsDeleted = isDeleted.Bool
	}
	if dueDate.Valid {
		issue.DueDate = &dueDate.Time
	}
	if createdDate.Valid {
		issue.CreatedDate = &createdDate.Time
	}
	if updatedDate.Valid {
		issue.UpdatedDate = &updatedDate.Time
	}
	if syncedAt.Valid {
		issue.SyncedAt = &syncedAt.Time
	}

	if labelsJSON.Valid {
		json.Unmarshal([]byte(labelsJSON.String), &issue.Labels)
	}
	if componentsJSON.Valid {
		json.Unmarshal([]byte(componentsJSON.String), &issue.Components)
	}
	if fixVersionsJSON.Valid {
		json.Unmarshal([]byte(fixVersionsJSON.String), &issue.FixVersions)
	}

	c.current = &issue
	return true
}

func (c *issueCursor) Issue() *models.Issue { return c.current }
func (c *issueCursor) Err() error           { return c.err }
func (c *issueCursor) Close() error         { return c.rows.Close() }

// scanIssueSummary scans an issue row without raw_data (for listing).
func scanIssueSummary(rows *sql.Rows) (*models.Issue, error) {
	var issue models.Issue
	var labelsJSON, componentsJSON, fixVersionsJSON sql.NullString
	var description, status, priority, assignee, reporter sql.NullString
	var issueType, resolution, sprint, team, parentKey sql.NullString
	var dueDate, createdDate, updatedDate, syncedAt sql.NullTime
	var isDeleted sql.NullBool

	err := rows.Scan(
		&issue.ID, &issue.ProjectID, &issue.Key, &issue.Summary,
		&description, &status, &priority,
		&assignee, &reporter, &issueType, &resolution,
		&labelsJSON, &componentsJSON, &fixVersionsJSON,
		&sprint, &team, &parentKey,
		&dueDate, &createdDate, &updatedDate,
		&syncedAt, &isDeleted,
	)
	if err != nil {
		return nil, err
	}

	issue.Description = description.String
	issue.Status = status.String
	issue.Priority = priority.String
	issue.Assignee = assignee.String
	issue.Reporter = reporter.String
	issue.IssueType = issueType.String
	issue.Resolution = resolution.String
	issue.Sprint = sprint.String
	issue.Team = team.String
	issue.ParentKey = parentKey.String
	if isDeleted.Valid {
		issue.IsDeleted = isDeleted.Bool
	}
	if dueDate.Valid {
		issue.DueDate = &dueDate.Time
	}
	if createdDate.Valid {
		issue.CreatedDate = &createdDate.Time
	}
	if updatedDate.Valid {
		issue.UpdatedDate = &updatedDate.Time
	}
	if syncedAt.Valid {
		issue.SyncedAt = &syncedAt.Time
	}
	if labelsJSON.Valid {
		json.Unmarshal([]byte(labelsJSON.String), &issue.Labels)
	}
	if componentsJSON.Valid {
		json.Unmarshal([]byte(componentsJSON.String), &issue.Components)
	}
	if fixVersionsJSON.Valid {
		json.Unmarshal([]byte(fixVersionsJSON.String), &issue.FixVersions)
	}
	return &issue, nil
}

// scanIssueFromRow scans a single issue row including raw_data.
func scanIssueFromRow(row *sql.Row) (*models.Issue, error) {
	var issue models.Issue
	var labelsJSON, componentsJSON, fixVersionsJSON sql.NullString
	var rawData sql.NullString
	var description, status, priority, assignee, reporter sql.NullString
	var issueType, resolution, sprint, team, parentKey sql.NullString
	var dueDate, createdDate, updatedDate, syncedAt sql.NullTime
	var isDeleted sql.NullBool

	err := row.Scan(
		&issue.ID, &issue.ProjectID, &issue.Key, &issue.Summary,
		&description, &status, &priority,
		&assignee, &reporter, &issueType, &resolution,
		&labelsJSON, &componentsJSON, &fixVersionsJSON,
		&sprint, &team, &parentKey,
		&dueDate, &createdDate, &updatedDate,
		&rawData, &syncedAt, &isDeleted,
	)
	if err != nil {
		return nil, err
	}

	issue.Description = description.String
	issue.Status = status.String
	issue.Priority = priority.String
	issue.Assignee = assignee.String
	issue.Reporter = reporter.String
	issue.IssueType = issueType.String
	issue.Resolution = resolution.String
	issue.Sprint = sprint.String
	issue.Team = team.String
	issue.ParentKey = parentKey.String
	issue.RawJSON = rawData.String
	if isDeleted.Valid {
		issue.IsDeleted = isDeleted.Bool
	}
	if dueDate.Valid {
		issue.DueDate = &dueDate.Time
	}
	if createdDate.Valid {
		issue.CreatedDate = &createdDate.Time
	}
	if updatedDate.Valid {
		issue.UpdatedDate = &updatedDate.Time
	}
	if syncedAt.Valid {
		issue.SyncedAt = &syncedAt.Time
	}
	if labelsJSON.Valid {
		json.Unmarshal([]byte(labelsJSON.String), &issue.Labels)
	}
	if componentsJSON.Valid {
		json.Unmarshal([]byte(componentsJSON.String), &issue.Components)
	}
	if fixVersionsJSON.Valid {
		json.Unmarshal([]byte(fixVersionsJSON.String), &issue.FixVersions)
	}
	return &issue, nil
}
