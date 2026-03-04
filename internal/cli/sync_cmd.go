package cli

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/ysksm/go-jira/core/domain/models"
	"github.com/ysksm/go-jira/core/service"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize JIRA data",
	}

	cmd.AddCommand(newSyncRunCmd())
	cmd.AddCommand(newSyncStatusCmd())
	return cmd
}

func newSyncRunCmd() *cobra.Command {
	var projectKey string
	var force bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run sync for enabled projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getConfigStore()
			if err != nil {
				return err
			}
			connMgr, err := getConnMgr()
			if err != nil {
				return err
			}
			if connMgr == nil {
				return fmt.Errorf("database not configured. Run 'go-jira config init' first")
			}
			defer connMgr.Close()

			svc := service.NewSyncService(store, connMgr, slog.Default())

			// Set up progress display.
			progress := newProgressDisplay()
			svc.SetProgressCallback(func(p models.SyncProgress) {
				progress.update(p)
			})

			results, err := svc.Execute(cmd.Context(), service.SyncOptions{
				ProjectKey: projectKey,
				Force:      force,
			})
			if err != nil {
				return err
			}

			// Print summary.
			fmt.Println()
			fmt.Println("Summary:")
			totalIssues := 0
			totalDuration := 0.0
			for _, r := range results {
				status := "OK"
				if !r.Success {
					status = fmt.Sprintf("FAILED: %s", r.Error)
				}
				fmt.Printf("  %s: %d issues (%.1fs) - %s\n", r.ProjectKey, r.IssueCount, r.Duration, status)
				totalIssues += r.IssueCount
				totalDuration += r.Duration
			}
			fmt.Printf("  Total: %d projects, %d issues, %.1fs\n", len(results), totalIssues, totalDuration)

			return nil
		},
	}

	cmd.Flags().StringVarP(&projectKey, "project", "p", "", "Sync specific project only")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force full sync (ignore incremental)")

	return cmd
}

func newSyncStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getConfigStore()
			if err != nil {
				return err
			}
			settings, err := store.Load()
			if err != nil {
				return err
			}

			fmt.Println("Sync status:")
			for _, pc := range settings.Projects {
				if !pc.SyncEnabled {
					continue
				}
				lastSynced := "never"
				if pc.LastSynced != nil {
					lastSynced = pc.LastSynced.Format("2006-01-02 15:04:05")
				}
				checkpoint := "none"
				if pc.SyncCheckpoint != nil {
					checkpoint = fmt.Sprintf("at %s (%d items)", pc.SyncCheckpoint.LastIssueKey, pc.SyncCheckpoint.ItemsProcessed)
				}
				fmt.Printf("  %s: last=%s, checkpoint=%s\n", pc.Key, lastSynced, checkpoint)
			}
			return nil
		},
	}
}
