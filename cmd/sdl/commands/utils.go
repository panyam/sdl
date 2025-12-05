package commands

import (
	"fmt"
	"io"

	"github.com/panyam/sdl/lib/decl"
	"github.com/panyam/sdl/lib/parser"
)

// SDLParserAdapter adapts the existing parser function to the loader.Parser interface.
type SDLParserAdapter struct{}

func (pa *SDLParserAdapter) Parse(input io.Reader, sourceName string) (*decl.FileDecl, error) {
	// The existing parser.Parse might not use sourceName directly,
	// but the lexer it creates might use it indirectly if errors occur early.
	// Or parser.Parse could be modified to accept it.
	_, ast, err := parser.Parse(input) // Assuming parser.Parse has func(io.Reader) (*Lexer, *FileDecl, error) signature
	// We ignore the lexer instance returned by the current parser.Parse
	if err != nil {
		// Wrap the error to include sourceName if the parser didn't already
		return nil, fmt.Errorf("in '%s': %w", sourceName, err)
	}
	if ast == nil {
		// Handle cases where parser succeeds but returns nil AST
		// (shouldn't happen ideally)
		return nil, fmt.Errorf("parser succeeded but returned nil AST for '%s'", sourceName)
	}
	return ast, nil
}
