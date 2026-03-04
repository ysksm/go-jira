package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ysksm/go-jira/core/service"
)

func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Execute SQL queries",
	}

	cmd.AddCommand(newQueryExecCmd())
	cmd.AddCommand(newQuerySchemaCmd())
	return cmd
}

func newQueryExecCmd() *cobra.Command {
	var project string
	var limit int

	cmd := &cobra.Command{
		Use:   "exec <sql>",
		Short: "Execute a SQL query",
		Args:  cobra.ExactArgs(1),
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

			svc := service.NewQueryService(connMgr)
			result, err := svc.Execute(cmd.Context(), project, args[0], limit)
			if err != nil {
				return err
			}

			if len(result.Columns) == 0 {
				fmt.Println("No results.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, strings.Join(result.Columns, "\t"))
			for _, row := range result.Rows {
				vals := make([]string, len(row))
				for i, v := range row {
					vals[i] = fmt.Sprintf("%v", v)
				}
				fmt.Fprintln(w, strings.Join(vals, "\t"))
			}
			w.Flush()
			fmt.Printf("\n%d rows (%dms)\n", result.RowCount, result.ExecutionTimeMs)
			return nil
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "Project key (required)")
	cmd.Flags().IntVarP(&limit, "limit", "l", 500, "Maximum rows")

	return cmd
}

func newQuerySchemaCmd() *cobra.Command {
	var project string

	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Show database schema",
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

			svc := service.NewQueryService(connMgr)
			tables, err := svc.GetSchema(cmd.Context(), project)
			if err != nil {
				return err
			}

			for _, table := range tables {
				fmt.Printf("Table: %s\n", table.Name)
				for _, col := range table.Columns {
					nullable := ""
					if col.IsNullable {
						nullable = " (nullable)"
					}
					fmt.Printf("  %-30s %s%s\n", col.Name, col.DataType, nullable)
				}
				fmt.Println()
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "Project key (required)")

	return cmd
}
