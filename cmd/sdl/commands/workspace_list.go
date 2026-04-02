package commands

import (
	"context"
	"fmt"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	v1s "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/spf13/cobra"
)

var workspaceListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all workspaces",
	Long:    `List all available workspaces in the SDL server.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return withWorkspaceClient(func(client v1s.WorkspaceServiceClient, ctx context.Context) error {
			resp, err := client.ListWorkspaces(ctx, &v1.ListWorkspacesRequest{})
			if err != nil {
				return fmt.Errorf("failed to list workspaces: %w", err)
			}

			if len(resp.Workspaces) == 0 {
				fmt.Println("No workspaces found")
				return nil
			}

			fmt.Printf("Workspacees (%d):\n", len(resp.Workspaces))
			for _, workspace := range resp.Workspaces {
				fmt.Printf("  - %s", workspace.Id)
				if "" != "" {
					fmt.Printf(" (active system: %s)", "")
				}
				fmt.Println()
			}

			return nil
		})
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceListCmd)
}
