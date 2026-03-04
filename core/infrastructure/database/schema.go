package database

import "database/sql"

// InitSchema creates all tables and indexes if they don't exist.
func InitSchema(db *sql.DB) error {
	statements := []string{
		createIssuesTable,
		createChangeHistoryTable,
		createSnapshotsTable,
		createStatusesTable,
		createPrioritiesTable,
		createIssueTypesTable,
		createLabelsTable,
		createComponentsTable,
		createFixVersionsTable,
		createSyncHistoryTable,
		createJiraFieldsTable,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

const createIssuesTable = `
CREATE TABLE IF NOT EXISTS issues (
    id VARCHAR PRIMARY KEY,
    project_id VARCHAR,
    key VARCHAR,
    summary TEXT,
    description TEXT,
    status VARCHAR,
    priority VARCHAR,
    assignee VARCHAR,
    reporter VARCHAR,
    issue_type VARCHAR,
    resolution VARCHAR,
    labels JSON,
    components JSON,
    fix_versions JSON,
    sprint VARCHAR,
    team VARCHAR,
    parent_key VARCHAR,
    due_date TIMESTAMPTZ,
    created_date TIMESTAMPTZ,
    updated_date TIMESTAMPTZ,
    raw_data JSON,
    synced_at TIMESTAMPTZ,
    is_deleted BOOLEAN DEFAULT false
);
CREATE INDEX IF NOT EXISTS idx_issues_project ON issues(project_id);
CREATE INDEX IF NOT EXISTS idx_issues_key ON issues(key);
CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);
`

const createChangeHistoryTable = `
CREATE TABLE IF NOT EXISTS issue_change_history (
    id INTEGER PRIMARY KEY DEFAULT nextval('seq_change_history_id'),
    issue_id VARCHAR,
    issue_key VARCHAR,
    history_id VARCHAR,
    author_account_id VARCHAR,
    author_display_name VARCHAR,
    field VARCHAR,
    field_type VARCHAR,
    from_value TEXT,
    from_string TEXT,
    to_value TEXT,
    to_string TEXT,
    changed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT current_timestamp
);
CREATE SEQUENCE IF NOT EXISTS seq_change_history_id START 1;
CREATE INDEX IF NOT EXISTS idx_change_history_issue_id ON issue_change_history(issue_id);
CREATE INDEX IF NOT EXISTS idx_change_history_issue_key ON issue_change_history(issue_key);
CREATE INDEX IF NOT EXISTS idx_change_history_field ON issue_change_history(field);
CREATE INDEX IF NOT EXISTS idx_change_history_changed_at ON issue_change_history(changed_at);
`

const createSnapshotsTable = `
CREATE TABLE IF NOT EXISTS issue_snapshots (
    issue_id VARCHAR,
    issue_key VARCHAR,
    project_id VARCHAR,
    version INTEGER,
    valid_from TIMESTAMPTZ,
    valid_to TIMESTAMPTZ,
    summary TEXT,
    description TEXT,
    status VARCHAR,
    priority VARCHAR,
    assignee VARCHAR,
    reporter VARCHAR,
    issue_type VARCHAR,
    resolution VARCHAR,
    labels JSON,
    components JSON,
    fix_versions JSON,
    sprint VARCHAR,
    parent_key VARCHAR,
    raw_data JSON,
    updated_date TIMESTAMPTZ,
    resolved_date TIMESTAMPTZ,
    due_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT current_timestamp,
    PRIMARY KEY (issue_id, version)
);
CREATE INDEX IF NOT EXISTS idx_snapshots_issue_key ON issue_snapshots(issue_key);
CREATE INDEX IF NOT EXISTS idx_snapshots_project_id ON issue_snapshots(project_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_valid_from ON issue_snapshots(valid_from);
CREATE INDEX IF NOT EXISTS idx_snapshots_valid_to ON issue_snapshots(valid_to);
`

const createStatusesTable = `
CREATE TABLE IF NOT EXISTS statuses (
    project_id VARCHAR,
    name VARCHAR,
    description TEXT,
    category VARCHAR,
    PRIMARY KEY (project_id, name)
);
`

const createPrioritiesTable = `
CREATE TABLE IF NOT EXISTS priorities (
    project_id VARCHAR,
    name VARCHAR,
    description TEXT,
    icon_url VARCHAR,
    PRIMARY KEY (project_id, name)
);
`

const createIssueTypesTable = `
CREATE TABLE IF NOT EXISTS issue_types (
    project_id VARCHAR,
    name VARCHAR,
    description TEXT,
    icon_url VARCHAR,
    subtask BOOLEAN,
    PRIMARY KEY (project_id, name)
);
`

const createLabelsTable = `
CREATE TABLE IF NOT EXISTS labels (
    project_id VARCHAR,
    name VARCHAR,
    PRIMARY KEY (project_id, name)
);
`

const createComponentsTable = `
CREATE TABLE IF NOT EXISTS components (
    project_id VARCHAR,
    name VARCHAR,
    description TEXT,
    lead VARCHAR,
    PRIMARY KEY (project_id, name)
);
`

const createFixVersionsTable = `
CREATE TABLE IF NOT EXISTS fix_versions (
    project_id VARCHAR,
    name VARCHAR,
    description TEXT,
    released BOOLEAN,
    release_date VARCHAR,
    PRIMARY KEY (project_id, name)
);
`

const createSyncHistoryTable = `
CREATE TABLE IF NOT EXISTS sync_history (
    id INTEGER PRIMARY KEY DEFAULT nextval('seq_sync_history_id'),
    project_id VARCHAR,
    sync_type VARCHAR,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    status VARCHAR,
    items_synced INTEGER,
    error_message TEXT
);
CREATE SEQUENCE IF NOT EXISTS seq_sync_history_id START 1;
CREATE INDEX IF NOT EXISTS idx_sync_history_project ON sync_history(project_id);
`

const createJiraFieldsTable = `
CREATE TABLE IF NOT EXISTS jira_fields (
    id VARCHAR PRIMARY KEY,
    key VARCHAR,
    name VARCHAR,
    custom BOOLEAN,
    searchable BOOLEAN,
    navigable BOOLEAN,
    schema_type VARCHAR,
    schema_items VARCHAR,
    schema_system VARCHAR,
    schema_custom VARCHAR,
    schema_custom_id BIGINT
);
`
