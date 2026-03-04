package service

import (
	"context"
	"fmt"

	"github.com/ysksm/go-jira/core/domain/models"
	"github.com/ysksm/go-jira/core/domain/repository"
)

// IssueService handles issue search and retrieval.
type IssueService struct {
	issueRepo         repository.IssueRepository
	changeHistoryRepo repository.ChangeHistoryRepository
}

// NewIssueService creates a new IssueService.
func NewIssueService(
	issueRepo repository.IssueRepository,
	changeHistoryRepo repository.ChangeHistoryRepository,
) *IssueService {
	return &IssueService{
		issueRepo:         issueRepo,
		changeHistoryRepo: changeHistoryRepo,
	}
}

// SearchResult holds paginated search results.
type SearchResult struct {
	Issues []models.Issue `json:"issues"`
	Total  int            `json:"total"`
}

// Search finds issues with filters and pagination.
func (s *IssueService) Search(ctx context.Context, projectID string, offset, limit int) (*SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}

	page, err := s.issueRepo.FindByProjectPaginated(ctx, projectID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}

	return &SearchResult{
		Issues: page.Issues,
		Total:  page.Total,
	}, nil
}

// Get returns a single issue by key.
func (s *IssueService) Get(ctx context.Context, key string) (*models.Issue, error) {
	issue, err := s.issueRepo.FindByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get issue %s: %w", key, err)
	}
	return issue, nil
}

// GetHistory returns change history for an issue.
func (s *IssueService) GetHistory(ctx context.Context, issueKey string) ([]models.ChangeHistoryItem, error) {
	items, err := s.changeHistoryRepo.FindByIssueKey(ctx, issueKey)
	if err != nil {
		return nil, fmt.Errorf("get history for %s: %w", issueKey, err)
	}
	return items, nil
}
