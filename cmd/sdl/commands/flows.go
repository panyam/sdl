package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/panyam/sdl/runtime"
	"github.com/spf13/cobra"
)

// flowsCmd represents the flows command
var flowsCmd = &cobra.Command{
	Use:   "flows",
	Short: "Manage and analyze system flow rates",
	Long: `The flows command provides subcommands to list available flow strategies,
evaluate flows using different strategies, apply flow strategies to the system,
and view current flow state.`,
}

// listStrategiesCmd lists available flow evaluation strategies
var listStrategiesCmd = &cobra.Command{
	Use:   "list",
	Short: "List available flow evaluation strategies",
	Long:  `Lists all registered flow evaluation strategies with their descriptions and status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		strategies := runtime.ListFlowStrategies()

		if outputFormat == "json" {
			data, err := json.MarshalIndent(strategies, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal strategies: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "STRATEGY\tSTATUS\tRECOMMENDED\tDESCRIPTION")
		fmt.Fprintln(w, "--------\t------\t-----------\t-----------")

		for name, info := range strategies {
			recommended := ""
			if info.Recommended {
				recommended = "✓"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, info.Status, recommended, info.Description)
		}
		w.Flush()

		return nil
	},
}

// evalFlowCmd evaluates flows using a specific strategy
var evalFlowCmd = &cobra.Command{
	Use:   "eval [strategy]",
	Short: "Evaluate system flows using a specific strategy",
	Long: `Evaluates the system flows using the specified strategy.
If no strategy is provided, uses the default (runtime) strategy.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		strategy := "runtime"
		if len(args) > 0 {
			strategy = args[0]
		}

		// Make API call to evaluate flows
		resp, err := makeAPICall[map[string]any]("GET", fmt.Sprintf("/api/flows/%s/eval", strategy), nil)
		if err != nil {
			return fmt.Errorf("failed to evaluate flows: %w", err)
		}

		// Convert response to FlowAnalysisResult
		var result runtime.FlowAnalysisResult
		if data, ok := resp["result"]; ok {
			// Marshal and unmarshal to convert map to struct
			jsonData, _ := json.Marshal(data)
			if err := json.Unmarshal(jsonData, &result); err != nil {
				return fmt.Errorf("failed to parse result: %w", err)
			}
		}

		if outputFormat == "json" {
			data, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal result: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Human-readable format
		fmt.Printf("Flow Analysis Results\n")
		fmt.Printf("====================\n")
		fmt.Printf("Strategy: %s\n", result.Strategy)
		fmt.Printf("Status: %s\n", result.Status)
		if result.Iterations > 0 {
			fmt.Printf("Iterations: %d\n", result.Iterations)
		}
		fmt.Printf("System: %s\n", result.System)

		if len(result.Generators) > 0 {
			fmt.Printf("\nGenerators:\n")
			for _, gen := range result.Generators {
				fmt.Printf("  - %s.%s @ %.2f RPS\n", gen.Component, gen.Method, gen.Rate)
			}
		}

		if len(result.Flows.ComponentRates) > 0 {
			fmt.Printf("\nComponent Rates:\n")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "COMPONENT.METHOD\tRATE (RPS)")
			fmt.Fprintln(w, "----------------\t----------")

			// Sort for consistent output
			var keys []string
			for k := range result.Flows.ComponentRates {
				keys = append(keys, k)
			}
			// Simple sort
			for i := range keys {
				for j := i + 1; j < len(keys); j++ {
					if keys[i] > keys[j] {
						keys[i], keys[j] = keys[j], keys[i]
					}
				}
			}

			for _, key := range keys {
				rate := result.Flows.ComponentRates[key]
				fmt.Fprintf(w, "%s\t%.2f\n", key, rate)
			}
			w.Flush()
		}

		if len(result.Warnings) > 0 {
			fmt.Printf("\nWarnings:\n")
			for _, warning := range result.Warnings {
				fmt.Printf("  ⚠ %s\n", warning)
			}
		}

		return nil
	},
}

// applyFlowCmd applies a flow strategy to the system
var applyFlowCmd = &cobra.Command{
	Use:   "apply [strategy]",
	Short: "Apply a flow evaluation strategy to the system",
	Long: `Applies the specified flow evaluation strategy to the system,
updating all component arrival rates based on the analysis.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		strategy := "runtime"
		if len(args) > 0 {
			strategy = args[0]
		}

		// Make API call to apply flow strategy
		_, err := makeAPICall[any]("POST", fmt.Sprintf("/api/flows/%s/apply", strategy), nil)
		if err != nil {
			return fmt.Errorf("failed to apply flow strategy: %w", err)
		}

		fmt.Printf("Successfully applied '%s' flow strategy\n", strategy)

		// Show current state if verbose
		if verbose {
			fmt.Println("\nCurrent flow state:")
			return showCurrentFlow()
		}

		return nil
	},
}

// showFlowCmd shows current flow state
var showFlowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current flow state and rates",
	Long:  `Displays the current flow state including active strategy, rates, and any manual overrides.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showCurrentFlow()
	},
}

func showCurrentFlow() error {
	resp, err := makeAPICall[map[string]any]("GET", "/api/flows/current", nil)
	if err != nil {
		return fmt.Errorf("failed to get current flow state: %w", err)
	}

	// Convert response to FlowState
	var state runtime.FlowState
	if data, ok := resp["state"]; ok {
		jsonData, _ := json.Marshal(data)
		if err := json.Unmarshal(jsonData, &state); err != nil {
			return fmt.Errorf("failed to parse state: %w", err)
		}
	}

	if outputFormat == "json" {
		data, err := json.MarshalIndent(state, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal state: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Human-readable format
	fmt.Printf("Current Flow State\n")
	fmt.Printf("==================\n")
	fmt.Printf("Strategy: %s\n", state.Strategy)

	if len(state.Rates) > 0 {
		fmt.Printf("\nComponent Rates:\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "COMPONENT.METHOD\tRATE (RPS)\tOVERRIDDEN")
		fmt.Fprintln(w, "----------------\t----------\t----------")

		// Sort keys
		var keys []string
		for k := range state.Rates {
			keys = append(keys, k)
		}
		for i := range keys {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}

		for _, key := range keys {
			rate := state.Rates[key]
			overridden := ""
			if _, ok := state.ManualOverrides[key]; ok {
				overridden = "✓"
			}
			fmt.Fprintf(w, "%s\t%.2f\t%s\n", key, rate, overridden)
		}
		w.Flush()
	}

	return nil
}

// setRateCmd sets arrival rate for a specific component method
var setRateCmd = &cobra.Command{
	Use:   "set-rate <component.method> <rate>",
	Short: "Set manual arrival rate override for a component method",
	Long: `Sets a manual arrival rate override for a specific component method.
This override will persist until cleared or the flow strategy is re-applied.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse component.method
		parts := strings.Split(args[0], ".")
		if len(parts) != 2 {
			return fmt.Errorf("invalid format: expected component.method, got %s", args[0])
		}

		component := parts[0]
		method := parts[1]

		// Parse rate
		var rate float64
		if _, err := fmt.Sscanf(args[1], "%f", &rate); err != nil {
			return fmt.Errorf("invalid rate: %s", args[1])
		}

		// Set the rate
		body := map[string]interface{}{
			"rate": rate,
		}
		_, err := makeAPICall[any]("PUT", fmt.Sprintf("/api/components/%s/methods/%s/arrival-rate", component, method), body)
		if err != nil {
			return fmt.Errorf("failed to set arrival rate: %w", err)
		}

		fmt.Printf("Set arrival rate for %s.%s to %.2f RPS\n", component, method, rate)

		if verbose {
			fmt.Println("\nCurrent flow state:")
			return showCurrentFlow()
		}

		return nil
	},
}

var (
	outputFormat string
	verbose      bool
)

func init() {
	// Add flows command to root
	AddCommand(flowsCmd)

	// Add subcommands
	flowsCmd.AddCommand(listStrategiesCmd)
	flowsCmd.AddCommand(evalFlowCmd)
	flowsCmd.AddCommand(applyFlowCmd)
	flowsCmd.AddCommand(showFlowCmd)
	flowsCmd.AddCommand(setRateCmd)

	// Add flags
	flowsCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json)")
	flowsCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}
