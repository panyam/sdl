package decl

import (
	"fmt"
	"log"
	"strings"
)

// FileDecl represents the top-level node of a parsed DSL file.
type FileDecl struct {
	NodeInfo
	FullPath     string
	Declarations []Node // ComponentDecl, SystemDecl, OptionsDecl, EnumDecl, ImportDecl

	// Resolved values so we can work with processed/loaded values instead of resolving
	// Identify expressions etc
	resolved       bool
	allDefinitions map[string]Node // All definitions by name, including components, enums, systems, imports
	components     map[string]*ComponentDecl
	enums          map[string]*EnumDecl
	imports        map[string]*ImportDecl
	aggregators    map[string]*AggregatorDecl
	nativeMethods  map[string]*MethodDecl
	importList     []*ImportDecl // Keep original list for iteration order if needed
	systems        map[string]*SystemDecl
}

func (f *FileDecl) PrettyPrint(cp CodePrinter) {
	for _, n := range f.Declarations {
		n.PrettyPrint(cp)
		cp.Println("")
	}
}

// Get a map of the all the components encountered in this FileDecl
func (f *FileDecl) GetComponents() (out map[string]*ComponentDecl, err error) {
	err = f.Resolve()
	out = f.components
	return
}

// Get a particular component by name in this FileDecl
func (f *FileDecl) GetComponent(name string) (out *ComponentDecl, err error) {
	components, err := f.GetComponents()
	if err == nil {
		out = components[name]
	}
	return
}

// Get a map of the all the enums encountered in this FileDecl
func (f *FileDecl) GetEnums() (out map[string]*EnumDecl, err error) {
	err = f.Resolve()
	out = f.enums
	return
}

// Get a particular enum by name in this FileDecl
func (f *FileDecl) GetEnum(name string) (out *EnumDecl, err error) {
	enums, err := f.GetEnums()
	if err == nil {
		out = enums[name]
	}
	return
}

// Get a map of the all the systems encountered in this FileDecl
func (f *FileDecl) GetSystems() (out map[string]*SystemDecl, err error) {
	err = f.Resolve()
	out = f.systems
	return
}

// Get a particular system by name in this FileDecl
func (f *FileDecl) GetSystem(name string) (out *SystemDecl, err error) {
	systems, err := f.GetSystems()
	if err == nil {
		out = systems[name]
	}
	return
}

// Get a map of the all the imports encountered in this FileDecl
func (f *FileDecl) Imports() (map[string]*ImportDecl, error) {
	if err := f.Resolve(); err != nil {
		return nil, err
	}
	return f.imports, nil
}

func (f *FileDecl) GetAggregator(name string) (out *AggregatorDecl, err error) {
	aggs, err := f.Aggregators()
	if err == nil {
		out = aggs[name]
	}
	return
}

// Get a map of the all the aggregators encountered in this FileDecl
func (f *FileDecl) Aggregators() (map[string]*AggregatorDecl, error) {
	if err := f.Resolve(); err != nil {
		return nil, err
	}
	return f.aggregators, nil
}

// Get a map of all the native methods encountered in this FileDecl
func (f *FileDecl) GetNativeMethods() (out map[string]*MethodDecl, err error) {
	err = f.Resolve()
	out = f.nativeMethods
	return
}

// Get a particular native method by name in this FileDecl
func (f *FileDecl) GetNativeMethod(name string) (out *MethodDecl, err error) {
	methods, err := f.GetNativeMethods()
	if err == nil {
		out = methods[name]
	}
	return
}

// Called to resolve specific AST aspects out of the parse tree
func (f *FileDecl) Resolve() error {
	if f == nil {
		return fmt.Errorf("cannot load nil file")
	}
	if f.resolved {
		return nil
	}
	// Add initializers for other registries (Enums, Options) if they exist

	// log.Printf("Loading definitions from File AST...")
	for _, decl := range f.Declarations {
		switch node := decl.(type) {
		case *ComponentDecl:
			// Process and register the component definition
			err := node.Resolve() // Use a helper function
			if err != nil {
				return fmt.Errorf("error processing component '%s' at pos %d: %w", node.Name.Value, node.Pos(), err)
			}
			if err := f.RegisterComponent(node); err != nil {
				return err
			}
			if err := f.RegisterDefinition(node.Name.Value, node); err != nil {
				return fmt.Errorf("error registering definition '%s': %w", node.Name.Value, err)
			}
		case *MethodDecl:
			if err := f.RegisterNativeMethod(node); err != nil {
				return err
			}
			if err := f.RegisterDefinition(node.Name.Value, node); err != nil {
				return fmt.Errorf("error registering definition '%s': %w", node.Name.Value, err)
			}
		case *SystemDecl:
			// Store the SystemDecl AST by name for later execution
			if err := f.RegisterSystem(node); err != nil {
				return err
			}
			if err := f.RegisterDefinition(node.Name.Value, node); err != nil {
				return fmt.Errorf("error registering definition '%s': %w", node.Name.Value, err)
			}
		case *EnumDecl:
			if err := f.RegisterEnum(node); err != nil {
				return err
			}
			if err := f.RegisterDefinition(node.Name.Value, node); err != nil {
				return fmt.Errorf("error registering definition '%s': %w", node.Name.Value, err)
			}

		case *OptionsDecl:
			log.Printf("Found OptionsDecl (TODO: Implement processing)")

		case *ImportDecl:
			if err := f.RegisterImport(node); err != nil {
				return err
			}
			if err := f.RegisterDefinition(node.Alias.Value, node); err != nil {
				return fmt.Errorf("error registering definition '%s': %w", node.Alias.Value, err)
			}

		case *AggregatorDecl:
			if err := f.RegisterAggregator(node); err != nil {
				return err
			}
			if err := f.RegisterDefinition(node.Name.Value, node); err != nil {
				return fmt.Errorf("error registering definition '%s': %w", node.Name.Value, err)
			}

		default:
			// Ignore other node types at the top level? Or error?
			log.Printf("Ignoring unsupported top-level declaration type %T at pos %d", node, node.Pos())
		}
	}
	// log.Printf("Finished loading definitions.")
	f.resolved = true
	return nil
}

// GetDefinition retrieves a definition by name from the FileDecl.
func (f *FileDecl) AllDefinitions() (map[string]Node, error) {
	if f.allDefinitions == nil {
		if err := f.Resolve(); err != nil {
			return nil, fmt.Errorf("error resolving file definitions: %w", err)
		}
	}
	return f.allDefinitions, nil
}

func (f *FileDecl) GetDefinition(name string) (Node, error) {
	if f.allDefinitions == nil {
		if err := f.Resolve(); err != nil {
			return nil, fmt.Errorf("error resolving file definitions: %w", err)
		}
	}
	if decl, exists := f.allDefinitions[name]; exists {
		return decl, nil
	}
	return nil, fmt.Errorf("definition '%s' not found in file '%s'", name, f.FullPath)
}

// RegisterDefinition registers a definition in the FileDecl.
func (f *FileDecl) RegisterDefinition(name string, decl Node) error {
	if f.allDefinitions == nil {
		f.allDefinitions = make(map[string]Node)
	}
	if _, exists := f.allDefinitions[name]; exists {
		return fmt.Errorf("definition '%s' already registered", name)
	}
	f.allDefinitions[name] = decl
	// log.Printf("Registered definition '%s' of type %T", name, decl)
	return nil
}

// RegisterComponent registers a component definition in the FileDecl.
// It checks for duplicates and returns an error if the component is already registered.
func (f *FileDecl) RegisterComponent(c *ComponentDecl) error {
	if f.components == nil {
		f.components = map[string]*ComponentDecl{}
	}
	if _, exists := f.components[c.Name.Value]; exists {
		return fmt.Errorf("component definition '%s' already registered", c.Name.Value)
	}
	f.components[c.Name.Value] = c
	c.ParentFileDecl = f
	return nil
}

func (f *FileDecl) RegisterSystem(c *SystemDecl) error {
	if f.systems == nil {
		f.systems = map[string]*SystemDecl{}
	}
	if _, exists := f.systems[c.Name.Value]; exists {
		return fmt.Errorf("system definition '%s' already registered", c.Name.Value)
	}
	f.systems[c.Name.Value] = c
	c.ParentFileDecl = f
	return nil
}

func (f *FileDecl) RegisterEnum(c *EnumDecl) error {
	if f.enums == nil {
		f.enums = map[string]*EnumDecl{}
	}
	if _, exists := f.enums[c.Name.Value]; exists {
		return fmt.Errorf("enum definition '%s' already registered", c.Name.Value)
	}
	f.enums[c.Name.Value] = c
	return nil
}

func (f *FileDecl) RegisterAggregator(c *AggregatorDecl) error {
	if f.aggregators == nil {
		f.aggregators = map[string]*AggregatorDecl{}
	}
	if _, exists := f.aggregators[c.Name.Value]; exists {
		err := fmt.Errorf("aggregator definition '%s' already registered", c.Name.Value)
		panic(err)
	}
	f.aggregators[c.Name.Value] = c
	return nil
}

func (f *FileDecl) RegisterNativeMethod(c *MethodDecl) error {
	if f.nativeMethods == nil {
		f.nativeMethods = map[string]*MethodDecl{}
	}
	if _, exists := f.nativeMethods[c.Name.Value]; exists {
		err := fmt.Errorf("native method definition '%s' already registered", c.Name.Value)
		panic(err)
	}
	f.nativeMethods[c.Name.Value] = c
	return nil
}

func (f *FileDecl) RegisterImport(c *ImportDecl) error {
	if f.imports == nil {
		f.imports = map[string]*ImportDecl{}
		f.importList = []*ImportDecl{}
	}
	if _, exists := f.imports[c.ImportedAs()]; exists {
		err := fmt.Errorf("import definition '%s' already registered", c.ImportedAs())
		panic(err)
	}
	f.imports[c.ImportedAs()] = c
	f.importList = append(f.importList, c)
	return nil
}

func (f *FileDecl) String() string {
	lines := []string{}
	for _, d := range f.Declarations {
		lines = append(lines, d.String())
	}
	return strings.Join(lines, "\n")
}

// When performing static checking we want to create an initial scope starting at this file
// and adding items to the scope for name resolution
func (f *FileDecl) AddToScope(currentScope *Env[Node]) (errors []error) {
	// 1. Add local symbols to the scope
	localEnums, err := f.GetEnums()
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting local enums for scope: %w", err))
	} else {
		for name, enumDecl := range localEnums {
			if existingRef := currentScope.GetRef(name); existingRef != nil {
				errors = append(errors, fmt.Errorf("duplicate definition for local enum '%s'", name))
			} else {
				currentScope.Set(name, enumDecl)
			}
		}
	}

	localComponents, err := f.GetComponents()
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting local components for scope: %w", err))
	} else {
		for name, compDecl := range localComponents {
			if existingRef := currentScope.GetRef(name); existingRef != nil {
				errors = append(errors, fmt.Errorf("duplicate definition for local component '%s'", name))
			} else {
				currentScope.Set(name, compDecl)
			}
		}
	}

	// Add aggregators and methods
	aggs, err := f.Aggregators()
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting local aggregators for scope: %w", err))
	} else {
		for name, aggDecl := range aggs {
			if existingRef := currentScope.GetRef(name); existingRef != nil {
				errors = append(errors, fmt.Errorf("duplicate definition for local aggregator '%s'", name))
			} else {
				currentScope.Set(name, aggDecl)
			}
		}
	}

	nativeMethods, err := f.GetNativeMethods()
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting local native methods for scope: %w", err))
	} else {
		for name, methodDecl := range nativeMethods {
			if existingRef := currentScope.GetRef(name); existingRef != nil {
				errors = append(errors, fmt.Errorf("duplicate definition for local native method '%s'", name))
			} else {
				currentScope.Set(name, methodDecl)
			}
		}
	}

	// ... Add other types - eg internal enums, internal constants etc when we enable them
	return
}
