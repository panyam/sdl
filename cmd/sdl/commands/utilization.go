package commands

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	v1s "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/spf13/cobra"
)

var utilizationCmd = &cobra.Command{
	Use:   "utilization [component]",
	Short: "Show resource utilization",
	Long: `Display resource utilization for components in the system.
	
Examples:
  # Show utilization for all components
  sdl utilization
  
  # Show utilization for a specific component
  sdl utilization database
  
  # Show utilization for multiple components
  sdl utilization database webserver`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		threshold, _ := cmd.Flags().GetFloat64("threshold")
		showUtilization(args, jsonOutput, threshold)
	},
}

func init() {
	// Add flags
	utilizationCmd.Flags().Bool("json", false, "Output as JSON")
	utilizationCmd.Flags().Float64("threshold", 0.0, "Only show resources above this utilization threshold")

	// Add to root
	AddCommand(utilizationCmd)
}

func showUtilization(components []string, jsonOutput bool, threshold float64) {

	err := withCanvasClient(func(client v1s.CanvasServiceClient, ctx context.Context) error {
		req := &v1.GetUtilizationRequest{
			CanvasId:   canvasID,
			Components: components,
		}

		resp, err := client.GetUtilization(ctx, req)
		if err != nil {
			return err
		}

		if jsonOutput {
			// JSON output would go here
			fmt.Println("JSON output not yet implemented")
			return nil
		}

		// Display utilization in a table
		displayUtilizationTable(resp.Utilizations, threshold)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}

func displayUtilizationTable(utilizations []*v1.UtilizationInfo, threshold float64) {
	// Filter by threshold
	var filtered []*v1.UtilizationInfo
	for _, u := range utilizations {
		if u.Utilization >= threshold {
			filtered = append(filtered, u)
		}
	}

	if len(filtered) == 0 {
		fmt.Printf("No resources found with utilization above %.1f%%\n", threshold*100)
		return
	}

	// Sort by utilization (highest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Utilization > filtered[j].Utilization
	})

	// Print header
	fmt.Println("\n📊 Resource Utilization Report")
	fmt.Println(strings.Repeat("─", 80))
	fmt.Printf("%-40s %-15s %-10s %-12s\n", "COMPONENT.RESOURCE", "TYPE", "UTIL %", "STATUS")
	fmt.Println(strings.Repeat("─", 80))

	// Print each resource
	for _, u := range filtered {
		status := getUtilizationStatus(u.Utilization, u.WarningThreshold, u.CriticalThreshold)
		utilPct := u.Utilization * 100

		// Color the utilization based on status
		utilStr := fmt.Sprintf("%.1f%%", utilPct)
		if u.Utilization >= u.CriticalThreshold {
			utilStr = fmt.Sprintf("\033[31m%.1f%%\033[0m", utilPct) // Red
		} else if u.Utilization >= u.WarningThreshold {
			utilStr = fmt.Sprintf("\033[33m%.1f%%\033[0m", utilPct) // Yellow
		}

		componentPath := u.ComponentPath
		if u.IsBottleneck {
			componentPath = "🔥 " + componentPath
		}

		fmt.Printf("%-40s %-15s %-10s %-12s\n",
			componentPath,
			u.ResourceName,
			utilStr,
			status,
		)

		// Show additional details for high utilization
		if u.Utilization >= u.WarningThreshold {
			fmt.Printf("  └─ Load: %.2f RPS, Capacity: %.0f\n", 
				u.CurrentLoad, u.Capacity)
		}
	}

	fmt.Println(strings.Repeat("─", 80))

	// Show legend
	fmt.Println("\nLegend:")
	fmt.Println("  🔥 = Bottleneck resource")
	fmt.Println("  \033[33mYellow\033[0m = Warning (>80% utilization)")
	fmt.Println("  \033[31mRed\033[0m = Critical (>95% utilization)")
}

func getUtilizationStatus(util, warning, critical float64) string {
	if util >= critical {
		return "⚠️  CRITICAL"
	} else if util >= warning {
		return "⚠️  WARNING"
	} else if util >= 0.5 {
		return "MODERATE"
	} else {
		return "OK"
	}
}
