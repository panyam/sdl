package commands

import (
	"context"
	"fmt"
	"strconv"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	v1s "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/spf13/cobra"
)

// Canvas state management commands use shared API client from api.go

// Canvas state management commands

var loadCmd = &cobra.Command{
	Use:   "load [file]",
	Short: "Load an SDL file into the server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := withWorkspaceClient(func(client v1s.WorkspaceServiceClient, ctx context.Context) error {
			_, err := client.LoadFile(ctx, &v1.LoadFileRequest{
				WorkspaceId:    canvasID,
				SdlFilePath: args[0],
			})
			return err
		})

		if err != nil {
			fmt.Printf("❌ Failed to load file: %v\n", err)
			if err.Error() == "cannot connect to SDL server: failed to connect to gRPC server at localhost:9090: context deadline exceeded" {
				fmt.Printf("\nTo use SDL commands, first start the server:\n")
				fmt.Printf("   sdl serve\n")
			}
			return
		}

		fmt.Printf("✅ Loaded %s successfully (canvas: %s)\n", args[0], canvasID)
	},
}

var useCmd = &cobra.Command{
	Use:   "use [system]",
	Short: "Select the active system",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := withWorkspaceClient(func(client v1s.WorkspaceServiceClient, ctx context.Context) error {
			_, err := client.UseSystem(ctx, &v1.UseSystemRequest{
				WorkspaceId:   canvasID,
				SystemName: args[0],
			})
			return err
		})

		if err != nil {
			fmt.Printf("❌ Failed to use system: %v\n", err)
			return
		}

		fmt.Printf("✅ Now using system: %s\n", args[0])
	},
}

var setCmd = &cobra.Command{
	Use:   "set [parameter] [value]",
	Short: "Set a parameter value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := withWorkspaceClient(func(client v1s.WorkspaceServiceClient, ctx context.Context) error {
			resp, err := client.SetParameter(ctx, &v1.SetParameterRequest{
				WorkspaceId: canvasID,
				Path:     args[0],
				NewValue: args[1],
			})
			if err != nil {
				return err
			}
			if !resp.Success {
				return fmt.Errorf("%s", resp.ErrorMessage)
			}
			return nil
		})

		if err != nil {
			fmt.Printf("❌ Failed to set parameter: %v\n", err)
			return
		}

		fmt.Printf("✅ Set %s = %s\n", args[0], args[1])
	},
}

var getCmd = &cobra.Command{
	Use:   "get [parameter]",
	Short: "View parameter values",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := withWorkspaceClient(func(client v1s.WorkspaceServiceClient, ctx context.Context) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			resp, err := client.GetParameters(ctx, &v1.GetParametersRequest{
				WorkspaceId: canvasID,
				Path:     path,
			})
			if err != nil {
				return err
			}

			if len(resp.Parameters) == 0 {
				if path == "" {
					fmt.Println("No parameters set")
				} else {
					fmt.Printf("Parameter '%s' not found\n", path)
				}
				return nil
			}

			// Display parameters
			for key, value := range resp.Parameters {
				fmt.Printf("%s = %s\n", key, value)
			}

			return nil
		})

		if err != nil {
			fmt.Printf("❌ Failed to get parameters: %v\n", err)
		}
	},
}

var runCanvasCmd = &cobra.Command{
	Use:   "run [name] [method] [calls]",
	Short: "Run a simulation",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		calls, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Printf("❌ Invalid call count '%s': must be a number\n", args[2])
			return
		}

		_, err = makeAPICall[any]("POST", "/api/console/run", map[string]any{
			"name":   args[0],
			"method": args[1],
			"calls":  calls,
		})
		if err == nil {
			fmt.Printf("✅ Running %s: %s (%d calls)\n", args[0], args[1], calls)
		}
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show current canvas state",
	Run: func(cmd *cobra.Command, args []string) {
		err := withWorkspaceClient(func(client v1s.WorkspaceServiceClient, ctx context.Context) error {
			resp, err := client.GetWorkspace(ctx, &v1.GetWorkspaceRequest{
				Id: canvasID,
			})
			if err != nil {
				return err
			}

			canvas := resp.Workspace
			fmt.Printf("SDL Canvas State:\n")
			fmt.Printf("🆔 Canvas ID: %s\n", canvas.Id)

			if canvas.ActiveDesign != "" {
				fmt.Printf("🎯 Active System: %s\n", canvas.ActiveDesign)
			}

			// TODO: When Canvas proto is updated to include generators and metrics
			// if len(canvas.Generators) > 0 {
			//     fmt.Printf("⚡ Generators: %d\n", len(canvas.Generators))
			// }
			// if len(canvas.Metrics) > 0 {
			//     fmt.Printf("📊 Metrics: %d\n", len(canvas.Metrics))
			// }

			return nil
		})

		if err != nil {
			fmt.Printf("❌ Failed to get canvas info: %v\n", err)
		}
	},
}

var executeCmd = &cobra.Command{
	Use:   "execute [recipe-file]",
	Short: "Execute a recipe file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := makeAPICall[any]("POST", "/api/console/execute", map[string]any{"filePath": args[0]})
		if err == nil {
			fmt.Printf("✅ Executed recipe: %s\n", args[0])
		}
	},
}

// HTTP client is provided by api.go

func init() {
	// Add commands to root (server flag is now persistent on root command)
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(runCanvasCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(executeCmd)
}
