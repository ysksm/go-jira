package cli

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/ysksm/go-jira/core/infrastructure/config"
	"github.com/ysksm/go-jira/core/infrastructure/database"
)

var (
	verbose bool
	quiet   bool
)

// NewRootCommand creates the root CLI command with all subcommands.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "go-jira",
		Short: "JIRA data synchronization and analysis tool",
		Long:  "Sync JIRA data locally for fast search, SQL queries, and analysis.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}
			if quiet {
				level = slog.LevelError
			}
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
		},
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error output")

	root.AddCommand(
		newConfigCmd(),
		newProjectCmd(),
		newSyncCmd(),
		newIssueCmd(),
		newQueryCmd(),
		newVersionCmd(),
	)

	return root
}

// getConfigStore returns the file config store.
func getConfigStore() (*config.FileConfigStore, error) {
	return config.NewFileConfigStore("")
}

// getConnMgr returns a database connection manager using current settings.
func getConnMgr() (*database.Connection, error) {
	store, err := getConfigStore()
	if err != nil {
		return nil, err
	}
	settings, err := store.Load()
	if err != nil {
		return nil, err
	}
	if settings.Database.DatabaseDir == "" {
		return nil, nil
	}
	return database.NewConnection(settings.Database.DatabaseDir), nil
}
