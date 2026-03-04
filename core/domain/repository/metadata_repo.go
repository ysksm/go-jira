package repository

import (
	"context"

	"github.com/ysksm/go-jira/core/domain/models"
)

// MetadataRepository defines operations for project metadata persistence.
type MetadataRepository interface {
	UpsertStatuses(ctx context.Context, projectID string, statuses []models.Status) error
	UpsertPriorities(ctx context.Context, projectID string, priorities []models.Priority) error
	UpsertIssueTypes(ctx context.Context, projectID string, issueTypes []models.IssueType) error
	UpsertLabels(ctx context.Context, projectID string, labels []models.Label) error
	UpsertComponents(ctx context.Context, projectID string, components []models.Component) error
	UpsertFixVersions(ctx context.Context, projectID string, versions []models.FixVersion) error

	// GetByProject returns all metadata for a project.
	GetByProject(ctx context.Context, projectID string) (*models.ProjectMetadata, error)
}
