package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ysksm/go-jira/core/domain/models"
	"github.com/ysksm/go-jira/core/domain/repository"
	"github.com/ysksm/go-jira/core/infrastructure/config"
	"github.com/ysksm/go-jira/core/infrastructure/database"
	"github.com/ysksm/go-jira/core/infrastructure/jira"
)

// SyncOptions controls sync behavior.
type SyncOptions struct {
	ProjectKey string // Empty = all enabled projects
	Force      bool   // Force full sync (ignore incremental)
}

// SyncService orchestrates the data synchronization process.
type SyncService struct {
	configStore *config.FileConfigStore
	connMgr     *database.Connection
	logger      *slog.Logger
	onProgress  models.ProgressCallback
}

// NewSyncService creates a new SyncService.
func NewSyncService(
	configStore *config.FileConfigStore,
	connMgr *database.Connection,
	logger *slog.Logger,
) *SyncService {
	if logger == nil {
		logger = slog.Default()
	}
	return &SyncService{
		configStore: configStore,
		connMgr:     connMgr,
		logger:      logger,
	}
}

// SetProgressCallback sets a callback for progress reporting.
func (s *SyncService) SetProgressCallback(cb models.ProgressCallback) {
	s.onProgress = cb
}

// Execute runs the sync process for enabled projects.
func (s *SyncService) Execute(ctx context.Context, opts SyncOptions) ([]models.SyncResult, error) {
	settings, err := s.configStore.Load()
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}

	ep := settings.GetActiveEndpoint()
	if ep == nil {
		return nil, fmt.Errorf("no JIRA endpoint configured")
	}

	syncSettings := settings.GetSyncSettings()
	client := jira.NewClient(ep, s.logger)

	var results []models.SyncResult

	for i := range settings.Projects {
		pc := &settings.Projects[i]

		if !pc.SyncEnabled {
			continue
		}
		if opts.ProjectKey != "" && pc.Key != opts.ProjectKey {
			continue
		}

		start := time.Now()
		result := s.syncProject(ctx, client, pc, syncSettings, opts.Force)
		result.Duration = time.Since(start).Seconds()

		// Update last synced time.
		if result.Success {
			now := time.Now().UTC()
			pc.LastSynced = &now
			pc.SyncCheckpoint = nil
			pc.SnapshotCheckpoint = nil
		}

		results = append(results, result)
	}

	// Save updated settings (last synced times, cleared checkpoints).
	if err := s.configStore.Save(settings); err != nil {
		s.logger.Error("failed to save settings after sync", "error", err)
	}

	return results, nil
}

// syncProject runs all sync phases for a single project.
func (s *SyncService) syncProject(
	ctx context.Context,
	client *jira.Client,
	pc *models.ProjectConfig,
	syncSettings models.SyncSettings,
	force bool,
) models.SyncResult {
	result := models.SyncResult{ProjectKey: pc.Key, Success: true}
	s.logger.Info("starting sync", "project", pc.Key)

	db, err := s.connMgr.GetDB(pc.Key)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("open database: %v", err)
		return result
	}

	issueRepo := database.NewIssueRepository(db)
	changeHistoryRepo := database.NewChangeHistoryRepository(db)
	snapshotRepo := database.NewSnapshotRepository(db)
	metadataRepo := database.NewMetadataRepository(db)
	syncHistoryRepo := database.NewSyncHistoryRepository(db)

	// Record sync start.
	syncType := "incremental"
	if force || pc.LastSynced == nil {
		syncType = "full"
	}
	syncHistoryID, _ := syncHistoryRepo.Insert(ctx, pc.ID, syncType, time.Now().UTC())

	// Phase 1: Fetch Issues
	issueCount, err := s.fetchIssues(ctx, client, pc, issueRepo, changeHistoryRepo, syncSettings, force)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("fetch issues: %v", err)
		syncHistoryRepo.UpdateFailed(ctx, syncHistoryID, result.Error, time.Now().UTC())
		return result
	}
	result.IssueCount = issueCount

	// Phase 2: Sync Metadata
	if err := s.syncMetadata(ctx, client, pc, metadataRepo); err != nil {
		s.logger.Warn("metadata sync failed (non-fatal)", "project", pc.Key, "error", err)
	} else {
		result.MetadataUpdated = true
	}

	// Phase 3: Generate Snapshots
	snapshotSvc := NewSnapshotService(issueRepo, changeHistoryRepo, snapshotRepo, s.logger)
	if err := snapshotSvc.GenerateForProject(ctx, pc, s.onProgress); err != nil {
		s.logger.Warn("snapshot generation failed (non-fatal)", "project", pc.Key, "error", err)
	}

	// Phase 4: Verify Integrity
	s.verifyIntegrity(ctx, client, pc, issueRepo)

	syncHistoryRepo.UpdateCompleted(ctx, syncHistoryID, issueCount, time.Now().UTC())
	s.logger.Info("sync completed", "project", pc.Key, "issues", issueCount)

	return result
}

// fetchIssues performs Phase 1: paginated issue fetching with checkpoints.
func (s *SyncService) fetchIssues(
	ctx context.Context,
	client *jira.Client,
	pc *models.ProjectConfig,
	issueRepo repository.IssueRepository,
	changeHistoryRepo repository.ChangeHistoryRepository,
	syncSettings models.SyncSettings,
	force bool,
) (int, error) {
	// Build JQL.
	jql := fmt.Sprintf("project = %s", pc.Key)
	isFullSync := force || pc.LastSynced == nil

	if !isFullSync && syncSettings.IncrementalSyncEnabled && pc.LastSynced != nil {
		margin := time.Duration(syncSettings.IncrementalSyncMarginMinutes) * time.Minute
		since := pc.LastSynced.Add(-margin)
		jql += fmt.Sprintf(` AND updated >= "%s"`, since.Format("2006-01-02 15:04"))
	}
	jql += " ORDER BY updated ASC, key ASC"

	syncedAt := time.Now().UTC()
	nextPageToken := ""
	totalProcessed := 0

	// Resume from checkpoint if available.
	if pc.SyncCheckpoint != nil && !force {
		totalProcessed = pc.SyncCheckpoint.ItemsProcessed
		s.logger.Info("resuming from checkpoint", "project", pc.Key, "processed", totalProcessed)
	}

	for {
		select {
		case <-ctx.Done():
			return totalProcessed, ctx.Err()
		default:
		}

		fetchResult, err := client.FetchIssues(ctx, jql, nextPageToken)
		if err != nil {
			return totalProcessed, err
		}

		if len(fetchResult.Issues) > 0 {
			// Batch insert issues → immediately written to DB.
			if err := issueRepo.BatchInsert(ctx, fetchResult.Issues); err != nil {
				return totalProcessed, fmt.Errorf("batch insert issues: %w", err)
			}
			// Batch insert change history → immediately written to DB.
			if len(fetchResult.ChangeHistory) > 0 {
				if err := changeHistoryRepo.BatchInsert(ctx, fetchResult.ChangeHistory); err != nil {
					return totalProcessed, fmt.Errorf("batch insert change history: %w", err)
				}
			}
			// Memory for this batch's slices is now eligible for GC.
		}

		totalProcessed += len(fetchResult.Issues)

		// Save checkpoint.
		if len(fetchResult.Issues) > 0 {
			lastIssue := fetchResult.Issues[len(fetchResult.Issues)-1]
			checkpoint := &models.SyncCheckpoint{
				LastIssueKey:   lastIssue.Key,
				ItemsProcessed: totalProcessed,
				TotalItems:     fetchResult.Total,
			}
			if lastIssue.UpdatedDate != nil {
				checkpoint.LastIssueUpdatedAt = *lastIssue.UpdatedDate
			}
			s.configStore.UpdateCheckpoint(pc.Key, checkpoint)
		}

		// Report progress.
		if s.onProgress != nil {
			s.onProgress(models.SyncProgress{
				ProjectKey: pc.Key,
				Phase:      models.PhaseFetchIssues,
				Current:    totalProcessed,
				Total:      fetchResult.Total,
				Message:    fmt.Sprintf("Fetched %d/%d issues", totalProcessed, fetchResult.Total),
			})
		}

		nextPageToken = fetchResult.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	// Full sync: soft-delete issues not in current sync (DB subquery, no memory).
	if isFullSync {
		deleted, err := issueRepo.MarkDeletedNotInCurrentSync(ctx, pc.ID, syncedAt)
		if err != nil {
			s.logger.Warn("soft delete failed", "project", pc.Key, "error", err)
		} else if deleted > 0 {
			s.logger.Info("soft-deleted stale issues", "project", pc.Key, "count", deleted)
		}
	}

	return totalProcessed, nil
}

// syncMetadata performs Phase 2: metadata synchronization.
func (s *SyncService) syncMetadata(
	ctx context.Context,
	client *jira.Client,
	pc *models.ProjectConfig,
	metadataRepo repository.MetadataRepository,
) error {
	s.reportProgress(pc.Key, models.PhaseSyncMetadata, 0, 6, "Syncing metadata...")

	// Fetch all metadata types. These are small (tens of items) so in-memory is OK.
	statuses, err := client.FetchStatuses(ctx, pc.Key)
	if err != nil {
		return fmt.Errorf("fetch statuses: %w", err)
	}
	if err := metadataRepo.UpsertStatuses(ctx, pc.ID, statuses); err != nil {
		return err
	}
	s.reportProgress(pc.Key, models.PhaseSyncMetadata, 1, 6, fmt.Sprintf("Statuses (%d)", len(statuses)))

	priorities, err := client.FetchPriorities(ctx)
	if err != nil {
		return fmt.Errorf("fetch priorities: %w", err)
	}
	if err := metadataRepo.UpsertPriorities(ctx, pc.ID, priorities); err != nil {
		return err
	}
	s.reportProgress(pc.Key, models.PhaseSyncMetadata, 2, 6, fmt.Sprintf("Priorities (%d)", len(priorities)))

	issueTypes, err := client.FetchIssueTypes(ctx, pc.ID)
	if err != nil {
		return fmt.Errorf("fetch issue types: %w", err)
	}
	if err := metadataRepo.UpsertIssueTypes(ctx, pc.ID, issueTypes); err != nil {
		return err
	}
	s.reportProgress(pc.Key, models.PhaseSyncMetadata, 3, 6, fmt.Sprintf("IssueTypes (%d)", len(issueTypes)))

	components, err := client.FetchComponents(ctx, pc.Key)
	if err != nil {
		return fmt.Errorf("fetch components: %w", err)
	}
	if err := metadataRepo.UpsertComponents(ctx, pc.ID, components); err != nil {
		return err
	}
	s.reportProgress(pc.Key, models.PhaseSyncMetadata, 4, 6, fmt.Sprintf("Components (%d)", len(components)))

	versions, err := client.FetchVersions(ctx, pc.Key)
	if err != nil {
		return fmt.Errorf("fetch versions: %w", err)
	}
	if err := metadataRepo.UpsertFixVersions(ctx, pc.ID, versions); err != nil {
		return err
	}
	s.reportProgress(pc.Key, models.PhaseSyncMetadata, 5, 6, fmt.Sprintf("Versions (%d)", len(versions)))

	s.reportProgress(pc.Key, models.PhaseSyncMetadata, 6, 6, "Metadata sync complete")
	return nil
}

// verifyIntegrity performs Phase 4: data integrity verification.
func (s *SyncService) verifyIntegrity(
	ctx context.Context,
	client *jira.Client,
	pc *models.ProjectConfig,
	issueRepo repository.IssueRepository,
) {
	s.reportProgress(pc.Key, models.PhaseVerifyIntegrity, 0, 1, "Verifying data integrity...")

	jql := fmt.Sprintf("project = %s", pc.Key)
	jiraCount, err := client.GetIssueCount(ctx, jql)
	if err != nil {
		s.logger.Warn("failed to get JIRA issue count", "project", pc.Key, "error", err)
		return
	}

	localCount, err := issueRepo.CountByProject(ctx, pc.ID)
	if err != nil {
		s.logger.Warn("failed to get local issue count", "project", pc.Key, "error", err)
		return
	}

	if jiraCount != localCount {
		s.logger.Warn("issue count mismatch",
			"project", pc.Key,
			"jira", jiraCount,
			"local", localCount,
			"diff", jiraCount-localCount,
		)
	} else {
		s.logger.Info("data integrity OK", "project", pc.Key, "count", localCount)
	}

	s.reportProgress(pc.Key, models.PhaseVerifyIntegrity, 1, 1,
		fmt.Sprintf("Local: %d, JIRA: %d", localCount, jiraCount))
}

func (s *SyncService) reportProgress(projectKey, phase string, current, total int, message string) {
	if s.onProgress != nil {
		s.onProgress(models.SyncProgress{
			ProjectKey: projectKey,
			Phase:      phase,
			Current:    current,
			Total:      total,
			Message:    message,
		})
	}
}
