package commands

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/panyam/sdl/console"
	"github.com/spf13/cobra"
)

// Global variables for the prompt context
var (
	currentCanvas  *console.Canvas
	commandHistory []string
	consolePort    = 8080
	historyFile    string
)

// Console command with go-prompt
var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Start interactive SDL console with web dashboard",
	Long: `Start an interactive REPL console that shares state with a web dashboard.
	
The console provides:
- Interactive REPL with tab completion and command history
- Rich auto-completion with descriptions
- Arrow key navigation and multi-line support
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
		fmt.Printf("💬 Type 'help' for available commands, Ctrl+D to quit\n\n")

		// Start HTTP server in background
		go func() {
			log.Fatal(http.ListenAndServe(addr, router))
		}()

		// Store canvas for global access
		currentCanvas = webServer.GetCanvas()

		// Initialize command history
		initializeHistory()

		// Setup signal handling to save history on exit
		setupSignalHandling()

		// Start enhanced REPL with go-prompt
		startEnhancedREPL()
	},
}

// Command structure for better organization
type commandInfo struct {
	Name        string
	Description string
	Usage       string
	MinArgs     int
}

var commands = []commandInfo{
	{Name: "help", Description: "Show help message", Usage: "help", MinArgs: 0},
	{Name: "load", Description: "Load an SDL file", Usage: "load <file_path>", MinArgs: 1},
	{Name: "use", Description: "Activate a system from loaded file", Usage: "use <system_name>", MinArgs: 1},
	{Name: "set", Description: "Set parameter value", Usage: "set <path> <value>", MinArgs: 2},
	{Name: "run", Description: "Run simulation", Usage: "run <var> <target> [runs]", MinArgs: 2},
	{Name: "execute", Description: "Execute commands from a recipe file", Usage: "execute <recipe_file>", MinArgs: 1},
	{Name: "state", Description: "Show current Canvas state", Usage: "state", MinArgs: 0},
	{Name: "exit", Description: "Exit the console", Usage: "exit", MinArgs: 0},
	{Name: "quit", Description: "Exit the console", Usage: "quit", MinArgs: 0},
}

func startEnhancedREPL() {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionTitle("SDL Console"),
		prompt.OptionPrefix(getPromptPrefix()),
		prompt.OptionLivePrefix(getLivePrefix),
		prompt.OptionHistory(commandHistory),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
		prompt.OptionDescriptionBGColor(prompt.DarkGray),
		prompt.OptionDescriptionTextColor(prompt.White),
		prompt.OptionCompletionWordSeparator(" "),
		// Control auto-suggest behavior
		prompt.OptionMaxSuggestion(10),
	)
	p.Run()
	
	// Save history on exit
	saveHistory()
}

func getPromptPrefix() string {
	if currentCanvas == nil {
		return "SDL> "
	}
	
	state, err := currentCanvas.Save()
	if err != nil || state.ActiveSystem == "" {
		return "SDL> "
	}
	
	// Show active system in prompt
	return fmt.Sprintf("SDL[%s]> ", state.ActiveSystem)
}

func getLivePrefix() (string, bool) {
	return getPromptPrefix(), true
}

// History management functions
func setupSignalHandling() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		fmt.Println("\n\n👋 Saving history and exiting...")
		saveHistory()
		os.Exit(0)
	}()
}

func initializeHistory() {
	// Get history file path
	historyFile = getHistoryFilePath()
	
	// Load existing history
	loadHistory()
	
	fmt.Printf("📚 Command history loaded from: %s (%d commands)\n", historyFile, len(commandHistory))
}

func getHistoryFilePath() string {
	// Try to get user's home directory
	usr, err := user.Current()
	if err != nil {
		// Fallback to current directory
		return ".sdl_history"
	}
	
	// Use ~/.sdl_history
	return filepath.Join(usr.HomeDir, ".sdl_history")
}

func loadHistory() {
	file, err := os.Open(historyFile)
	if err != nil {
		// File doesn't exist yet, start with empty history
		commandHistory = []string{}
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	history := []string{}
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			history = append(history, line)
		}
	}
	
	commandHistory = history
}

func saveHistory() {
	if historyFile == "" {
		return
	}
	
	// Limit history size to last 1000 commands
	maxHistory := 1000
	startIdx := 0
	if len(commandHistory) > maxHistory {
		startIdx = len(commandHistory) - maxHistory
	}
	
	file, err := os.Create(historyFile)
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not save command history: %v\n", err)
		return
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	for i := startIdx; i < len(commandHistory); i++ {
		writer.WriteString(commandHistory[i] + "\n")
	}
	writer.Flush()
	
	fmt.Printf("📚 Command history saved to: %s (%d commands)\n", historyFile, len(commandHistory)-startIdx)
}

func completer(d prompt.Document) []prompt.Suggest {
	// Get the current line and word
	line := d.CurrentLine()
	word := d.GetWordBeforeCursor()
	args := strings.Fields(line)
	
	// Don't show suggestions for empty input unless user explicitly pressed Tab
	if line == "" {
		return []prompt.Suggest{}
	}
	
	// Handle shell commands (prefixed with !)
	if strings.HasPrefix(line, "!") {
		return getShellCommandSuggestions(word, line)
	}
	
	// If we're at the beginning, suggest commands (only if there's some input)
	if len(args) <= 1 && word != "" {
		return getCommandSuggestions(word)
	}
	
	// Context-aware completions based on the command
	command := args[0]
	argIndex := len(args) - 1
	if strings.HasSuffix(line, " ") {
		argIndex++
	}
	
	switch command {
	case "load":
		if argIndex == 1 {
			return getFileSuggestions(word, ".sdl")
		}
	case "use":
		if argIndex == 1 {
			return getSystemSuggestions(word)
		}
	case "set":
		if argIndex == 1 {
			return getParameterPathSuggestions(word)
		} else if argIndex == 2 {
			// Could suggest common values based on parameter type
			return getValueSuggestions(args[1])
		}
	case "run":
		if argIndex == 1 {
			return []prompt.Suggest{
				{Text: "latest", Description: "Use latest measurements"},
				{Text: "baseline", Description: "Use baseline measurements"},
			}
		} else if argIndex == 2 {
			return getTargetSuggestions(word)
		} else if argIndex == 3 {
			return []prompt.Suggest{
				{Text: "100", Description: "Quick test"},
				{Text: "1000", Description: "Default runs"},
				{Text: "5000", Description: "Extended test"},
				{Text: "10000", Description: "Comprehensive test"},
			}
		}
	case "execute":
		if argIndex == 1 {
			return getFileSuggestions(word, ".txt")
		}
	}
	
	return []prompt.Suggest{}
}

func getCommandSuggestions(prefix string) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	for _, cmd := range commands {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        cmd.Name,
			Description: cmd.Description,
		})
	}
	
	// Add shell command suggestion
	suggestions = append(suggestions, prompt.Suggest{
		Text:        "!",
		Description: "Execute shell command (e.g., !ls, !git status)",
	})
	
	return prompt.FilterHasPrefix(suggestions, prefix, true)
}

func getShellCommandSuggestions(word string, line string) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	
	// Common shell commands with descriptions
	commonCommands := []struct {
		cmd  string
		desc string
	}{
		{"ls", "List directory contents"},
		{"pwd", "Print working directory"},
		{"cd", "Change directory"},
		{"cat", "Display file contents"},
		{"grep", "Search text patterns"},
		{"find", "Find files and directories"},
		{"git", "Git version control"},
		{"make", "Build using Makefile"},
		{"go", "Go programming language tools"},
		{"ps", "List running processes"},
		{"top", "Display running processes"},
		{"curl", "Transfer data from servers"},
		{"wget", "Download files"},
		{"docker", "Container management"},
		{"kubectl", "Kubernetes control"},
	}
	
	// If line is just "!", suggest common commands
	if strings.TrimSpace(line) == "!" {
		for _, cmd := range commonCommands {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        "!" + cmd.cmd,
				Description: cmd.desc,
			})
		}
		return suggestions
	}
	
	// For more complex shell commands, suggest file completion after the command
	// Extract the shell command part (after !)
	shellPart := strings.TrimSpace(line[1:])
	parts := strings.Fields(shellPart)
	
	if len(parts) >= 1 {
		// For commands that typically work with files, suggest file completion
		cmd := parts[0]
		if cmd == "cat" || cmd == "less" || cmd == "more" || cmd == "head" || cmd == "tail" || 
		   cmd == "cp" || cmd == "mv" || cmd == "rm" || cmd == "chmod" || cmd == "ls" {
			// Get current word for file completion
			currentWord := ""
			if strings.HasSuffix(line, " ") {
				currentWord = ""
			} else if len(parts) > 1 {
				currentWord = parts[len(parts)-1]
			}
			
			// Use file suggestions without extension filter
			fileSuggestions := getFileSuggestions(currentWord, "")
			for _, fs := range fileSuggestions {
				suggestions = append(suggestions, prompt.Suggest{
					Text:        "!" + shellPart[:len(shellPart)-len(currentWord)] + fs.Text,
					Description: fs.Description,
				})
			}
		}
	}
	
	return suggestions
}

func getFileSuggestions(prefix string, extension string) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	
	// Start from current directory or the directory in prefix
	searchDir := "."
	
	if strings.Contains(prefix, "/") {
		dir := filepath.Dir(prefix)
		searchDir = dir
	}
	
	// Read directory
	files, err := os.ReadDir(searchDir)
	if err != nil {
		return suggestions
	}
	
	for _, file := range files {
		name := file.Name()
		fullPath := filepath.Join(searchDir, name)
		if searchDir == "." {
			fullPath = name
		}
		
		// Include directories and files with the right extension
		if file.IsDir() {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        fullPath + "/",
				Description: "Directory",
			})
		} else if extension == "" || strings.HasSuffix(name, extension) {
			info, _ := file.Info()
			size := float64(0)
			if info != nil {
				size = float64(info.Size()) / 1024
			}
			suggestions = append(suggestions, prompt.Suggest{
				Text:        fullPath,
				Description: fmt.Sprintf("File (%.1fKB)", size),
			})
		}
	}
	
	// Also suggest common SDL example paths
	if extension == ".sdl" && strings.HasPrefix("examples/", prefix) {
		exampleDirs := []string{
			"examples/contacts/contacts.sdl",
			"examples/kafka/kafka.sdl",
			"examples/hotel/hotel.sdl",
		}
		for _, path := range exampleDirs {
			if strings.HasPrefix(path, prefix) {
				suggestions = append(suggestions, prompt.Suggest{
					Text:        path,
					Description: "Example SDL file",
				})
			}
		}
	}
	
	return prompt.FilterHasPrefix(suggestions, prefix, true)
}

func getSystemSuggestions(prefix string) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	
	if currentCanvas == nil {
		return suggestions
	}
	
	// Get system names directly from Canvas using the new public method
	systemNames := currentCanvas.GetAvailableSystemNames()
	
	for _, systemName := range systemNames {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        systemName,
			Description: "System definition from loaded SDL files",
		})
	}
	
	return prompt.FilterHasPrefix(suggestions, prefix, true)
}


func getParameterPathSuggestions(prefix string) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	
	// Common parameter paths
	paths := []struct {
		path string
		desc string
	}{
		{"server.pool.ArrivalRate", "Request arrival rate"},
		{"server.pool.Size", "Connection pool size"},
		{"server.db.pool.Size", "Database pool size"},
		{"server.db.CacheHitRate", "Cache hit ratio (0-1)"},
		{"server.db.QueryTimeout", "Query timeout in ms"},
		{"client.RequestTimeout", "Client request timeout"},
		{"client.RetryCount", "Number of retries"},
		{"cache.Size", "Cache size in entries"},
		{"cache.TTL", "Cache TTL in seconds"},
	}
	
	for _, p := range paths {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        p.path,
			Description: p.desc,
		})
	}
	
	return prompt.FilterHasPrefix(suggestions, prefix, true)
}

func getValueSuggestions(paramPath string) []prompt.Suggest {
	// Suggest common values based on parameter type
	if strings.Contains(paramPath, "Rate") {
		return []prompt.Suggest{
			{Text: "5", Description: "Low rate"},
			{Text: "10", Description: "Medium rate"},
			{Text: "25", Description: "High rate"},
			{Text: "50", Description: "Very high rate"},
		}
	} else if strings.Contains(paramPath, "Size") {
		return []prompt.Suggest{
			{Text: "5", Description: "Small"},
			{Text: "10", Description: "Medium"},
			{Text: "20", Description: "Large"},
			{Text: "50", Description: "Very large"},
		}
	} else if strings.Contains(paramPath, "CacheHitRate") {
		return []prompt.Suggest{
			{Text: "0.4", Description: "40% hit rate"},
			{Text: "0.6", Description: "60% hit rate"},
			{Text: "0.8", Description: "80% hit rate"},
			{Text: "0.95", Description: "95% hit rate"},
		}
	}
	return []prompt.Suggest{}
}

func getTargetSuggestions(prefix string) []prompt.Suggest {
	// Common targets in SDL systems
	targets := []prompt.Suggest{
		{Text: "server.HandleLookup", Description: "Lookup handler latency"},
		{Text: "server.HandleCreate", Description: "Create handler latency"},
		{Text: "server.HandleUpdate", Description: "Update handler latency"},
		{Text: "server.HandleDelete", Description: "Delete handler latency"},
		{Text: "db.Query", Description: "Database query latency"},
		{Text: "cache.Get", Description: "Cache get latency"},
		{Text: "cache.Set", Description: "Cache set latency"},
	}
	
	return prompt.FilterHasPrefix(targets, prefix, true)
}

func executor(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	
	// Add to history (avoid duplicates of the last command)
	if len(commandHistory) == 0 || commandHistory[len(commandHistory)-1] != line {
		commandHistory = append(commandHistory, line)
	}
	
	// Handle exit commands
	if line == "exit" || line == "quit" {
		fmt.Println("👋 Goodbye!")
		os.Exit(0)
	}
	
	// Handle shell commands (prefixed with !)
	if strings.HasPrefix(line, "!") {
		shellCmd := strings.TrimSpace(line[1:])
		if shellCmd == "" {
			fmt.Println("❌ Error: empty shell command")
			return
		}
		if err := executeShellCommand(shellCmd); err != nil {
			fmt.Printf("❌ Shell error: %v\n", err)
		}
		return
	}
	
	// Execute SDL command
	if err := executeCommand(currentCanvas, line); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	}
}

func executeShellCommand(cmd string) error {
	// Parse the command and arguments
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	
	// Create the command
	var shellCmd *exec.Cmd
	if len(parts) == 1 {
		shellCmd = exec.Command(parts[0])
	} else {
		shellCmd = exec.Command(parts[0], parts[1:]...)
	}
	
	// Set up command to use current stdio
	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr
	shellCmd.Stdin = os.Stdin
	
	// Run the command
	fmt.Printf("🐚 Running: %s\n", cmd)
	err := shellCmd.Run()
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	
	return nil
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
		valueStr := strings.Join(args[1:], " ") // Allow spaces in values
		
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
		fmt.Printf("✅ Simulation completed: %d runs of %s\n", runs, target)
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
		if len(state.SystemParameters) > 0 {
			fmt.Printf("  Modified Parameters:\n")
			for path, value := range state.SystemParameters {
				fmt.Printf("    %s = %v\n", path, value)
			}
		}
		return nil
		
	case "execute":
		if len(args) < 1 {
			return fmt.Errorf("usage: execute <recipe_file>")
		}
		return executeRecipe(canvas, args[0])
		
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
  execute <recipe_file>      Execute commands from a recipe file
  state                      Show current Canvas state
  !<shell_command>           Execute shell command (e.g., !ls, !git status)
  exit, quit                 Exit the console (or press Ctrl+D)

Navigation:
  ↑↓                         Navigate through command history (persistent across sessions)
  ←→                         Move cursor within line
  Tab                        Auto-complete commands, paths, and parameters
  Ctrl+A/E                   Jump to beginning/end of line
  Ctrl+K/U                   Delete to end/beginning of line
  Ctrl+W                     Delete word before cursor
  Ctrl+C                     Exit console (saves history)
  Ctrl+D                     Exit console

History:
  Commands are automatically saved to ~/.sdl_history
  Up to 1000 commands are preserved across console restarts

Examples:
  SDL> load examples/contacts/contacts.sdl
  SDL> use ContactsSystem
  SDL> set server.pool.ArrivalRate 15
  SDL> run latest server.HandleLookup 2000
  SDL> execute examples/demo_recipe.txt
  SDL> !ls -la
  SDL> !git status
  SDL> !ps aux | grep sdl

`)
}

func executeRecipe(canvas *console.Canvas, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open recipe file '%s': %w", filePath, err)
	}
	defer file.Close()
	
	fmt.Printf("🍳 Executing recipe: %s\n", filePath)
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Handle special commands
		if strings.HasPrefix(line, "sleep ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if duration, err := time.ParseDuration(parts[1]); err == nil {
					fmt.Printf("⏳ Sleeping for %s...\n", duration)
					time.Sleep(duration)
					continue
				}
			}
		}
		
		fmt.Printf("SDL[%d]> %s\n", lineNum, line)
		
		if err := executeCommand(canvas, line); err != nil {
			return fmt.Errorf("recipe failed at line %d: %w", lineNum, err)
		}
		
		// Small delay between commands for demo effect
		time.Sleep(100 * time.Millisecond)
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading recipe file: %w", err)
	}
	
	fmt.Printf("✅ Recipe completed successfully\n")
	return nil
}

func init() {
	consoleCmd.Flags().IntVar(&consolePort, "port", 8080, "Port to serve the web interface on")
	rootCmd.AddCommand(consoleCmd)
}