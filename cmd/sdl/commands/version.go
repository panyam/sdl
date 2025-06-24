package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print SDL version information",
	Long:  `Print detailed version information including version number, git commit, and build date.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("SDL %s\n", Version)
		if GitCommit != "none" {
			fmt.Printf("Git commit: %s\n", GitCommit)
		}
		if BuildDate != "unknown" {
			fmt.Printf("Build date: %s\n", BuildDate)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}