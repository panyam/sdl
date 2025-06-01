package commands

import (
	"bytes"
	"fmt"
	"os"

	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader" // Added loader import
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

		var diagramContent string

		if diagramType == "static" {
			// 1. Load the SDL file
			sdlLoader := loader.NewLoader(nil, nil, 10) // Use default parser & resolver, depth 10
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
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error finding system '%s' in '%s': %v\n", systemName, dslFilePath, err)
				os.Exit(1)
			}
			if sysDecl == nil {
				fmt.Fprintf(os.Stderr, "Error: System '%s' not found in '%s'.\n", systemName, dslFilePath)
				os.Exit(1)
			}

			// 3. For 'static':
			//    - Analyze SystemDecl for InstanceDecls and their Overrides.
			//    - Generate DOT or Mermaid content.
			var b bytes.Buffer
			instanceNames := make(map[string]string) // Map instance name to component type for node labels

			if format == "dot" {
				b.WriteString(fmt.Sprintf("digraph \"%s\" {\n", sysDecl.Name.Value))
				b.WriteString("  rankdir=LR;\n") // Left to right ranking
				b.WriteString(fmt.Sprintf("  label=\"Static Diagram for System: %s\";\n", sysDecl.Name.Value))
				b.WriteString("  node [shape=record];\n")

				// First pass: define all nodes (instances)
				for _, item := range sysDecl.Body {
					if instDecl, ok := item.(*decl.InstanceDecl); ok {
						instanceNames[instDecl.Name.Value] = instDecl.ComponentName.Value
						// Define the node with its type
						b.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n(%s)\"];\n", instDecl.Name.Value, instDecl.Name.Value, instDecl.ComponentName.Value))
					}
				}

				// Second pass: define edges from overrides
				for _, item := range sysDecl.Body {
					if instDecl, ok := item.(*decl.InstanceDecl); ok {
						for _, assignment := range instDecl.Overrides {
							if targetIdent, okIdent := assignment.Value.(*decl.IdentifierExpr); okIdent {
								// Check if the targetIdent is one of the defined instances
								if _, isInstance := instanceNames[targetIdent.Value]; isInstance {
									b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", instDecl.Name.Value, targetIdent.Value, assignment.Var.Value))
								}
							}
						}
					}
				}
				b.WriteString("}\n")
			} else if format == "mermaid" {
				b.WriteString("graph TD;\n")
				b.WriteString(fmt.Sprintf("  subgraph System %s\n", sysDecl.Name.Value))
				// First pass: define all nodes (instances)
				for _, item := range sysDecl.Body {
					if instDecl, ok := item.(*decl.InstanceDecl); ok {
						instanceNames[instDecl.Name.Value] = instDecl.ComponentName.Value
						b.WriteString(fmt.Sprintf("    %s[\"%s (%s)\"];\n", instDecl.Name.Value, instDecl.Name.Value, instDecl.ComponentName.Value))
					}
				}
				// Second pass: define edges
				for _, item := range sysDecl.Body {
					if instDecl, ok := item.(*decl.InstanceDecl); ok {
						for _, assignment := range instDecl.Overrides {
							if targetIdent, okIdent := assignment.Value.(*decl.IdentifierExpr); okIdent {
								if _, isInstance := instanceNames[targetIdent.Value]; isInstance {
									b.WriteString(fmt.Sprintf("    %s -- \"%s\" --> %s;\n", instDecl.Name.Value, assignment.Var.Value, targetIdent.Value))
								}
							}
						}
					}
				}
				b.WriteString("  end\n")
			} else {
				fmt.Fprintf(os.Stderr, "Static diagram for format '%s' placeholder or not supported by this basic implementation.\n", format)
			}
			diagramContent = b.String()

		} else if diagramType == "dynamic" {
			// Placeholder for dynamic diagrams
			if format == "mermaid" {
				diagramContent = fmt.Sprintf("sequenceDiagram\n  participant User\n  User->>ServiceA: %s\n  ServiceA->>ServiceB: call\n", dynamicTarget)
			} else if format == "dot" {
				diagramContent = fmt.Sprintf("digraph %s_dynamic {\n label=\"Dynamic Diagram for %s (Placeholder)\";\n  User -> ServiceA [label=\"%s\"];\n ServiceA -> ServiceB;\n}", systemName, dynamicTarget, dynamicTarget)
			} else {
				fmt.Fprintf(os.Stderr, "Dynamic diagram for format '%s' placeholder.\n", format)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: Unknown diagram type '%s'.\n", diagramType)
			os.Exit(1)
		}

		if diagramContent != "" {
			if outputFile != "" {
				err := os.WriteFile(outputFile, []byte(diagramContent), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing diagram to %s: %v\n", outputFile, err)
					os.Exit(1)
				}
				fmt.Printf("Diagram content written to %s\n", outputFile)
			} else {
				fmt.Println("\nDiagram Content (stdout):")
				fmt.Println(diagramContent)
			}
		}
	},
}

func init() {
	AddCommand(diagramCmd)
	diagramCmd.Flags().StringP("output", "o", "", "Output file path for the diagram")
	diagramCmd.Flags().String("format", "dot", "Output format (dot, mermaid, png, svg)")
}
