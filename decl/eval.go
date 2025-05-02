package decl

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/panyam/leetcoach/sdl/core"
)

type Value = any

var (
	ErrNotImplemented       = errors.New("evaluation for this node type not implemented")
	ErrNotFound             = errors.New("identifier not found")
	ErrInternalFuncNotFound = errors.New("internal function not found")
	ErrUnsupportedType      = errors.New("unsupported type for operation")
)

// Starts the execution of a single expression
func Eval(node Node, env *Env[any], v *VM) (OpNode, error) {
	// fmt.Printf("Eval entry: %T - %s\n", node, node) // Debug entry
	switch n := node.(type) {
	// We'll add cases here in subsequent phases
	case *LiteralExpr:
		return evalLiteral(n, env, v)
	case *IdentifierExpr:
		return evalIdentifier(n, env, v)

	// --- Statement Nodes ---
	case *BlockStmt:
		// When Eval is called directly on a BlockStmt (e.g., top level),
		// there is no initial context from an outer structure like IfStmt.
		// evalBlockStmt now returns the result directly, doesn't leave on stack implicitly
		return evalBlockStmt(n, NewEnv(env), nil) // Pass nil context
	case *LetStmt:
		return evalLetStmt(n, env, nil)
	case *BinaryExpr:
		return evalBinaryExpr(n, env, v)
	case *ExprStmt:
		return evalExprStmt(n, env, v)
	case *IfStmt: // <-- Will be implemented now
		return evalIfStmt(n, env, v)
	case *SystemDecl: // <<< Added case
		return evalSystemDecl(n, env, v)
	case *InstanceDecl: // <<< Added case
		return evalInstanceDecl(n, env, v)
	case *ComponentDecl: // <<< Added case
		return evalComponentDecl(n, env, v)
	/* - TODO
	case *AssignmentStmt:
		return evalAssignmentStmt(n, env, v)
	case *SwitchStmt: // <-- Will be implemented now
		return evalSwitchStmt(n, env, v)
	case *CallExpr:
		return evalCall(n, env, v)
	case *MemberAccessExpr:
		// Member access is often handled *within* evalCallExpr,
		// but we might need a stub if it can be evaluated alone.
		return evalMemberAccess(n, env, v) // <-- Call the actual implementation
	*/

	default:
		return nil, fmt.Errorf("Eval not implemented for node type %T", node)
	}
}

/** Evaluate a literal and return its value */
func evalLiteral(expr *LiteralExpr, env *Env[any], v *VM) (OpNode, error) {
	rawValue, err := ParseLiteralValue(expr) // Use existing helper
	if err != nil {
		return nil, fmt.Errorf("failed to parse literal '%s': %w", expr.Value, err)
	}

	var valueOutcome any
	latencyOutcome := ZeroLatencyOutcome()

	switch expr.Kind {
	case "INT":
		val, ok := rawValue.(int64)
		if !ok {
			return nil, fmt.Errorf("internal error: parsed INT literal '%s' resulted in %T, expected int64", expr.Value, rawValue)
		}
		valueOutcome = (&core.Outcomes[int64]{}).Add(1.0, val)
	case "FLOAT":
		val, ok := rawValue.(float64)
		if !ok {
			return nil, fmt.Errorf("internal error: parsed FLOAT literal '%s' resulted in %T, expected float64", expr.Value, rawValue)
		}
		valueOutcome = (&core.Outcomes[float64]{}).Add(1.0, val)
	case "BOOL":
		val, ok := rawValue.(bool)
		if !ok {
			return nil, fmt.Errorf("internal error: parsed BOOL literal '%s' resulted in %T, expected bool", expr.Value, rawValue)
		}
		valueOutcome = (&core.Outcomes[bool]{}).Add(1.0, val)
	case "STRING":
		val, ok := rawValue.(string)
		if !ok {
			return nil, fmt.Errorf("internal error: parsed STRING literal '%s' resulted in %T, expected string", expr.Value, rawValue)
		}
		valueOutcome = (&core.Outcomes[string]{}).Add(1.0, val)
	// TODO: case "DURATION": Handle duration parsing and create *core.Outcomes[core.Duration] value outcome
	default:
		return nil, fmt.Errorf("unsupported literal kind '%s' in evalLiteral", expr.Kind)
	}

	state := &VarState{
		ValueOutcome:   valueOutcome,
		LatencyOutcome: latencyOutcome,
	}
	return &LeafNode{State: state}, nil
}

/** Evaluate a Identifier and return its value */
func evalIdentifier(expr *IdentifierExpr, env *Env[any], v *VM) (OpNode, error) {
	name := expr.Name
	value, ok := env.Get(name)
	if !ok {
		return nil, fmt.Errorf("%w: identifier '%s'", ErrNotFound, name)
	}

	// The environment should store OpNodes for variables during evaluation
	opNode, ok := value.(OpNode)
	if !ok {
		// This indicates an internal inconsistency - something other than an OpNode
		// was stored for a variable during evaluation.
		return nil, fmt.Errorf("internal error: expected OpNode for identifier '%s', but found type %T in environment", name, value)
	}

	return opNode, nil
}

func evalLetStmt(stmt *LetStmt, env *Env[any], v *VM) (OpNode, error) {
	varName := stmt.Variable.Name
	valueOpNode, err := Eval(stmt.Value, env, v)
	if err != nil {
		return nil, fmt.Errorf("evaluating value for let statement '%s': %w", varName, err)
	}

	// Store the resulting OpNode in the current environment
	env.Set(varName, valueOpNode)

	// 'let' itself doesn't produce a value for subsequent sequence steps
	return theNilNode, nil
}

/** Evaluate a Call and return its value */
func evalCall(expr *CallExpr, env *Env[any], v *VM) (val Value, err error) {
	return
}

/** Evaluate a MemberAccess and return its value */
func evalMemberAccess(expr *MemberAccessExpr, env *Env[any], v *VM) (val Value, err error) {
	return
}

/** Evaluate a Block and return its value */
func evalBlockStmt(stmt *BlockStmt, env *Env[any], v *VM) (OpNode, error) {
	blockEnv := NewEnv[any](env) // Create block scope
	steps := make([]OpNode, 0, len(stmt.Statements))

	for _, statement := range stmt.Statements {
		resultNode, err := Eval(statement, blockEnv, v)
		if err != nil {
			// TODO: Improve error reporting with position info from statement
			return nil, fmt.Errorf("error in block statement: %w", err)
		}

		// Only include non-nil nodes in the sequence
		if _, isNil := resultNode.(*NilNode); !isNil {
			steps = append(steps, resultNode)
		}
	}

	// Determine return value based on collected steps
	if len(steps) == 0 {
		return theNilNode, nil // Empty block or only let statements
	}
	if len(steps) == 1 {
		return steps[0], nil // Single effective statement, return its node
	}
	return &SequenceNode{Steps: steps}, nil // Multiple effective statements
}

/** Evaluate a If and return its value */
func evalIfStmt(stmt *IfStmt, env *Env[any], v *VM) (val OpNode, err error) {
	// Evaluate the condition expression to get its OpNode representation
	conditionNode, err := Eval(stmt.Condition, env, v)
	if err != nil {
		// TODO: Improve error reporting
		return nil, fmt.Errorf("evaluating condition for if statement: %w", err)
	}

	// Evaluate the 'then' block to get its OpNode representation
	// Note: Use the *same* environment level as the if statement itself.
	// Scoping for variables *inside* the block is handled by evalBlockStmt.
	thenNode, err := Eval(stmt.Then, env, v)
	if err != nil {
		// TODO: Improve error reporting
		return nil, fmt.Errorf("evaluating 'then' block for if statement: %w", err)
	}

	// Evaluate the 'else' block/statement, if it exists
	var elseNode OpNode = theNilNode // Default to NilNode if no else
	if stmt.Else != nil {
		elseNode, err = Eval(stmt.Else, env, v)
		if err != nil {
			// TODO: Improve error reporting
			return nil, fmt.Errorf("evaluating 'else' block for if statement: %w", err)
		}
	}

	// Construct and return the IfChoiceNode representing the structure
	return &IfChoiceNode{
		Condition: conditionNode,
		Then:      thenNode,
		Else:      elseNode,
	}, nil
}

/** Evaluate a Switch and return its value */
func evalSwitchStmt(stmt *SwitchStmt, env *Env[any], v *VM) (val Value, err error) {
	return
}

/** Evaluate a Expr as a statement and return its value */
func evalExprStmt(stmt *ExprStmt, env *Env[any], v *VM) (OpNode, error) {
	// Evaluate the expression and return its OpNode result
	return Eval(stmt.Expression, env, v)
}

/** Evaluate a Assignment as a statement and return its value */
func evalAssignmentStmt(stmt *AssignmentStmt, env *Env[any], v *VM) (val Value, err error) {
	return
}

func evalBinaryExpr(expr *BinaryExpr, env *Env[any], v *VM) (OpNode, error) {
	// Recursively evaluate left and right operands
	leftNode, err := Eval(expr.Left, env, v)
	if err != nil {
		// TODO: Improve error reporting with position info
		return nil, fmt.Errorf("evaluating left operand for '%s': %w", expr.Operator, err)
	}

	rightNode, err := Eval(expr.Right, env, v)
	if err != nil {
		// TODO: Improve error reporting with position info
		return nil, fmt.Errorf("evaluating right operand for '%s': %w", expr.Operator, err)
	}

	// Check operator validity if needed (parser should handle this mostly)
	// switch expr.Operator {
	// case "+", "-", "*", "/", "%", "&&", "||", "==", "!=", "<", "<=", ">", ">=":
	// 	// Valid operator
	// default:
	// 	return nil, fmt.Errorf("unsupported binary operator '%s'", expr.Operator)
	// }

	// Construct and return the BinaryOpNode representing the operation
	return &BinaryOpNode{
		Op:    expr.Operator,
		Left:  leftNode,
		Right: rightNode,
	}, nil
}

// --- evalComponentDecl (Registers definition) ---
func evalComponentDecl(stmt *ComponentDecl, env *Env[any], v *VM) (OpNode, error) {
	compName := stmt.Name.Name
	// Check if definition already exists in VM registry
	if _, exists := v.ComponentDefRegistry[compName]; exists {
		return nil, fmt.Errorf("component '%s' already defined", compName)
	}

	compDef := &ComponentDefinition{
		Node:    stmt,
		Params:  make(map[string]*ParamDecl),
		Uses:    make(map[string]*UsesDecl),
		Methods: make(map[string]*MethodDef),
	}

	// Process body to populate definition details
	for _, item := range stmt.Body {
		switch bodyNode := item.(type) {
		case *ParamDecl:
			paramName := bodyNode.Name.Name
			if _, exists := compDef.Params[paramName]; exists {
				return nil, fmt.Errorf("duplicate parameter '%s' in component '%s'", paramName, compName)
			}
			compDef.Params[paramName] = bodyNode
		case *UsesDecl:
			usesName := bodyNode.Name.Name
			if _, exists := compDef.Uses[usesName]; exists {
				return nil, fmt.Errorf("duplicate uses declaration '%s' in component '%s'", usesName, compName)
			}
			compDef.Uses[usesName] = bodyNode
		case *MethodDef:
			methodName := bodyNode.Name.Name
			if _, exists := compDef.Methods[methodName]; exists {
				return nil, fmt.Errorf("duplicate method definition '%s' in component '%s'", methodName, compName)
			}
			compDef.Methods[methodName] = bodyNode
		case *ComponentDecl:
			// Handle nested definitions
			_, err := evalComponentDecl(bodyNode, env, v)
			if err != nil {
				return nil, fmt.Errorf("error defining nested component '%s' within '%s': %w", bodyNode.Name.Name, compName, err)
			}
		default:
			// Ignore other items for now
		}
	}

	// Register the processed definition with the VM
	err := v.RegisterComponentDef(compDef)
	if err != nil {
		return nil, err // Should be caught by initial check, but good practice
	}

	// Defining a component doesn't produce an executable OpNode
	return theNilNode, nil
}

// --- evalSystemDecl (Processes system body) ---
func evalSystemDecl(stmt *SystemDecl, env *Env[any], v *VM) (OpNode, error) {
	// Systems define a scope, but for now, let's use the passed env.
	// A system run might eventually need its own top-level env.
	// systemEnv := NewEnv(env) // Option for later

	for _, item := range stmt.Body {
		// Evaluate each item within the system's context
		// For now, InstanceDecl modifies the passed env.
		_, err := Eval(item, env, v) // Use passed env
		if err != nil {
			return nil, fmt.Errorf("error evaluating item in system '%s': %w", stmt.Name.Name, err)
		}
		// Ignore the OpNode returned by body items (e.g., InstanceDecl returns NilNode)
	}

	// System declaration itself doesn't produce a value OpNode
	return theNilNode, nil
}

// --- evalInstanceDecl (Instantiates Native or DSL component) ---
func evalInstanceDecl(stmt *InstanceDecl, env *Env[any], v *VM) (OpNode, error) {
	instanceName := stmt.Name.Name
	componentTypeName := stmt.ComponentType.Name

	// Check if instance name already exists in the current scope
	if _, exists := env.Get(instanceName); exists {
		return nil, fmt.Errorf("identifier '%s' already exists in the current scope", instanceName)
	}

	// Get Component Definition from VM Registry
	compDef, foundDef := v.ComponentDefRegistry[componentTypeName]
	if !foundDef {
		return nil, fmt.Errorf("unknown component type definition '%s' for instance '%s'", componentTypeName, instanceName)
	}

	// Check for Native Go Constructor (Prefer native if constructor exists)
	constructor, foundConst := v.ComponentRegistry[componentTypeName]

	var runtimeInstance ComponentRuntime // The instance to store in the env

	// --- Branch: Instantiate Native Go Component ---
	if foundConst {
		// Prepare maps for constructor params and dependencies
		overrideParamValues := make(map[string]any)              // Raw Go values for constructor
		dependencyInstances := make(map[string]ComponentRuntime) // Map uses_name -> ComponentRuntime

		// Evaluate overrides provided in the InstanceDecl block
		processedOverrides := make(map[string]bool)
		for _, assignStmt := range stmt.Overrides {
			assignVarName := assignStmt.Var.Name
			processedOverrides[assignVarName] = true

			valueOpNode, err := Eval(assignStmt.Value, env, v) // Eval RHS -> OpNode
			if err != nil {
				return nil, fmt.Errorf("evaluating override '%s' for native instance '%s': %w", assignVarName, instanceName, err)
			}

			// Check if override targets a parameter or a dependency
			if _, isParam := compDef.Params[assignVarName]; isParam {
				// --- TEMPORARY WORKAROUND for Param ---
				// Requires immediate evaluation of valueOpNode to a simple Go value.
				leaf, ok := valueOpNode.(*LeafNode)
				if !ok {
					return nil, fmt.Errorf("param override '%s' for native instance '%s' did not evaluate to a simple value (got %T)", assignVarName, instanceName, valueOpNode)
				}
				rawValue, err := extractLeafValue(leaf) // Use helper
				if err != nil {
					return nil, fmt.Errorf("extracting value for param override '%s' for native instance '%s': %w", assignVarName, instanceName, err)
				}
				overrideParamValues[assignVarName] = rawValue
				// --- END TEMPORARY WORKAROUND ---
			} else if _, isUses := compDef.Uses[assignVarName]; isUses {
				// Dependency assignment: RHS must be an identifier
				identExpr, okIdent := assignStmt.Value.(*IdentifierExpr)
				if !okIdent {
					return nil, fmt.Errorf("value for 'uses' override '%s' must be an identifier, got %T", assignVarName, assignStmt.Value)
				}
				depInstanceName := identExpr.Name
				depInstanceAny, foundDep := env.Get(depInstanceName)
				if !foundDep {
					return nil, fmt.Errorf("dependency instance '%s' (for '%s.%s') not found", depInstanceName, instanceName, assignVarName)
				}
				// Assert the dependency is a ComponentRuntime
				depRuntime, okRuntime := depInstanceAny.(ComponentRuntime)
				if !okRuntime {
					return nil, fmt.Errorf("dependency '%s' resolved to non-runtime type %T", depInstanceName, depInstanceAny)
				}
				dependencyInstances[assignVarName] = depRuntime
			} else {
				return nil, fmt.Errorf("unknown native override target '%s' for instance '%s'", assignVarName, instanceName)
			}
		}

		// Check if all 'uses' dependencies were satisfied by overrides
		for usesName := range compDef.Uses {
			if _, satisfied := dependencyInstances[usesName]; !satisfied {
				return nil, fmt.Errorf("missing override to satisfy 'uses %s: %s' dependency for native instance '%s'", usesName, compDef.Uses[usesName].ComponentType.Name, instanceName)
			}
		}

		// Instantiate Component using Constructor
		goInstance, err := constructor(instanceName, overrideParamValues)
		if err != nil {
			return nil, fmt.Errorf("failed to construct native component '%s': %w", instanceName, err)
		}

		// Inject Dependencies (Conceptual - requires reflection/setters)
		// Pass the ComponentRuntime map for injection
		err = injectDependencies(goInstance, dependencyInstances)
		if err != nil {
			return nil, fmt.Errorf("failed to inject dependencies into native '%s': %w", instanceName, err)
		}

		// Wrap native instance in adapter
		adapter := &NativeComponentAdapter{
			InstanceName: instanceName,
			TypeName:     componentTypeName,
			GoInstance:   goInstance,
		}
		runtimeInstance = adapter // Store the adapter

		// --- Branch: Instantiate DSL ComponentInstance ---
	} else {
		// Create the DSL instance
		dslInstance := &ComponentInstance{
			Definition:   compDef,
			InstanceName: instanceName,
			Params:       make(map[string]OpNode),
			Dependencies: make(map[string]ComponentRuntime),
		}

		processedOverrides := make(map[string]bool)

		// Process overrides
		for _, assignStmt := range stmt.Overrides {
			assignVarName := assignStmt.Var.Name
			processedOverrides[assignVarName] = true

			valueOpNode, err := Eval(assignStmt.Value, env, v) // Eval RHS -> OpNode
			if err != nil {
				return nil, fmt.Errorf("evaluating override '%s' for DSL instance '%s': %w", assignVarName, instanceName, err)
			}

			if _, isParam := compDef.Params[assignVarName]; isParam {
				// Store the evaluated OpNode directly for parameters
				dslInstance.Params[assignVarName] = valueOpNode
			} else if _, isUses := compDef.Uses[assignVarName]; isUses {
				// Dependency assignment: RHS must be an identifier
				identExpr, okIdent := assignStmt.Value.(*IdentifierExpr)
				if !okIdent {
					return nil, fmt.Errorf("value for 'uses' override '%s' must be an identifier, got %T", assignVarName, assignStmt.Value)
				}
				depInstanceName := identExpr.Name
				depInstanceAny, foundDep := env.Get(depInstanceName)
				if !foundDep {
					return nil, fmt.Errorf("dependency instance '%s' (for '%s.%s') not found", depInstanceName, instanceName, assignVarName)
				}

				// Assert the dependency is a ComponentRuntime before storing
				depRuntime, okRuntime := depInstanceAny.(ComponentRuntime)
				if !okRuntime {
					return nil, fmt.Errorf("dependency '%s' resolved to non-runtime type %T", depInstanceName, depInstanceAny)
				}

				// Store dependency based on its underlying type (Native vs DSL)
				if _, isNative := depRuntime.(*NativeComponentAdapter); isNative {
					dslInstance.Dependencies[assignVarName] = depRuntime
				} else {
					// Should not happen if only adapters and instances are stored
					return nil, fmt.Errorf("internal error: dependency '%s' resolved to unknown ComponentRuntime type %T", depInstanceName, depRuntime)
				}
			} else {
				return nil, fmt.Errorf("unknown DSL override target '%s' for instance '%s'", assignVarName, instanceName)
			}
		}

		// Process default parameter values for those not overridden
		for paramName, paramAST := range compDef.Params {
			if _, overridden := processedOverrides[paramName]; !overridden {
				if paramAST.DefaultValue != nil {
					defaultOpNode, err := Eval(paramAST.DefaultValue, env, v)
					if err != nil {
						return nil, fmt.Errorf("evaluating default value for param '%s' in DSL instance '%s': %w", paramName, instanceName, err)
					}
					dslInstance.Params[paramName] = defaultOpNode
				} else {
					return nil, fmt.Errorf("missing required parameter '%s' for DSL instance '%s'", paramName, instanceName)
				}
			}
		}

		// Check if all 'uses' dependencies were satisfied by overrides
		for usesName := range compDef.Uses {
			_, found := dslInstance.Dependencies[usesName]
			if !found {
				return nil, fmt.Errorf("missing override to satisfy 'uses %s: %s' dependency for DSL instance '%s'", usesName, compDef.Uses[usesName].ComponentType.Name, instanceName)
			}
		}
		runtimeInstance = dslInstance // Store the DSL instance
	}

	// Store the resulting ComponentRuntime (either adapter or DSL instance)
	env.Set(instanceName, runtimeInstance)

	// Instance declaration itself doesn't produce a value OpNode
	return theNilNode, nil
}

// --- Helper to extract simple value from LeafNode (Used in Temp Workaround) ---
func extractLeafValue(leaf *LeafNode) (any, error) {
	if leaf == nil || leaf.State == nil || leaf.State.ValueOutcome == nil {
		return nil, fmt.Errorf("cannot extract value from nil leaf/state/outcome")
	}
	// Try to extract the raw Go value using GetValue()
	var rawValue any
	var extractOk bool
	switch outcome := leaf.State.ValueOutcome.(type) {
	case *core.Outcomes[int64]:
		rawValue, extractOk = outcome.GetValue()
	case *core.Outcomes[float64]:
		rawValue, extractOk = outcome.GetValue()
	case *core.Outcomes[bool]:
		rawValue, extractOk = outcome.GetValue()
	case *core.Outcomes[string]:
		rawValue, extractOk = outcome.GetValue()
	// TODO: Add case for *core.Outcomes[core.Duration]?
	default:
		return nil, fmt.Errorf("unsupported outcome type %T for value extraction", outcome)
	}
	if !extractOk {
		return nil, fmt.Errorf("value is probabilistic or empty, cannot extract single value")
	}
	return rawValue, nil
}

// --- Placeholder injectDependencies (remains the same, uses reflection conceptually) ---
func injectDependencies(targetInstance any, dependencies map[string]ComponentRuntime) error {
	targetVal := reflect.ValueOf(targetInstance)
	// Check if targetInstance is valid pointer to struct
	if targetVal.Kind() != reflect.Ptr || targetVal.IsNil() {
		return fmt.Errorf("targetInstance must be a non-nil pointer")
	}
	targetElem := targetVal.Elem()
	if targetElem.Kind() != reflect.Struct {
		return fmt.Errorf("targetInstance must point to a struct")
	}

	// log.Printf("Conceptual Injection into %T:", targetInstance)
	for name, depRuntime := range dependencies {
		// log.Printf("  Injecting '%s' (%T)", name, depRuntime)

		// Find field in targetInstance struct (Simple approach: match name case-insensitively?)
		// Real approach needs tags or better mapping. Assume field name matches `uses` name but capitalized.
		fieldName := strings.Title(name)
		field := targetElem.FieldByName(fieldName)
		if !field.IsValid() {
			// log.Printf("    Warning: No field '%s' found in target %T (Skipping injection)", fieldName, targetInstance)
			continue // Skip if no matching field found (could be error?)
			// return fmt.Errorf("no field '%s' found in target %T", fieldName, targetInstance)
		}
		if !field.CanSet() {
			return fmt.Errorf("cannot set field '%s' in target %T", fieldName, targetInstance)
		}

		// Get the actual underlying value (Go instance or *ComponentInstance)
		var depValueToInject any
		if adapter, ok := depRuntime.(*NativeComponentAdapter); ok {
			depValueToInject = adapter.GoInstance
		} else if dslInst, ok := depRuntime.(*ComponentInstance); ok {
			depValueToInject = dslInst
		} else {
			return fmt.Errorf("dependency %s has unknown ComponentRuntime type %T", name, depRuntime)
		}

		depVal := reflect.ValueOf(depValueToInject)

		// Check type compatibility before setting
		if !depVal.Type().AssignableTo(field.Type()) {
			// Allow assigning concrete types to interface fields
			if field.Type().Kind() == reflect.Interface && depVal.Type().AssignableTo(field.Type()) {
				// This check seems redundant with AssignableTo?
				// Let AssignableTo handle interface checks.
			} else if field.Type().Kind() == reflect.Interface && depVal.Type().Implements(field.Type()) {
				// Explicit check for interface implementation
			} else if field.Type().Kind() == reflect.Interface && field.Type().NumMethod() == 0 {
				// Allow assigning anything to empty interface{}
			} else {
				return fmt.Errorf("type mismatch: cannot assign dependency '%s' type %T to field '%s' type %s", name, depValueToInject, fieldName, field.Type())
			}
		}
		field.Set(depVal)
		// log.Printf("    Successfully injected '%s' into field '%s'", name, fieldName)
	}
	return nil
}
