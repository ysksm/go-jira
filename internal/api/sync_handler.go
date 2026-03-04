package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ysksm/go-jira/core/domain/models"
	"github.com/ysksm/go-jira/core/service"
)

// SyncHandler handles sync API endpoints.
type SyncHandler struct {
	svc        *service.SyncService
	mu         sync.Mutex
	inProgress bool
	progress   *models.SyncProgress
}

// NewSyncHandler creates a new SyncHandler.
func NewSyncHandler(svc *service.SyncService) *SyncHandler {
	return &SyncHandler{svc: svc}
}

// Execute handles POST /api/sync.execute.
func (h *SyncHandler) Execute(w http.ResponseWriter, r *http.Request) {
	var req SyncExecuteRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	h.mu.Lock()
	if h.inProgress {
		h.mu.Unlock()
		writeError(w, http.StatusConflict, fmt.Errorf("sync already in progress"))
		return
	}
	h.inProgress = true
	h.mu.Unlock()

	// Set progress callback.
	h.svc.SetProgressCallback(func(p models.SyncProgress) {
		h.mu.Lock()
		h.progress = &p
		h.mu.Unlock()
	})

	opts := service.SyncOptions{
		ProjectKey: req.ProjectKey,
		Force:      req.Force,
	}

	results, err := h.svc.Execute(r.Context(), opts)

	h.mu.Lock()
	h.inProgress = false
	h.progress = nil
	h.mu.Unlock()

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, SyncExecuteResponse{Results: results})
}

// Status handles POST /api/sync.status.
func (h *SyncHandler) Status(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	resp := SyncStatusResponse{
		InProgress: h.inProgress,
		Progress:   h.progress,
	}
	h.mu.Unlock()

	writeJSON(w, http.StatusOK, resp)
}

// Progress handles GET /api/sync.progress (SSE).
func (h *SyncHandler) Progress(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("streaming not supported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ctx := r.Context()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.mu.Lock()
			inProgress := h.inProgress
			progress := h.progress
			h.mu.Unlock()

			data := struct {
				InProgress bool                 `json:"inProgress"`
				Progress   *models.SyncProgress `json:"progress,omitempty"`
			}{
				InProgress: inProgress,
				Progress:   progress,
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()

			if !inProgress {
				return
			}
		}
	}
}
