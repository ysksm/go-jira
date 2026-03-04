package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// Connection manages DuckDB database connections per project.
type Connection struct {
	databaseDir string
	connections map[string]*sql.DB
}

// NewConnection creates a new connection manager.
func NewConnection(databaseDir string) *Connection {
	return &Connection{
		databaseDir: databaseDir,
		connections: make(map[string]*sql.DB),
	}
}

// GetDB returns a database connection for the given project key.
// Creates the database directory and initializes schema if needed.
func (c *Connection) GetDB(projectKey string) (*sql.DB, error) {
	if db, ok := c.connections[projectKey]; ok {
		return db, nil
	}

	dbDir := filepath.Join(c.databaseDir, projectKey)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("create db dir for %s: %w", projectKey, err)
	}

	dbPath := filepath.Join(dbDir, "data.duckdb")
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open duckdb for %s: %w", projectKey, err)
	}

	if err := InitSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema for %s: %w", projectKey, err)
	}

	c.connections[projectKey] = db
	return db, nil
}

// Close closes all open database connections.
func (c *Connection) Close() error {
	var lastErr error
	for key, db := range c.connections {
		if err := db.Close(); err != nil {
			lastErr = fmt.Errorf("close db for %s: %w", key, err)
		}
		delete(c.connections, key)
	}
	return lastErr
}
