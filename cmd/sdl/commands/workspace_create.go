package commands

import (
	"context"
	"fmt"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	v1s "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/spf13/cobra"
)

var workspaceCreateCmd = &cobra.Command{
	Use:   "create <workspaceId>",
	Short: "Create a new workspace",
	Long:  `Create a new workspace with the specified ID. If a workspace with the given ID already exists, it will display a message indicating so.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceID := args[0]

		return withWorkspaceClient(func(client v1s.WorkspaceServiceClient, ctx context.Context) error {
			// First, try to get the workspace to see if it exists
			getResp, err := client.GetWorkspace(ctx, &v1.GetWorkspaceRequest{
				Id: workspaceID,
			})

			if err == nil && getResp.Workspace != nil {
				// Workspace already exists
				fmt.Printf("Workspace '%s' already exists\n", workspaceID)
				return nil
			}

			// Workspace doesn't exist, create it
			workspace := &v1.Workspace{
				Id: workspaceID,
			}

			_, err = client.CreateWorkspace(ctx, &v1.CreateWorkspaceRequest{
				Workspace: workspace,
			})

			if err != nil {
				return fmt.Errorf("failed to create workspace: %w", err)
			}

			fmt.Printf("Workspace '%s' created successfully\n", workspaceID)
			return nil
		})
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceCreateCmd)
}
