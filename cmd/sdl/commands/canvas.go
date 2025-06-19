package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// Canvas state management commands use shared API client from api.go

// Canvas state management commands

var loadCmd = &cobra.Command{
	Use:   "load [file]",
	Short: "Load an SDL file into the server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkServerConnection(); err != nil {
			return
		}

		_, err := makeAPICall[any]("POST", "/api/console/load", map[string]any{"filePath": args[0]})
		if err == nil {
			fmt.Printf("✅ Loaded %s successfully\n", args[0])
		}
	},
}

var useCmd = &cobra.Command{
	Use:   "use [system]",
	Short: "Select the active system",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := makeAPICall[any]("POST", "/api/console/use", map[string]any{"systemName": args[0]})
		if err == nil {
			fmt.Printf("✅ Now using system: %s\n", args[0])
		}
	},
}

var setCmd = &cobra.Command{
	Use:   "set [parameter] [value]",
	Short: "Set a parameter value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := makeAPICall[any]("POST", "/api/console/set", map[string]any{
			"parameter": args[0],
			"value":     args[1],
		})
		if err == nil {
			fmt.Printf("✅ Set %s = %s\n", args[0], args[1])
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get [parameter]",
	Short: "View parameter values",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := makeAPICall[any]("GET", "/api/console/state", nil)
		if err != nil {
			return
		}

		state := result.(map[string]any)["state"].(map[string]any)
		params := state["systemParameters"].(map[string]any)

		if len(args) == 0 {
			// Show all parameters
			if len(params) == 0 {
				fmt.Println("No parameters set")
				return
			}
			for key, value := range params {
				fmt.Printf("%s = %v\n", key, value)
			}
		} else {
			// Show specific parameter
			if value, exists := params[args[0]]; exists {
				fmt.Printf("%s = %v\n", args[0], value)
			} else {
				fmt.Printf("Parameter '%s' not found\n", args[0])
			}
		}
	},
}

var runCanvasCmd = &cobra.Command{
	Use:   "run [name] [method] [calls]",
	Short: "Run a simulation",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		calls, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Printf("❌ Invalid call count '%s': must be a number\n", args[2])
			return
		}

		_, err = makeAPICall[any]("POST", "/api/console/run", map[string]any{
			"name":   args[0],
			"method": args[1],
			"calls":  calls,
		})
		if err == nil {
			fmt.Printf("✅ Running %s: %s (%d calls)\n", args[0], args[1], calls)
		}
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current canvas state",
	Run: func(cmd *cobra.Command, args []string) {
		result, err := makeAPICall[any]("GET", "/api/console/state", nil)
		if err != nil {
			return
		}

		state := result.(map[string]any)["state"].(map[string]any)

		fmt.Printf("SDL Canvas State:\n")
		if activeFile := state["activeFile"]; activeFile != nil && activeFile != "" {
			fmt.Printf("📁 Active File: %s\n", activeFile)
		}
		if activeSystem := state["activeSystem"]; activeSystem != nil && activeSystem != "" {
			fmt.Printf("🎯 Active System: %s\n", activeSystem)
		}

		// Show generators
		if generators := state["generators"]; generators != nil {
			genMap := generators.(map[string]any)
			if len(genMap) > 0 {
				fmt.Printf("⚡ Generators: %d\n", len(genMap))
			}
		}

		// Show measurements
		if measurements := state["measurements"]; measurements != nil {
			measMap := measurements.(map[string]any)
			if len(measMap) > 0 {
				fmt.Printf("📊 Measurements: %d\n", len(measMap))
			}
		}
	},
}

var executeCmd = &cobra.Command{
	Use:   "execute [recipe-file]",
	Short: "Execute a recipe file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := makeAPICall[any]("POST", "/api/console/execute", map[string]any{"filePath": args[0]})
		if err == nil {
			fmt.Printf("✅ Executed recipe: %s\n", args[0])
		}
	},
}

// HTTP client is provided by api.go

func init() {
	// Add commands to root (server flag is now persistent on root command)
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(runCanvasCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(executeCmd)
}
