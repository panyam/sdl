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
func (f *FileInstance) NewComponent(name string) (*ComponentInstance, Value, error) {
	compDecl, err := f.GetComponentDecl(name)
	if compDecl == nil || err != nil {
		return nil, Nil, err
	}
	return NewComponentInstance(f, compDecl)
}

// GetComponentDecl returns the ComponentDecl for the given name even if it is an import by resolving to the original source
func (f *FileInstance) GetComponentDecl(name string) (*ComponentDecl, error) {
	def, err := f.Decl.GetDefinition(name)
	if err != nil {
		log.Println("error getting component definition: ", err)
		return nil, err
	}

	compDecl, ok := def.(*decl.ComponentDecl)
	if !ok {
		// see if it is an import
		importDecl, ok := def.(*decl.ImportDecl)
		if ok {
			importedFS, err := f.Runtime.Loader.LoadFile(importDecl.ResolvedFullPath, f.Decl.FullPath, 0)
			if err != nil {
				return nil, err
			}
			def, _ = importedFS.FileDecl.GetDefinition(importDecl.ImportedItem.Value)
			compDecl, _ = def.(*decl.ComponentDecl)
		}
	}

	if compDecl == nil {
		log.Println("error getting component definition: ", name, " is not a component")
		return nil, fmt.Errorf("definition is not a component")
	}
	return compDecl, nil
}

// GetEnumDecl returns the EnumDecl for the given name even if it is an import by resolving to the original source
func (f *FileInstance) GetEnumDecl(name string) (*EnumDecl, error) {
	def, err := f.Decl.GetDefinition(name)
	if err != nil {
		log.Println("error getting enum definition: ", err)
		return nil, err
	}

	enumDecl, ok := def.(*decl.EnumDecl)
	if !ok {
		// see if it is an import
		importDecl, ok := def.(*decl.ImportDecl)
		if ok {
			importedFS, err := f.Runtime.Loader.LoadFile(importDecl.ResolvedFullPath, f.Decl.FullPath, 0)
			if err != nil {
				return nil, err
			}
			def, _ = importedFS.FileDecl.GetDefinition(importDecl.ImportedItem.Value)
			enumDecl, _ = def.(*decl.EnumDecl)
		}
	}

	if enumDecl == nil {
		log.Println("error getting enum definition: ", name, " is not a enum")
		return nil, fmt.Errorf("definition is not a enum")
	}
	return enumDecl, nil
}
