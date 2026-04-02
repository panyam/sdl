package commands

import (
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Workspace management commands",
	Long:  `Commands for creating and managing SDL workspaces.`,
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
}