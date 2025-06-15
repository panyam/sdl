package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Measurement management commands

var measureCmd = &cobra.Command{
	Use:   "measure",
	Short: "Manage performance measurements",
	Long:  "Create, monitor, and analyze performance measurements",
}

var measureAddCmd = &cobra.Command{
	Use:   "add [id] [target] [type]",
	Short: "Create a new measurement",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		target := args[1]
		metricType := args[2]

		// Validate metric type
		validMetrics := map[string]bool{
			"latency":    true,
			"throughput": true,
			"error_rate": true,
		}

		if !validMetrics[metricType] {
			fmt.Printf("❌ Invalid metric type '%s'. Valid types: latency, throughput, error_rate\n", metricType)
			return
		}

		_, err := makeAPICall("POST", "/api/canvas/measurements", map[string]interface{}{
			"id":         id,
			"name":       fmt.Sprintf("Measurement-%s", id),
			"target":     target,
			"metricType": metricType,
			"enabled":    true,
		})
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		fmt.Printf("✅ Measurement '%s' created\n", id)
		fmt.Printf("🎯 Target: %s\n", target)
		fmt.Printf("📊 Metric: %s\n", metricType)
		fmt.Printf("💾 Storage: DuckDB time-series database\n")
	},
}

var measureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all measurements",
	Run: func(cmd *cobra.Command, args []string) {
		result, err := makeAPICall("GET", "/api/canvas/measurements", nil)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		measurements, ok := result["data"].(map[string]interface{})
		if !ok || len(measurements) == 0 {
			fmt.Println("Active Measurements:")
			fmt.Println("┌─────────────┬─────────────────────┬─────────────┬────────────┐")
			fmt.Println("│ Name        │ Target              │ Type        │ Data Points│")
			fmt.Println("├─────────────┼─────────────────────┼─────────────┼────────────┤")
			fmt.Println("│ (none)      │                     │             │            │")
			fmt.Println("└─────────────┴─────────────────────┴─────────────┴────────────┘")
			return
		}

		fmt.Println("Active Measurements:")
		fmt.Println("┌─────────────┬─────────────────────┬─────────────┬────────────┐")
		fmt.Println("│ Name        │ Target              │ Type        │ Data Points│")
		fmt.Println("├─────────────┼─────────────────────┼─────────────┼────────────┤")
		
		for id, measData := range measurements {
			meas := measData.(map[string]interface{})
			target := meas["target"].(string)
			metricType := meas["metricType"].(string)
			dataPoints := int(meas["dataPoints"].(float64))
			
			fmt.Printf("│ %-11s │ %-19s │ %-11s │ %10d │\n", id, target, metricType, dataPoints)
		}
		fmt.Println("└─────────────┴─────────────────────┴─────────────┴────────────┘")
	},
}

var measureDataCmd = &cobra.Command{
	Use:   "data [id]",
	Short: "View measurement data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		
		result, err := makeAPICall("GET", fmt.Sprintf("/api/measurements/%s/data", id), nil)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		data, ok := result["data"].([]interface{})
		if !ok || len(data) == 0 {
			fmt.Printf("📊 No data points available for measurement '%s'\n", id)
			return
		}

		fmt.Printf("Recent data for '%s' (last %d points):\n", id, len(data))
		fmt.Println("┌─────────────────────┬─────────────────────┬─────────┐")
		fmt.Println("│ Timestamp           │ Target              │ Value   │")
		fmt.Println("├─────────────────────┼─────────────────────┼─────────┤")
		
		for _, pointData := range data {
			point := pointData.(map[string]interface{})
			timestamp := point["timestamp"].(string)
			target := point["target"].(string)
			value := point["value"].(float64)
			
			fmt.Printf("│ %-19s │ %-19s │ %7.1f │\n", timestamp, target, value)
		}
		fmt.Println("└─────────────────────┴─────────────────────┴─────────┘")
	},
}

var measureStatsCmd = &cobra.Command{
	Use:   "stats [id]",
	Short: "Show measurement statistics",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		
		result, err := makeAPICall("GET", fmt.Sprintf("/api/canvas/measurements/%s", id), nil)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		stats := result["stats"].(map[string]interface{})
		
		fmt.Printf("Statistics for '%s':\n", id)
		fmt.Printf("📊 Total Samples: %.0f\n", stats["totalSamples"].(float64))
		fmt.Printf("⏱️  Average: %.1fms\n", stats["average"].(float64))
		fmt.Printf("📈 95th Percentile: %.1fms\n", stats["p95"].(float64))
		fmt.Printf("📉 Min: %.1fms\n", stats["min"].(float64))
		fmt.Printf("📊 Max: %.1fms\n", stats["max"].(float64))
	},
}

var measureRemoveCmd = &cobra.Command{
	Use:   "remove [id...]",
	Short: "Remove measurements",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("❌ Must specify measurement ID(s) to remove")
			return
		}
		
		for _, id := range args {
			_, err := makeAPICall("DELETE", fmt.Sprintf("/api/canvas/measurements/%s", id), nil)
			if err != nil {
				fmt.Printf("❌ Error removing '%s': %v\n", id, err)
				continue
			}
			fmt.Printf("✅ Measurement '%s' removed\n", id)
		}
	},
}

func init() {
	// Add subcommands to measure
	measureCmd.AddCommand(measureAddCmd)
	measureCmd.AddCommand(measureListCmd)
	measureCmd.AddCommand(measureDataCmd)
	measureCmd.AddCommand(measureStatsCmd)
	measureCmd.AddCommand(measureRemoveCmd)
	
	// Add to root command (server flag is now persistent on root command)
	rootCmd.AddCommand(measureCmd)
}