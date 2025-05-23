package commands

import (
	"fmt"
	"os"

	// Assuming parser is in decl
	"github.com/panyam/sdl/loader"
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
		sdlParser := &SDLParserAdapter{}
		fileResolver := loader.NewDefaultFileResolver()
		sdlLoader := loader.NewLoader(sdlParser, fileResolver, 10) // Max depth 10
		allValid, results := sdlLoader.LoadFiles(args...)
		if allValid {
			fmt.Println("All specified files validated successfully.")
		} else {
			fmt.Fprintln(os.Stderr, "One or more files failed validation.")
			for path, result := range results {
				fmt.Printf("Errors in: %s\n", path)
				for _, err := range result.Errors {
					fmt.Println("    ", err)
				}
			}
		}
	},
}

func init() {
	AddCommand(validateCmd)
	// validateCmd.Flags().BoolP("strict", "s", false, "Enable stricter validation checks")
}
