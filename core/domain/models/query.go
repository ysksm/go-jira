package models

// SavedQuery represents a saved SQL query.
type SavedQuery struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Query       string `json:"query"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// SqlTable represents a database table in the schema.
type SqlTable struct {
	Name    string      `json:"name"`
	Columns []SqlColumn `json:"columns,omitempty"`
}

// SqlColumn represents a column in a database table.
type SqlColumn struct {
	Name       string `json:"name"`
	DataType   string `json:"dataType"`
	IsNullable bool   `json:"isNullable"`
}
