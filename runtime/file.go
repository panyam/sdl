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
	env     *Env[Value]
}

func NewFileInstance(r *Runtime, file *FileDecl) *FileInstance {
	inst := &FileInstance{Runtime: r, Decl: file}
	return inst
}

func (f *FileInstance) Env() *Env[Value] {
	if f.env == nil {
		f.env = decl.NewEnv[Value](nil)
		// Now push all the methods etc here
		defns, err := f.Decl.AllDefinitions()
		ensureNoErr(err)
		for defname, defn := range defns {
			var methodDecl *MethodDecl
			if n, ok := defn.(*MethodDecl); ok {
				methodDecl = n
			} else if n, ok := defn.(*ImportDecl); ok {
				// log.Println("TODO - How to handle imports of methods here??: ", n.ResolvedItem)
				if _, ok := n.ResolvedItem.(*ImportDecl); ok {
					panic("where should we handle recursive imports - in the loader?")
				}
				if n2, ok := n.ResolvedItem.(*MethodDecl); ok {
					methodDecl = n2
				}
			}
			if methodDecl != nil {
				methodType := decl.MethodType(methodDecl)
				methodVal := &decl.MethodValue{
					Method: methodDecl, SavedEnv: f.env.Push(), IsNative: methodDecl.IsNative,
				}
				val, err := NewValue(methodType, methodVal)
				ensureNoErr(err)
				f.env.Set(defname, val)
			}
		}
	}
	return f.env.Push()
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
