package runtime

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	cd "github.com/panyam/sdl/components/decl"
	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader"
)

type NativeMethod func(eval *SimpleEval, env *Env[Value], currTime *core.Duration, args ...Value) (result Value, returned bool)

type Runtime struct {
	Loader         *loader.Loader
	NativeObjects  []any
	fileInstances  map[string]*FileInstance
	nativeMethods  map[string]NativeMethod
	nativeAggrs    map[string]Aggregator
	nativeComps    map[string]any
	nativeCompCons map[string]func(name string) any
}

func NewRuntime(loader *loader.Loader) (r *Runtime) {
	r = &Runtime{
		Loader:        loader,
		fileInstances: make(map[string]*FileInstance),
		nativeMethods: make(map[string]NativeMethod),
	}
	r.RegisterNativeMethod("log", Native_log)
	r.RegisterNativeMethod("delay", Native_delay)
	return
}

func (r *Runtime) RegisterNativeMethod(name string, f NativeMethod) {
	r.nativeMethods[name] = f
}

// Gets the initial run time environment for a File which would include its parameters and component creators
func (r *Runtime) LoadFile(filePath string) (*FileInstance, error) {
	if inst, ok := r.fileInstances[filePath]; ok && inst != nil {
		return inst, nil
	}

	fileStatus, err := r.Loader.LoadFile(filePath, "", 0)
	if err != nil {
		return nil, err
	}
	r.Loader.Validate(fileStatus)
	if fileStatus.HasErrors() {
		log.Printf("\nError Validating File %s\n", fileStatus.FullPath)
		fileStatus.PrintErrors()
	} else {
		log.Printf("\nFile %s - Validated Successfully at: %v\n", fileStatus.FullPath, fileStatus.LastValidated)
	}

	file := fileStatus.FileDecl
	out := NewFileInstance(r, file)
	r.fileInstances[fileStatus.FullPath] = out
	return out, nil
}

// Gets the value of a parameter given by a path "comp1.comp2...compN.ParamName" starting at a given System
// and returns its Value
func (r *Runtime) GetParam(system *SystemInstance, paramPath string) (value decl.Value, err error) {
	// 1. resolve comp1.comp2.compN...ParamName to a nested component
	// 2. Get the ParamName at compN (look up the component instance's env)
	// 3. Return the value (if any)

	parts := strings.Split(paramPath, ".")
	if len(parts) < 2 {
		return value, fmt.Errorf("invalid parameter path: %s", paramPath)
	}

	// Extract parameter name (last part)
	paramName := parts[len(parts)-1]
	componentPath := strings.Join(parts[:len(parts)-1], ".")

	// Find the component instance using FindComponent
	componentInstance := system.FindComponent(componentPath)
	if componentInstance == nil {
		return value, fmt.Errorf("component '%s' not found", componentPath)
	}

	// Get the parameter value from the component's environment
	paramValue, ok := componentInstance.Get(paramName)
	if !ok {
		return value, fmt.Errorf("parameter '%s' not found in component '%s'", paramName, componentPath)
	}
	return paramValue, nil
}

func (r *Runtime) BatchSetParams(system *SystemInstance, paramPaths []string, newValues []decl.Value, oldValues map[string]decl.Value) (err error) {
	// 1. resolve comp1.comp2.compN...ParamName to a nested component
	// 2. Get the ParamName at compN (look up the component instance's env)
	// 3. Eval a SetStmt(ParamName, newValue) at compN using compN's Env
	// 4. Return oldValue and error (if any)

	for i, paramPath := range paramPaths {
		parts := strings.Split(paramPath, ".")
		if len(parts) < 2 {
			err = fmt.Errorf("invalid parameter path: %s", paramPath)
			return
		}
		newValue := newValues[i]

		// Extract parameter name (last part)
		paramName := parts[len(parts)-1]
		componentPath := strings.Join(parts[:len(parts)-1], ".")

		// Find the component instance using FindComponent
		componentInstance := system.FindComponent(componentPath)
		if componentInstance == nil {
			err = fmt.Errorf("component '%s' not found", componentPath)
			return
		}

		oldValue, _ := componentInstance.Get(paramName)
		/* instead of SetStmt, just call teh .Set method on the Component instance
		// Create a SetStmt to set the parameter
		setStmt := &decl.SetStmt{
			TargetExpr: &decl.IdentifierExpr{Value: paramName},
			Value:      &decl.LiteralExpr{Value: newValue},
		}

		// Evaluate the SetStmt in the component's environment
		eval := NewSimpleEval(system.File, nil)
		eval.Eval(setStmt, componentInstance.InitialEnv, nil)
		if eval.HasErrors() {
			err = eval.Errors[0]
		}
		*/
		oldValues[paramPath] = oldValue
		err = componentInstance.Set(paramName, newValue)
	}
	return
}

// Sets the value of a parameter given by a path "comp1.comp2...compN.ParamName" starting at a given System
// and sets its new value.  Returns the oldValue (whether success or failure) and returns an error
// if setting failed
func (r *Runtime) SetParam(system *SystemInstance, paramPath string, newValue decl.Value) (oldValue decl.Value, err error) {
	oldValues := map[string]decl.Value{}
	err = r.BatchSetParams(system, []string{paramPath}, []decl.Value{newValue}, oldValues)
	if err != nil {
		return decl.Nil, err
	}
	return oldValues[paramPath], err
}

// Get all available system declarations across all file instnces as a map
func (r *Runtime) AvailableSystems() (out map[string]*SystemDecl) {
	for _, finst := range r.fileInstances {
		sysDecls, _ := finst.Decl.GetSystems()
		for name, sysDecl := range sysDecls {
			out[name] = sysDecl
		}
	}
	return out
}

// Looks up all the files for the system maching the given name and initializes it
func (r *Runtime) NewSystem(systemName string) (sysInst *SystemInstance, err error) {
	log.Println("File Instances: ", r.fileInstances)
	for _, finst := range r.fileInstances {
		sysDecl, _ := finst.Decl.GetSystem(systemName)
		slog.Debug("Decl: ", finst.Decl.FullPath, systemName, sysDecl)
		if sysDecl == nil {
			continue
		}
		sysInst, _ = finst.NewSystem(systemName, true)
		if sysInst != nil {
			return sysInst, nil
		}
	}
	return nil, fmt.Errorf("system '%s' not found in any loaded file", systemName)
}

func (r *Runtime) CreateNativeComponent(compDecl *ComponentDecl) NativeObject {
	name := compDecl.Name.Value
	switch name {
	case "Disk", "NativeDisk":
		return cd.NewDisk(name)
	case "DiskWithContention":
		return cd.NewDiskWithContention("SSD") // Default to SSD
	case "HashIndex":
		return cd.NewHashIndex(name)
	case "BTreeIndex":
		return cd.NewBTreeIndex(name)
	case "BitmapIndex":
		return cd.NewBitmapIndex(name)
	case "Cache":
		return cd.NewCache(name)
	case "LSMTree":
		return cd.NewLSMTree(name)
	case "MM1Queue":
		return cd.NewMM1Queue(name)
	case "MMCKQueue":
		return cd.NewQueue(name)
	case "ResourcePool":
		return cd.NewResourcePool(name)
	case "Link":
		return cd.NewNetworkLink(name)
	case "SortedFile":
		return cd.NewSortedFile(name)
	case "HeapFile":
		return cd.NewHeapFile(name)
	case "TestNative":
		return NewTestNative(name)
	}
	panic(fmt.Sprintf("Native component not registered: %s", name))
}

func Native_log(eval *SimpleEval, env *Env[Value], currTime *core.Duration, args ...Value) (result Value, returned bool) {
	for _, arg := range args {
		fmt.Printf("LOG: %s\n", arg.String())
	}
	return Nil, false
}

func Native_delay(eval *SimpleEval, env *Env[Value], currTime *core.Duration, args ...Value) (result Value, returned bool) {
	if len(args) != 1 {
		panic("delay expects exactly one argument")
	}
	result = args[0]
	if i, err := result.GetInt(); err == nil {
		*currTime += core.Duration(i)
		result.Time += core.Duration(i)
	} else if f, err := result.GetFloat(); err == nil {
		*currTime += f
		result.Time += f
	} else {
		panic("delay value should have been int or float. type checking failed")
	}
	return
}
