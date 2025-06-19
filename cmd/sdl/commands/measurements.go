package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Measurement management commands

var measureCmd = &cobra.Command{
	Use:   "measure",
	Short: "Manage performance measurements",
	Long:  "Create, monitor, and analyze performance measurements",
}

var (
	measureAggregation string
	measureWindow      string
	measureResultValue string
)

var measureAddCmd = &cobra.Command{
	Use:   "add [id] [component.method] [metric]",
	Short: "Create a new measurement",
	Long: `Create a new measurement to track performance metrics.

Metric types:
  count    - Number of events (for throughput, error counts)
  latency  - Duration measurements (for response times)

Aggregation types:
  For count:   sum, rate
  For latency: avg, min, max, p50, p90, p95, p99

Examples:
  # Throughput measurement
  sdl measure add server_qps server.Lookup count --aggregation rate

  # Latency measurement
  sdl measure add server_p95 server.Lookup latency --aggregation p95

  # Error rate measurement
  sdl measure add server_errors server.Lookup count --aggregation rate --result-value "Val(Bool: false)"`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		target := args[1]
		metricType := args[2]

		// Parse target into component and method
		parts := strings.Split(target, ".")
		if len(parts) != 2 {
			fmt.Printf("❌ Invalid target format '%s'. Expected: component.method\n", target)
			return
		}
		component := parts[0]
		method := parts[1]

		// Validate metric type
		validMetrics := map[string]bool{
			"count":   true,
			"latency": true,
		}

		if !validMetrics[metricType] {
			fmt.Printf("❌ Invalid metric type '%s'. Valid types: count, latency\n", metricType)
			return
		}

		// Set defaults based on metric type
		if measureAggregation == "" {
			if metricType == "count" {
				measureAggregation = "rate"
			} else {
				measureAggregation = "p95"
			}
		}

		// Default window if not specified
		if measureWindow == "" {
			measureWindow = "10s"
		}

		// Default result value if not specified
		if measureResultValue == "" {
			measureResultValue = "*"
		}

		body := map[string]any{
			"id":          id,
			"name":        fmt.Sprintf("%s %s", strings.Title(component), strings.Title(method)),
			"component":   component,
			"methods":     []string{method},
			"resultValue": measureResultValue,
			"metric":      metricType,
			"aggregation": measureAggregation,
			"window":      measureWindow,
		}

		result, err := makeAPICall[map[string]any]("POST", "/api/measurements", body)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		// Parse response to get created measurement details
		if data, ok := result["data"].(string); ok {
			var measurement map[string]any
			if err := json.Unmarshal([]byte(data), &measurement); err == nil {
				fmt.Printf("✅ Measurement '%s' created\n", id)
				fmt.Printf("🎯 Target: %s.%s\n", component, method)
				fmt.Printf("📊 Metric: %s (%s)\n", metricType, measureAggregation)
				fmt.Printf("⏱️  Window: %s\n", measureWindow)
				if measureResultValue != "*" {
					fmt.Printf("🔍 Filter: %s\n", measureResultValue)
				}
				return
			}
		}

		fmt.Printf("✅ Measurement '%s' created\n", id)
		fmt.Printf("🎯 Target: %s.%s\n", component, method)
		fmt.Printf("📊 Metric: %s (%s)\n", metricType, measureAggregation)
	},
}

var measureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all measurements",
	Run: func(cmd *cobra.Command, args []string) {
		result, err := makeAPICall[[]any]("GET", "/api/measurements", nil)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		// Parse the JSON response
		var measurements []map[string]any
		measList := result
		// Handle direct array response
		for _, m := range measList {
			if meas, ok := m.(map[string]any); ok {
				measurements = append(measurements, meas)
			}
		}

		if len(measurements) == 0 {
			fmt.Println("No active measurements")
			fmt.Println("\nUse 'sdl measure add' to create a measurement:")
			fmt.Println("  sdl measure add server_qps server.Lookup count")
			fmt.Println("  sdl measure add server_p95 server.Lookup latency --aggregation p95")
			return
		}

		fmt.Println("Active Measurements:")
		fmt.Println("┌────────────────┬──────────────────────┬─────────────┬────────────┬────────────┐")
		fmt.Println("│ ID             │ Target               │ Metric      │ Aggregation│ Points     │")
		fmt.Println("├────────────────┼──────────────────────┼─────────────┼────────────┼────────────┤")

		for _, meas := range measurements {
			id := meas["id"].(string)
			component := meas["component"].(string)
			methods := meas["methods"].([]any)
			metric := meas["metric"].(string)
			aggregation := meas["aggregation"].(string)
			pointCount := 0
			if pc, ok := meas["pointCount"].(float64); ok {
				pointCount = int(pc)
			}

			// Build target string
			methodList := make([]string, len(methods))
			for i, m := range methods {
				methodList[i] = m.(string)
			}
			target := fmt.Sprintf("%s.%s", component, strings.Join(methodList, ","))

			// Truncate if too long
			if len(target) > 20 {
				target = target[:17] + "..."
			}
			if len(id) > 14 {
				id = id[:11] + "..."
			}

			fmt.Printf("│ %-14s │ %-20s │ %-11s │ %-10s │ %10d │\n",
				id, target, metric, aggregation, pointCount)
		}
		fmt.Println("└────────────────┴──────────────────────┴─────────────┴────────────┴────────────┘")
	},
}

var measureDataCmd = &cobra.Command{
	Use:   "data [id]",
	Short: "View measurement data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		result, err := makeAPICall[map[string]any]("GET", fmt.Sprintf("/api/measurements/%s/data?limit=20", id), nil)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		// Parse the JSON response
		var dataPoints []map[string]any
		if data, ok := result["data"].(string); ok {
			if err := json.Unmarshal([]byte(data), &dataPoints); err != nil {
				fmt.Printf("❌ Error parsing response: %v\n", err)
				return
			}
		} else if pointList, ok := result["data"].([]any); ok {
			for _, p := range pointList {
				if point, ok := p.(map[string]any); ok {
					dataPoints = append(dataPoints, point)
				}
			}
		}

		if len(dataPoints) == 0 {
			fmt.Printf("📊 No data points available for measurement '%s'\n", id)
			fmt.Println("\nGenerate some traffic first:")
			fmt.Println("  sdl run results server.Lookup --runs 100")
			fmt.Println("  sdl gen add test server.Lookup 10")
			return
		}

		// Get measurement info to determine metric type
		infoResult, _ := makeAPICall[map[string]any]("GET", fmt.Sprintf("/api/measurements/%s", id), nil)
		metricType := "count"
		if infoData, ok := infoResult["data"].(string); ok {
			var info map[string]any
			if err := json.Unmarshal([]byte(infoData), &info); err == nil {
				if mt, ok := info["metric"].(string); ok {
					metricType = mt
				}
			}
		}

		fmt.Printf("Recent data for '%s' (last %d points):\n", id, len(dataPoints))
		fmt.Println("┌────────────────┬──────────────────────┬──────────────┐")
		if metricType == "latency" {
			fmt.Println("│ Time (sec)     │ Component.Method     │ Latency (ms) │")
		} else {
			fmt.Println("│ Time (sec)     │ Component.Method     │ Count        │")
		}
		fmt.Println("├────────────────┼──────────────────────┼──────────────┤")

		for _, point := range dataPoints {
			timestamp := point["timestamp"].(float64)
			component := point["component"].(string)
			method := point["method"].(string)
			value := point["value"].(float64)

			target := fmt.Sprintf("%s.%s", component, method)
			if len(target) > 20 {
				target = target[:17] + "..."
			}

			if metricType == "latency" {
				// Convert nanoseconds to milliseconds
				fmt.Printf("│ %14.3f │ %-20s │ %12.3f │\n", timestamp, target, value/1e6)
			} else {
				fmt.Printf("│ %14.3f │ %-20s │ %12.0f │\n", timestamp, target, value)
			}
		}
		fmt.Println("└────────────────┴──────────────────────┴──────────────┘")
	},
}

var measureStatsCmd = &cobra.Command{
	Use:   "stats [id]",
	Short: "Show measurement statistics",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		// Get aggregated data
		result, err := makeAPICall[map[string]any]("GET", fmt.Sprintf("/api/measurements/%s/aggregated", id), nil)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		// Parse response
		var aggData map[string]any
		if data, ok := result["data"].(string); ok {
			if err := json.Unmarshal([]byte(data), &aggData); err != nil {
				fmt.Printf("❌ Error parsing response: %v\n", err)
				return
			}
		} else if agg, ok := result["data"].(map[string]any); ok {
			aggData = agg
		}

		// Get measurement info for context
		infoResult, _ := makeAPICall[map[string]any]("GET", fmt.Sprintf("/api/measurements/%s", id), nil)
		var measurementInfo map[string]any
		if infoData, ok := infoResult["data"].(string); ok {
			json.Unmarshal([]byte(infoData), &measurementInfo)
		}

		fmt.Printf("\nStatistics for '%s':\n", id)
		fmt.Println("─────────────────────────────────────")

		if measurementInfo != nil {
			component := measurementInfo["component"].(string)
			methods := measurementInfo["methods"].([]any)
			metric := measurementInfo["metric"].(string)
			aggregation := measurementInfo["aggregation"].(string)

			methodList := make([]string, len(methods))
			for i, m := range methods {
				methodList[i] = m.(string)
			}

			fmt.Printf("🎯 Target: %s.%s\n", component, strings.Join(methodList, ","))
			fmt.Printf("📊 Metric: %s\n", metric)
			fmt.Printf("📈 Aggregation: %s\n", aggregation)
		}

		if aggData != nil {
			value := aggData["value"].(float64)
			unit := aggData["unit"].(string)
			window := aggData["window"].(string)
			pointCount := int(aggData["pointCount"].(float64))

			fmt.Printf("⏱️  Window: %s\n", window)
			fmt.Printf("📊 Data Points: %d\n", pointCount)
			fmt.Printf("💯 Value: %.3f %s\n", value, unit)

			if lastUpdate, ok := aggData["lastUpdate"].(string); ok && lastUpdate != "" {
				fmt.Printf("🕐 Last Update: %s\n", lastUpdate)
			}
		}

		fmt.Println("─────────────────────────────────────")
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
			_, err := makeAPICall[any]("DELETE", fmt.Sprintf("/api/measurements/%s", id), nil)
			if err != nil {
				fmt.Printf("❌ Error removing '%s': %v\n", id, err)
				continue
			}
			fmt.Printf("✅ Measurement '%s' removed\n", id)
		}
	},
}

var measureClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove all measurements",
	Run: func(cmd *cobra.Command, args []string) {
		// First get the list of measurements
		result, err := makeAPICall[map[string]any]("GET", "/api/measurements", nil)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		var measurements []map[string]any
		if data, ok := result["data"].(string); ok {
			json.Unmarshal([]byte(data), &measurements)
		} else if measList, ok := result["data"].([]any); ok {
			for _, m := range measList {
				if meas, ok := m.(map[string]any); ok {
					measurements = append(measurements, meas)
				}
			}
		}

		if len(measurements) == 0 {
			fmt.Println("No measurements to remove")
			return
		}

		// Remove each measurement
		for _, meas := range measurements {
			id := meas["id"].(string)
			_, err := makeAPICall[any]("DELETE", fmt.Sprintf("/api/measurements/%s", id), nil)
			if err != nil {
				fmt.Printf("❌ Error removing '%s': %v\n", id, err)
				continue
			}
			fmt.Printf("✅ Removed '%s'\n", id)
		}

		fmt.Printf("\n🧹 Cleared %d measurements\n", len(measurements))
	},
}

func init() {
	// Add flags to measure add command
	measureAddCmd.Flags().StringVar(&measureAggregation, "aggregation", "", "Aggregation type (e.g., rate, p95, avg)")
	measureAddCmd.Flags().StringVar(&measureWindow, "window", "10s", "Time window for aggregation")
	measureAddCmd.Flags().StringVar(&measureResultValue, "result-value", "*", "Filter by result value")

	// Add subcommands to measure
	measureCmd.AddCommand(measureAddCmd)
	measureCmd.AddCommand(measureListCmd)
	measureCmd.AddCommand(measureDataCmd)
	measureCmd.AddCommand(measureStatsCmd)
	measureCmd.AddCommand(measureRemoveCmd)
	measureCmd.AddCommand(measureClearCmd)

	// Add to root command
	rootCmd.AddCommand(measureCmd)
}
