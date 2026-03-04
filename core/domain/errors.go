package domain

import (
	"errors"
	"fmt"
)

// Sentinel errors for domain operations.
var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrInvalidInput   = errors.New("invalid input")
	ErrNotConfigured  = errors.New("not configured")
	ErrSyncInProgress = errors.New("sync already in progress")
)

// JiraAPIError represents an error from the JIRA REST API.
type JiraAPIError struct {
	StatusCode int
	Message    string
	URL        string
}

func (e *JiraAPIError) Error() string {
	return fmt.Sprintf("JIRA API error (HTTP %d) %s: %s", e.StatusCode, e.URL, e.Message)
}

// DatabaseError represents a database operation error.
type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error in %s: %v", e.Operation, e.Err)
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// ConfigError represents a configuration error.
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error [%s]: %s", e.Field, e.Message)
}
