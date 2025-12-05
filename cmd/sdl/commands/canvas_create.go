package commands

import (
	"context"
	"fmt"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	v1s "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/spf13/cobra"
)

var canvasCreateCmd = &cobra.Command{
	Use:   "create <canvasId>",
	Short: "Create a new canvas",
	Long:  `Create a new canvas with the specified ID. If a canvas with the given ID already exists, it will display a message indicating so.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		canvasID := args[0]

		return withCanvasClient(func(client v1s.CanvasServiceClient, ctx context.Context) error {
			// First, try to get the canvas to see if it exists
			getResp, err := client.GetCanvas(ctx, &v1.GetCanvasRequest{
				Id: canvasID,
			})

			if err == nil && getResp.Canvas != nil {
				// Canvas already exists
				fmt.Printf("Canvas '%s' already exists\n", canvasID)
				return nil
			}

			// Canvas doesn't exist, create it
			canvas := &v1.Canvas{
				Id: canvasID,
			}

			_, err = client.CreateCanvas(ctx, &v1.CreateCanvasRequest{
				Canvas: canvas,
			})

			if err != nil {
				return fmt.Errorf("failed to create canvas: %w", err)
			}

			fmt.Printf("Canvas '%s' created successfully\n", canvasID)
			return nil
		})
	},
}

func init() {
	canvasCmd.AddCommand(canvasCreateCmd)
}
