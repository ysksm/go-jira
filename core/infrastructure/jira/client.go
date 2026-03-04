package jira

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/ysksm/go-jira/core/domain"
	"github.com/ysksm/go-jira/core/domain/models"
)

const (
	maxRetries     = 3
	requestTimeout = 30 * time.Second
	batchSize      = 100
)

// Client is a JIRA REST API client with retry and authentication.
type Client struct {
	baseURL    string
	authHeader string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewClient creates a new JIRA API client.
func NewClient(endpoint *models.JiraEndpoint, logger *slog.Logger) *Client {
	auth := base64.StdEncoding.EncodeToString(
		[]byte(endpoint.Username + ":" + endpoint.APIKey),
	)
	if logger == nil {
		logger = slog.Default()
	}
	return &Client{
		baseURL:    endpoint.Endpoint,
		authHeader: "Basic " + auth,
		httpClient: &http.Client{Timeout: requestTimeout},
		logger:     logger,
	}
}

// FetchIssuesResult holds the result of a FetchIssues call.
type FetchIssuesResult struct {
	Issues        []models.Issue
	ChangeHistory []models.ChangeHistoryItem
	NextPageToken string
	Total         int
}

// FetchIssues fetches issues from JIRA using JQL with token-based pagination.
func (c *Client) FetchIssues(ctx context.Context, jql string, nextPageToken string) (*FetchIssuesResult, error) {
	params := url.Values{}
	params.Set("jql", jql)
	params.Set("maxResults", fmt.Sprintf("%d", batchSize))
	params.Set("fields", "*navigable,created,updated")
	params.Set("expand", "changelog")
	if nextPageToken != "" {
		params.Set("nextPageToken", nextPageToken)
	}

	reqURL := fmt.Sprintf("%s/rest/api/3/search/jql?%s", c.baseURL, params.Encode())
	body, err := c.doGetWithRetry(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch issues: %w", err)
	}
	defer body.Close()

	var resp JiraSearchResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	result := &FetchIssuesResult{
		NextPageToken: resp.NextPageToken,
		Total:         resp.Total,
	}

	now := time.Now().UTC()
	for i := range resp.Issues {
		issue, err := ParseIssue(&resp.Issues[i], now)
		if err != nil {
			c.logger.Warn("failed to parse issue", "key", resp.Issues[i].Key, "error", err)
			continue
		}
		result.Issues = append(result.Issues, *issue)

		if resp.Issues[i].Changelog != nil {
			items := ExtractChangeHistory(&resp.Issues[i])
			result.ChangeHistory = append(result.ChangeHistory, items...)
		}
	}

	return result, nil
}

// FetchProjects fetches all projects from JIRA.
func (c *Client) FetchProjects(ctx context.Context) ([]models.Project, error) {
	reqURL := fmt.Sprintf("%s/rest/api/3/project", c.baseURL)
	body, err := c.doGetWithRetry(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch projects: %w", err)
	}
	defer body.Close()

	var jiraProjects []JiraProject
	if err := json.NewDecoder(body).Decode(&jiraProjects); err != nil {
		return nil, fmt.Errorf("decode projects: %w", err)
	}

	projects := make([]models.Project, len(jiraProjects))
	for i, jp := range jiraProjects {
		projects[i] = models.Project{
			ID:          jp.ID,
			Key:         jp.Key,
			Name:        jp.Name,
			Description: jp.Description,
		}
	}
	return projects, nil
}

// FetchStatuses fetches all statuses for a project.
func (c *Client) FetchStatuses(ctx context.Context, projectKey string) ([]models.Status, error) {
	reqURL := fmt.Sprintf("%s/rest/api/3/project/%s/statuses", c.baseURL, url.PathEscape(projectKey))
	body, err := c.doGetWithRetry(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch statuses: %w", err)
	}
	defer body.Close()

	var issueTypes []JiraIssueTypeWithStatuses
	if err := json.NewDecoder(body).Decode(&issueTypes); err != nil {
		return nil, fmt.Errorf("decode statuses: %w", err)
	}

	seen := make(map[string]bool)
	var statuses []models.Status
	for _, it := range issueTypes {
		for _, s := range it.Statuses {
			if seen[s.Name] {
				continue
			}
			seen[s.Name] = true
			category := ""
			if s.StatusCategory != nil {
				category = s.StatusCategory.Name
			}
			statuses = append(statuses, models.Status{
				Name:        s.Name,
				Description: s.Description,
				Category:    category,
			})
		}
	}
	return statuses, nil
}

// FetchPriorities fetches all priorities.
func (c *Client) FetchPriorities(ctx context.Context) ([]models.Priority, error) {
	reqURL := fmt.Sprintf("%s/rest/api/3/priority", c.baseURL)
	body, err := c.doGetWithRetry(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch priorities: %w", err)
	}
	defer body.Close()

	var jiraPriorities []JiraNamedField
	if err := json.NewDecoder(body).Decode(&jiraPriorities); err != nil {
		return nil, fmt.Errorf("decode priorities: %w", err)
	}

	priorities := make([]models.Priority, len(jiraPriorities))
	for i, jp := range jiraPriorities {
		priorities[i] = models.Priority{
			Name:        jp.Name,
			Description: jp.Description,
			IconURL:     jp.IconURL,
		}
	}
	return priorities, nil
}

// FetchIssueTypes fetches issue types for a project.
func (c *Client) FetchIssueTypes(ctx context.Context, projectID string) ([]models.IssueType, error) {
	reqURL := fmt.Sprintf("%s/rest/api/3/issuetype/project?projectId=%s", c.baseURL, url.QueryEscape(projectID))
	body, err := c.doGetWithRetry(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch issue types: %w", err)
	}
	defer body.Close()

	var jiraTypes []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IconURL     string `json:"iconUrl"`
		Subtask     bool   `json:"subtask"`
	}
	if err := json.NewDecoder(body).Decode(&jiraTypes); err != nil {
		return nil, fmt.Errorf("decode issue types: %w", err)
	}

	issueTypes := make([]models.IssueType, len(jiraTypes))
	for i, jt := range jiraTypes {
		issueTypes[i] = models.IssueType{
			Name:        jt.Name,
			Description: jt.Description,
			IconURL:     jt.IconURL,
			Subtask:     jt.Subtask,
		}
	}
	return issueTypes, nil
}

// FetchComponents fetches components for a project.
func (c *Client) FetchComponents(ctx context.Context, projectKey string) ([]models.Component, error) {
	reqURL := fmt.Sprintf("%s/rest/api/3/project/%s/components", c.baseURL, url.PathEscape(projectKey))
	body, err := c.doGetWithRetry(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch components: %w", err)
	}
	defer body.Close()

	var jiraComponents []JiraComponent
	if err := json.NewDecoder(body).Decode(&jiraComponents); err != nil {
		return nil, fmt.Errorf("decode components: %w", err)
	}

	components := make([]models.Component, len(jiraComponents))
	for i, jc := range jiraComponents {
		lead := ""
		if jc.Lead != nil {
			lead = jc.Lead.DisplayName
		}
		components[i] = models.Component{
			Name:        jc.Name,
			Description: jc.Description,
			Lead:        lead,
		}
	}
	return components, nil
}

// FetchVersions fetches versions for a project.
func (c *Client) FetchVersions(ctx context.Context, projectKey string) ([]models.FixVersion, error) {
	reqURL := fmt.Sprintf("%s/rest/api/3/project/%s/versions", c.baseURL, url.PathEscape(projectKey))
	body, err := c.doGetWithRetry(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch versions: %w", err)
	}
	defer body.Close()

	var jiraVersions []JiraVersion
	if err := json.NewDecoder(body).Decode(&jiraVersions); err != nil {
		return nil, fmt.Errorf("decode versions: %w", err)
	}

	versions := make([]models.FixVersion, len(jiraVersions))
	for i, jv := range jiraVersions {
		versions[i] = models.FixVersion{
			Name:        jv.Name,
			Description: jv.Description,
			Released:    jv.Released,
			ReleaseDate: jv.ReleaseDate,
		}
	}
	return versions, nil
}

// FetchFields fetches all field definitions from JIRA.
func (c *Client) FetchFields(ctx context.Context) ([]JiraField, error) {
	reqURL := fmt.Sprintf("%s/rest/api/3/field", c.baseURL)
	body, err := c.doGetWithRetry(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("fetch fields: %w", err)
	}
	defer body.Close()

	var fields []JiraField
	if err := json.NewDecoder(body).Decode(&fields); err != nil {
		return nil, fmt.Errorf("decode fields: %w", err)
	}
	return fields, nil
}

// GetIssueCount returns the total issue count for a JQL query.
func (c *Client) GetIssueCount(ctx context.Context, jql string) (int, error) {
	params := url.Values{}
	params.Set("jql", jql)
	params.Set("maxResults", "0")

	reqURL := fmt.Sprintf("%s/rest/api/3/search/jql?%s", c.baseURL, params.Encode())
	body, err := c.doGetWithRetry(ctx, reqURL)
	if err != nil {
		return 0, fmt.Errorf("get issue count: %w", err)
	}
	defer body.Close()

	var resp JiraSearchResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return 0, fmt.Errorf("decode count response: %w", err)
	}
	return resp.Total, nil
}

// doGetWithRetry performs an HTTP GET with exponential backoff retry.
func (c *Client) doGetWithRetry(ctx context.Context, reqURL string) (io.ReadCloser, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			c.logger.Debug("retrying request", "attempt", attempt, "delay", delay, "url", reqURL)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", c.authHeader)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp.Body, nil
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = &domain.JiraAPIError{
				StatusCode: resp.StatusCode,
				Message:    string(respBody),
				URL:        reqURL,
			}
			continue
		}

		return nil, &domain.JiraAPIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
			URL:        reqURL,
		}
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
