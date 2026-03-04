package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Helper ---

func doRequest(handler http.HandlerFunc, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(http.MethodPost, "/test", &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, dst interface{}) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(dst); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

// --- Tests ---

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var result map[string]string
	decodeJSON(t, w, &result)
	if result["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", result["key"])
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusBadRequest, &testError{"bad input"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var result errorResponse
	decodeJSON(t, w, &result)
	if result.Error != "Bad Request" {
		t.Errorf("expected error 'Bad Request', got '%s'", result.Error)
	}
	if result.Message != "bad input" {
		t.Errorf("expected message 'bad input', got '%s'", result.Message)
	}
}

func TestDecodeRequest(t *testing.T) {
	body := map[string]string{"key": "PROJ-123"}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)

	req := httptest.NewRequest(http.MethodPost, "/test", &buf)
	var dst IssueGetRequest
	if err := decodeRequest(req, &dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.Key != "PROJ-123" {
		t.Errorf("expected key PROJ-123, got %s", dst.Key)
	}
}

func TestDecodeRequestNilBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	var dst IssueGetRequest
	if err := decodeRequest(req, &dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractProjectKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PROJ-123", "PROJ"},
		{"ABC-1", "ABC"},
		{"NOHYPHEN", "NOHYPHEN"},
	}
	for _, tt := range tests {
		result := extractProjectKey(tt.input)
		if result != tt.expected {
			t.Errorf("extractProjectKey(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
