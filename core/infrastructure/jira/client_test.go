package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ysksm/go-jira/core/domain/models"
)

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient(&models.JiraEndpoint{
		Name:     "test",
		Endpoint: server.URL,
		Username: "user",
		APIKey:   "key",
	}, nil)
	return server, client
}

func TestFetchProjects(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/project" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}

		projects := []JiraProject{
			{ID: "1", Key: "PROJ", Name: "Project One"},
			{ID: "2", Key: "TEST", Name: "Test Project"},
		}
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	projects, err := client.FetchProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].Key != "PROJ" {
		t.Errorf("projects[0].Key: got %s, want PROJ", projects[0].Key)
	}
	if projects[1].Name != "Test Project" {
		t.Errorf("projects[1].Name: got %s, want Test Project", projects[1].Name)
	}
}

func TestFetchPriorities(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		priorities := []JiraNamedField{
			{Name: "Highest", Description: "Critical"},
			{Name: "High"},
			{Name: "Medium"},
		}
		json.NewEncoder(w).Encode(priorities)
	})
	defer server.Close()

	priorities, err := client.FetchPriorities(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(priorities) != 3 {
		t.Fatalf("expected 3 priorities, got %d", len(priorities))
	}
	if priorities[0].Name != "Highest" {
		t.Errorf("got %s, want Highest", priorities[0].Name)
	}
}

func TestFetchIssues(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		fields := JiraIssueFields{
			Summary:   "Bug fix",
			Status:    &JiraNamedField{Name: "Open"},
			Priority:  &JiraNamedField{Name: "High"},
			IssueType: &JiraNamedField{Name: "Bug"},
			Created:   "2024-01-15T10:00:00.000Z",
			Updated:   "2024-01-20T15:00:00.000Z",
			Project:   &JiraProjectField{ID: "10001", Key: "PROJ"},
		}
		fieldsJSON, _ := json.Marshal(fields)

		resp := JiraSearchResponse{
			Issues: []JiraIssue{
				{ID: "100", Key: "PROJ-1", Fields: fieldsJSON},
			},
			Total:         1,
			NextPageToken: "",
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	result, err := client.FetchIssues(context.Background(), "project = PROJ", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(result.Issues))
	}
	if result.Issues[0].Key != "PROJ-1" {
		t.Errorf("Key: got %s, want PROJ-1", result.Issues[0].Key)
	}
	if result.Issues[0].Summary != "Bug fix" {
		t.Errorf("Summary: got %s, want Bug fix", result.Issues[0].Summary)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if result.NextPageToken != "" {
		t.Errorf("NextPageToken should be empty, got %s", result.NextPageToken)
	}
}

func TestRetryOn500(t *testing.T) {
	attempts := 0
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
			return
		}
		json.NewEncoder(w).Encode([]JiraProject{{ID: "1", Key: "OK"}})
	})
	defer server.Close()

	projects, err := client.FetchProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestFailOn401(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
	})
	defer server.Close()

	_, err := client.FetchProjects(context.Background())
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestGetIssueCount(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("maxResults") != "0" {
			t.Errorf("expected maxResults=0, got %s", r.URL.Query().Get("maxResults"))
		}
		json.NewEncoder(w).Encode(JiraSearchResponse{Total: 42})
	})
	defer server.Close()

	count, err := client.GetIssueCount(context.Background(), "project = PROJ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 42 {
		t.Errorf("got %d, want 42", count)
	}
}
