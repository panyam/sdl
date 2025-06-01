package runtime

import (
	"fmt"
	"log"

	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/loader"
)

type Runtime struct {
	Loader        *loader.Loader
	NativeObjects []any
	fileInstances map[string]*FileInstance
}

func NewRuntime(loader *loader.Loader) (r *Runtime) {
	r = &Runtime{Loader: loader, fileInstances: make(map[string]*FileInstance)}
	return
}

func (r *Runtime) CreateNativeComponent(compDecl *ComponentDecl) NativeObject {
	name := compDecl.Name.Value
	if name == "HashIndex" {
		return &NativeWrapper{name, (&components.HashIndex{}).Init()}
	}
	if name == "TestNative" {
		return &NativeWrapper{name, NewTestNative()}
	}
	panic(fmt.Sprintf("Native component not registered: %s", name))
}

// Gets the initial run time environment for a File which would include its parameters and component creators
func (r *Runtime) LoadFile(filePath string) *FileInstance {
	if env, ok := r.fileInstances[filePath]; ok && env != nil {
		return env
	}

	fs, err := r.Loader.LoadFile(filePath, "", 0)
	if err != nil {
		panic(err)
	}
	r.Loader.Validate(fs)
	if fs.HasErrors() {
		log.Printf("\nError Validating File %s\n", fs.FullPath)
		fs.PrintErrors()
	} else {
		log.Printf("\nFile %s - Validated Successfully at: %v\n", fs.FullPath, fs.LastValidated)
	}

	file := fs.FileDecl
	out := NewFileInstance(r, file)
	r.fileInstances[fs.FullPath] = out
	return out
}
