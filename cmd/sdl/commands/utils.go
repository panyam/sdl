package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/panyam/leetcoach/sdl/decl"
	"github.com/panyam/leetcoach/sdl/parser"
)

func parse(input io.Reader) (*decl.FileDecl, error) {
	file := &decl.FileDecl{}
	l := parser.NewLexer(input)
	p := parser.NewLLParser(l)
	err := p.Parse(file)
	return file, err
}

func parseFile(filePath string) (*decl.FileDecl, error) {
	fmt.Printf("- %s\n", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error opening %s: %v\n", filePath, err)
		return nil, err
	}
	defer file.Close()
	fileDecl, err := parse(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error parsing %s: %v\n", filePath, err)
	}
	return fileDecl, err
}
