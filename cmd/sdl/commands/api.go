package commands

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	ghttp "github.com/panyam/goutils/http"
)

// Global server configuration - variables declared in root.go

// getServerURL returns the server URL using the priority:
// 1. Command line flag (--server)
// 2. Environment variable (CANVAS_SERVER_URL)
// 3. Default (http://localhost:8080)
func getServerURL() string {
	if serverURL != "" {
		return serverURL
	}

	if envURL := os.Getenv("CANVAS_SERVER_URL"); envURL != "" {
		return envURL
	}

	return "http://localhost:8080"
}

// getServeConfig returns the host and port for the serve command using:
// 1. Command line flags (--host, --port)
// 2. Environment variables (CANVAS_SERVE_HOST, CANVAS_SERVE_PORT)
// 3. Defaults (localhost, 8080)
func getServeConfig() (host string, port int) {
	// Use provided values if set
	if serveHost != "" && servePort != 0 {
		return serveHost, servePort
	}

	// Check environment variables
	envHost := os.Getenv("CANVAS_SERVE_HOST")
	if envHost == "" {
		envHost = "localhost"
	}

	envPort := 8080
	if envPortStr := os.Getenv("CANVAS_SERVE_PORT"); envPortStr != "" {
		if parsed, err := strconv.Atoi(envPortStr); err == nil {
			envPort = parsed
		}
	}

	// Use environment or command line values
	if serveHost == "" {
		serveHost = envHost
	}
	if servePort == 0 {
		servePort = envPort
	}

	return serveHost, servePort
}

func apiEndpoint(endpoint string) string {
	return strings.TrimSuffix(getServerURL(), "/") + endpoint
}

// makeAPICall makes HTTP requests to the SDL server
func makeAPICall(method, endpoint string, body map[string]any) (map[string]any, error) {
	req, _ := ghttp.NewJsonRequest(method, apiEndpoint(endpoint), body)
	resp, err := ghttp.Call(req, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return nil, err
	}
	return resp.(map[string]any), err
}

// testServerConnection verifies the server is reachable
func testServerConnection() error {
	client := &http.Client{Timeout: 5 * time.Second}

	// Parse server URL to build health check endpoint
	baseURL := getServerURL()
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %v", err)
	}

	healthURL := fmt.Sprintf("%s://%s/api/console/help", parsedURL.Scheme, parsedURL.Host)
	resp, err := client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("cannot connect to SDL server at %s: %v", baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// checkServerConnection checks if server is available and provides helpful guidance
func checkServerConnection() error {
	if err := testServerConnection(); err != nil {
		baseURL := getServerURL()
		fmt.Printf("❌ Cannot connect to SDL server at %s\n\n", baseURL)
		fmt.Printf("To use SDL commands, first start the server:\n\n")
		fmt.Printf("🚀 Terminal 1: Start SDL server\n")
		fmt.Printf("   sdl serve\n\n")
		fmt.Printf("🔌 Terminal 2: Use CLI commands\n")
		fmt.Printf("   sdl load examples/contacts/contacts.sdl\n\n")
		fmt.Printf("Or connect to a different server:\n")
		fmt.Printf("   export CANVAS_SERVER_URL=http://other-host:8080\n")
		fmt.Printf("   sdl load examples/contacts/contacts.sdl\n\n")
		fmt.Printf("💡 The server hosts the Canvas engine and web dashboard.\n")
		return err
	}
	return nil
}
