package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1"
	"github.com/spf13/cobra"
)

var traceCmd = &cobra.Command{
	Use:   "trace <component.method>",
	Short: "Traces the execution of a specific operation",
	Long: `Executes a specific method call within the active system (e.g., "server.Lookup")
and outputs a detailed execution trace. The output is a JSON file that can be
used by other commands like 'diagram dynamic' to generate visualizations.

Prerequisites:
- SDL server must be running (sdl serve)
- A file must be loaded (sdl load <file>)
- A system must be active (sdl use <system>)`,
	Args: cobra.ExactArgs(1), // component.method string
	Run: func(cmd *cobra.Command, args []string) {
		methodCallString := args[0]

		outputFile, _ := cmd.Flags().GetString("out")
		if outputFile == "" {
			fmt.Fprintln(os.Stderr, "Error: Output file must be specified with --out or -o.")
			os.Exit(1)
		}

		// Parse the method call string
		parts := strings.Split(methodCallString, ".")
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Error: Invalid method call format. Expected 'component.method', got '%s'\n", methodCallString)
			os.Exit(1)
		}
		componentName := parts[0]
		methodName := parts[1]

		// Execute trace via gRPC
		err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			req := &v1.ExecuteTraceRequest{
				CanvasId:  canvasID,
				Component: componentName,
				Method:    methodName,
			}

			resp, err := client.ExecuteTrace(ctx, req)
			if err != nil {
				return fmt.Errorf("trace execution failed: %v", err)
			}

			// Convert proto TraceData to JSON
			jsonData, err := json.MarshalIndent(resp.TraceData, "", "  ")
			if err != nil {
				return fmt.Errorf("error marshalling trace data to JSON: %v", err)
			}

			// Write to output file
			err = os.WriteFile(outputFile, jsonData, 0644)
			if err != nil {
				return fmt.Errorf("error writing trace data to %s: %v", outputFile, err)
			}

			fmt.Printf("Trace data successfully written to %s\n", outputFile)
			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	AddCommand(traceCmd)
	traceCmd.Flags().StringP("out", "o", "", "Output detailed trace data to a JSON file (required)")
	traceCmd.Flags().Int("depth", 0, "Limit trace depth (0 for unlimited)")
}
