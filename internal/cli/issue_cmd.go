package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ysksm/go-jira/core/infrastructure/database"
	"github.com/ysksm/go-jira/core/service"
)

func newIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Search and view issues",
	}

	cmd.AddCommand(newIssueSearchCmd())
	cmd.AddCommand(newIssueGetCmd())
	cmd.AddCommand(newIssueHistoryCmd())
	return cmd
}

func newIssueSearchCmd() *cobra.Command {
	var project string
	var limit, offset int

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return fmt.Errorf("--project is required")
			}

			connMgr, err := getConnMgr()
			if err != nil {
				return err
			}
			if connMgr == nil {
				return fmt.Errorf("database not configured")
			}
			defer connMgr.Close()

			db, err := connMgr.GetDB(project)
			if err != nil {
				return err
			}

			issueRepo := database.NewIssueRepository(db)
			changeHistoryRepo := database.NewChangeHistoryRepository(db)
			svc := service.NewIssueService(issueRepo, changeHistoryRepo)

			result, err := svc.Search(cmd.Context(), project, offset, limit)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "KEY\tSTATUS\tPRIORITY\tASSIGNEE\tSUMMARY")
			for _, issue := range result.Issues {
				summary := issue.Summary
				if len(summary) > 60 {
					summary = summary[:57] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					issue.Key, issue.Status, issue.Priority, issue.Assignee, summary)
			}
			w.Flush()
			fmt.Printf("\nShowing %d of %d issues\n", len(result.Issues), result.Total)
			return nil
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "Project key (required)")
	cmd.Flags().IntVarP(&limit, "limit", "l", 50, "Maximum results")
	cmd.Flags().IntVarP(&offset, "offset", "o", 0, "Result offset")

	return cmd
}

func newIssueGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get issue details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			connMgr, err := getConnMgr()
			if err != nil {
				return err
			}
			if connMgr == nil {
				return fmt.Errorf("database not configured")
			}
			defer connMgr.Close()

			// Extract project key from issue key (e.g., "PROJ-123" → "PROJ").
			issueKey := args[0]
			projectKey := extractProjectKey(issueKey)
			if projectKey == "" {
				return fmt.Errorf("invalid issue key: %s", issueKey)
			}

			db, err := connMgr.GetDB(projectKey)
			if err != nil {
				return err
			}

			issueRepo := database.NewIssueRepository(db)
			changeHistoryRepo := database.NewChangeHistoryRepository(db)
			svc := service.NewIssueService(issueRepo, changeHistoryRepo)

			issue, err := svc.Get(cmd.Context(), issueKey)
			if err != nil {
				return err
			}
			if issue == nil {
				return fmt.Errorf("issue %s not found", issueKey)
			}

			fmt.Printf("Key:        %s\n", issue.Key)
			fmt.Printf("Summary:    %s\n", issue.Summary)
			fmt.Printf("Status:     %s\n", issue.Status)
			fmt.Printf("Priority:   %s\n", issue.Priority)
			fmt.Printf("Type:       %s\n", issue.IssueType)
			fmt.Printf("Assignee:   %s\n", issue.Assignee)
			fmt.Printf("Reporter:   %s\n", issue.Reporter)
			if issue.ParentKey != "" {
				fmt.Printf("Parent:     %s\n", issue.ParentKey)
			}
			if len(issue.Labels) > 0 {
				fmt.Printf("Labels:     %v\n", issue.Labels)
			}
			if issue.Description != "" {
				fmt.Printf("Description:\n%s\n", issue.Description)
			}
			return nil
		},
	}
}

func newIssueHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history <key>",
		Short: "Show issue change history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			connMgr, err := getConnMgr()
			if err != nil {
				return err
			}
			if connMgr == nil {
				return fmt.Errorf("database not configured")
			}
			defer connMgr.Close()

			issueKey := args[0]
			projectKey := extractProjectKey(issueKey)
			if projectKey == "" {
				return fmt.Errorf("invalid issue key: %s", issueKey)
			}

			db, err := connMgr.GetDB(projectKey)
			if err != nil {
				return err
			}

			issueRepo := database.NewIssueRepository(db)
			changeHistoryRepo := database.NewChangeHistoryRepository(db)
			svc := service.NewIssueService(issueRepo, changeHistoryRepo)

			history, err := svc.GetHistory(cmd.Context(), issueKey)
			if err != nil {
				return err
			}

			if len(history) == 0 {
				fmt.Println("No change history found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "DATE\tFIELD\tFROM\tTO\tAUTHOR")
			for _, h := range history {
				from := h.FromString
				if len(from) > 30 {
					from = from[:27] + "..."
				}
				to := h.ToString
				if len(to) > 30 {
					to = to[:27] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					h.ChangedAt.Format("2006-01-02 15:04"),
					h.Field, from, to, h.AuthorDisplayName)
			}
			w.Flush()
			return nil
		},
	}
}

func extractProjectKey(issueKey string) string {
	for i, c := range issueKey {
		if c == '-' && i > 0 {
			return issueKey[:i]
		}
	}
	return ""
}
