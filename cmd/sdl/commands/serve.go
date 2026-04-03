package commands

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	goal "github.com/panyam/goapplib"
	skhttp "github.com/panyam/servicekit/http"
	"github.com/panyam/sdl/lib/loader"
	"github.com/panyam/sdl/services/devenvbe"
	"github.com/panyam/sdl/web/server"
	"github.com/spf13/cobra"
)

var (
	showLogs      = true
	showStats     = true
	statsInterval = 30 * time.Second
	loadFiles     []string
)

// Serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the SDL server with web dashboard",
	Long: `Start the SDL server that hosts the simulation engine, web dashboard, and API endpoints.

The server provides:
- Workspace simulation engine for SDL system execution
- REST API for all workspace operations
- RESTful API for traffic generation and measurement management
- WebSocket connection for real-time updates
- Web dashboard for visualization and control
- Traffic generator and measurement logging
- Server statistics and health monitoring

Use the server with direct CLI commands for a clean shell experience.

Example:
  # Terminal 1: Start server
  sdl serve

  # Terminal 2: Use CLI commands
  sdl load examples/contacts/contacts.sdl
  sdl use ContactsSystem
  sdl gen add load1 server.HandleLookup 10
  sdl gen start load1
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create filesystem with @stdlib/ mounted and local filesystem as fallback
		cfs := loader.NewCompositeFS()
		cfs.SetFallback(loader.NewLocalFS("")) // handles relative and absolute paths
		stdlibPath := findStdlibPath()
		if stdlibPath != "" {
			cfs.Mount("@stdlib/", loader.NewLocalFS(stdlibPath))
			log.Printf("Mounted @stdlib/ from %s", stdlibPath)
		}
		fsResolver := loader.NewFileSystemResolver(cfs)
		wsSvc := devenvbe.NewWorkspaceService(fsResolver)

		// Start gRPC server in background
		log.Println("gRPC address:", grpcAddress)
		grpcSrv := &server.Server{Address: grpcAddress, WorkspaceService: wsSvc}
		grpcErr := make(chan error, 1)
		grpcStop := make(chan bool, 1)
		if err := grpcSrv.Start(cmd.Context(), grpcErr, grpcStop); err != nil {
			slog.Error("Failed to start gRPC server", "error", err)
			os.Exit(1)
		}

		// Build HTTP handler
		sdlApp, _, _ := server.NewSdlApp(grpcAddress)
		handler := goal.CORS(sdlApp.Handler())

		log.Println("HTTP address:", gatewayAddress)
		goal.PrintStartupMessage(gatewayAddress)

		srv := &http.Server{
			Addr:    gatewayAddress,
			Handler: handler,
		}

		// ListenAndServeGraceful owns the process lifetime:
		// on SIGTERM/SIGINT → stop gRPC → drain HTTP → exit
		if err := skhttp.ListenAndServeGraceful(srv,
			skhttp.WithDrainTimeout(10*time.Second),
			skhttp.WithOnShutdown(func() {
				slog.Info("Stopping gRPC server...")
				grpcStop <- true
			}),
		); err != nil {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
		slog.Info("Server stopped")
	},
}

// findStdlibPath looks for the stdlib directory in common locations.
func findStdlibPath() string {
	candidates := []string{
		"cmd/wasm/stdlib",    // running from project root
		"../cmd/wasm/stdlib", // running from cmd/sdl
		"stdlib",             // running from cmd/wasm
	}
	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(path)
			return abs
		}
	}
	return ""
}

func init() {
	// Port and host are now handled by persistent flags in root.go
	serveCmd.Flags().BoolVar(&showLogs, "logs", true, "Show server logs")
	serveCmd.Flags().BoolVar(&showStats, "stats", true, "Show periodic statistics")
	serveCmd.Flags().DurationVar(&statsInterval, "stats-interval", 5*time.Second, "Statistics display interval")
	serveCmd.Flags().StringSliceVar(&loadFiles, "load", []string{}, "Initial SDL files to load on server startup")
	rootCmd.AddCommand(serveCmd)
}
