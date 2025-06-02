package commands

import (
	"bytes"
	"fmt"
	"os"

	// "encoding/json" // Added for Excalidraw
	// "math/rand"     // Added for Excalidraw element IDs
	// "strconv"       // Added for Excalidraw element IDs
	// "time"          // Added for Excalidraw element IDs

	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader"
	"github.com/spf13/cobra"
)

var diagramCmd = &cobra.Command{
	Use:   "diagram <diagram_type> <system_name>",
	Short: "Generates diagrams of system structure or behavior",
	Long: `Generates visual representations of system components and their interactions.
Diagram types:
  static: Shows component instances and their declared dependencies.
  dynamic <analysis_name_or_method>: Shows component interactions for a specific operation (requires trace).`,
	Args: cobra.MinimumNArgs(2), // diagram_type and system_name are required
	Run: func(cmd *cobra.Command, args []string) {
		diagramType := args[0]
		systemName := args[1]
		dynamicTarget := ""
		if diagramType == "dynamic" {
			if len(args) < 3 {
				fmt.Fprintln(os.Stderr, "Error: 'dynamic' diagram type requires an <analysis_name_or_method_call> argument.")
				os.Exit(1)
			}
			dynamicTarget = args[2]
		}

		outputFile, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		if dslFilePath == "" {
			fmt.Fprintln(os.Stderr, "Error: DSL file path must be specified with -f or --file.")
			os.Exit(1)
		}

		fmt.Printf("Generating '%s' diagram for system '%s' from '%s'\n", diagramType, systemName, dslFilePath)
		if dynamicTarget != "" {
			fmt.Printf("Dynamic diagram target: '%s'\n", dynamicTarget)
		}
		fmt.Printf("Format: %s, Output: %s\n", format, outputFile)

		var diagramOutput string // Will store string for dot/mermaid, or JSON string for excalidraw
		var errGen error = nil

		if diagramType == "static" {
			// 1. Load the SDL file
			sdlLoader := loader.NewLoader(nil, nil, 10)
			fileStatus, err := sdlLoader.LoadFile(dslFilePath, "", 0)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading SDL file '%s': %v\n", dslFilePath, err)
				if fileStatus != nil {
					for _, e := range fileStatus.Errors {
						fmt.Fprintln(os.Stderr, "  ", e)
					}
				}
				os.Exit(1)
			}
			if fileStatus.HasErrors() {
				fmt.Fprintf(os.Stderr, "Errors found while loading SDL file '%s':\n", dslFilePath)
				for _, e := range fileStatus.Errors {
					fmt.Fprintln(os.Stderr, "  ", e)
				}
				os.Exit(1)
			}

			astRoot := fileStatus.FileDecl
			if astRoot == nil {
				fmt.Fprintf(os.Stderr, "Error: Parsed AST is nil for '%s'.\n", dslFilePath)
				os.Exit(1)
			}

			// 2. Find SystemDecl
			sysDecl, err := astRoot.GetSystem(systemName)
			if err != nil || sysDecl == nil {
				fmt.Fprintf(os.Stderr, "Error finding system '%s' in '%s': %v\n", systemName, dslFilePath, err)
				os.Exit(1)
			}

			// 3. Populate DiagramNode and DiagramEdge slices
			var diagramNodes []DiagramNode
			var diagramEdges []DiagramEdge
			instanceNameToID := make(map[string]string)

			for _, item := range sysDecl.Body {
				if instDecl, ok := item.(*decl.InstanceDecl); ok {
					nodeID := instDecl.Name.Value
					instanceNameToID[instDecl.Name.Value] = nodeID
					diagramNodes = append(diagramNodes, DiagramNode{
						ID:   nodeID,
						Name: instDecl.Name.Value,
						Type: instDecl.ComponentName.Value,
					})
				}
			}

			for _, item := range sysDecl.Body {
				if instDecl, ok := item.(*decl.InstanceDecl); ok {
					fromNodeID := instanceNameToID[instDecl.Name.Value]
					for _, assignment := range instDecl.Overrides {
						if targetIdent, okIdent := assignment.Value.(*decl.IdentifierExpr); okIdent {
							if toNodeID, isInstance := instanceNameToID[targetIdent.Value]; isInstance {
								diagramEdges = append(diagramEdges, DiagramEdge{
									FromID: fromNodeID,
									ToID:   toNodeID,
									Label:  assignment.Var.Value,
								})
							}
						}
					}
				}
			}

			// 4. Generate output based on format
			switch format {
			case "dot":
				diagramOutput = generateDotOutput(sysDecl.Name.Value, diagramNodes, diagramEdges)
			case "mermaid":
				diagramOutput = generateMermaidOutput(sysDecl.Name.Value, diagramNodes, diagramEdges)
			case "excalidraw":
				diagramOutput, errGen = generateExcalidrawOutput(sysDecl.Name.Value, diagramNodes, diagramEdges)
			// case "svg":
			// 	diagramOutput, errGen = generateSvgOutput(sysDecl.Name.Value, diagramNodes, diagramEdges)
			default:
				fmt.Fprintf(os.Stderr, "Static diagram for format '%s' not supported or placeholder.\n", format)
				os.Exit(1)
			}
			if errGen != nil {
				fmt.Fprintf(os.Stderr, "Error generating %s diagram: %v\n", format, errGen)
				os.Exit(1)
			}

		} else if diagramType == "dynamic" {
			var b bytes.Buffer
			if format == "mermaid" {
				b.WriteString(fmt.Sprintf("sequenceDiagram\n  participant User\n  User->>ServiceA: %s\n  ServiceA->>ServiceB: call\n", dynamicTarget))
				diagramOutput = b.String()
			} else if format == "dot" {
				b.WriteString(fmt.Sprintf("digraph %s_dynamic {\n label=\"Dynamic Diagram for %s (Placeholder)\";\n  User -> ServiceA [label=\"%s\"];\n ServiceA -> ServiceB;\n}", systemName, dynamicTarget, dynamicTarget))
				diagramOutput = b.String()
			} else {
				fmt.Fprintf(os.Stderr, "Dynamic diagram for format '%s' placeholder.\n", format)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: Unknown diagram type '%s'.\n", diagramType)
			os.Exit(1)
		}

		if diagramOutput != "" {
			if outputFile != "" {
				err := os.WriteFile(outputFile, []byte(diagramOutput), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing diagram to %s: %v\n", outputFile, err)
					os.Exit(1)
				}
				fmt.Printf("Diagram content written to %s\n", outputFile)
			} else {
				fmt.Println("\nDiagram Content (stdout):")
				fmt.Println(diagramOutput)
			}
		}
	},
}

func init() {
	AddCommand(diagramCmd)
	diagramCmd.Flags().StringP("output", "o", "", "Output file path for the diagram")
	// Updated to include excalidraw, and eventually svg, png
	diagramCmd.Flags().String("format", "dot", "Output format (dot, mermaid, excalidraw, svg, png)")
}
