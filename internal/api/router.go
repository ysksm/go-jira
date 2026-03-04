package api

import (
	"net/http"

	"github.com/ysksm/go-jira/core/service"
)

// NewRouter creates the HTTP router with all API routes registered.
func NewRouter(cfg ServerConfig) http.Handler {
	mux := http.NewServeMux()

	// --- Services ---
	configSvc := service.NewConfigService(cfg.ConfigStore)
	projectSvc := service.NewProjectService(cfg.ConfigStore, cfg.Logger)
	syncSvc := service.NewSyncService(cfg.ConfigStore, cfg.ConnMgr, cfg.Logger)

	// --- Handlers ---
	configH := NewConfigHandler(configSvc)
	projectH := NewProjectHandler(projectSvc)
	syncH := NewSyncHandler(syncSvc)
	issueH := NewIssueHandler(cfg.ConnMgr, cfg.ConfigStore)
	metadataH := NewMetadataHandler(cfg.ConnMgr)
	queryH := NewQueryHandler(cfg.ConnMgr)

	// --- Config ---
	mux.HandleFunc("POST /api/config.get", configH.Get)
	mux.HandleFunc("POST /api/config.update", configH.Update)
	mux.HandleFunc("POST /api/config.initialize", configH.Initialize)

	// --- Projects ---
	mux.HandleFunc("POST /api/projects.list", projectH.List)
	mux.HandleFunc("POST /api/projects.initialize", projectH.Initialize)
	mux.HandleFunc("POST /api/projects.enable", projectH.Enable)
	mux.HandleFunc("POST /api/projects.disable", projectH.Disable)

	// --- Sync ---
	mux.HandleFunc("POST /api/sync.execute", syncH.Execute)
	mux.HandleFunc("POST /api/sync.status", syncH.Status)
	mux.HandleFunc("GET /api/sync.progress", syncH.Progress)

	// --- Issues ---
	mux.HandleFunc("POST /api/issues.search", issueH.Search)
	mux.HandleFunc("POST /api/issues.get", issueH.Get)
	mux.HandleFunc("POST /api/issues.history", issueH.History)

	// --- Metadata ---
	mux.HandleFunc("POST /api/metadata.get", metadataH.Get)

	// --- SQL ---
	mux.HandleFunc("POST /api/sql.execute", queryH.Execute)
	mux.HandleFunc("POST /api/sql.get-schema", queryH.GetSchema)

	// --- Static files (embedded Svelte build) ---
	if hasEmbeddedWeb() {
		mux.Handle("/", staticHandler())
	}

	return mux
}
