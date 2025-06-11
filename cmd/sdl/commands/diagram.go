package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader"
	"github.com/panyam/sdl/runtime"
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

	var diagramOutput string
	switch format {
	case "mermaid":
		diagramOutput = generateMermaidSequenceDiagram(&traceData)
	default:
		fmt.Fprintf(os.Stderr, "Dynamic diagram for format '%s' not supported. Choose 'mermaid'.\n", format)
		os.Exit(1)
	}

	writeOutput(outputFile, diagramOutput)
}

// getParticipantAndMethod extracts the participant and method from a target string like "instance.method".
func getParticipantAndMethod(target string) (participant, method string) {
	parts := strings.SplitN(target, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return target, ""
}

func generateMermaidSequenceDiagram(trace *runtime.TraceData) string {
	var b bytes.Buffer
	b.WriteString("sequenceDiagram\n")

	// 1. Discover all participants and order them correctly
	participantSet := make(map[string]bool)
	participantSet["User"] = true // The initiator is always present

	eventMap := make(map[int]*runtime.TraceEvent)
	for _, event := range trace.Events {
		eventMap[event.ID] = event
		if event.Kind == runtime.EventEnter {
			participant, _ := getParticipantAndMethod(event.Target)
			if participant != "" {
				participantSet[participant] = true
			}
		}
	}

	// Create a sorted list of participants with "User" first
	participantList := make([]string, 0, len(participantSet))
	for p := range participantSet {
		if p != "User" {
			participantList = append(participantList, p)
		}
	}
	sort.Strings(participantList)
	participantList = append([]string{"User"}, participantList...)

	// 2. Declare all participants in the correct order
	for _, p := range participantList {
		b.WriteString(fmt.Sprintf("  participant %s\n", p))
	}
	b.WriteString("\n")

	// 3. Process events to generate the sequence
	// Map of active event IDs to their "from" participant
	activeEvents := make(map[int]string)
	activeEvents[0] = "User" // Root event's parent is the User

	for _, event := range trace.Events {
		from, parentFound := activeEvents[event.ParentID]
		if !parentFound {
			from = "User" // Default to User if parent not found (should not happen for non-root)
		}

		if event.Kind == runtime.EventEnter {
			to, method := getParticipantAndMethod(event.Target)
			if to == "self" {
				to = from // A call to "self" is a call to the current participant
			}

			b.WriteString(fmt.Sprintf("  %s->>%s: %s(%s)\n", from, to, method, strings.Join(event.Arguments, ", ")))
			b.WriteString(fmt.Sprintf("  activate %s\n", to))
			activeEvents[event.ID] = to // The callee is now the "from" for its children
		} else if event.Kind == runtime.EventExit {
			// Find the corresponding Enter event to know who to deactivate
			// This linear scan is inefficient but simple for now.
			var enterTarget string
			for _, e := range trace.Events {
				if e.Kind == runtime.EventEnter && e.ParentID == event.ParentID && e.Timestamp < event.Timestamp {
					// This is a heuristic match, a real solution needs to link exit to enter
					enterTarget = e.Target
					break
				}
			}
			if enterTarget != "" {
				to, _ := getParticipantAndMethod(enterTarget)
				if to == "self" {
					to = from
				}
				b.WriteString(fmt.Sprintf("  deactivate %s\n", to))
			}
		}
	}

	return b.String()
}

func generateStaticDiagram(systemName, outputFile, format string) {
	var diagramOutput string
	var errGen error

	// ... (rest of the static diagram logic remains the same)
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
	case "svg":
		diagramOutput, errGen = generateSvgOutput(sysDecl.Name.Value, diagramNodes, diagramEdges)
	default:
		fmt.Fprintf(os.Stderr, "Static diagram for format '%s' not supported or placeholder.\n", format)
		os.Exit(1)
	}
	if errGen != nil {
		fmt.Fprintf(os.Stderr, "Error generating %s diagram: %v\n", format, errGen)
		os.Exit(1)
	}

	writeOutput(outputFile, diagramOutput)
}

func writeOutput(outputFile, content string) {
	if content == "" {
		return // Nothing to write
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
	// Updated to include excalidraw and svg
	diagramCmd.Flags().String("format", "dot", "Output format (dot, mermaid, excalidraw, svg)")
}
