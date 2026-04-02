package commands

import (
	"github.com/spf13/cobra"
)

var canvasCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Canvas management commands",
	Long:  `Commands for creating and managing SDL canvases.`,
}

func init() {
	rootCmd.AddCommand(canvasCmd)
}