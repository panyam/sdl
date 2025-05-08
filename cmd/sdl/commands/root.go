package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sdl",
	Short: "SDL is a System Design Language processor and analyzer",
	Long: `SDL (System Design Language) provides tools to define, model,
and analyze the performance characteristics of distributed systems.`,
	// Run: func(cmd *cobra.Command, args []string) { }, // No action for root if subcommands are used
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// init will be called when the package is loaded.
// We can add global flags here if needed.
func init() {
	// Example global flag:
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sdl.yaml)")
}

// AddCommand allows adding subcommands from other files.
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
