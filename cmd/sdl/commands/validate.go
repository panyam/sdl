package commands

import (
	"fmt"

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
		sdlLoader := loader.NewLoader(nil, nil, 10) // Max depth 10
		sdlLoader.LoadFilesAndValidate(args...)
	},
}

func init() {
	AddCommand(validateCmd)
	// validateCmd.Flags().BoolP("strict", "s", false, "Enable stricter validation checks")
}
