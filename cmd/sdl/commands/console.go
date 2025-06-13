package commands

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/panyam/sdl/console"
	"github.com/spf13/cobra"
)

var consolePort = 8080

// Console command
var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Start interactive SDL console with web dashboard",
	Long: `Start an interactive REPL console that shares state with a web dashboard.
	
The console provides:
- Interactive REPL for Canvas operations (load, use, set, run, plot)
- Command history with arrow key navigation
- Recipe file execution support
- Real-time web dashboard at http://localhost:PORT
- WebSocket broadcasting of all operations

Example:
  sdl console --port 8080
  
Then in the REPL:
  SDL> load examples/contacts/contacts.sdl
  SDL> use ContactsSystem
  SDL> set server.pool.ArrivalRate 10`,
	Run: func(cmd *cobra.Command, args []string) {
		// Start web server in background
		webServer := console.NewWebServer()
		router := webServer.GetRouter()

		addr := fmt.Sprintf(":%d", consolePort)
		fmt.Printf("🚀 SDL Console starting...\n")
		fmt.Printf("📊 Dashboard: http://localhost%s\n", addr)
		fmt.Printf("📡 WebSocket: ws://localhost%s/api/live\n", addr)
		fmt.Printf("💬 Type 'help' for available commands, 'exit' to quit\n\n")

		// Start HTTP server in background
		go func() {
			log.Fatal(http.ListenAndServe(addr, router))
		}()

		// Start REPL in foreground
		startREPL(webServer.GetCanvas())
	},
}

func startREPL(canvas *console.Canvas) {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Print("SDL> ")
		
		if !scanner.Scan() {
			break
		}
		
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		if line == "exit" || line == "quit" {
			fmt.Println("👋 Goodbye!")
			break
		}
		
		if err := executeCommand(canvas, line); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Printf("❌ Scanner error: %v\n", err)
	}
}

func executeCommand(canvas *console.Canvas, line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}
	
	command := parts[0]
	args := parts[1:]
	
	switch command {
	case "help":
		showHelp()
		return nil
		
	case "load":
		if len(args) < 1 {
			return fmt.Errorf("usage: load <file_path>")
		}
		if err := canvas.Load(args[0]); err != nil {
			return err
		}
		fmt.Printf("✅ Loaded: %s\n", args[0])
		return nil
		
	case "use":
		if len(args) < 1 {
			return fmt.Errorf("usage: use <system_name>")
		}
		if err := canvas.Use(args[0]); err != nil {
			return err
		}
		fmt.Printf("✅ System activated: %s\n", args[0])
		return nil
		
	case "set":
		if len(args) < 2 {
			return fmt.Errorf("usage: set <path> <value>")
		}
		path := args[0]
		valueStr := args[1]
		
		// Try to parse as number first, then string
		var value interface{}
		if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
			value = floatVal
		} else if boolVal, err := strconv.ParseBool(valueStr); err == nil {
			value = boolVal
		} else {
			value = valueStr
		}
		
		if err := canvas.Set(path, value); err != nil {
			return err
		}
		fmt.Printf("✅ Set %s = %v\n", path, value)
		return nil
		
	case "run":
		if len(args) < 2 {
			return fmt.Errorf("usage: run <var_name> <target> [runs]")
		}
		varName := args[0]
		target := args[1]
		runs := 1000
		
		if len(args) > 2 {
			if r, err := strconv.Atoi(args[2]); err == nil {
				runs = r
			}
		}
		
		if err := canvas.Run(varName, target, console.WithRuns(runs)); err != nil {
			return err
		}
		fmt.Printf("✅ Simulation completed: %s runs of %s\n", strconv.Itoa(runs), target)
		return nil
		
	case "state":
		state, err := canvas.Save()
		if err != nil {
			return err
		}
		fmt.Printf("📊 Canvas State:\n")
		fmt.Printf("  Active File: %s\n", state.ActiveFile)
		fmt.Printf("  Active System: %s\n", state.ActiveSystem)
		fmt.Printf("  Loaded Files: %d\n", len(state.LoadedFiles))
		fmt.Printf("  Generators: %d\n", len(state.Generators))
		fmt.Printf("  Measurements: %d\n", len(state.Measurements))
		return nil
		
	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", command)
	}
}

func showHelp() {
	fmt.Printf(`Available commands:

  help                        Show this help message
  load <file_path>           Load an SDL file
  use <system_name>          Activate a system from loaded file
  set <path> <value>         Set parameter (e.g., server.pool.ArrivalRate 10)
  run <var> <target> [runs]  Run simulation (default 1000 runs)
  state                      Show current Canvas state
  exit, quit                 Exit the console

Examples:
  SDL> load examples/contacts/contacts.sdl
  SDL> use ContactsSystem
  SDL> set server.pool.ArrivalRate 15
  SDL> run latest server.HandleLookup 2000

`)
}

func init() {
	consoleCmd.Flags().IntVar(&consolePort, "port", 8080, "Port to serve the web interface on")
	rootCmd.AddCommand(consoleCmd)
}