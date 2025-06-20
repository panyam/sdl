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
		parts := strings.Split(target, ".")
		if len(parts) < 2 {
			fmt.Printf("❌ Error: Target must be of the form comp1.comp2.comp3...compN.MethodName")
			return
		}
		err = withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			_, err := client.AddGenerator(ctx, &v1.AddGeneratorRequest{
				Generator: &v1.Generator{
					Id:        id,
					Name:      fmt.Sprintf("Generator-%s", id),
					CanvasId:  canvasID,
					Component: strings.Join(parts[:len(parts)-1], "."),
					Method:    parts[len(parts)-1],
					Rate:      float64(rate),
					Enabled:   false,
				},
			})
			return err
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
			err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
				_, err := client.StartAllGenerators(ctx, &v1.StartAllGeneratorsRequest{
					CanvasId: canvasID,
				})
				return err
			})
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}
			fmt.Println("✅ All generators started")
		} else {
			// Start specific generators
			for _, id := range args {
				err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
					_, err := client.ResumeGenerator(ctx, &v1.ResumeGeneratorRequest{
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
			err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
				_, err := client.StopAllGenerators(ctx, &v1.StopAllGeneratorsRequest{
					CanvasId: canvasID,
				})
				return err
			})
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}
			fmt.Println("✅ All generators stopped")
			fmt.Println("🛑 All traffic generation halted")
		} else {
			// Stop specific generators
			for _, id := range args {
				err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
					_, err := client.PauseGenerator(ctx, &v1.PauseGeneratorRequest{
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

var genSetCmd = &cobra.Command{
	Use:   "set [id] [property] [value]",
	Short: "Modify generator properties",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		property := args[1]
		value := args[2]

		err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			// Create a generator with just the fields to update
			gen := &v1.Generator{
				Id:       id,
				CanvasId: canvasID,
			}

			// Set the appropriate field based on property
			switch property {
			case "rate":
				rate, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("invalid rate value: %v", err)
				}
				gen.Rate = float64(rate)
			case "component":
				gen.Component = value
			case "method":
				gen.Method = value
			case "enabled":
				enabled, err := strconv.ParseBool(value)
				if err != nil {
					return fmt.Errorf("invalid boolean value: %v", err)
				}
				gen.Enabled = enabled
			default:
				return fmt.Errorf("unknown property: %s", property)
			}

			_, err := client.UpdateGenerator(ctx, &v1.UpdateGeneratorRequest{
				Generator: gen,
			})
			return err
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
