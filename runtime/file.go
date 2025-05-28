package runtime

import (
	"fmt"
	"log"

	"github.com/panyam/sdl/decl"
)

// Runtime information about a Filedecl and its initial environment
type FileInstance struct {
	Runtime *Runtime
	Decl    *FileDecl
	Env     *Env[Value]
}

func NewFileInstance(r *Runtime, file *FileDecl) *FileInstance {
	inst := &FileInstance{Runtime: r, Decl: file}
	inst.Env = decl.NewEnv[Value](nil)
	return inst
}

// Initialize a new system with the given name.
// Returns nil if system name is invalid
func (f *FileInstance) NewSystem(systemName string) *SystemInstance {
	system, err := f.Decl.GetSystem(systemName)
	if err != nil {
		log.Println("error getting system: ", err)
		return nil
	}

	return NewSystemInstance(f, system)
}

// Creates a new instance of a component
// Note that a component can be native or user defined.
// We want the same semantics regardless
func (f *FileInstance) NewComponent(name string) (*ComponentInstance, error) {
	def, err := f.Decl.GetDefinition(name)
	if err != nil {
		log.Println("error getting component definition: ", err)
		return nil, err
	}

	compDecl := def.(*decl.ComponentDecl)
	if compDecl == nil {
		log.Println("error getting component definition: ", name, " is not a component")
		return nil, fmt.Errorf("definition is not a component")
	}

	return NewComponentInstance(f, compDecl)
}
