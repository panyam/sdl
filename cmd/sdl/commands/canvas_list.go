package commands

import (
	"context"
	"fmt"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	v1s "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/spf13/cobra"
)

var canvasListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all canvases",
	Long:    `List all available canvases in the SDL server.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return withCanvasClient(func(client v1s.CanvasServiceClient, ctx context.Context) error {
			resp, err := client.ListCanvases(ctx, &v1.ListCanvasesRequest{})
			if err != nil {
				return fmt.Errorf("failed to list canvases: %w", err)
			}

			if len(resp.Canvases) == 0 {
				fmt.Println("No canvases found")
				return nil
			}

			fmt.Printf("Canvases (%d):\n", len(resp.Canvases))
			for _, canvas := range resp.Canvases {
				fmt.Printf("  - %s", canvas.Id)
				if canvas.ActiveSystem != "" {
					fmt.Printf(" (active system: %s)", canvas.ActiveSystem)
				}
				fmt.Println()
			}

			return nil
		})
	},
}

func init() {
	canvasCmd.AddCommand(canvasListCmd)
}
