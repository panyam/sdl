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

func evalSystemDecl(stmt *SystemDecl, env *Env[any], v *VM) (OpNode, error) {
	// Systems start a new scope so you can assume they will be given a new scope
	for _, item := range stmt.Body {
		// Evaluate each item within the system's context
		_, err := Eval(item, env, v) // Use passed env for now
		if err != nil {
			// TODO: Improve error reporting
			return nil, fmt.Errorf("error evaluating item in system '%s': %w", stmt.Name.Name, err)
		}
		// We ignore the OpNode returned by body items (like InstanceDecl returning NilNode)
	}

	// System declaration itself doesn't produce a value OpNode
	return theNilNode, nil
}

func evalComponentDecl(stmt *ComponentDecl, vm *Env[any], v *VM) (OpNode, error) {
	compName := stmt.Name.Name
	if _, exists := v.ComponentDefRegistry[compName]; exists {
		// TODO: Allow redefinition or error? Error for now.
		return nil, fmt.Errorf("component '%s' already defined", compName)
	}

	// Process the body to build the definition
	compDef := &ComponentDefinition{
		Node:    stmt,
		Params:  make(map[string]*ParamDecl),
		Uses:    make(map[string]*UsesDecl),
		Methods: make(map[string]*MethodDef),
	}

	for _, item := range stmt.Body {
		switch bodyNode := item.(type) {
		case *ParamDecl:
			paramName := bodyNode.Name.Name
			if _, exists := compDef.Params[paramName]; exists {
				return nil, fmt.Errorf("duplicate parameter '%s' in component '%s'", paramName, compName)
			}
			compDef.Params[paramName] = bodyNode
			// Note: DefaultValue expression isn't evaluated here, only stored.
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
			// TODO: Process method parameters if needed for definition?
		case *ComponentDecl:
			// Nested component definition? Allowed by grammar.
			// Recursively evaluate/register the nested component.
			_, err := evalComponentDecl(bodyNode, vm, v)
			if err != nil {
				return nil, fmt.Errorf("error defining nested component '%s' within '%s': %w", bodyNode.Name.Name, compName, err)
			}
			// case *LetStmt: // Allow internal constants?
			// case *Options: // Allow internal options?
		default:
			// Ignore/error on unknown body items?
			// return nil, fmt.Errorf("unexpected item type %T in component body '%s'", item, compName)
		}
	}

	// Register the fully processed definition
	err := v.RegisterComponentDef(compDef)
	if err != nil {
		return nil, err // Error if already registered
	}

	// Defining a component doesn't produce an executable OpNode
	return theNilNode, nil
}

func evalInstanceDecl(stmt *InstanceDecl, env *Env[any], v *VM) (OpNode, error) {
	instanceName := stmt.Name.Name
	componentTypeName := stmt.ComponentType.Name

	// --- Get Component Definition ---
	compDef, foundDef := v.ComponentDefRegistry[componentTypeName]
	if !foundDef {
		return nil, fmt.Errorf("unknown component type definition '%s' for instance '%s'", componentTypeName, instanceName)
	}

	// --- Find Component Constructor ---
	constructor, foundConst := v.ComponentRegistry[componentTypeName]
	if !foundConst {
		// If definition exists but constructor doesn't, implies user-defined component
		// We might need a generic constructor later, but error for now if no Go counterpart.
		// OR maybe the definition *is* the constructor for DSL-only components?
		// Let's assume for now that instantiable components need a registered Go constructor.
		return nil, fmt.Errorf("no registered Go constructor found for component type '%s' (instance '%s')", componentTypeName, instanceName)
	}

	// --- Evaluate & Prepare Overrides (Temporary Workaround) ---
	overrideValues := make(map[string]any)      // For component constructor params
	dependencyInstances := make(map[string]any) // For injecting dependencies map[local_uses_name]goInstance
	processedOverrides := make(map[string]bool) // Track processed overrides

	for _, assignStmt := range stmt.Overrides {
		assignVarName := assignStmt.Var.Name
		processedOverrides[assignVarName] = true // Mark as processed

		// --- Evaluate override value ---
		valueOpNode, err := Eval(assignStmt.Value, env, v)
		if err != nil {
			return nil, fmt.Errorf("evaluating override '%s' for instance '%s': %w", assignVarName, instanceName, err)
		}

		// --- Check if this override targets a 'param' or a 'uses' dependency ---
		if _, isParam := compDef.Params[assignVarName]; isParam {
			// --- TEMPORARY WORKAROUND for Param ---
			// (Same logic as before: assume LeafNode, extract raw value)
			leaf /*ok*/, _ := valueOpNode.(*LeafNode)
			// ... (value extraction logic from previous step) ...
			// Handle error if not LeafNode or probabilistic
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
			default:
				return nil, fmt.Errorf("unsupported outcome type %T for param override '%s'", outcome, assignVarName)
			}
			if !extractOk {
				return nil, fmt.Errorf("param override value '%s' is probabilistic", assignVarName)
			}
			overrideValues[assignVarName] = rawValue
			// --- END TEMPORARY WORKAROUND ---

		} else if /*usesDecl*/ _, isUses := compDef.Uses[assignVarName]; isUses {
			// This assignment is meant to satisfy a 'uses' dependency.
			// The RHS expression should evaluate to an identifier referencing another instance.
			identExpr, okIdent := assignStmt.Value.(*IdentifierExpr)
			if !okIdent {
				return nil, fmt.Errorf("value for 'uses' override '%s' must be an identifier referencing another instance, got %T", assignVarName, assignStmt.Value)
			}
			dependencyInstanceName := identExpr.Name

			// Look up the dependency instance in the current environment
			depGoInstance, foundDep := env.Get(dependencyInstanceName)
			if !foundDep {
				return nil, fmt.Errorf("dependency instance '%s' (for '%s.%s') not found in current scope", dependencyInstanceName, instanceName, assignVarName)
			}

			// TODO: Type check if depGoInstance is compatible with usesDecl.ComponentType? Requires reflection/registry info.
			// For now, assume the name lookup is sufficient.

			dependencyInstances[assignVarName] = depGoInstance // Store the Go instance to be injected

		} else {
			// The override name doesn't match a param or a uses declaration
			return nil, fmt.Errorf("unknown override target '%s' for instance '%s' of type '%s'", assignVarName, instanceName, componentTypeName)
		}
	}

	// --- Check if all 'uses' dependencies were satisfied by overrides ---
	for usesName := range compDef.Uses {
		if _, satisfied := dependencyInstances[usesName]; !satisfied {
			return nil, fmt.Errorf("missing override to satisfy 'uses %s: %s' dependency for instance '%s'", usesName, compDef.Uses[usesName].ComponentType.Name, instanceName)
		}
	}

	// --- Instantiate Component using Constructor ---
	// Pass only the 'param' overrides to the constructor
	goInstance, err := constructor(instanceName, overrideValues)
	if err != nil {
		return nil, fmt.Errorf("failed to construct component '%s' of type '%s': %w", instanceName, componentTypeName, err)
	}

	// --- Inject Dependencies ---
	// Use reflection or specific setter methods on the goInstance to inject
	// the instances from the dependencyInstances map.
	// This is complex and requires knowledge of the Go component struct fields/setters.
	// For now, let's skip the actual injection code, but assume it happens conceptually.
	// log.Printf("TODO: Inject dependencies into '%s': %v", instanceName, dependencyInstances)
	err = injectDependencies(goInstance, dependencyInstances) // Placeholder function
	if err != nil {
		return nil, fmt.Errorf("failed to inject dependencies into '%s': %w", instanceName, err)
	}

	// --- Store Go Instance in Environment ---
	if _, exists := env.Get(instanceName); exists {
		return nil, fmt.Errorf("identifier '%s' already exists in the current scope", instanceName)
	}
	env.Set(instanceName, goInstance)

	return theNilNode, nil
}

// Updated injection placeholder to actually set a field for testing
func injectDependencies(targetInstance any, dependencies map[string]any) error {
	targetVal := reflect.ValueOf(targetInstance)
	if targetVal.Kind() != reflect.Ptr || targetVal.IsNil() {
		return fmt.Errorf("targetInstance must be a non-nil pointer")
	}
	targetElem := targetVal.Elem()
	if targetElem.Kind() != reflect.Struct {
		return fmt.Errorf("targetInstance must point to a struct")
	}

	for name, dep := range dependencies {
		// Find field by matching name (case-sensitive) - simple approach
		// Real approach uses tags: field := findFieldByTag(targetElem, "sdluses", name)
		field := targetElem.FieldByName(strings.Title(name)) // Assumes CamelCase field names
		if !field.IsValid() {
			return fmt.Errorf("no field '%s' found in target %T", strings.Title(name), targetInstance)
		}
		if !field.CanSet() {
			return fmt.Errorf("cannot set field '%s' in target %T", strings.Title(name), targetInstance)
		}

		depVal := reflect.ValueOf(dep)
		if !depVal.Type().AssignableTo(field.Type()) {
			// Allow assigning MockDisk to interface{} field 'DB'
			if field.Type().Kind() == reflect.Interface && depVal.Type().Implements(field.Type()) {
				// Ok
			} else if field.Type().Kind() == reflect.Interface && field.Type().NumMethod() == 0 {
				// Ok to assign anything to empty interface{}
			} else {
				return fmt.Errorf("type mismatch: cannot assign dependency '%s' type %T to field type %s", name, dep, field.Type())
			}
		}
		field.Set(depVal)
	}
	return nil
}
