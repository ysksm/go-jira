package service

import (
	"fmt"

	"github.com/ysksm/go-jira/core/domain/models"
	"github.com/ysksm/go-jira/core/infrastructure/config"
)

// ConfigService manages application configuration.
type ConfigService struct {
	store *config.FileConfigStore
}

// NewConfigService creates a new ConfigService.
func NewConfigService(store *config.FileConfigStore) *ConfigService {
	return &ConfigService{store: store}
}

// Get returns the current settings.
func (s *ConfigService) Get() (*models.Settings, error) {
	return s.store.Load()
}

// Initialize creates a new configuration with JIRA credentials.
func (s *ConfigService) Initialize(endpoint, username, apiKey, dbPath string) (*models.Settings, error) {
	if endpoint == "" || username == "" || apiKey == "" {
		return nil, fmt.Errorf("endpoint, username, and apiKey are required")
	}
	return s.store.Initialize(endpoint, username, apiKey, dbPath)
}

// Update modifies the current settings.
func (s *ConfigService) Update(updateFn func(settings *models.Settings)) (*models.Settings, error) {
	settings, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	updateFn(settings)
	if err := s.store.Save(settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// AddEndpoint adds a new JIRA endpoint.
func (s *ConfigService) AddEndpoint(ep models.JiraEndpoint) (*models.Settings, error) {
	return s.Update(func(settings *models.Settings) {
		settings.JiraEndpoints = append(settings.JiraEndpoints, ep)
		if settings.ActiveEndpoint == "" {
			settings.ActiveEndpoint = ep.Name
		}
	})
}

// RemoveEndpoint removes a JIRA endpoint by name.
func (s *ConfigService) RemoveEndpoint(name string) (*models.Settings, error) {
	return s.Update(func(settings *models.Settings) {
		filtered := settings.JiraEndpoints[:0]
		for _, ep := range settings.JiraEndpoints {
			if ep.Name != name {
				filtered = append(filtered, ep)
			}
		}
		settings.JiraEndpoints = filtered
		if settings.ActiveEndpoint == name && len(filtered) > 0 {
			settings.ActiveEndpoint = filtered[0].Name
		}
	})
}

// SetActiveEndpoint switches the active endpoint.
func (s *ConfigService) SetActiveEndpoint(name string) (*models.Settings, error) {
	return s.Update(func(settings *models.Settings) {
		settings.ActiveEndpoint = name
	})
}
