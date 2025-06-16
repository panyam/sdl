package commands

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/panyam/sdl/console"
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
	Short: "Start the SDL Canvas server with web dashboard",
	Long: `Start the SDL Canvas server that hosts the simulation engine, web dashboard, and API endpoints.
	
The server provides:
- Canvas simulation engine for SDL system execution
- REST API for all Canvas operations 
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
  
  # Or start server with initial files loaded
  sdl serve --load examples/contacts/contacts.sdl examples/common.sdl
  
  # Or start server on custom port
  sdl serve --port 9090 --no-logs
  sdl load --server http://localhost:9090 examples/contacts/contacts.sdl`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create web server with Canvas
		webServer := console.NewWebServer()
		router := webServer.GetRouter()
		canvas := webServer.GetCanvas()

		// Setup graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle shutdown signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Get server configuration
		host, port := getServeConfig()
		addr := fmt.Sprintf("%s:%d", host, port)
		server := &http.Server{
			Addr:    addr,
			Handler: router,
		}

		// Server startup message
		baseURL := fmt.Sprintf("http://%s", addr)
		fmt.Printf("🚀 SDL Canvas Server v1.0\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("📊 Dashboard:    %s\n", baseURL)
		fmt.Printf("🛠️  REST API:     %s/api/canvas\n", baseURL)
		fmt.Printf("📡 WebSocket:    ws://%s/api/live\n", addr)
		fmt.Printf("💻 CLI Commands: sdl load/use/gen/measure --server %s\n", baseURL)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		// Start server in goroutine
		go func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server failed to start: %v", err)
			}
		}()

		// Start statistics display if enabled
		if showStats {
			go displayServerStats(ctx, canvas)
		}

		// Show initial server status
		if showLogs {
			log.Printf("✅ Server started successfully on port %d", servePort)
			log.Printf("📝 Logging enabled (use --no-logs to disable)")
			if showStats {
				log.Printf("📈 Statistics display enabled (updates every %v)", statsInterval)
			}
		}

		// Load initial files if specified
		if len(loadFiles) > 0 {
			go loadInitialFiles(canvas, loadFiles)
		}

		// Wait for shutdown signal
		<-sigChan

		fmt.Println("\n🛑 Shutting down server...")

		// Graceful shutdown with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("⚠️  Server shutdown error: %v", err)
		} else {
			fmt.Println("✅ Server stopped gracefully")
		}
	},
}

// displayServerStats shows periodic server statistics
func displayServerStats(ctx context.Context, canvas *console.Canvas) {
	ticker := time.NewTicker(statsInterval)
	defer ticker.Stop()

	lastStats := canvas.GetStats()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := canvas.GetStats()
			if showLogs {
				if lastStats.LoadedFiles != stats.LoadedFiles ||
					lastStats.ActiveSystems != stats.ActiveSystems ||
					lastStats.ActiveGenerators != stats.ActiveGenerators ||
					lastStats.ActiveMeasurements != stats.ActiveMeasurements ||
					lastStats.TotalRuns != stats.TotalRuns {
					log.Printf("📊 Stats: Files=%d Systems=%d Generators=%d Measurements=%d Runs=%d",
						stats.LoadedFiles,
						stats.ActiveSystems,
						stats.ActiveGenerators,
						stats.ActiveMeasurements,
						stats.TotalRuns)
				}
				lastStats = stats
			}
		}
	}
}

// loadInitialFiles loads SDL files into the canvas on server startup
func loadInitialFiles(canvas *console.Canvas, files []string) {
	// Give the server a moment to fully start
	time.Sleep(1 * time.Second)

	if showLogs {
		log.Printf("📂 Loading %d initial file(s)...", len(files))
	}

	for _, file := range files {
		if showLogs {
			log.Printf("📂 Loading file: %s", file)
		}

		err := canvas.Load(file)
		if err != nil {
			if showLogs {
				log.Printf("❌ Failed to load file %s: %v", file, err)
			}
			continue
		}

		if showLogs {
			log.Printf("✅ Successfully loaded: %s", file)
		}
	}

	if showLogs {
		log.Printf("📂 Initial file loading completed")
	}
}

func init() {
	// Port and host are now handled by persistent flags in root.go
	serveCmd.Flags().BoolVar(&showLogs, "logs", true, "Show server logs")
	serveCmd.Flags().BoolVar(&showStats, "stats", true, "Show periodic statistics")
	serveCmd.Flags().DurationVar(&statsInterval, "stats-interval", 5*time.Second, "Statistics display interval")
	serveCmd.Flags().StringSliceVar(&loadFiles, "load", []string{}, "Initial SDL files to load on server startup")
	rootCmd.AddCommand(serveCmd)
}
