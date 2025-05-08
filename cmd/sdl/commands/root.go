package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Global flags can be defined here if needed
// var cfgFile string
var dslFilePath string

var rootCmd = &cobra.Command{
	Use:   "sdl",
	Short: "SDL is a System Design Language processor and analyzer",
	Long: `SDL (System Design Language) provides tools to define, model,
and analyze the performance characteristics of distributed systems.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// Example:
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sdl.yaml)")

	// A common flag for specifying the DSL file, can be made persistent if most commands need it.
	// Alternatively, each command can define its own file flag if preferred.
	// Let's make it persistent for now.
	rootCmd.PersistentFlags().StringVarP(&dslFilePath, "file", "f", "", "Path to the DSL file (required by many commands)")
}

// AddCommand allows adding subcommands from other files.
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
