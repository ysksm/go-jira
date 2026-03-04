package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ysksm/go-jira/core/domain/models"
)

func TestLoadNonExistentFile(t *testing.T) {
	store, err := NewFileConfigStore(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settings, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settings == nil {
		t.Fatal("expected non-nil default settings")
	}
	if len(settings.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(settings.Projects))
	}
}

func TestInitializeAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store, err := NewFileConfigStore(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settings, err := store.Initialize(
		"https://jira.example.com",
		"user@example.com",
		"api-key-123",
		filepath.Join(t.TempDir(), "data"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if settings.Jira == nil {
		t.Fatal("Jira config should not be nil")
	}
	if settings.Jira.Endpoint != "https://jira.example.com" {
		t.Errorf("Endpoint: got %s", settings.Jira.Endpoint)
	}
	if len(settings.JiraEndpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(settings.JiraEndpoints))
	}
	if settings.ActiveEndpoint != "default" {
		t.Errorf("ActiveEndpoint: got %s, want default", settings.ActiveEndpoint)
	}
	if settings.Sync == nil || !settings.Sync.IncrementalSyncEnabled {
		t.Error("incremental sync should be enabled by default")
	}

	// Verify file was written.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file should exist")
	}

	// Load and verify.
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.Jira.Username != "user@example.com" {
		t.Errorf("loaded Username: got %s", loaded.Jira.Username)
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store, _ := NewFileConfigStore(path)

	settings := &models.Settings{
		Jira: &models.JiraConfig{
			Endpoint: "https://jira.test.com",
			Username: "test",
			APIKey:   "key",
		},
		Database: models.DatabaseConfig{
			DatabaseDir: "/tmp/db",
		},
		Projects: []models.ProjectConfig{
			{ID: "1", Key: "PROJ", Name: "Project", SyncEnabled: true},
		},
	}

	if err := store.Save(settings); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if loaded.Database.DatabaseDir != "/tmp/db" {
		t.Errorf("DatabaseDir: got %s", loaded.Database.DatabaseDir)
	}
	if len(loaded.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(loaded.Projects))
	}
	if loaded.Projects[0].Key != "PROJ" {
		t.Errorf("Project Key: got %s", loaded.Projects[0].Key)
	}
}

func TestGetActiveEndpoint(t *testing.T) {
	settings := &models.Settings{
		JiraEndpoints: []models.JiraEndpoint{
			{Name: "prod", Endpoint: "https://prod.jira.com", Username: "u", APIKey: "k"},
			{Name: "dev", Endpoint: "https://dev.jira.com", Username: "u", APIKey: "k"},
		},
		ActiveEndpoint: "dev",
	}

	ep := settings.GetActiveEndpoint()
	if ep == nil {
		t.Fatal("expected non-nil endpoint")
	}
	if ep.Name != "dev" {
		t.Errorf("got %s, want dev", ep.Name)
	}
	if ep.Endpoint != "https://dev.jira.com" {
		t.Errorf("got %s, want https://dev.jira.com", ep.Endpoint)
	}
}

func TestGetActiveEndpointFallback(t *testing.T) {
	// Falls back to legacy JiraConfig when no endpoints.
	settings := &models.Settings{
		Jira: &models.JiraConfig{
			Endpoint: "https://legacy.jira.com",
			Username: "legacy",
			APIKey:   "legacykey",
		},
	}

	ep := settings.GetActiveEndpoint()
	if ep == nil {
		t.Fatal("expected non-nil endpoint")
	}
	if ep.Endpoint != "https://legacy.jira.com" {
		t.Errorf("got %s, want https://legacy.jira.com", ep.Endpoint)
	}
}

func TestUpdateCheckpoint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store, _ := NewFileConfigStore(path)

	settings := &models.Settings{
		Projects: []models.ProjectConfig{
			{ID: "1", Key: "PROJ", Name: "Project", SyncEnabled: true},
		},
	}
	store.Save(settings)

	checkpoint := &models.SyncCheckpoint{
		LastIssueKey:   "PROJ-100",
		ItemsProcessed: 100,
		TotalItems:     500,
	}
	if err := store.UpdateCheckpoint("PROJ", checkpoint); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, _ := store.Load()
	if loaded.Projects[0].SyncCheckpoint == nil {
		t.Fatal("checkpoint should not be nil")
	}
	if loaded.Projects[0].SyncCheckpoint.LastIssueKey != "PROJ-100" {
		t.Errorf("got %s, want PROJ-100", loaded.Projects[0].SyncCheckpoint.LastIssueKey)
	}
}
