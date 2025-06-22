package commands

import (
	"context"
	"fmt"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1"
	"github.com/spf13/cobra"
)

var canvasResetCmd = &cobra.Command{
	Use:   "reset <canvasId>",
	Short: "Reset a canvas - clear all state, generators, and metrics",
	Long: `Reset the specified canvas to a clean state. This will:
- Stop all running generators
- Clear all metrics
- Remove the active system
- Clear all loaded systems
- Reset flow analysis state

This is useful for starting fresh without restarting the server.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		canvasID := args[0]
		
		return withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			req := &v1.ResetCanvasRequest{
				CanvasId: canvasID,
			}

			resp, err := client.ResetCanvas(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to reset canvas '%s': %w", canvasID, err)
			}

			if !resp.Success {
				return fmt.Errorf("reset failed: %s", resp.Message)
			}

			fmt.Printf("Canvas '%s' reset successfully\n", canvasID)
			return nil
		})
	},
}

func init() {
	canvasCmd.AddCommand(canvasResetCmd)
}