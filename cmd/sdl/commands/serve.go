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
	servePort    = 8080
	showLogs     = true
	showStats    = true
	statsInterval = 5 * time.Second
)

// Serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the SDL Canvas server with web dashboard",
	Long: `Start the SDL Canvas server that hosts the simulation engine, web dashboard, and API endpoints.
	
The server provides:
- Canvas simulation engine for SDL system execution
- REST API for all Canvas operations (load, use, set, run, plot, etc.)
- RESTful API for traffic generation and measurement management  
- WebSocket connection for real-time updates
- Web dashboard for visualization and control
- Traffic generator and measurement logging
- Server statistics and health monitoring

This server can be used standalone or with the SDL console client for a clean REPL experience.

Example:
  # Start server with default settings
  sdl serve
  
  # Start server on custom port without logs
  sdl serve --port 9090 --no-logs
  
  # Then connect with console client
  sdl console --server http://localhost:9090`,
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

		// Start server
		addr := fmt.Sprintf(":%d", servePort)
		server := &http.Server{
			Addr:    addr,
			Handler: router,
		}

		// Server startup message
		fmt.Printf("🚀 SDL Canvas Server v1.0\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("📊 Dashboard:    http://localhost%s\n", addr)
		fmt.Printf("🛠️  REST API:     http://localhost%s/api/canvas\n", addr)
		fmt.Printf("📡 WebSocket:    ws://localhost%s/api/live\n", addr) 
		fmt.Printf("💻 Console:      sdl console --server http://localhost%s\n", addr)
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

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := canvas.GetStats()
			if showLogs {
				log.Printf("📊 Stats: Files=%d Systems=%d Generators=%d Measurements=%d Runs=%d",
					stats.LoadedFiles,
					stats.ActiveSystems,
					stats.ActiveGenerators,
					stats.ActiveMeasurements,
					stats.TotalRuns)
			}
		}
	}
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port to serve on")
	serveCmd.Flags().BoolVar(&showLogs, "logs", true, "Show server logs")
	serveCmd.Flags().BoolVar(&showStats, "stats", true, "Show periodic statistics")
	serveCmd.Flags().DurationVar(&statsInterval, "stats-interval", 5*time.Second, "Statistics display interval")
	rootCmd.AddCommand(serveCmd)
}