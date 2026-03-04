package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ysksm/go-jira/core/domain/models"
)

const (
	configDirName  = "go-jira"
	configFileName = "settings.json"
)

// FileConfigStore manages settings persistence to a JSON file.
type FileConfigStore struct {
	path string
}

// NewFileConfigStore creates a new file-based config store.
// Uses ~/.config/go-jira/settings.json by default.
func NewFileConfigStore(path string) (*FileConfigStore, error) {
	if path == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("get config dir: %w", err)
		}
		path = filepath.Join(configDir, configDirName, configFileName)
	}
	return &FileConfigStore{path: path}, nil
}

// Path returns the config file path.
func (s *FileConfigStore) Path() string {
	return s.path
}

// Load reads settings from the config file.
// Returns default settings if the file doesn't exist.
func (s *FileConfigStore) Load() (*models.Settings, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.Settings{
				Database: models.DatabaseConfig{},
			}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var settings models.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}
	return &settings, nil
}

// Save writes settings to the config file.
func (s *FileConfigStore) Save(settings *models.Settings) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

// Initialize creates a new config with the given JIRA credentials.
func (s *FileConfigStore) Initialize(endpoint, username, apiKey, dbPath string) (*models.Settings, error) {
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		dbPath = filepath.Join(homeDir, ".local", "share", configDirName, "data")
	}

	settings := &models.Settings{
		Jira: &models.JiraConfig{
			Endpoint: endpoint,
			Username: username,
			APIKey:   apiKey,
		},
		JiraEndpoints: []models.JiraEndpoint{
			{
				Name:     "default",
				Endpoint: endpoint,
				Username: username,
				APIKey:   apiKey,
			},
		},
		ActiveEndpoint: "default",
		Database: models.DatabaseConfig{
			DatabaseDir: dbPath,
		},
		Sync: &models.SyncSettings{
			IncrementalSyncEnabled:       true,
			IncrementalSyncMarginMinutes: 5,
		},
	}

	if err := s.Save(settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// UpdateCheckpoint updates the sync checkpoint for a project.
func (s *FileConfigStore) UpdateCheckpoint(projectKey string, checkpoint *models.SyncCheckpoint) error {
	settings, err := s.Load()
	if err != nil {
		return err
	}

	for i := range settings.Projects {
		if settings.Projects[i].Key == projectKey {
			settings.Projects[i].SyncCheckpoint = checkpoint
			return s.Save(settings)
		}
	}
	return fmt.Errorf("project %s not found in config", projectKey)
}

// UpdateSnapshotCheckpoint updates the snapshot checkpoint for a project.
func (s *FileConfigStore) UpdateSnapshotCheckpoint(projectKey string, checkpoint *models.SnapshotCheckpoint) error {
	settings, err := s.Load()
	if err != nil {
		return err
	}

	for i := range settings.Projects {
		if settings.Projects[i].Key == projectKey {
			settings.Projects[i].SnapshotCheckpoint = checkpoint
			return s.Save(settings)
		}
	}
	return fmt.Errorf("project %s not found in config", projectKey)
}
