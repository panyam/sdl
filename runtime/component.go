package runtime

import (
	"github.com/panyam/sdl/decl"
)

// The runtime instance of a component. This could be Native or a UserDefined component
type ComponentInstance struct {
	ObjectInstance

	// The specs about the component
	ComponentDecl *ComponentDecl
}

// NewComponentInstance creates a new component instanceof the given type.
func NewComponentInstance(file *FileInstance, compDecl *ComponentDecl) (*ComponentInstance, Value, error) {
	// Create the component instance
	var nativeValue NativeObject
	if compDecl.IsNative {
		nativeValue = file.Runtime.CreateNativeComponent(compDecl)
	}

	originFile := file.Runtime.LoadFile(compDecl.ParentFileDecl.FullPath)
	compInst := &ComponentInstance{
		ObjectInstance: ObjectInstance{
			File:           originFile,
			IsNative:       compDecl.IsNative,
			InitialEnv:     originFile.Env(), // should parent be File.Env?
			NativeInstance: nativeValue,
		},
		ComponentDecl: compDecl,
	}
	compType := decl.ComponentType(compDecl)
	compValue, err := NewValue(compType, compInst)
	ensureNoErr(err)
	compInst.InitialEnv.Set("self", compValue)

	// Initialize the runtime based on whether it is native or user-defined
	if !compInst.IsNative {
		// Create a ComponentInstance instance
		compInst.params = make(map[string]Value) // Evaluated parameter Values (override or default)
	}
	return compInst, compValue, nil
}

// A component declaration contains instantiations of components, params, methods etc
// Specifically when a component is initialized in initializers it is important to not be bound by order.
// This method compiles the System into a set of statements that can be executed so that
// all components are intantiated first and then their properties/params are set.
func (ci *ComponentInstance) Initializer() (blockStmt *BlockStmt, err error) {
	var stmts []Stmt
	var usesDecls []*decl.UsesDecl

	// Phase 1 - Create all dependencies that have overrides on them
	params, _ := ci.ComponentDecl.Params()
	for _, param := range params {
		if param.DefaultValue != nil {
			stmts = append(stmts, &decl.SetStmt{
				TargetExpr: &MemberAccessExpr{
					Receiver: decl.NewIdent("self"),
					Member:   param.Name,
				},
				Value: param.DefaultValue,
			})
		}
	}

	deps, _ := ci.ComponentDecl.Dependencies()
	for _, usesdecl := range deps {
		if usesdecl.Overrides == nil {
			// For a dependency that is not overridden - it is not meant to be constructed
			// If a dependency is not initialized it will be reporeted when a system is initialized
			continue
		}

		usesDecls = append(usesDecls, usesdecl)
		stmts = append(stmts, &decl.SetStmt{
			TargetExpr: &MemberAccessExpr{
				Receiver: decl.NewIdent("self"),
				Member:   usesdecl.Name,
			},
			Value: NewNewExpr(usesdecl.ResolvedComponent),
		})
	}

	// Phase 2 - For each dependency that was created (it had overrides), set parameters too
	for _, it := range usesDecls {
		for _, assign := range it.Overrides {
			stmts = append(stmts, &decl.SetStmt{
				TargetExpr: &MemberAccessExpr{
					Receiver: it.Name,
					Member:   assign.Var,
				},
				Value: assign.Value,
			})
		}
	}
	return &BlockStmt{Statements: stmts}, nil
}

/*
func (ci *ComponentInstance) InvokeMethod(methodName string, args []Value, vm *VM, callFrame *Frame) (val Value, err error) {
	// 1. Find the Method Definition in the ComponentDefinition
	methodDef, err := ci.ComponentDecl.GetMethod(methodName)
	ensureNoErr(err) {
		return
	}
	if methodDef == nil {
		err = fmt.Errorf("method not found")
		return
	}

	// 2. Create a new frame for the method call.
	//    The outer frame should be the frame where the *component instance* lives?
	//    Or should it be the frame where the *call* is made? Let's use callFrame for now.
	methodFrame := NewFrame(nil)

	// 3. Bind parameters (args) to local variables in methodFrame.
	//    (Needs implementation: check arg count, types?, store Values in methodFrame)
	if len(args) != len(methodDef.Parameters) {
		err = fmt.Errorf("argument count mismatch for method '%s': expected %d, got %d", methodName, len(methodDef.Parameters), len(args))
		return
	}
	for i, paramDef := range methodDef.Parameters {
		// Store the provided argument Value under the parameter's name
		methodFrame.Set(paramDef.Name.Name, args[i])
	}

	// 4. Bind 'self'/'this' maybe? Store ci (*ComponentInstance) itself?
	val, err = NewValue(ComponentType, ci)
	ensureNoErr(err) {
		return
	}
	methodFrame.Set("self", val) // Allow methods to access instance params/deps via self.

	// 5. Bind dependencies ('uses') to local variables in methodFrame.
	for depName, depInstance := range ci.Dependencies {
		depVal, err := NewValue(ComponentType, depInstance)
		ensureNoErr(err) {
			return nil, err
		}
		methodFrame.Set(depName, depVal) // Store the ComponentRuntime dependency
	}

	// 6. Evaluate the method body (BlockStmt) using the methodFrame.
	//    The result of the block is the result of the method.
	resultValue, err := Eval(methodDef.Body, methodFrame, vm)
	ensureNoErr(err) {
		err = fmt.Errorf("error executing method '%s' body for instance '%s': %w", methodName, ci.InstanceName, err)
		return
	}

	// TODO: Handle 'return' statements within the block? Eval needs modification.
	// TODO: Check return type compatibility?

	return resultValue, nil
}
*/
