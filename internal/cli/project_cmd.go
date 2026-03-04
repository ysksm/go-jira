package cli

import (
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ysksm/go-jira/core/service"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage JIRA projects",
	}

	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectFetchCmd())
	cmd.AddCommand(newProjectEnableCmd())
	cmd.AddCommand(newProjectDisableCmd())
	return cmd
}

func newProjectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getConfigStore()
			if err != nil {
				return err
			}
			svc := service.NewProjectService(store, slog.Default())
			projects, err := svc.List()
			if err != nil {
				return err
			}

			if len(projects) == 0 {
				fmt.Println("No projects configured. Run 'go-jira project fetch' to get projects from JIRA.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "KEY\tNAME\tENABLED\tLAST SYNCED")
			for _, p := range projects {
				enabled := "no"
				if p.Enabled {
					enabled = "yes"
				}
				lastSynced := "-"
				if p.LastSyncedAt != "" {
					lastSynced = p.LastSyncedAt
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Key, p.Name, enabled, lastSynced)
			}
			w.Flush()
			return nil
		},
	}
}

func newProjectFetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fetch",
		Short: "Fetch projects from JIRA",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getConfigStore()
			if err != nil {
				return err
			}
			svc := service.NewProjectService(store, slog.Default())
			projects, newCount, err := svc.FetchFromJira(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Printf("Found %d projects (%d new)\n", len(projects), newCount)
			for _, p := range projects {
				fmt.Printf("  %s - %s\n", p.Key, p.Name)
			}
			return nil
		},
	}
}

func newProjectEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <key>",
		Short: "Enable sync for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getConfigStore()
			if err != nil {
				return err
			}
			svc := service.NewProjectService(store, slog.Default())
			project, err := svc.Enable(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Enabled sync for project %s (%s)\n", project.Key, project.Name)
			return nil
		},
	}
}

func newProjectDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <key>",
		Short: "Disable sync for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getConfigStore()
			if err != nil {
				return err
			}
			svc := service.NewProjectService(store, slog.Default())
			project, err := svc.Disable(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Disabled sync for project %s (%s)\n", project.Key, project.Name)
			return nil
		},
	}
}
