package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ysksm/go-jira/core/infrastructure/config"
	"github.com/ysksm/go-jira/core/infrastructure/database"
	"github.com/ysksm/go-jira/core/service"
)

// IssueHandler handles issue API endpoints.
type IssueHandler struct {
	connMgr     *database.Connection
	configStore *config.FileConfigStore
}

// NewIssueHandler creates a new IssueHandler.
func NewIssueHandler(connMgr *database.Connection, configStore *config.FileConfigStore) *IssueHandler {
	return &IssueHandler{
		connMgr:     connMgr,
		configStore: configStore,
	}
}

// getIssueService creates an IssueService for the given project.
func (h *IssueHandler) getIssueService(projectKey string) (*service.IssueService, error) {
	db, err := h.connMgr.GetDB(projectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get database for project %s: %w", projectKey, err)
	}
	issueRepo := database.NewIssueRepository(db)
	changeHistoryRepo := database.NewChangeHistoryRepository(db)
	return service.NewIssueService(issueRepo, changeHistoryRepo), nil
}

// Search handles POST /api/issues.search.
func (h *IssueHandler) Search(w http.ResponseWriter, r *http.Request) {
	var req IssueSearchRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if req.Project == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("project is required"))
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	svc, err := h.getIssueService(req.Project)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	result, err := svc.Search(r.Context(), req.Project, req.Offset, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, IssueSearchResponse{
		Issues: result.Issues,
		Total:  result.Total,
	})
}

// Get handles POST /api/issues.get.
func (h *IssueHandler) Get(w http.ResponseWriter, r *http.Request) {
	var req IssueGetRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if req.Key == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("key is required"))
		return
	}

	projectKey := extractProjectKey(req.Key)
	svc, err := h.getIssueService(projectKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	issue, err := svc.Get(r.Context(), req.Key)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, IssueGetResponse{Issue: *issue})
}

// History handles POST /api/issues.history.
func (h *IssueHandler) History(w http.ResponseWriter, r *http.Request) {
	var req IssueHistoryRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if req.Key == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("key is required"))
		return
	}

	projectKey := extractProjectKey(req.Key)
	svc, err := h.getIssueService(projectKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	history, err := svc.GetHistory(r.Context(), req.Key)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, IssueHistoryResponse{History: history})
}

// extractProjectKey extracts the project key from an issue key (e.g., "PROJ-123" -> "PROJ").
func extractProjectKey(issueKey string) string {
	parts := strings.SplitN(issueKey, "-", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return issueKey
}
