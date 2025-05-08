package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/panyam/leetcoach/sdl/parser" // Assuming parser is in decl
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <entity_type>",
	Short: "Lists defined entities within DSL file(s)",
	Long: `Lists components, systems, analyses, or enums defined in the specified DSL file.
Entity types: components, systems, analyses, enums.`,
	Args: cobra.ExactArgs(1), // Expects exactly one argument: entity_type
	Run: func(cmd *cobra.Command, args []string) {
		entityType := args[0]
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

		var items []string
		var dataToMarshal interface{}

		switch entityType {
		case "components":
			fmt.Printf("Components in %s:\n", dslFilePath)
			components, _ := astRoot.GetComponents()
			for _, comp := range components {
				items = append(items, comp.NameNode.Name)
			}
			dataToMarshal = items
		case "systems":
			fmt.Printf("Systems in %s:\n", dslFilePath)
			systems, _ := astRoot.GetSystems()
			for _, sys := range systems {
				items = append(items, sys.NameNode.Name)
			}
			dataToMarshal = items
		case "analyses":
			systemName, _ := cmd.Flags().GetString("system")
			fmt.Printf("Analyses in %s", dslFilePath)
			if systemName != "" {
				fmt.Printf(" (System: %s):\n", systemName)
			} else {
				fmt.Println(" (all systems):")
			}
			// Placeholder: Need to iterate through systems and then their analyze blocks
			items = append(items, "Analysis1_Placeholder", "Analysis2_Placeholder")
			dataToMarshal = items // Update when actual data structure is available
		case "enums":
			fmt.Printf("Enums in %s:\n", dslFilePath)
			enums, _ := astRoot.GetEnums()
			for _, enum := range enums {
				items = append(items, enum.NameNode.Name)
			}
			dataToMarshal = items
		default:
			fmt.Fprintf(os.Stderr, "Error: Unknown entity type '%s'. Valid types: components, systems, analyses, enums.\n", entityType)
			os.Exit(1)
		}

		if outputJSON {
			jsonData, err := json.MarshalIndent(dataToMarshal, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshalling to JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonData))
		} else {
			if len(items) == 0 {
				fmt.Println("  No items found.")
			}
			for _, item := range items {
				fmt.Printf("- %s\n", item)
			}
		}
	},
}

func init() {
	AddCommand(listCmd)
	listCmd.Flags().Bool("json", false, "Output in JSON format")
	listCmd.Flags().StringP("system", "s", "", "Filter analyses by system name (for 'analyses' type)")
	// Note: The global -f/--file flag is inherited from rootCmd
}

// Helper methods to be added to decl.FileDecl (or similar AST structure access)
// These are placeholders for what decl.FileDecl should provide after Resolve()
// func (f *decl.FileDecl) GetComponents() []*decl.ComponentDecl { /* ... */ return f.componentsSlice }
// func (f *decl.FileDecl) GetSystems()    []*decl.SystemDecl    { /* ... */ return f.systemsSlice }
// func (f *decl.FileDecl) GetEnums()      []*decl.EnumDecl      { /* ... */ return f.enumsSlice }
// func (f *decl.FileDecl) GetAnalyses(systemNameFilter string) []AnalysisInfo { /* ... */ }
