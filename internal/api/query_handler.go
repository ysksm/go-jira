package api

import (
	"fmt"
	"net/http"

	"github.com/ysksm/go-jira/core/infrastructure/database"
	"github.com/ysksm/go-jira/core/service"
)

// QueryHandler handles SQL query API endpoints.
type QueryHandler struct {
	connMgr *database.Connection
}

// NewQueryHandler creates a new QueryHandler.
func NewQueryHandler(connMgr *database.Connection) *QueryHandler {
	return &QueryHandler{connMgr: connMgr}
}

func (h *QueryHandler) getQueryService() *service.QueryService {
	return service.NewQueryService(h.connMgr)
}

// Execute handles POST /api/sql.execute.
func (h *QueryHandler) Execute(w http.ResponseWriter, r *http.Request) {
	var req SqlExecuteRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if req.Query == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("query is required"))
		return
	}
	if req.ProjectKey == "" && !req.AllProjects {
		writeError(w, http.StatusBadRequest, fmt.Errorf("projectKey or allProjects is required"))
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	svc := h.getQueryService()
	result, err := svc.Execute(r.Context(), req.ProjectKey, req.Query, limit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, SqlExecuteResponse{
		Columns:         result.Columns,
		Rows:            result.Rows,
		RowCount:        result.RowCount,
		ExecutionTimeMs: result.ExecutionTimeMs,
	})
}

// GetSchema handles POST /api/sql.get-schema.
func (h *QueryHandler) GetSchema(w http.ResponseWriter, r *http.Request) {
	var req SqlGetSchemaRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if req.ProjectKey == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("projectKey is required"))
		return
	}

	svc := h.getQueryService()
	tables, err := svc.GetSchema(r.Context(), req.ProjectKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, SqlGetSchemaResponse{Tables: tables})
}
