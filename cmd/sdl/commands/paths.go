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

var pathsCmd = &cobra.Command{
	Use:   "paths <component.method>",
	Short: "Traces all possible execution paths from a specific operation",
	Long: `Performs breadth-first traversal to discover and display all possible 
execution paths from a specific method call within the active system (e.g., "server.Lookup").
Unlike 'trace' which shows a single execution, this command enumerates all possible 
paths through the system, including conditional branches and loops.

By default, it shows a human-readable tree view of all execution paths.
Use -o to save the complete path data as JSON for visualization tools.

Prerequisites:
- SDL server must be running (sdl serve)
- A file must be loaded (sdl load <file>)
- A system must be active (sdl use <system>)`,
	Args: cobra.ExactArgs(1), // component.method string
	Run: func(cmd *cobra.Command, args []string) {
		methodCallString := args[0]

		outputFile, _ := cmd.Flags().GetString("out")
		maxDepth, _ := cmd.Flags().GetInt32("max-depth")

		// Parse the method call string
		parts := strings.Split(methodCallString, ".")
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Error: Invalid method call format. Expected 'component.method', got '%s'\n", methodCallString)
			os.Exit(1)
		}
		componentName := parts[0]
		methodName := parts[1]

		// Execute path traversal via gRPC
		err := withCanvasClient(func(client v1.CanvasServiceClient, ctx context.Context) error {
			req := &v1.TraceAllPathsRequest{
				CanvasId:  canvasID,
				Component: componentName,
				Method:    methodName,
				MaxDepth:  maxDepth,
			}

			resp, err := client.TraceAllPaths(ctx, req)
			if err != nil {
				return fmt.Errorf("path traversal failed: %v", err)
			}

			if outputFile != "" {
				// Convert proto AllPathsTraceData to JSON
				jsonData, err := json.MarshalIndent(resp.TraceData, "", "  ")
				if err != nil {
					return fmt.Errorf("error marshalling path data to JSON: %v", err)
				}

				// Write to output file
				err = os.WriteFile(outputFile, jsonData, 0644)
				if err != nil {
					return fmt.Errorf("error writing path data to %s: %v", outputFile, err)
				}

				fmt.Printf("Path data successfully written to %s\n", outputFile)
			} else {
				// Display human-readable output
				displayAllPathsOutput(resp.TraceData)
			}
			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	AddCommand(pathsCmd)
	pathsCmd.Flags().StringP("out", "o", "", "Output detailed path data to a JSON file (optional)")
	pathsCmd.Flags().Int32("max-depth", 0, "Maximum traversal depth (0 for unlimited)")
}

// displayAllPathsOutput displays all paths data in a human-readable format
func displayAllPathsOutput(traceData *v1.AllPathsTraceData) {
	if traceData == nil || traceData.Root == nil {
		fmt.Println("No path data recorded")
		return
	}

	fmt.Printf("\n=== All Execution Paths: %s ===\n\n", traceData.Root.StartingTarget)
	fmt.Printf("Trace ID: %s\n\n", traceData.TraceId)

	// Display the tree starting from root
	displayTraceNode(traceData.Root, 0, []bool{})

	// Display summary
	pathCount := countAllPaths(traceData.Root)
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total unique execution paths: %d\n", pathCount)
	
	if len(traceData.Root.Groups) > 0 {
		fmt.Printf("Branch points: %d\n", len(traceData.Root.Groups))
	}
}

// displayTraceNode displays a trace node and its edges in tree format
func displayTraceNode(node *v1.TraceNode, depth int, pipes []bool) {
	if node == nil {
		return
	}

	// Build the tree prefix with pipes
	prefix := ""
	if depth > 0 {
		// Draw vertical lines for parent levels
		for i := 0; i < depth-1; i++ {
			if i < len(pipes) && pipes[i] {
				prefix += "│   "
			} else {
				prefix += "    "
			}
		}
		
		// Add the branch for this level
		if depth-1 < len(pipes) && !pipes[depth-1] {
			prefix += "└── "
		} else {
			prefix += "├── "
		}
	}

	// Display the current node
	fmt.Printf("%s%s\n", prefix, node.StartingTarget)

	// Display groups if any
	for _, group := range node.Groups {
		groupPrefix := prefix
		if depth > 0 {
			if depth-1 < len(pipes) && !pipes[depth-1] {
				groupPrefix += "    "
			} else {
				groupPrefix += "│   "
			}
		}
		fmt.Printf("%s[%s: edges %d-%d]\n", groupPrefix, group.GroupLabel, group.GroupStart, group.GroupEnd)
	}

	// Display edges
	for i, edge := range node.Edges {
		isLast := i == len(node.Edges)-1
		
		// Build edge label with conditions
		edgeLabel := edge.Label
		if edge.IsConditional && edge.Condition != "" {
			edgeLabel += fmt.Sprintf(" [%s]", edge.Condition)
		}
		if edge.Probability != "" {
			edgeLabel += fmt.Sprintf(" (p=%s)", edge.Probability)
		}
		if edge.IsAsync {
			edgeLabel += " [async]"
		}
		if edge.IsReverse {
			edgeLabel += " [wait]"
		}

		// Build prefix for edge
		edgePrefix := ""
		for j := 0; j < depth; j++ {
			if j < len(pipes) && pipes[j] {
				edgePrefix += "│   "
			} else {
				edgePrefix += "    "
			}
		}
		
		if isLast {
			edgePrefix += "└─→ "
		} else {
			edgePrefix += "├─→ "
		}

		fmt.Printf("%s%s\n", edgePrefix, edgeLabel)

		// Recursively display the next node
		newPipes := append([]bool{}, pipes...)
		if !isLast {
			newPipes = append(newPipes, true)
		} else {
			newPipes = append(newPipes, false)
		}
		
		if edge.NextNode != nil {
			displayTraceNode(edge.NextNode, depth+1, newPipes)
		}
	}
}

// countAllPaths counts the total number of unique execution paths
func countAllPaths(node *v1.TraceNode) int {
	if node == nil || len(node.Edges) == 0 {
		return 1 // Leaf node represents one complete path
	}

	totalPaths := 0
	for _, edge := range node.Edges {
		if edge.NextNode != nil {
			totalPaths += countAllPaths(edge.NextNode)
		} else {
			totalPaths += 1
		}
	}

	return totalPaths
}