package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("go-jira version %s\n", version)
		},
	}
}
