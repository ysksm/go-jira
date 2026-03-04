package api

import (
	"fmt"
	"net/http"

	"github.com/ysksm/go-jira/core/infrastructure/database"
)

// MetadataHandler handles metadata API endpoints.
type MetadataHandler struct {
	connMgr *database.Connection
}

// NewMetadataHandler creates a new MetadataHandler.
func NewMetadataHandler(connMgr *database.Connection) *MetadataHandler {
	return &MetadataHandler{connMgr: connMgr}
}

// Get handles POST /api/metadata.get.
func (h *MetadataHandler) Get(w http.ResponseWriter, r *http.Request) {
	var req MetadataGetRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if req.ProjectKey == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("projectKey is required"))
		return
	}

	db, err := h.connMgr.GetDB(req.ProjectKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to get database: %w", err))
		return
	}

	metadataRepo := database.NewMetadataRepository(db)
	metadata, err := metadataRepo.GetByProject(r.Context(), req.ProjectKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, MetadataGetResponse{Metadata: *metadata})
}
