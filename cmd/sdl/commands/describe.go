package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/panyam/sdl/parser" // Assuming parser is in decl
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe <entity_type> <entity_name>",
	Short: "Shows detailed information about a specific entity",
	Long: `Describes a component definition, system configuration, or analysis block
from the specified DSL file.
Entity types: component, system, analysis.`,
	Args: cobra.ExactArgs(2), // Expects entity_type and entity_name
	Run: func(cmd *cobra.Command, args []string) {
		entityType := args[0]
		entityName := args[1]
		outputJSON, _ := cmd.Flags().GetBool("json")

		if dslFilePath == "" {
			fmt.Fprintln(os.Stderr, "Error: DSL file path must be specified with -f or --file.")
			os.Exit(1)
		}

		file, err := os.Open(dslFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", dslFilePath, err)
			os.Exit(1)
		}
		defer file.Close()

		_, astRoot, err := parser.Parse(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", dslFilePath, err)
			os.Exit(1)
		}
		if err := astRoot.Resolve(); err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving AST for %s: %v\n", dslFilePath, err)
			os.Exit(1)
		}

		var description any // To hold the data for JSON output or formatted print

		switch entityType {
		case "component":
			comp, err := astRoot.GetComponent(entityName)
			if err != nil || comp == nil {
				fmt.Fprintf(os.Stderr, "Error: Component '%s' not found in %s.\n", entityName, dslFilePath)
				os.Exit(1)
			}
			description = comp // The AST node itself can be marshalled or pretty-printed
			if !outputJSON {
				fmt.Printf("Description for Component '%s':\n", entityName)
				// Placeholder for pretty printing component details
				fmt.Printf("  Name: %s\n", comp.Name.Value)
				fmt.Println("  Params:")
				params, _ := comp.Params()
				for _, p := range params { // Assuming GetParams()
					fmt.Printf("    - %s: %s\n", p.Name.Value, p.TypeDecl.Name)
				}
				fmt.Println("  Uses:")
				uses, _ := comp.Dependencies()
				for _, u := range uses { // Assuming GetUses()
					fmt.Printf("    - %s: %s\n", u.Name.Value, u.ComponentName.Value)
				}
				fmt.Println("  Methods:")
				methods, _ := comp.Methods()
				for _, m := range methods {
					fmt.Printf("    - %s()\n", m.Name.Value)
				}
			}
		case "system":
			sys, err := astRoot.GetSystem(entityName)
			if err != nil || sys == nil {
				fmt.Fprintf(os.Stderr, "Error: System '%s' not found in %s.\n", entityName, dslFilePath)
				os.Exit(1)
			}
			description = sys
			if !outputJSON {
				fmt.Printf("Description for System '%s':\n", entityName)
				// Placeholder for pretty printing system details
				fmt.Printf("  Name: %s\n", sys.Name.Value)
				fmt.Println("  Instances:")
				// Iterate through sys.Body for InstanceDecls
				// for _, item := range sys.Body {
				// 	if inst, ok := item.(*decl.InstanceDecl); ok {
				// 		fmt.Printf("    - %s: %s\n", inst.Name.Value, inst.ComponentType.Name)
				// 	}
				// }
				fmt.Println("  (Instance details placeholder)")
			}
		case "analysis":
			// Placeholder: Need to find the specific analysis block
			fmt.Printf("Description for Analysis '%s' (Placeholder):\n", entityName)
			description = map[string]string{"name": entityName, "status": "placeholder"}
		default:
			fmt.Fprintf(os.Stderr, "Error: Unknown entity type '%s'. Valid types: component, system, analysis.\n", entityType)
			os.Exit(1)
		}

		if outputJSON {
			jsonData, err := json.MarshalIndent(description, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshalling to JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonData))
		}
	},
}

func init() {
	AddCommand(describeCmd)
	describeCmd.Flags().Bool("json", false, "Output in JSON format")
}

// Placeholder accessors to be added to decl.ComponentDecl
// func (cd *decl.ComponentDecl) GetParams()  []*decl.ParamDecl  { /* ... */ }
// func (cd *decl.ComponentDecl) GetUses()    []*decl.UsesDecl   { /* ... */ }
// func (cd *decl.ComponentDecl) GetMethods() []*decl.MethodDecl { /* ... */ }
