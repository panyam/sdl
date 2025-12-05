package commands

import (
	"encoding/json"
	"fmt"
	"os"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/lib/decl"
	"github.com/panyam/sdl/lib/loader"
	"github.com/panyam/sdl/lib/runtime"
	"github.com/panyam/sdl/lib/viz"
	"github.com/spf13/cobra"
)

var diagramCmd = &cobra.Command{
	Use:   "diagram <diagram_type> [system_name]",
	Short: "Generates diagrams of system structure or behavior",
	Long: `Generates visual representations of system components and their interactions.
Diagram types:
  static: Shows component instances and their declared dependencies. Requires a <system_name>.
  dynamic: Shows component interactions from a trace file. Requires the --from flag.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		diagramType := args[0]
		fromFile, _ := cmd.Flags().GetString("from")
		outputFile, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		if diagramType == "dynamic" {
			if fromFile == "" {
				fmt.Fprintln(os.Stderr, "Error: 'dynamic' diagrams require a trace file specified with --from.")
				os.Exit(1)
			}
			fmt.Printf("Generating 'dynamic' diagram from trace file '%s'\n", fromFile)
			generateDynamicDiagram(fromFile, outputFile, format)

		} else if diagramType == "static" {
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "Error: 'static' diagram requires a <system_name> argument.")
				os.Exit(1)
			}
			systemName := args[1]
			if dslFilePath == "" {
				fmt.Fprintln(os.Stderr, "Error: DSL file path must be specified with -f or --file for a static diagram.")
				os.Exit(1)
			}
			fmt.Printf("Generating 'static' diagram for system '%s' from '%s'\n", systemName, dslFilePath)
			generateStaticDiagram(systemName, outputFile, format)
		} else {
			fmt.Fprintf(os.Stderr, "Error: Unknown diagram type '%s'. Choose 'static' or 'dynamic'.\n", diagramType)
			os.Exit(1)
		}
	},
}

func generateDynamicDiagram(fromFile, outputFile, format string) {
	data, err := os.ReadFile(fromFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading trace file %s: %v\n", fromFile, err)
		os.Exit(1)
	}

	var traceData runtime.TraceData
	if err := json.Unmarshal(data, &traceData); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON from trace file %s: %v\n", fromFile, err)
		os.Exit(1)
	}

	var generator viz.SequenceDiagramGenerator
	switch format {
	case "mermaid":
		generator = &viz.MermaidSequenceGenerator{}
	case "dot":
		generator = &viz.DotTraceGenerator{}
	default:
		fmt.Fprintf(os.Stderr, "Dynamic diagram for format '%s' not supported. Choose 'mermaid' or 'dot'.\n", format)
		os.Exit(1)
	}

	diagramOutput, err := generator.Generate(&traceData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating dynamic diagram: %v\n", err)
		os.Exit(1)
	}

	writeOutput(outputFile, diagramOutput)
}

func generateStaticDiagram(systemName, outputFile, format string) {
	// 1. Load the SDL file
	sdlLoader := loader.NewLoader(nil, nil, 10)
	fileStatus, err := sdlLoader.LoadFile(dslFilePath, "", 0)
	if err != nil || fileStatus.HasErrors() {
		fmt.Fprintf(os.Stderr, "Error loading or parsing SDL file '%s':\n", dslFilePath)
		if fileStatus != nil {
			fileStatus.PrintErrors()
		} else {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	astRoot := fileStatus.FileDecl
	sysDecl, err := astRoot.GetSystem(systemName)
	if err != nil || sysDecl == nil {
		fmt.Fprintf(os.Stderr, "Error finding system '%s' in '%s': %v\n", systemName, dslFilePath, err)
		os.Exit(1)
	}

	// 2. Populate proto DiagramNode and DiagramEdge slices
	var nodes []*protos.DiagramNode
	var edges []*protos.DiagramEdge
	instanceNameToID := make(map[string]string)

	for _, item := range sysDecl.Body {
		if instDecl, ok := item.(*decl.InstanceDecl); ok {
			nodeID := instDecl.Name.Value
			instanceNameToID[nodeID] = nodeID
			nodes = append(nodes, &protos.DiagramNode{
				Id:   nodeID,
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
						edges = append(edges, &protos.DiagramEdge{
							FromId: fromNodeID,
							ToId:   toNodeID,
							Label:  assignment.Var.Value,
						})
					}
				}
			}
		}
	}

	// 3. Get the correct generator based on format
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
		fmt.Fprintf(os.Stderr, "Static diagram for format '%s' is not supported.\n", format)
		os.Exit(1)
	}

	// 4. Create proto SystemDiagram and generate output
	diagram := &protos.SystemDiagram{
		SystemName: systemName,
		Nodes:      nodes,
		Edges:      edges,
	}

	diagramOutput, err := generator.Generate(diagram)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating %s diagram: %v\n", format, err)
		os.Exit(1)
	}

	writeOutput(outputFile, diagramOutput)
}

func writeOutput(outputFile, content string) {
	if content == "" {
		return
	}
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(content), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing diagram to %s: %v\n", outputFile, err)
			os.Exit(1)
		}
		fmt.Printf("Diagram content written to %s\n", outputFile)
	} else {
		fmt.Println("\nDiagram Content (stdout):")
		fmt.Println(content)
	}
}

func init() {
	AddCommand(diagramCmd)
	diagramCmd.Flags().StringP("output", "o", "", "Output file path for the diagram")
	diagramCmd.Flags().String("from", "", "Path to a JSON trace file (for dynamic diagrams)")
	diagramCmd.Flags().String("format", "dot", "Output format (dot, mermaid, excalidraw, svg)")
}
