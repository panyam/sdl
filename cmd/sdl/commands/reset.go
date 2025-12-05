package commands

import (
	"context"
	"fmt"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	v1s "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/spf13/cobra"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the canvas - clear all state, generators, and metrics",
	Long: `Reset the canvas to a clean state. This will:
- Stop all running generators
- Clear all metrics
- Remove the active system
- Clear all loaded systems
- Reset flow analysis state

This is useful for starting fresh without restarting the server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReset()
	},
}

// DEPRECATED: This command has been moved to "canvas reset <canvasId>"
// Keeping the code for backward compatibility but not registering it
func init() {
	// rootCmd.AddCommand(resetCmd)
}

func runReset() error {
	return withCanvasClient(func(client v1s.CanvasServiceClient, ctx context.Context) error {
		req := &v1.ResetCanvasRequest{
			CanvasId: canvasID,
		}

		resp, err := client.ResetCanvas(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to reset canvas: %w", err)
		}

		if !resp.Success {
			return fmt.Errorf("reset failed: %s", resp.Message)
		}

		fmt.Println(resp.Message)
		return nil
	})
}
