package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// Generator management commands

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Manage traffic generators",
	Long:  "Create, control, and monitor traffic generators for load testing",
}

var genAddCmd = &cobra.Command{
	Use:   "add [id] [target] [rate]",
	Short: "Create a new traffic generator",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		target := args[1]
		rateStr := args[2]

		rate, err := strconv.Atoi(rateStr)
		if err != nil {
			fmt.Printf("❌ Invalid rate '%s': must be a number\n", rateStr)
			return
		}

		_, err = makeAPICall("POST", "/api/canvas/generators", map[string]interface{}{
			"id":      id,
			"name":    fmt.Sprintf("Generator-%s", id),
			"target":  target,
			"rate":    rate,
			"enabled": false,
		})
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		fmt.Printf("✅ Generator '%s' created\n", id)
		fmt.Printf("🎯 Target: %s\n", target)
		fmt.Printf("⚡ Rate: %d calls/second\n", rate)
		fmt.Printf("🔄 Status: Stopped\n")
	},
}

var genListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all traffic generators",
	Run: func(cmd *cobra.Command, args []string) {
		result, err := makeAPICall("GET", "/api/canvas/generators", nil)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		generators, ok := result["data"].(map[string]interface{})
		if !ok || len(generators) == 0 {
			fmt.Println("Active Traffic Generators:")
			fmt.Println("┌─────────────┬─────────────────────┬──────┬─────────┐")
			fmt.Println("│ Name        │ Target              │ Rate │ Status  │")
			fmt.Println("├─────────────┼─────────────────────┼──────┼─────────┤")
			fmt.Println("│ (none)      │                     │      │         │")
			fmt.Println("└─────────────┴─────────────────────┴──────┴─────────┘")
			return
		}

		fmt.Println("Active Traffic Generators:")
		fmt.Println("┌─────────────┬─────────────────────┬──────┬─────────┐")
		fmt.Println("│ Name        │ Target              │ Rate │ Status  │")
		fmt.Println("├─────────────┼─────────────────────┼──────┼─────────┤")
		
		for id, genData := range generators {
			gen := genData.(map[string]interface{})
			target := gen["target"].(string)
			rate := gen["rate"].(float64)
			enabled := gen["enabled"].(bool)
			
			status := "Stopped"
			if enabled {
				status = "Running"
			}
			
			fmt.Printf("│ %-11s │ %-19s │ %4.0f │ %-7s │\n", id, target, rate, status)
		}
		fmt.Println("└─────────────┴─────────────────────┴──────┴─────────┘")
	},
}

var genStartCmd = &cobra.Command{
	Use:   "start [id...]",
	Short: "Start traffic generators",
	Long:  "Start specific generators by ID, or all generators if no IDs provided",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Start all generators
			_, err := makeAPICall("POST", "/api/canvas/generators/start", nil)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}
			fmt.Println("✅ All generators started")
		} else {
			// Start specific generators
			for _, id := range args {
				_, err := makeAPICall("POST", fmt.Sprintf("/api/canvas/generators/%s/resume", id), nil)
				if err != nil {
					fmt.Printf("❌ Error starting '%s': %v\n", id, err)
					continue
				}
				fmt.Printf("✅ Generator '%s' started\n", id)
			}
		}
	},
}

var genStopCmd = &cobra.Command{
	Use:   "stop [id...]",
	Short: "Stop traffic generators",
	Long:  "Stop specific generators by ID, or all generators if no IDs provided",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Stop all generators
			_, err := makeAPICall("POST", "/api/canvas/generators/stop", nil)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}
			fmt.Println("✅ All generators stopped")
			fmt.Println("🛑 All traffic generation halted")
		} else {
			// Stop specific generators
			for _, id := range args {
				_, err := makeAPICall("POST", fmt.Sprintf("/api/canvas/generators/%s/pause", id), nil)
				if err != nil {
					fmt.Printf("❌ Error stopping '%s': %v\n", id, err)
					continue
				}
				fmt.Printf("✅ Generator '%s' stopped\n", id)
			}
		}
	},
}

var genStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show generator status",
	Run: func(cmd *cobra.Command, args []string) {
		result, err := makeAPICall("GET", "/api/canvas/generators", nil)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		generators, ok := result["data"].(map[string]interface{})
		if !ok || len(generators) == 0 {
			fmt.Println("Generator Status:")
			fmt.Println("📊 Total Generators: 0")
			return
		}

		runningCount := 0
		for _, genData := range generators {
			gen := genData.(map[string]interface{})
			if gen["enabled"].(bool) {
				runningCount++
			}
		}

		fmt.Println("Generator Status (Live):")
		fmt.Println("┌─────────────┬─────────────────────┬──────┬─────────┬───────────┐")
		fmt.Println("│ Name        │ Target              │ Rate │ Status  │ Uptime    │")
		fmt.Println("├─────────────┼─────────────────────┼──────┼─────────┼───────────┤")
		
		for id, genData := range generators {
			gen := genData.(map[string]interface{})
			target := gen["target"].(string)
			rate := gen["rate"].(float64)
			enabled := gen["enabled"].(bool)
			
			status := "Stopped"
			uptime := "--"
			if enabled {
				status = "Running"
				uptime = "00:00:00" // Placeholder - would need actual uptime from server
			}
			
			fmt.Printf("│ %-11s │ %-19s │ %4.0f │ %-7s │ %-9s │\n", id, target, rate, status, uptime)
		}
		fmt.Println("└─────────────┴─────────────────────┴──────┴─────────┴───────────┘")
		fmt.Printf("Total Active Load: %d generators running\n", runningCount)
	},
}

var genSetCmd = &cobra.Command{
	Use:   "set [id] [property] [value]",
	Short: "Modify generator properties",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		property := args[1]
		value := args[2]
		
		_, err := makeAPICall("PUT", fmt.Sprintf("/api/canvas/generators/%s", id), map[string]interface{}{
			property: value,
		})
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		fmt.Printf("✅ Generator '%s' %s updated to %s\n", id, property, value)
	},
}

var genRemoveCmd = &cobra.Command{
	Use:   "remove [id...]",
	Short: "Remove traffic generators",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("❌ Must specify generator ID(s) to remove")
			return
		}
		
		for _, id := range args {
			_, err := makeAPICall("DELETE", fmt.Sprintf("/api/canvas/generators/%s", id), nil)
			if err != nil {
				fmt.Printf("❌ Error removing '%s': %v\n", id, err)
				continue
			}
			fmt.Printf("✅ Generator '%s' removed\n", id)
		}
	},
}

func init() {
	// Add subcommands to gen
	genCmd.AddCommand(genAddCmd)
	genCmd.AddCommand(genListCmd)
	genCmd.AddCommand(genStartCmd)
	genCmd.AddCommand(genStopCmd)
	genCmd.AddCommand(genStatusCmd)
	genCmd.AddCommand(genSetCmd)
	genCmd.AddCommand(genRemoveCmd)
	
	// Add to root command (server flag is now persistent on root command)
	rootCmd.AddCommand(genCmd)
}