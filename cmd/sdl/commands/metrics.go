package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Manage and query metrics",
	Long:  `Commands for creating, managing, and querying metrics from the MetricStore.`,
}

var addMetricCmd = &cobra.Command{
	Use:   "add <id> <component>.<method>",
	Short: "Add a new metric",
	Long: `Add a new metric to collect data for specific component methods.
	
Examples:
  # Track latency for server.Lookup
  sdl metrics add server_latency server.Lookup --type latency
  
  # Track count for multiple methods
  sdl metrics add db_calls database Query Update Insert --type count`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		component := args[1]
		methods := strings.Split(args[2], ",")

		metricType, _ := cmd.Flags().GetString("type")
		aggregation, _ := cmd.Flags().GetString("aggregation")
		window, _ := cmd.Flags().GetFloat64("window")

		// Validate metric type
		if metricType != "count" && metricType != "latency" {
			fmt.Fprintf(os.Stderr, "Error: Invalid metric type '%s'. Must be 'count' or 'latency'\n", metricType)
			os.Exit(1)
		}

		err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			req := &v1.AddMetricRequest{
				Metric: &v1.Metric{
					Id:                id,
					CanvasId:          canvasID,
					Name:              fmt.Sprintf("%s.(%s)", component, strings.Join(methods, ",")),
					Component:         component,
					Methods:           methods,
					MetricType:        metricType,
					Aggregation:       aggregation,
					AggregationWindow: window,
					Enabled:           true,
				},
			}

			_, err := client.AddMetric(ctx, req)
			return err
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Added metric '%s' for %s.(%s) (%s)\n", id, component, strings.Join(methods, ","), metricType)
	},
}

var removeMetricCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove a metric",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		metricID := args[0]

		err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			req := &v1.DeleteMetricRequest{
				CanvasId: canvasID,
				MetricId: metricID,
			}

			_, err := client.DeleteMetric(ctx, req)
			return err
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Removed metric '%s'\n", metricID)
	},
}

var listMetricsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available metrics",
	Run: func(cmd *cobra.Command, args []string) {
		err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			req := &v1.ListMetricsRequest{
				CanvasId: canvasID,
			}

			resp, err := client.ListMetrics(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to list metrics: %v", err)
			}

			if len(resp.Metrics) == 0 {
				fmt.Println("No metrics available")
				fmt.Println("\nStart some generators to collect metrics:")
				fmt.Println("  sdl gen add test server.Lookup 10")
				fmt.Println("  sdl gen start test")
				return nil
			}

			// Display metrics in a table
			fmt.Printf("%-40s %-20s %-15s %10s %15s %15s\n",
				"ID", "Target", "Type", "Points", "Oldest", "Newest")
			fmt.Println(strings.Repeat("-", 115))

			for _, m := range resp.Metrics {
				oldestTime := "-"
				newestTime := "-"
				if m.NumDataPoints > 0 {
					oldestTime = time.Unix(int64(m.OldestTimestamp), 0).Format("15:04:05")
					newestTime = time.Unix(int64(m.NewestTimestamp), 0).Format("15:04:05")
				}
				fmt.Printf("%-40s %-20s %-15s %10d %15s %15s\n",
					m.Id, m.Component+".("+strings.Join(m.Methods, ",")+")", m.MetricType, m.NumDataPoints, oldestTime, newestTime)
			}

			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var queryMetricsCmd = &cobra.Command{
	Use:   "query <metric-id>",
	Short: "Query metric data points",
	Long:  "Query metric data points. The data is already aggregated according to the metric's configuration.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		metricID := args[0]

		// Get time range flags
		duration, _ := cmd.Flags().GetDuration("duration")
		limit, _ := cmd.Flags().GetInt32("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		endTime := time.Now()
		startTime := endTime.Add(-duration)

		err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			req := &v1.QueryMetricsRequest{
				CanvasId:  canvasID,
				MetricId:  metricID,
				StartTime: float64(startTime.Unix()),
				EndTime:   float64(endTime.Unix()),
				Limit:     limit,
			}

			resp, err := client.QueryMetrics(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to query metrics: %v", err)
			}

			if outputJSON {
				// Output as JSON
				data, err := json.MarshalIndent(resp.Points, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}
				fmt.Println(string(data))
			} else {
				// Human-readable output
				fmt.Printf("Metric: %s\n", metricID)
				fmt.Printf("Time Range: %s to %s\n", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
				fmt.Printf("Data Points: %d\n", len(resp.Points))

				// Note about pre-aggregated data
				fmt.Println("\nNote: Values are pre-aggregated according to the metric's configuration")
				fmt.Println("(e.g., if metric is configured for p95 with 10s windows, each point is a p95 value)")
				fmt.Println("")

				if len(resp.Points) > 0 {
					fmt.Printf("%-30s %15s\n", "Timestamp", "Value")
					fmt.Println(strings.Repeat("-", 47))
					for _, p := range resp.Points {
						ts := time.Unix(int64(p.Timestamp), 0)
						fmt.Printf("%-30s %15.3f\n", ts.Format(time.RFC3339), p.Value)
					}
				}
			}

			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Add subcommands
	metricsCmd.AddCommand(addMetricCmd)
	metricsCmd.AddCommand(removeMetricCmd)
	metricsCmd.AddCommand(listMetricsCmd)
	metricsCmd.AddCommand(queryMetricsCmd)

	// Add metric command flags
	addMetricCmd.Flags().String("type", "latency", "Metric type: 'count' or 'latency'")
	addMetricCmd.Flags().String("aggregation", "avg", "Aggregation function (e.g., sum, avg, p95)")
	addMetricCmd.Flags().Float64("window", 10.0, "Aggregation window in seconds")

	// Query command flags
	queryMetricsCmd.Flags().Duration("duration", 5*time.Minute, "Time duration to query (e.g., 5m, 1h)")
	queryMetricsCmd.Flags().Int32("limit", 100, "Maximum number of points to return")
	queryMetricsCmd.Flags().Bool("json", false, "Output as JSON")

	// Add to root
	AddCommand(metricsCmd)
}
