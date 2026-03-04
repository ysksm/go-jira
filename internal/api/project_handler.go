package api

import (
	"fmt"
	"net/http"

	"github.com/ysksm/go-jira/core/service"
)

// ProjectHandler handles project API endpoints.
type ProjectHandler struct {
	svc *service.ProjectService
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// List handles POST /api/projects.list.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.svc.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, ProjectListResponse{Projects: projects})
}

// Initialize handles POST /api/projects.initialize.
func (h *ProjectHandler) Initialize(w http.ResponseWriter, r *http.Request) {
	projects, newCount, err := h.svc.FetchFromJira(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, ProjectInitResponse{
		Projects: projects,
		NewCount: newCount,
	})
}

// Enable handles POST /api/projects.enable.
func (h *ProjectHandler) Enable(w http.ResponseWriter, r *http.Request) {
	var req ProjectEnableRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if req.Key == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("key is required"))
		return
	}

	project, err := h.svc.Enable(req.Key)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, ProjectEnableResponse{Project: *project})
}

// Disable handles POST /api/projects.disable.
func (h *ProjectHandler) Disable(w http.ResponseWriter, r *http.Request) {
	var req ProjectDisableRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if req.Key == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("key is required"))
		return
	}

	project, err := h.svc.Disable(req.Key)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, ProjectDisableResponse{Project: *project})
}
