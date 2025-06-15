package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Global flags can be defined here if needed
// var cfgFile string
var dslFilePath string

// Imported from api.go - these need to be accessible for flag binding
var (
	serverURL string
	serveHost string
	servePort int
)

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
	// Global persistent flags
	rootCmd.PersistentFlags().StringVarP(&dslFilePath, "file", "f", "", "Path to the DSL file (required by many commands)")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "SDL server URL (default: CANVAS_SERVER_URL env var or http://localhost:8080)")
	
	// Serve command flags
	rootCmd.PersistentFlags().StringVar(&serveHost, "host", "", "Server host (default: CANVAS_SERVE_HOST env var or localhost)")
	rootCmd.PersistentFlags().IntVar(&servePort, "port", 0, "Server port (default: CANVAS_SERVE_PORT env var or 8080)")
}

// AddCommand allows adding subcommands from other files.
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
