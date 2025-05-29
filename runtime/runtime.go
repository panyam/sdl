package runtime

import (
	"github.com/panyam/sdl/loader"
)

type Runtime struct {
	Loader           *loader.Loader
	NativeComponents []any
	fileInstances    map[string]*FileInstance
}

func NewRuntime(loader *loader.Loader) (r *Runtime) {
	r = &Runtime{Loader: loader, fileInstances: make(map[string]*FileInstance)}
	return
}

func (r *Runtime) CreateNativeComponent(name string) *NativeComponent {
	return nil
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

	file := fs.FileDecl
	out := NewFileInstance(r, file)
	r.fileInstances[fs.FullPath] = out
	return out
}
