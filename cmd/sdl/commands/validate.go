package commands

import (
	"fmt"
	"os"

	"github.com/panyam/leetcoach/sdl/parser" // Assuming parser is in decl
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <dsl_file_path...>",
	Short: "Parses and semantically checks DSL file(s)",
	Long: `The validate command parses one or more DSL files to check for syntactic
correctness and basic semantic validity. It does not run any simulations.`,
	Args: cobra.MinimumNArgs(1), // Require at least one file path
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Validating DSL files:")
		allValid := true
		for _, filePath := range args {
			fmt.Printf("- %s\n", filePath)

			file, err := os.Open(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error opening %s: %v\n", filePath, err)
				allValid = false
				continue
			}
			defer file.Close()

			// Assume decl.Parse is available
			_, astRoot, err := parser.Parse(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error parsing %s: %v\n", filePath, err)
				allValid = false
				continue
			}

			// Perform semantic checks (e.g., vm.LoadFile(ast) or astRoot.Resolve())
			// For now, let's assume FileDecl.Resolve() handles initial semantic checks.
			if astRoot != nil { // Check if parsing returned a valid AST
				err = astRoot.Resolve()
				if err != nil {
					fmt.Fprintf(os.Stderr, "  Error validating semantics in %s: %v\n", filePath, err)
					allValid = false
					continue
				}
				fmt.Printf("  Successfully validated %s\n", filePath)
			} else {
				// This case should ideally be caught by parser error, but defensive check.
				fmt.Fprintf(os.Stderr, "  Parsing %s did not return a valid AST.\n", filePath)
				allValid = false
			}
		}

		if !allValid {
			fmt.Fprintln(os.Stderr, "One or more files failed validation.")
			os.Exit(1)
		}
		fmt.Println("All specified files validated successfully.")
	},
}

func init() {
	AddCommand(validateCmd)
	// validateCmd.Flags().BoolP("strict", "s", false, "Enable stricter validation checks")
}
