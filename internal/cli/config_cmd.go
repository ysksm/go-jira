package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ysksm/go-jira/core/service"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigShowCmd())
	return cmd
}

func newConfigInitCmd() *cobra.Command {
	var endpoint, username, apiKey, dbPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration with JIRA credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getConfigStore()
			if err != nil {
				return err
			}
			svc := service.NewConfigService(store)
			settings, err := svc.Initialize(endpoint, username, apiKey, dbPath)
			if err != nil {
				return err
			}
			fmt.Println("Configuration initialized successfully.")
			fmt.Printf("Config file: %s\n", store.Path())
			fmt.Printf("Database dir: %s\n", settings.Database.DatabaseDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&endpoint, "endpoint", "", "JIRA endpoint URL (required)")
	cmd.Flags().StringVar(&username, "username", "", "JIRA username (required)")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "JIRA API key (required)")
	cmd.Flags().StringVar(&dbPath, "db-path", "", "Database directory path (optional)")
	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("api-key")

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getConfigStore()
			if err != nil {
				return err
			}
			svc := service.NewConfigService(store)
			settings, err := svc.Get()
			if err != nil {
				return err
			}

			data, _ := json.MarshalIndent(settings, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}
