package models

// SyncResult represents the result of syncing a single project.
type SyncResult struct {
	ProjectKey      string  `json:"projectKey"`
	IssueCount      int     `json:"issueCount"`
	MetadataUpdated bool    `json:"metadataUpdated"`
	Duration        float64 `json:"duration"`
	Success         bool    `json:"success"`
	Error           string  `json:"error,omitempty"`
}

// SyncProgress reports real-time sync progress.
type SyncProgress struct {
	ProjectKey string `json:"projectKey"`
	Phase      string `json:"phase"`
	Current    int    `json:"current"`
	Total      int    `json:"total"`
	Message    string `json:"message"`
}

// Sync phase constants.
const (
	PhaseFetchIssues      = "fetch_issues"
	PhaseSyncMetadata     = "sync_metadata"
	PhaseGenerateSnapshots = "generate_snapshots"
	PhaseVerifyIntegrity  = "verify_integrity"
)

// ProgressCallback is a function called to report sync progress.
type ProgressCallback func(progress SyncProgress)
