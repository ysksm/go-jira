package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestJiraAPIError(t *testing.T) {
	err := &JiraAPIError{
		StatusCode: 401,
		Message:    "Unauthorized",
		URL:        "https://jira.example.com/rest/api/3/project",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
	if !contains(msg, "401") {
		t.Errorf("error should contain status code: %s", msg)
	}
}

func TestDatabaseError(t *testing.T) {
	inner := fmt.Errorf("connection refused")
	err := &DatabaseError{
		Operation: "batch_insert",
		Err:       inner,
	}

	msg := err.Error()
	if !contains(msg, "batch_insert") {
		t.Errorf("error should contain operation: %s", msg)
	}

	// Test Unwrap.
	if !errors.Is(err, inner) {
		t.Error("should unwrap to inner error")
	}
}

func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Field:   "endpoint",
		Message: "cannot be empty",
	}

	msg := err.Error()
	if !contains(msg, "endpoint") || !contains(msg, "cannot be empty") {
		t.Errorf("error should contain field and message: %s", msg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
