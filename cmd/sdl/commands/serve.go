package commands

import (
	"context"
	"log"
	"time"

	"github.com/panyam/sdl/services"
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
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create shared CanvasService instance for gRPC server
		canvasService := services.NewCanvasService()

		// Create servers
		log.Println("Grpc, Address: ", grpcAddress)
		log.Println("gateway, Address: ", gatewayAddress)
		app := App{Ctx: context.Background()}
		app.AddServer(&server.Server{Address: grpcAddress, CanvasService: canvasService})
		app.AddServer(&server.WebAppServer{
			WebAppServer: server.NewWebAppServerConfig(gatewayAddress, grpcAddress, true),
		})
		app.Start()
		app.Done(nil)
	},
}

// displayServerStats shows periodic server statistics
/*
func displayServerStats(ctx context.Context, canvas *services.Canvas) {
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
*/

// loadInitialFiles loads SDL files into the canvas on server startup
func loadInitialFiles(canvas *services.Canvas, files []string) {
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
