package commands

import (
	"fmt"
	"log"
	"net/http"

	"github.com/panyam/sdl/console"
	"github.com/spf13/cobra"
)

var servePort = 8080

// Serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the SDL web visualization server",
	Long: `Start an HTTP server that provides a web interface for interactive SDL system visualization.
	
The server provides:
- REST API for Canvas operations (load, use, set, run, plot)
- RESTful API for traffic generation and measurement management
- WebSocket connection for real-time updates
- Static file serving for the frontend dashboard

Example:
  sdl serve --port 8080`,
	Run: func(cmd *cobra.Command, args []string) {
		webServer := console.NewWebServer()
		router := webServer.GetRouter()

		addr := fmt.Sprintf(":%d", servePort)
		fmt.Printf("🚀 SDL Web Server starting on http://localhost%s\n", addr)
		fmt.Printf("📊 Dashboard: http://localhost%s\n", addr)
		fmt.Printf("🔌 Legacy API: http://localhost%s/api\n", addr)
		fmt.Printf("🛠️  RESTful API: http://localhost%s/api/canvas\n", addr)
		fmt.Printf("📡 WebSocket: ws://localhost%s/api/live\n", addr)

		log.Fatal(http.ListenAndServe(addr, router))
	},
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 8080, "Port to serve the web interface on")
	rootCmd.AddCommand(serveCmd)
}