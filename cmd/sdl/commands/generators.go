package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1"
	"github.com/spf13/cobra"
)

// Generator management commands

func splitTarget(target string) (component string, method string, ok bool) {
	parts := strings.Split(target, ".")
	if len(parts) < 2 {
		fmt.Printf("❌ Error: Target must be of the form comp1.comp2.comp3...compN.MethodName")
		return
	}
	component = strings.Join(parts[:len(parts)-1], ".")
	method = parts[len(parts)-1]
	ok = true
	return
}

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
		component, method, ok := splitTarget(args[1])
		if !ok {
			return
		}
		rateStr := args[2]

		rate, err := strconv.ParseFloat(rateStr, 64)
		if err != nil {
			fmt.Printf("❌ Invalid rate '%s': must be a number\n", rateStr)
			return
		}
		err = withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			_, err := client.AddGenerator(ctx, &v1.AddGeneratorRequest{
				Generator: &v1.Generator{
					Id:        id,
					Name:      fmt.Sprintf("Generator-%s", id),
					CanvasId:  canvasID,
					Component: component,
					Method:    method,
					Rate:      rate,
					Enabled:   false,
				},
			})
			if err != nil {
				return err
			}

			// Apply flows if requested
			return applyFlowsIfRequested(client, ctx)
		})

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		fmt.Printf("✅ Generator '%s' created\n", id)
		fmt.Printf("🎯 Component: %s, Method: %s\n", component, method)
		fmt.Printf("⚡ Rate: %.2f calls/second\n", rate)
		fmt.Printf("🔄 Status: Stopped\n")
	},
}

var genListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all traffic generators",
	Run: func(cmd *cobra.Command, args []string) {
		err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			resp, err := client.ListGenerators(ctx, &v1.ListGeneratorsRequest{
				CanvasId: canvasID,
			})
			if err != nil {
				return err
			}

			fmt.Println("Traffic Generators:")
			fmt.Println("┌─────────────┬─────────────────────┬────────────┬─────────┐")
			fmt.Println("│ Name        │ Target              │    Rate    │ Status  │")
			fmt.Println("├─────────────┼─────────────────────┼────────────┼─────────┤")
			for _, gen := range resp.Generators {
				status := "Stopped"
				if gen.Enabled {
					status = "Running"
				}
				fmt.Printf("│ %-11s │ %-19s │ %10s │ %-7s │\n", gen.Id, gen.Component+"."+gen.Method, fmt.Sprintf("%0.2f", gen.Rate), status)
			}
			fmt.Println("└─────────────┴─────────────────────┴────────────┴─────────┘")
			_ = resp // Silence unused variable warning for now
			return nil
		})

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
	},
}

var genStartCmd = &cobra.Command{
	Use:   "start [id...]",
	Short: "Start traffic generators",
	Long:  "Start specific generators by ID, or all generators if no IDs provided",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Start all generators
			var resp *v1.StartAllGeneratorsResponse
			err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
				var err error
				resp, err = client.StartAllGenerators(ctx, &v1.StartAllGeneratorsRequest{
					CanvasId: canvasID,
				})
				if err != nil {
					return err
				}

				// Apply flows if requested
				return applyFlowsIfRequested(client, ctx)
			})
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}

			// Display detailed results
			if resp.TotalGenerators == 0 {
				fmt.Println("⚠️  No generators configured")
				return
			}

			fmt.Printf("✅ Generator batch operation completed:\n")
			fmt.Printf("   📊 Total generators: %d\n", resp.TotalGenerators)
			if resp.StartedCount > 0 {
				fmt.Printf("   ▶️  Started: %d\n", resp.StartedCount)
			}
			if resp.AlreadyRunningCount > 0 {
				fmt.Printf("   🔄 Already running: %d\n", resp.AlreadyRunningCount)
			}
			if resp.FailedCount > 0 {
				fmt.Printf("   ❌ Failed: %d\n", resp.FailedCount)
				for _, id := range resp.FailedIds {
					fmt.Printf("      - %s\n", id)
				}
			}
		} else {
			// Start specific generators
			for _, id := range args {
				err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
					_, err := client.StartGenerator(ctx, &v1.StartGeneratorRequest{
						CanvasId:    canvasID,
						GeneratorId: id,
					})
					return err
				})
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
			var resp *v1.StopAllGeneratorsResponse
			err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
				var err error
				resp, err = client.StopAllGenerators(ctx, &v1.StopAllGeneratorsRequest{
					CanvasId: canvasID,
				})
				if err != nil {
					return err
				}

				// Apply flows if requested
				return applyFlowsIfRequested(client, ctx)
			})
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}

			// Display detailed results
			if resp.TotalGenerators == 0 {
				fmt.Println("⚠️  No generators configured")
				return
			}

			fmt.Printf("✅ Generator batch operation completed:\n")
			fmt.Printf("   📊 Total generators: %d\n", resp.TotalGenerators)
			if resp.StoppedCount > 0 {
				fmt.Printf("   ⏹️  Stopped: %d\n", resp.StoppedCount)
			}
			if resp.AlreadyStoppedCount > 0 {
				fmt.Printf("   💤 Already stopped: %d\n", resp.AlreadyStoppedCount)
			}
			if resp.FailedCount > 0 {
				fmt.Printf("   ❌ Failed: %d\n", resp.FailedCount)
				for _, id := range resp.FailedIds {
					fmt.Printf("      - %s\n", id)
				}
			}
		} else {
			// Stop specific generators
			for _, id := range args {
				err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
					_, err := client.StopGenerator(ctx, &v1.StopGeneratorRequest{
						CanvasId:    canvasID,
						GeneratorId: id,
					})
					return err
				})
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
		err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			resp, err := client.GetCanvas(ctx, &v1.GetCanvasRequest{
				Id: canvasID,
			})
			if err != nil {
				return err
			}

			// TODO: Once Canvas proto includes generators field
			fmt.Println("Generator Status:")
			fmt.Println("📊 Total Generators: 0")

			// When available:
			// runningCount := 0
			// for _, gen := range resp.Canvas.Generators {
			//     if gen.Enabled {
			//         runningCount++
			//     }
			// }
			// fmt.Printf("Total Active Load: %d generators running\n", runningCount)

			_ = resp // Silence unused variable warning
			return nil
		})

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
	},
}

var genUpdateCmd = &cobra.Command{
	Use:   "update [id] [rate]",
	Short: "Update generator rate",
	Long:  "Update the rate of an existing generator. This is more efficient than removing and re-adding generators.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		rateStr := args[1]

		rate, err := strconv.ParseFloat(rateStr, 64)
		if err != nil {
			fmt.Printf("❌ Invalid rate '%s': must be a number\n", rateStr)
			return
		}

		err = withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			// First get the generator to validate it exists
			resp, err := client.GetGenerator(ctx, &v1.GetGeneratorRequest{
				CanvasId:    canvasID,
				GeneratorId: id,
			})
			if err != nil {
				return fmt.Errorf("generator '%s' not found: %v", id, err)
			}

			// Update only the rate
			gen := resp.Generator
			gen.Rate = rate

			_, err = client.UpdateGenerator(ctx, &v1.UpdateGeneratorRequest{
				Generator: gen,
			})
			if err != nil {
				return err
			}

			// Apply flows if requested
			return applyFlowsIfRequested(client, ctx)
		})

		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}

		fmt.Printf("✅ Generator '%s' rate updated to %.2f RPS\n", id, rate)
		if applyFlows {
			fmt.Println("✅ Flow rates automatically recalculated and applied")
		}
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

		removedCount := 0
		for _, id := range args {
			err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
				_, err := client.DeleteGenerator(ctx, &v1.DeleteGeneratorRequest{
					CanvasId:    canvasID,
					GeneratorId: id,
				})
				return err
			})
			if err != nil {
				fmt.Printf("❌ Error removing '%s': %v\n", id, err)
				continue
			}
			fmt.Printf("✅ Generator '%s' removed\n", id)
			removedCount++
		}

		// Apply flows if requested and at least one generator was removed
		if removedCount > 0 {
			err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
				return applyFlowsIfRequested(client, ctx)
			})
			if err != nil {
				fmt.Printf("❌ Error applying flows: %v\n", err)
			}
		}
	},
}

var applyFlows bool

func init() {
	// Add subcommands to gen
	genCmd.AddCommand(genAddCmd)
	genCmd.AddCommand(genListCmd)
	genCmd.AddCommand(genStartCmd)
	genCmd.AddCommand(genStopCmd)
	genCmd.AddCommand(genStatusCmd)
	genCmd.AddCommand(genUpdateCmd)
	genCmd.AddCommand(genRemoveCmd)

	// Add --apply-flows flag to commands that modify generators
	genAddCmd.Flags().BoolVar(&applyFlows, "apply-flows", false, "Automatically evaluate and apply flow rates after adding generator")
	genRemoveCmd.Flags().BoolVar(&applyFlows, "apply-flows", false, "Automatically evaluate and apply flow rates after removing generator")
	genUpdateCmd.Flags().BoolVar(&applyFlows, "apply-flows", false, "Automatically evaluate and apply flow rates after updating generator")
	genStartCmd.Flags().BoolVar(&applyFlows, "apply-flows", false, "Automatically evaluate and apply flow rates after starting generator")
	genStopCmd.Flags().BoolVar(&applyFlows, "apply-flows", false, "Automatically evaluate and apply flow rates after stopping generator")

	// Add to root command (server flag is now persistent on root command)
	rootCmd.AddCommand(genCmd)
}

// applyFlowsIfRequested evaluates and applies flows if the --apply-flows flag is set
func applyFlowsIfRequested(client v1.CanvasServiceClient, ctx context.Context) error {
	if !applyFlows {
		return nil
	}

	fmt.Println("🔄 Evaluating and applying flow rates...")

	// Evaluate flows with default strategy
	evalResp, err := client.EvaluateFlows(ctx, &v1.EvaluateFlowsRequest{
		CanvasId: canvasID,
		Strategy: "runtime",
	})
	if err != nil {
		return fmt.Errorf("failed to evaluate flows: %w", err)
	}

	// Convert component rates to parameter updates
	var updates []*v1.ParameterUpdate
	for componentMethod, rate := range evalResp.ComponentRates {
		// Skip zero rates
		if rate == 0 {
			continue
		}

		// Format as ArrivalRate parameter
		parts := strings.Split(componentMethod, ".")
		if len(parts) >= 2 {
			paramPath := fmt.Sprintf("%s.ArrivalRate", parts[0])
			updates = append(updates, &v1.ParameterUpdate{
				Path:     paramPath,
				NewValue: fmt.Sprintf("%g", rate),
			})
		}
	}

	if len(updates) == 0 {
		fmt.Println("   ℹ️  No arrival rates to apply")
		return nil
	}

	// Apply the rates using batch set
	batchResp, err := client.BatchSetParameters(ctx, &v1.BatchSetParametersRequest{
		CanvasId: canvasID,
		Updates:  updates,
	})
	if err != nil {
		return fmt.Errorf("failed to apply flow rates: %w", err)
	}

	if !batchResp.Success {
		return fmt.Errorf("failed to apply some parameters: %s", batchResp.ErrorMessage)
	}

	fmt.Printf("   ✅ Applied %d arrival rates\n", len(updates))
	return nil
}
