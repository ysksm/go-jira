package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// decodeRequest reads and decodes a JSON request body into dst.
func decodeRequest(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return nil
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// errorResponse is the standard error response format.
type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, err error) {
	slog.Error("API error", "status", status, "error", err)
	writeJSON(w, status, errorResponse{
		Error:   http.StatusText(status),
		Message: err.Error(),
	})
}
