package commands

import (
	"context"
	"fmt"
	"os"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/models"
	v1s "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"github.com/panyam/sdl/lib/viz"
	"github.com/spf13/cobra"
)

var canvasDiagramCmd = &cobra.Command{
	Use:   "diagram",
	Short: "Generate a diagram of the current canvas system",
	Long: `Generates a visual representation of the active system in the canvas,
showing runtime component instances and their relationships.

This diagram is based on actual runtime instances, properly handling shared components.`,
	Run: func(cmd *cobra.Command, args []string) {
		outputFile, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		err := withCanvasClient(func(client v1s.CanvasServiceClient, ctx context.Context) error {
			// Get the system diagram from the canvas
			resp, err := client.GetSystemDiagram(ctx, &v1.GetSystemDiagramRequest{
				CanvasId: canvasID,
			})
			if err != nil {
				return fmt.Errorf("failed to get system diagram: %v", err)
			}

			diagram := resp.Diagram
			if diagram == nil {
				return fmt.Errorf("no diagram returned")
			}

			// Get the appropriate generator based on format
			var generator viz.StaticDiagramGenerator
			switch format {
			case "dot":
				generator = &viz.DotGenerator{}
			case "mermaid":
				generator = &viz.MermaidStaticGenerator{}
			case "excalidraw":
				generator = &viz.ExcalidrawGenerator{}
			case "svg":
				generator = &viz.SvgGenerator{}
			default:
				return fmt.Errorf("unsupported format '%s'. Choose 'dot', 'mermaid', 'excalidraw', or 'svg'", format)
			}

			// Generate the diagram
			diagramOutput, err := generator.Generate(diagram)
			if err != nil {
				return fmt.Errorf("error generating %s diagram: %v", format, err)
			}

			// Write output
			if diagramOutput == "" {
				return fmt.Errorf("no diagram content generated")
			}

			if outputFile != "" {
				err := os.WriteFile(outputFile, []byte(diagramOutput), 0644)
				if err != nil {
					return fmt.Errorf("error writing diagram to %s: %v", outputFile, err)
				}
				fmt.Printf("✅ Diagram written to %s\n", outputFile)
			} else {
				fmt.Println("\nSystem Diagram:")
				fmt.Println(diagramOutput)
			}

			return nil
		})

		if err != nil {
			fmt.Printf("❌ Failed to generate diagram: %v\n", err)
			if err.Error() == "cannot connect to SDL server: failed to connect to gRPC server at localhost:9090: context deadline exceeded" {
				fmt.Printf("\nTo use SDL commands, first start the server:\n")
				fmt.Printf("   sdl serve\n")
			}
		}
	},
}

func init() {
	canvasCmd.AddCommand(canvasDiagramCmd)
	canvasDiagramCmd.Flags().StringP("output", "o", "", "Output file path for the diagram")
	canvasDiagramCmd.Flags().String("format", "dot", "Output format (dot, mermaid, excalidraw, svg)")
}
