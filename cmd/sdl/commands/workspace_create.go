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
		workspaceID := args[0]

		return withWorkspaceClient(func(client v1s.WorkspaceServiceClient, ctx context.Context) error {
			// First, try to get the canvas to see if it exists
			getResp, err := client.GetWorkspace(ctx, &v1.GetWorkspaceRequest{
				Id: workspaceID,
			})

			if err == nil && getResp.Workspace != nil {
				// Canvas already exists
				fmt.Printf("Workspace '%s' already exists\n", workspaceID)
				return nil
			}

			// Canvas doesn't exist, create it
			canvas := &v1.Workspace{
				Id: workspaceID,
			}

			_, err = client.CreateWorkspace(ctx, &v1.CreateWorkspaceRequest{
				Workspace: canvas,
			})

			if err != nil {
				return fmt.Errorf("failed to create canvas: %w", err)
			}

			fmt.Printf("Workspace '%s' created successfully\n", workspaceID)
			return nil
		})
	},
}

func init() {
	canvasCmd.AddCommand(canvasCreateCmd)
}
