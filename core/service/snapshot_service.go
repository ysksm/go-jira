package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/ysksm/go-jira/core/domain/models"
	"github.com/ysksm/go-jira/core/domain/repository"
)

// SnapshotService handles issue snapshot generation.
type SnapshotService struct {
	issueRepo         repository.IssueRepository
	changeHistoryRepo repository.ChangeHistoryRepository
	snapshotRepo      repository.SnapshotRepository
	logger            *slog.Logger
}

// NewSnapshotService creates a new SnapshotService.
func NewSnapshotService(
	issueRepo repository.IssueRepository,
	changeHistoryRepo repository.ChangeHistoryRepository,
	snapshotRepo repository.SnapshotRepository,
	logger *slog.Logger,
) *SnapshotService {
	if logger == nil {
		logger = slog.Default()
	}
	return &SnapshotService{
		issueRepo:         issueRepo,
		changeHistoryRepo: changeHistoryRepo,
		snapshotRepo:      snapshotRepo,
		logger:            logger,
	}
}

// GenerateForProject generates snapshots for all issues in a project.
// Processes one issue at a time via cursor to minimize memory usage.
func (s *SnapshotService) GenerateForProject(
	ctx context.Context,
	pc *models.ProjectConfig,
	onProgress models.ProgressCallback,
) error {
	totalIssues, err := s.issueRepo.CountByProject(ctx, pc.ID)
	if err != nil {
		return fmt.Errorf("count issues: %w", err)
	}

	if totalIssues == 0 {
		return nil
	}

	if err := s.snapshotRepo.BeginTransaction(ctx); err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Use cursor to iterate one issue at a time (memory efficient).
	cursor, err := s.issueRepo.FindByProjectCursor(ctx, pc.ID)
	if err != nil {
		s.snapshotRepo.RollbackTransaction(ctx)
		return fmt.Errorf("open cursor: %w", err)
	}
	defer cursor.Close()

	processed := 0
	snapshotsGenerated := 0

	// Resume from checkpoint if available.
	skipUntil := ""
	if pc.SnapshotCheckpoint != nil {
		skipUntil = pc.SnapshotCheckpoint.LastIssueID
		processed = pc.SnapshotCheckpoint.IssuesProcessed
		snapshotsGenerated = pc.SnapshotCheckpoint.SnapshotsGenerated
	}

	skipping := skipUntil != ""

	for cursor.Next() {
		select {
		case <-ctx.Done():
			s.snapshotRepo.RollbackTransaction(ctx)
			return ctx.Err()
		default:
		}

		issue := cursor.Issue()

		// Skip already-processed issues when resuming.
		if skipping {
			if issue.ID == skipUntil {
				skipping = false
			}
			continue
		}

		// Process one issue: fetch history → build snapshots → write to DB.
		snapshots, err := s.generateForIssue(ctx, issue)
		if err != nil {
			s.logger.Warn("failed to generate snapshots", "issue", issue.Key, "error", err)
			processed++
			continue
		}

		if len(snapshots) > 0 {
			// Delete old snapshots and insert new ones.
			if err := s.snapshotRepo.DeleteByIssueID(ctx, issue.ID); err != nil {
				s.logger.Warn("failed to delete old snapshots", "issue", issue.Key, "error", err)
			}
			if err := s.snapshotRepo.BatchInsert(ctx, snapshots); err != nil {
				s.logger.Warn("failed to insert snapshots", "issue", issue.Key, "error", err)
			}
			snapshotsGenerated += len(snapshots)
		}

		processed++

		// Report progress and save checkpoint every 100 issues.
		if processed%100 == 0 {
			if onProgress != nil {
				onProgress(models.SyncProgress{
					ProjectKey: pc.Key,
					Phase:      models.PhaseGenerateSnapshots,
					Current:    processed,
					Total:      totalIssues,
					Message:    fmt.Sprintf("Generated snapshots: %d/%d issues", processed, totalIssues),
				})
			}
		}
		// snapshots slice is now eligible for GC.
	}

	if err := cursor.Err(); err != nil {
		s.snapshotRepo.RollbackTransaction(ctx)
		return fmt.Errorf("cursor error: %w", err)
	}

	if err := s.snapshotRepo.CommitTransaction(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	if onProgress != nil {
		onProgress(models.SyncProgress{
			ProjectKey: pc.Key,
			Phase:      models.PhaseGenerateSnapshots,
			Current:    processed,
			Total:      totalIssues,
			Message:    fmt.Sprintf("Snapshot generation complete (%d snapshots)", snapshotsGenerated),
		})
	}

	return nil
}

// generateForIssue builds time-series snapshots for a single issue.
// Only the current issue's change history is held in memory.
func (s *SnapshotService) generateForIssue(ctx context.Context, issue *models.Issue) ([]models.IssueSnapshot, error) {
	history, err := s.changeHistoryRepo.FindByIssueKey(ctx, issue.Key)
	if err != nil {
		return nil, fmt.Errorf("find history for %s: %w", issue.Key, err)
	}

	// Group changes by timestamp.
	groups := groupChangesByTime(history)
	if len(groups) == 0 {
		// No change history → single snapshot (current state).
		return []models.IssueSnapshot{
			buildSnapshotFromIssue(issue, 1, issue.CreatedDate, nil),
		}, nil
	}

	// Build initial state by reversing all changes from current state.
	state := currentState(issue)
	for i := len(groups) - 1; i >= 0; i-- {
		for _, change := range groups[i].changes {
			reverseChange(&state, &change)
		}
	}

	var snapshots []models.IssueSnapshot

	// Version 1: initial state.
	validFrom := issue.CreatedDate
	var validTo *time.Time
	if len(groups) > 0 {
		validTo = &groups[0].timestamp
	}
	snapshots = append(snapshots, buildSnapshotFromState(issue, 1, validFrom, validTo, &state))

	// Versions 2+: apply each change group forward.
	for i, group := range groups {
		for _, change := range group.changes {
			applyChange(&state, &change)
		}
		version := i + 2
		vFrom := group.timestamp
		var vTo *time.Time
		if i+1 < len(groups) {
			vTo = &groups[i+1].timestamp
		}
		snapshots = append(snapshots, buildSnapshotFromState(issue, version, &vFrom, vTo, &state))
	}

	return snapshots, nil
}

// issueState tracks mutable field values for snapshot construction.
type issueState struct {
	Summary     string
	Description string
	Status      string
	Priority    string
	Assignee    string
	Reporter    string
	IssueType   string
	Resolution  string
	Sprint      string
	Labels      []string
	Components  []string
	FixVersions []string
}

// changeGroup groups changes that happened at the same timestamp.
type changeGroup struct {
	timestamp time.Time
	changes   []models.ChangeHistoryItem
}

func currentState(issue *models.Issue) issueState {
	return issueState{
		Summary:     issue.Summary,
		Description: issue.Description,
		Status:      issue.Status,
		Priority:    issue.Priority,
		Assignee:    issue.Assignee,
		Reporter:    issue.Reporter,
		IssueType:   issue.IssueType,
		Resolution:  issue.Resolution,
		Sprint:      issue.Sprint,
		Labels:      issue.Labels,
		Components:  issue.Components,
		FixVersions: issue.FixVersions,
	}
}

func groupChangesByTime(history []models.ChangeHistoryItem) []changeGroup {
	if len(history) == 0 {
		return nil
	}

	sort.Slice(history, func(i, j int) bool {
		return history[i].ChangedAt.Before(history[j].ChangedAt)
	})

	var groups []changeGroup
	var current *changeGroup

	for _, item := range history {
		if current == nil || !current.timestamp.Equal(item.ChangedAt) {
			if current != nil {
				groups = append(groups, *current)
			}
			current = &changeGroup{timestamp: item.ChangedAt}
		}
		current.changes = append(current.changes, item)
	}
	if current != nil {
		groups = append(groups, *current)
	}
	return groups
}

func reverseChange(state *issueState, change *models.ChangeHistoryItem) {
	switch change.Field {
	case "summary":
		state.Summary = change.FromString
	case "description":
		state.Description = change.FromString
	case "status":
		state.Status = change.FromString
	case "priority":
		state.Priority = change.FromString
	case "assignee":
		state.Assignee = change.FromString
	case "reporter":
		state.Reporter = change.FromString
	case "issuetype":
		state.IssueType = change.FromString
	case "resolution":
		state.Resolution = change.FromString
	case "Sprint":
		state.Sprint = change.FromString
	}
}

func applyChange(state *issueState, change *models.ChangeHistoryItem) {
	switch change.Field {
	case "summary":
		state.Summary = change.ToString
	case "description":
		state.Description = change.ToString
	case "status":
		state.Status = change.ToString
	case "priority":
		state.Priority = change.ToString
	case "assignee":
		state.Assignee = change.ToString
	case "reporter":
		state.Reporter = change.ToString
	case "issuetype":
		state.IssueType = change.ToString
	case "resolution":
		state.Resolution = change.ToString
	case "Sprint":
		state.Sprint = change.ToString
	}
}

func buildSnapshotFromIssue(issue *models.Issue, version int, validFrom *time.Time, validTo *time.Time) models.IssueSnapshot {
	now := time.Now().UTC()
	vf := now
	if validFrom != nil {
		vf = *validFrom
	}
	return models.IssueSnapshot{
		IssueID:     issue.ID,
		IssueKey:    issue.Key,
		ProjectID:   issue.ProjectID,
		Version:     version,
		ValidFrom:   vf,
		ValidTo:     validTo,
		Summary:     issue.Summary,
		Description: issue.Description,
		Status:      issue.Status,
		Priority:    issue.Priority,
		Assignee:    issue.Assignee,
		Reporter:    issue.Reporter,
		IssueType:   issue.IssueType,
		Resolution:  issue.Resolution,
		Labels:      issue.Labels,
		Components:  issue.Components,
		FixVersions: issue.FixVersions,
		Sprint:      issue.Sprint,
		ParentKey:   issue.ParentKey,
		UpdatedDate: issue.UpdatedDate,
		DueDate:     issue.DueDate,
		CreatedAt:   now,
	}
}

func buildSnapshotFromState(issue *models.Issue, version int, validFrom *time.Time, validTo *time.Time, state *issueState) models.IssueSnapshot {
	now := time.Now().UTC()
	vf := now
	if validFrom != nil {
		vf = *validFrom
	}
	return models.IssueSnapshot{
		IssueID:     issue.ID,
		IssueKey:    issue.Key,
		ProjectID:   issue.ProjectID,
		Version:     version,
		ValidFrom:   vf,
		ValidTo:     validTo,
		Summary:     state.Summary,
		Description: state.Description,
		Status:      state.Status,
		Priority:    state.Priority,
		Assignee:    state.Assignee,
		Reporter:    state.Reporter,
		IssueType:   state.IssueType,
		Resolution:  state.Resolution,
		Labels:      state.Labels,
		Components:  state.Components,
		FixVersions: state.FixVersions,
		Sprint:      state.Sprint,
		ParentKey:   issue.ParentKey,
		UpdatedDate: issue.UpdatedDate,
		DueDate:     issue.DueDate,
		CreatedAt:   now,
	}
}
