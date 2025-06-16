package runtime

import (
	"github.com/panyam/sdl/decl"
)

// The runtime instance of a component. This could be Native or a UserDefined component
type ComponentInstance struct {
	ObjectInstance

	// The specs about the component
	ComponentDecl *ComponentDecl
	
	// Arrival rates for SDL components (native components handle their own)
	arrivalRates map[string]float64
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
		compInst.arrivalRates = make(map[string]float64)
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

// SetArrivalRate sets the arrival rate for a specific method on this component.
// For native components, this delegates to the native implementation if supported.
// For SDL components, stores the rate internally.
// Returns nil on success (even if component doesn't support arrival rates).
func (ci *ComponentInstance) SetArrivalRate(methodName string, rate float64) error {
	if ci.IsNative {
		// For native components, check if they implement arrival rate setting
		if setter, ok := ci.NativeInstance.(interface{ SetArrivalRate(string, float64) error }); ok {
			return setter.SetArrivalRate(methodName, rate)
		}
		// No error if not supported - just means infinite bandwidth
		return nil
	}
	
	// For SDL components, store the rate
	if ci.arrivalRates == nil {
		ci.arrivalRates = make(map[string]float64)
	}
	ci.arrivalRates[methodName] = rate
	return nil
}

// GetArrivalRate returns the arrival rate for a specific method.
// Returns -1 if the component doesn't support arrival rates (infinite bandwidth).
func (ci *ComponentInstance) GetArrivalRate(methodName string) float64 {
	if ci.IsNative {
		// For native components, check if they implement arrival rate getting
		if getter, ok := ci.NativeInstance.(interface{ GetArrivalRate(string) float64 }); ok {
			return getter.GetArrivalRate(methodName)
		}
		return -1 // Not supported = infinite bandwidth
	}
	
	// For SDL components, return stored rate
	if ci.arrivalRates != nil {
		if rate, ok := ci.arrivalRates[methodName]; ok {
			return rate
		}
	}
	return 0 // No rate set yet
}

// GetTotalArrivalRate returns the sum of all method arrival rates.
// Returns -1 if the component doesn't support arrival rates.
func (ci *ComponentInstance) GetTotalArrivalRate() float64 {
	if ci.IsNative {
		// First try native implementation of GetTotalArrivalRate
		if getter, ok := ci.NativeInstance.(interface{ GetTotalArrivalRate() float64 }); ok {
			return getter.GetTotalArrivalRate()
		}
		// Otherwise try to sum individual rates
		if _, ok := ci.NativeInstance.(interface{ GetArrivalRate(string) float64 }); ok {
			// We'd need to know all method names - return -1 for now
			return -1
		}
		return -1 // Not supported
	}
	
	// For SDL components, sum all rates
	if ci.arrivalRates == nil || len(ci.arrivalRates) == 0 {
		return 0
	}
	
	total := 0.0
	for _, rate := range ci.arrivalRates {
		total += rate
	}
	return total
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
