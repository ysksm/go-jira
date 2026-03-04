package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ysksm/go-jira/core/domain/models"
	"github.com/ysksm/go-jira/core/infrastructure/config"
	"github.com/ysksm/go-jira/core/infrastructure/jira"
)

// ProjectService manages JIRA projects.
type ProjectService struct {
	configStore *config.FileConfigStore
	logger      *slog.Logger
}

// NewProjectService creates a new ProjectService.
func NewProjectService(configStore *config.FileConfigStore, logger *slog.Logger) *ProjectService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProjectService{configStore: configStore, logger: logger}
}

// List returns all configured projects.
func (s *ProjectService) List() ([]models.Project, error) {
	settings, err := s.configStore.Load()
	if err != nil {
		return nil, err
	}

	projects := make([]models.Project, len(settings.Projects))
	for i, pc := range settings.Projects {
		projects[i] = models.Project{
			ID:      pc.ID,
			Key:     pc.Key,
			Name:    pc.Name,
			Enabled: pc.SyncEnabled,
		}
		if pc.LastSynced != nil {
			projects[i].LastSyncedAt = pc.LastSynced.Format("2006-01-02T15:04:05Z")
		}
	}
	return projects, nil
}

// FetchFromJira fetches projects from JIRA and adds new ones to config.
func (s *ProjectService) FetchFromJira(ctx context.Context) ([]models.Project, int, error) {
	settings, err := s.configStore.Load()
	if err != nil {
		return nil, 0, err
	}

	ep := settings.GetActiveEndpoint()
	if ep == nil {
		return nil, 0, fmt.Errorf("no JIRA endpoint configured")
	}

	client := jira.NewClient(ep, s.logger)
	jiraProjects, err := client.FetchProjects(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch projects from JIRA: %w", err)
	}

	existingKeys := make(map[string]bool)
	for _, pc := range settings.Projects {
		existingKeys[pc.Key] = true
	}

	newCount := 0
	for _, jp := range jiraProjects {
		if !existingKeys[jp.Key] {
			settings.Projects = append(settings.Projects, models.ProjectConfig{
				ID:          jp.ID,
				Key:         jp.Key,
				Name:        jp.Name,
				SyncEnabled: false,
				Endpoint:    ep.Name,
			})
			newCount++
		}
	}

	if newCount > 0 {
		if err := s.configStore.Save(settings); err != nil {
			return nil, 0, err
		}
	}

	return jiraProjects, newCount, nil
}

// Enable enables sync for a project.
func (s *ProjectService) Enable(key string) (*models.Project, error) {
	settings, err := s.configStore.Load()
	if err != nil {
		return nil, err
	}

	for i := range settings.Projects {
		if settings.Projects[i].Key == key {
			settings.Projects[i].SyncEnabled = true
			if err := s.configStore.Save(settings); err != nil {
				return nil, err
			}
			return &models.Project{
				ID:      settings.Projects[i].ID,
				Key:     settings.Projects[i].Key,
				Name:    settings.Projects[i].Name,
				Enabled: true,
			}, nil
		}
	}
	return nil, fmt.Errorf("project %s not found", key)
}

// Disable disables sync for a project.
func (s *ProjectService) Disable(key string) (*models.Project, error) {
	settings, err := s.configStore.Load()
	if err != nil {
		return nil, err
	}

	for i := range settings.Projects {
		if settings.Projects[i].Key == key {
			settings.Projects[i].SyncEnabled = false
			if err := s.configStore.Save(settings); err != nil {
				return nil, err
			}
			return &models.Project{
				ID:      settings.Projects[i].ID,
				Key:     settings.Projects[i].Key,
				Name:    settings.Projects[i].Name,
				Enabled: false,
			}, nil
		}
	}
	return nil, fmt.Errorf("project %s not found", key)
}
