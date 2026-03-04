package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ysksm/go-jira/core/domain/models"
)

// MetadataRepositoryImpl implements repository.MetadataRepository.
type MetadataRepositoryImpl struct {
	db *sql.DB
}

// NewMetadataRepository creates a new MetadataRepositoryImpl.
func NewMetadataRepository(db *sql.DB) *MetadataRepositoryImpl {
	return &MetadataRepositoryImpl{db: db}
}

func (r *MetadataRepositoryImpl) UpsertStatuses(ctx context.Context, projectID string, statuses []models.Status) error {
	r.db.ExecContext(ctx, `DELETE FROM statuses WHERE project_id = ?`, projectID)
	for _, s := range statuses {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO statuses (project_id, name, description, category) VALUES (?, ?, ?, ?)`,
			projectID, s.Name, s.Description, s.Category)
		if err != nil {
			return fmt.Errorf("upsert status %s: %w", s.Name, err)
		}
	}
	return nil
}

func (r *MetadataRepositoryImpl) UpsertPriorities(ctx context.Context, projectID string, priorities []models.Priority) error {
	r.db.ExecContext(ctx, `DELETE FROM priorities WHERE project_id = ?`, projectID)
	for _, p := range priorities {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO priorities (project_id, name, description, icon_url) VALUES (?, ?, ?, ?)`,
			projectID, p.Name, p.Description, p.IconURL)
		if err != nil {
			return fmt.Errorf("upsert priority %s: %w", p.Name, err)
		}
	}
	return nil
}

func (r *MetadataRepositoryImpl) UpsertIssueTypes(ctx context.Context, projectID string, issueTypes []models.IssueType) error {
	r.db.ExecContext(ctx, `DELETE FROM issue_types WHERE project_id = ?`, projectID)
	for _, t := range issueTypes {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO issue_types (project_id, name, description, icon_url, subtask) VALUES (?, ?, ?, ?, ?)`,
			projectID, t.Name, t.Description, t.IconURL, t.Subtask)
		if err != nil {
			return fmt.Errorf("upsert issue type %s: %w", t.Name, err)
		}
	}
	return nil
}

func (r *MetadataRepositoryImpl) UpsertLabels(ctx context.Context, projectID string, labels []models.Label) error {
	r.db.ExecContext(ctx, `DELETE FROM labels WHERE project_id = ?`, projectID)
	for _, l := range labels {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO labels (project_id, name) VALUES (?, ?)`,
			projectID, l.Name)
		if err != nil {
			return fmt.Errorf("upsert label %s: %w", l.Name, err)
		}
	}
	return nil
}

func (r *MetadataRepositoryImpl) UpsertComponents(ctx context.Context, projectID string, components []models.Component) error {
	r.db.ExecContext(ctx, `DELETE FROM components WHERE project_id = ?`, projectID)
	for _, c := range components {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO components (project_id, name, description, lead) VALUES (?, ?, ?, ?)`,
			projectID, c.Name, c.Description, c.Lead)
		if err != nil {
			return fmt.Errorf("upsert component %s: %w", c.Name, err)
		}
	}
	return nil
}

func (r *MetadataRepositoryImpl) UpsertFixVersions(ctx context.Context, projectID string, versions []models.FixVersion) error {
	r.db.ExecContext(ctx, `DELETE FROM fix_versions WHERE project_id = ?`, projectID)
	for _, v := range versions {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO fix_versions (project_id, name, description, released, release_date) VALUES (?, ?, ?, ?, ?)`,
			projectID, v.Name, v.Description, v.Released, v.ReleaseDate)
		if err != nil {
			return fmt.Errorf("upsert fix version %s: %w", v.Name, err)
		}
	}
	return nil
}

func (r *MetadataRepositoryImpl) GetByProject(ctx context.Context, projectID string) (*models.ProjectMetadata, error) {
	meta := &models.ProjectMetadata{ProjectKey: projectID}

	// Statuses
	rows, err := r.db.QueryContext(ctx, `SELECT name, description, category FROM statuses WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var s models.Status
		rows.Scan(&s.Name, &s.Description, &s.Category)
		meta.Statuses = append(meta.Statuses, s)
	}
	rows.Close()

	// Priorities
	rows, err = r.db.QueryContext(ctx, `SELECT name, description, icon_url FROM priorities WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var p models.Priority
		rows.Scan(&p.Name, &p.Description, &p.IconURL)
		meta.Priorities = append(meta.Priorities, p)
	}
	rows.Close()

	// Issue Types
	rows, err = r.db.QueryContext(ctx, `SELECT name, description, icon_url, subtask FROM issue_types WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var t models.IssueType
		rows.Scan(&t.Name, &t.Description, &t.IconURL, &t.Subtask)
		meta.IssueTypes = append(meta.IssueTypes, t)
	}
	rows.Close()

	// Labels
	rows, err = r.db.QueryContext(ctx, `SELECT name FROM labels WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var l models.Label
		rows.Scan(&l.Name)
		meta.Labels = append(meta.Labels, l)
	}
	rows.Close()

	// Components
	rows, err = r.db.QueryContext(ctx, `SELECT name, description, lead FROM components WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var c models.Component
		rows.Scan(&c.Name, &c.Description, &c.Lead)
		meta.Components = append(meta.Components, c)
	}
	rows.Close()

	// Fix Versions
	rows, err = r.db.QueryContext(ctx, `SELECT name, description, released, release_date FROM fix_versions WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var v models.FixVersion
		rows.Scan(&v.Name, &v.Description, &v.Released, &v.ReleaseDate)
		meta.FixVersions = append(meta.FixVersions, v)
	}
	rows.Close()

	return meta, nil
}
