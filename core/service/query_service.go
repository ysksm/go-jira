package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ysksm/go-jira/core/domain/models"
	"github.com/ysksm/go-jira/core/infrastructure/database"
)

// QueryService handles SQL query execution.
type QueryService struct {
	connMgr *database.Connection
}

// NewQueryService creates a new QueryService.
func NewQueryService(connMgr *database.Connection) *QueryService {
	return &QueryService{connMgr: connMgr}
}

// QueryResult holds the result of a SQL query execution.
type QueryResult struct {
	Columns         []string        `json:"columns"`
	Rows            [][]interface{} `json:"rows"`
	RowCount        int             `json:"rowCount"`
	ExecutionTimeMs int64           `json:"executionTimeMs"`
}

// Execute runs a read-only SQL query.
func (s *QueryService) Execute(ctx context.Context, projectKey, query string, limit int) (*QueryResult, error) {
	db, err := s.connMgr.GetDB(projectKey)
	if err != nil {
		return nil, fmt.Errorf("get db for %s: %w", projectKey, err)
	}

	if limit <= 0 {
		limit = 500
	}

	start := time.Now()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	var resultRows [][]interface{}
	count := 0
	for rows.Next() && count < limit {
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		// Convert sql types to JSON-friendly values.
		row := make([]interface{}, len(columns))
		for i, v := range values {
			switch val := v.(type) {
			case []byte:
				row[i] = string(val)
			case sql.NullString:
				if val.Valid {
					row[i] = val.String
				}
			default:
				row[i] = val
			}
		}
		resultRows = append(resultRows, row)
		count++
	}

	return &QueryResult{
		Columns:         columns,
		Rows:            resultRows,
		RowCount:        count,
		ExecutionTimeMs: time.Since(start).Milliseconds(),
	}, nil
}

// GetSchema returns the database schema for a project.
func (s *QueryService) GetSchema(ctx context.Context, projectKey string) ([]models.SqlTable, error) {
	db, err := s.connMgr.GetDB(projectKey)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, `SELECT table_name FROM information_schema.tables WHERE table_schema = 'main' ORDER BY table_name`)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []models.SqlTable
	for rows.Next() {
		var name string
		rows.Scan(&name)

		// Get columns for this table.
		colRows, err := db.QueryContext(ctx,
			`SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = ? ORDER BY ordinal_position`, name)
		if err != nil {
			tables = append(tables, models.SqlTable{Name: name})
			continue
		}

		var columns []models.SqlColumn
		for colRows.Next() {
			var col models.SqlColumn
			var nullable string
			colRows.Scan(&col.Name, &col.DataType, &nullable)
			col.IsNullable = nullable == "YES"
			columns = append(columns, col)
		}
		colRows.Close()

		tables = append(tables, models.SqlTable{Name: name, Columns: columns})
	}
	return tables, nil
}
