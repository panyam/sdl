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
and displays the execution trace. By default, it shows a human-readable tree view
of the execution. Use -o to save the trace data as JSON for other commands like
'diagram dynamic' to generate visualizations.

Prerequisites:
- SDL server must be running (sdl serve)
- A file must be loaded (sdl load <file>)
- A system must be active (sdl use <system>)`,
	Args: cobra.ExactArgs(1), // component.method string
	Run: func(cmd *cobra.Command, args []string) {
		methodCallString := args[0]

		outputFile, _ := cmd.Flags().GetString("out")

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

			if outputFile != "" {
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
			} else {
				// Display human-readable output
				displayTraceOutput(resp.TraceData)
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
	AddCommand(traceCmd)
	traceCmd.Flags().StringP("out", "o", "", "Output detailed trace data to a JSON file (optional)")
	traceCmd.Flags().Int("depth", 0, "Limit trace depth (0 for unlimited)")
}

// displayCallTree displays a single call and its children in tree format
func displayCallTree(event *v1.TraceEvent, exitMap map[int64]*v1.TraceEvent, childrenMap map[int64][]*v1.TraceEvent, depth int, pipes []bool) {
	// Get exit event for timing info
	exitEvent := exitMap[event.Id]
	
	// Format timing information
	startTime := fmt.Sprintf("%7.2f", event.Timestamp*1000)
	duration := "        "
	if exitEvent != nil && exitEvent.Duration > 0 {
		duration = fmt.Sprintf("%7.2fms", exitEvent.Duration*1000)
	}
	
	// Build the tree prefix with pipes
	prefix := ""
	// Don't draw pipes for the root level
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
	
	// Format the call
	component := event.Component
	if component == "" {
		component = "[native]"
	}
	
	call := fmt.Sprintf("%s.%s", component, event.Method)
	if len(event.Args) > 0 {
		call += "(" + strings.Join(event.Args, ", ") + ")"
	}
	
	// Add return value or error if present
	if exitEvent != nil {
		if exitEvent.ReturnValue != "" && exitEvent.ReturnValue != "null" {
			call += " → " + exitEvent.ReturnValue
		}
		if exitEvent.ErrorMessage != "" {
			call += " ERROR: " + exitEvent.ErrorMessage
		}
	}
	
	// Print the line
	fmt.Printf("%s  %s  %s%s\n", startTime, duration, prefix, call)
	
	// Display children
	children := childrenMap[event.Id]
	for i, child := range children {
		isLast := i == len(children)-1
		newPipes := append([]bool{}, pipes...)
		// Only add a pipe if this is not the last child
		if !isLast {
			newPipes = append(newPipes, true)
		} else {
			newPipes = append(newPipes, false)
		}
		displayCallTree(child, exitMap, childrenMap, depth+1, newPipes)
	}
}

// displayTraceOutput displays trace data in a human-readable format
func displayTraceOutput(trace *v1.TraceData) {
	if trace == nil || len(trace.Events) == 0 {
		fmt.Println("No trace events recorded")
		return
	}

	fmt.Printf("\n=== Call Tree: %s.%s ===\n\n", trace.System, trace.EntryPoint)

	// Build event relationships
	eventMap := make(map[int64]*v1.TraceEvent)
	exitMap := make(map[int64]*v1.TraceEvent)
	childrenMap := make(map[int64][]*v1.TraceEvent)
	
	for _, event := range trace.Events {
		if event.Kind == "enter" {
			eventMap[event.Id] = event
			if event.ParentId > 0 {
				childrenMap[event.ParentId] = append(childrenMap[event.ParentId], event)
			}
		} else if event.Kind == "exit" {
			exitMap[event.Id] = event
		}
	}

	// Find root events
	var roots []*v1.TraceEvent
	for _, event := range trace.Events {
		if event.ParentId == 0 && event.Kind == "enter" {
			roots = append(roots, event)
		}
	}

	// Display the tree
	fmt.Println("Time(ms)  Duration   Call")
	fmt.Println("--------  --------   " + strings.Repeat("-", 60))
	
	for _, root := range roots {
		displayCallTree(root, exitMap, childrenMap, 0, []bool{})
	}

	// Display summary statistics
	fmt.Printf("\n=== Performance Summary ===\n")
	var totalDuration float64
	callCount := make(map[string]int)
	totalLatency := make(map[string]float64)
	
	for _, event := range trace.Events {
		if event.Kind == "exit" {
			key := fmt.Sprintf("%s.%s", event.Component, event.Method)
			callCount[key]++
			totalLatency[key] += event.Duration
			if event.ParentId == 0 {
				totalDuration = event.Duration
			}
		}
	}

	fmt.Printf("Total Duration: %.3fms\n", totalDuration*1000)
	fmt.Printf("\nComponent Performance:\n")
	
	// Sort by component name for consistent output
	var components []string
	for key := range callCount {
		components = append(components, key)
	}
	for i := 0; i < len(components); i++ {
		for j := i + 1; j < len(components); j++ {
			if components[i] > components[j] {
				components[i], components[j] = components[j], components[i]
			}
		}
	}
	
	for _, key := range components {
		count := callCount[key]
		avgLatency := totalLatency[key] / float64(count)
		totalTime := totalLatency[key]
		fmt.Printf("  %-30s: %3d calls, %7.3fms avg, %7.3fms total\n", 
			key, count, avgLatency*1000, totalTime*1000)
	}
}

