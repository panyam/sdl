package loader

import (
	"log"
	"testing"

	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
)

func TestBitly(t *testing.T) {
	l := NewLoader(nil, nil, 10) // Max depth 10

	sourceFiles := []string{"../examples/disk.sdl", "../examples/db.sdl", "../examples/common.sdl", "../examples/bitly.sdl"}
	// sourceFiles := []string{"../examples/disk.sdl"}
	for _, f := range sourceFiles {
		fs, err := l.LoadFile(f, "", 0)
		if err != nil {
			log.Println("Error loading file: ", f, err)
			continue
		}
		l.Validate(fs)
		if fs.HasErrors() {
			log.Printf("\nError Validating File %s\n", fs.FullPath)
			fs.PrintErrors()
		} else {
			log.Printf("\nFile %s - Validated Successfully at: %v\n", fs.FullPath, fs.LastValidated)
		}
	}
}
