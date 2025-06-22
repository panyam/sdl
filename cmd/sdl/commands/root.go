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
	serverURL      string
	grpcAddress    string
	gatewayAddress string
	canvasID       string
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
	
	// Canvas ID defaults to SDL_CANVAS_ID env var, then "default"
	defaultCanvasID := os.Getenv("SDL_CANVAS_ID")
	if defaultCanvasID == "" {
		defaultCanvasID = "default"
	}
	rootCmd.PersistentFlags().StringVar(&canvasID, "canvas", defaultCanvasID, "Canvas ID to use for operations (default: SDL_CANVAS_ID env var or 'default')")

	// Serve command flags
	rootCmd.PersistentFlags().StringVar(&gatewayAddress, "gwaddr", DefaultGatewayAddress(), "Host/Port of the Gateway Server (default: CANVAS_GATEWAY_SERVER_ADDRESS env var or localhost)")
	rootCmd.PersistentFlags().StringVar(&grpcAddress, "grpcaddr", DefaultServiceAddress(), "Host/Port of the GRPC Server (default: CANVAS_GRPC_SERVER_ADDRESS env var or localhost)")
}

// AddCommand allows adding subcommands from other files.
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

func DefaultGatewayAddress() string {
	gateway_addr := os.Getenv("CANVAS_GATEWAY_SERVER_ADDRESS")
	if gateway_addr != "" {
		return gateway_addr
	}
	return ":8080"
}

func DefaultServiceAddress() string {
	port := os.Getenv("CANVAS_GRPC_SERVER_ADDRESS")
	if port != "" {
		return port
	}
	return ":9090"
}
