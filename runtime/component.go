package runtime

import (
	"fmt"
	"log"
	"maps"
	"slices"

	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
)

// The runtime instance of a component. This could be Native or a UserDefined component
type ComponentInstance struct {
	ObjectInstance

	// The specs about the component
	ComponentDecl *ComponentDecl

	// Arrival rates for SDL components (native components handle their own)
	arrivalRates map[string]float64

	id string
}

func (c *ComponentInstance) ID() string {
	return c.id
}

// NewComponentInstance creates a new component instanceof the given type.
func NewComponentInstance(id string, file *FileInstance, compDecl *ComponentDecl) (comp *ComponentInstance, result Value, err error) {
	// Create the component instance
	var nativeValue NativeObject
	if compDecl.IsNative {
		nativeValue = file.Runtime.CreateNativeComponent(compDecl)
	}

	originFile, err := file.Runtime.LoadFile(compDecl.ParentFileDecl.FullPath)
	if err != nil {
		return
	}
	compInst := &ComponentInstance{
		ObjectInstance: ObjectInstance{
			File:           originFile,
			IsNative:       compDecl.IsNative,
			Env:            originFile.Env(), // should parent be File.Env?
			NativeInstance: nativeValue,
		},
		ComponentDecl: compDecl,
		id:            id,
	}
	compType := decl.ComponentType(compDecl)
	compValue, err := NewValue(compType, compInst)
	ensureNoErr(err)
	compInst.Env.Set("self", compValue)

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

			e := setter.SetArrivalRate(methodName, rate)
			if e != nil {
				panic(e)
			}
			return e
		}
		log.Printf("Component %T Does not support SetArrivalRate", ci.NativeInstance)
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

func (c *ComponentInstance) GetFlowPattern(method string, inputRate float64) components.FlowPattern {
	type analyzable interface {
		GetFlowPattern(method string, inputRate float64) components.FlowPattern
	}
	result := c.NativeInstance.(analyzable)
	return result.GetFlowPattern(method, inputRate)
}

// GetUtilizationInfo returns utilization information for this component and its children.
// For SDL components, it aggregates info from all child components that support utilization tracking.
func (ci *ComponentInstance) GetUtilizationInfo() []components.UtilizationInfo {
	var allInfos []components.UtilizationInfo

	if ci.IsNative {
		// For native components, check if they implement UtilizationProvider
		if provider, ok := ci.NativeInstance.(components.UtilizationProvider); ok {
			return provider.GetUtilizationInfo()
		}
		return allInfos // Empty if not supported
	}

	// For SDL components, collect utilization from all dependent components
	if ci.ComponentDecl != nil {
		deps, _ := ci.ComponentDecl.Dependencies()
		for _, dep := range deps {
			// Look up the component instance in the environment
			if binding, ok := ci.Env.Get(dep.Name.Value); ok {
				// Check if the binding is a component instance
				if childComp, ok := binding.Value.(*ComponentInstance); ok && childComp != nil {
					// Get utilization info from child
					childInfos := childComp.GetUtilizationInfo()
					// Update paths to include this component's hierarchy
					for i := range childInfos {
						childInfos[i].ComponentPath = dep.Name.Value + "." + childInfos[i].ComponentPath
					}
					allInfos = append(allInfos, childInfos...)
				}
			}
		}
	}

	// Find and mark the bottleneck resource
	if bottleneck := components.GetBottleneckUtilization(allInfos); bottleneck != nil {
		// The helper already marks the bottleneck
	}

	return allInfos
}

type NeighborMethod struct {
	Component  *ComponentInstance
	MethodName string
}

// Returns all the neighbouring component+method from a given component method
func (ci *ComponentInstance) NeighborsFromMethod(methodName string) []*NeighborMethod {
	neighbors := map[string]*NeighborMethod{}

	methodDecl, _ := ci.ComponentDecl.GetMethod(methodName)
	// analyze the body to see what methods are being called from this method

	if methodDecl == nil || methodDecl.Body == nil {
		return slices.Collect(maps.Values(neighbors))
	}

	nodes := []Node{methodDecl.Body}
	appendNode := func(n Node) {
		if n == nil {
			panic("node cannot be nil")
		}
		nodes = append(nodes, n)
	}
	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		switch n := node.(type) {
		case *decl.LetStmt:
			appendNode(n.Value)
			break
		case *decl.IfStmt:
			appendNode(n.Condition)
			appendNode(n.Then)
			if n.Else != nil {
				appendNode(n.Else)
			}
		case *decl.ForStmt:
			appendNode(n.Body)
			break
		case *decl.ReturnStmt:
			if n.ReturnValue != nil {
				appendNode(n.ReturnValue)
			}
		case *decl.UnaryExpr:
			appendNode(n.Right)
		case *decl.BinaryExpr:
			appendNode(n.Left)
			appendNode(n.Right)
		case *decl.BlockStmt:
			for _, s := range n.Statements {
				appendNode(s)
			}
		case *decl.ExprStmt:
			appendNode(n.Expression)
		case *decl.CallExpr:
			// See if args themselves call any functions
			for _, arg := range n.ArgList {
				appendNode(arg)
			}
			for _, arg := range n.ArgMap {
				appendNode(arg)
			}

			// May be move this into a helper method in MemberAccessExpr as .FullPath() method
			if mae, ok := n.Function.(*decl.MemberAccessExpr); ok {
				methodName := mae.Member.Value
				var fqn []string
				// var rootComponent *ComponentInstance
				for mae != nil {
					receiver := mae.Receiver
					switch recv := receiver.(type) {
					case *decl.IdentifierExpr:
						if fqn == nil {
							fqn = []string{recv.Value}
						} else {
							fqn = append([]string{recv.Value}, fqn...)
						}
						mae = nil
						break
					case *decl.MemberAccessExpr:
						if fqn == nil {
							fqn = []string{recv.Member.Value}
						} else {
							fqn = append([]string{recv.Member.Value}, fqn...)
						}
						mae = recv
						break
					default:
						panic(fmt.Sprintf("Invalid receiver type: %T", recv))
					}
				}

				// TODO - Move this to a "FindNestedValue" method in ComponentInstance
				targetComponent := ci
				for _, part := range fqn {
					if part == "self" {
						continue
					}
					if next, ok := targetComponent.Get(part); ok && next.Type.Tag == decl.TypeTagComponent {
						targetComponent = next.Value.(*ComponentInstance)
					}
				}

				// TODO: Now handle the Function Expr - and only process if it is not a native method call
				if targetComponent != nil {
					// we can add this to our map if not alreay done so
					neighKey := fmt.Sprintf("%p:%s", targetComponent, methodName)
					if neighbors[neighKey] == nil {
						neighbors[neighKey] = &NeighborMethod{Component: targetComponent, MethodName: methodName}
					}
				}
			}
			break
		case *decl.IdentifierExpr:
			// Resolve from the env
			val, ok := ci.Env.Get(n.Value)
			// log.Println("Value: ", n.Value, val, reflect.TypeOf(val.Value))
			if ok && !val.IsNil() {
				// log.Println("Value: ", val, reflect.TypeOf(val.Value))
			}
		case *decl.LiteralExpr:
		case *decl.SampleExpr:
		case *decl.MemberAccessExpr:
			// Ignore these
			break
		default:
			panic(fmt.Sprintf("Invalid node type: %T", n))
		}
	}

	return slices.Collect(maps.Values(neighbors))
}
