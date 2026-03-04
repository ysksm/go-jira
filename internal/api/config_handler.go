package api

import (
	"fmt"
	"net/http"

	"github.com/ysksm/go-jira/core/domain/models"
	"github.com/ysksm/go-jira/core/service"
)

// ConfigHandler handles config API endpoints.
type ConfigHandler struct {
	svc *service.ConfigService
}

// NewConfigHandler creates a new ConfigHandler.
func NewConfigHandler(svc *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

// Get handles POST /api/config.get.
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := h.svc.Get()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, ConfigGetResponse{Settings: *settings})
}

// Update handles POST /api/config.update.
func (h *ConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req ConfigUpdateRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Handle endpoint operations first.
	if req.AddEndpoint != nil {
		if _, err := h.svc.AddEndpoint(*req.AddEndpoint); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
	}
	if req.RemoveEndpoint != "" {
		if _, err := h.svc.RemoveEndpoint(req.RemoveEndpoint); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
	}
	if req.SetActiveEndpoint != "" {
		if _, err := h.svc.SetActiveEndpoint(req.SetActiveEndpoint); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
	}

	// Update config fields.
	settings, err := h.svc.Update(func(s *models.Settings) {
		if req.Jira != nil {
			s.Jira = req.Jira
		}
		if req.Database != nil {
			s.Database = *req.Database
		}
		if req.Embeddings != nil {
			s.Embeddings = req.Embeddings
		}
		if req.Log != nil {
			s.Log = req.Log
		}
		if req.Sync != nil {
			s.Sync = req.Sync
		}
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, ConfigUpdateResponse{
		Success:  true,
		Settings: *settings,
	})
}

// Initialize handles POST /api/config.initialize.
func (h *ConfigHandler) Initialize(w http.ResponseWriter, r *http.Request) {
	var req ConfigInitRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if req.Endpoint == "" || req.Username == "" || req.APIKey == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("endpoint, username, and apiKey are required"))
		return
	}

	settings, err := h.svc.Initialize(req.Endpoint, req.Username, req.APIKey, req.DatabasePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, ConfigInitResponse{
		Success:  true,
		Settings: *settings,
	})
}
