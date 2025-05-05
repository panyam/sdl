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
	ErrInvalidType          = errors.New("invalid type")
)

// Starts the execution of a single expression
func Eval(node Node, frame *Frame, v *VM) (OpNode, error) {
	// fmt.Printf("Eval entry: %T - %s\n", node, node) // Debug entry
	switch n := node.(type) {
	// We'll add cases here in subsequent phases
	case *LiteralExpr:
		return evalLiteral(n, frame, v)
	case *IdentifierExpr:
		return evalIdentifier(n, frame, v)

	// --- Statement Nodes ---
	case *BlockStmt:
		// When Eval is called directly on a BlockStmt (e.g., top level),
		// there is no initial context from an outer structure like IfStmt.
		// evalBlockStmt now returns the result directly, doesn't leave on stack implicitly
		return evalBlockStmt(n, NewFrame(frame), nil) // Pass nil context
	case *LetStmt:
		return evalLetStmt(n, frame, nil)
	case *BinaryExpr:
		return evalBinaryExpr(n, frame, v)
	case *ExprStmt:
		return evalExprStmt(n, frame, v)
	case *IfStmt: // <-- Will be implemented now
		return evalIfStmt(n, frame, v)
	case *SystemDecl: // <<< Added case
		return evalSystemDecl(n, frame, v)
	case *InstanceDecl: // <<< Added case
		return evalInstanceDecl(n, frame, v)
	case *CallExpr:
		return evalCallExpr(n, frame, v)
	case *MemberAccessExpr:
		// Member access is often handled *within* evalCallExpr,
		// but we might need a stub if it can be evaluated alone.
		return evalMemberAccess(n, frame, v) // <-- Call the actual implementation
	/* - TODO
	case *AssignmentStmt:
		return evalAssignmentStmt(n, frame, v)
	case *SwitchStmt: // <-- Will be implemented now
		return evalSwitchStmt(n, frame, v)
	*/

	default:
		return nil, fmt.Errorf("Eval not implemented for node type %T", node)
	}
}

/** Evaluate a literal and return its value */
func evalLiteral(expr *LiteralExpr, frame *Frame, v *VM) (OpNode, error) {
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
func evalIdentifier(expr *IdentifierExpr, frame *Frame, v *VM) (OpNode, error) {
	name := expr.Name
	value, ok := frame.Get(name)
	if !ok {
		return nil, fmt.Errorf("%w: identifier '%s'", ErrNotFound, name)
	}

	// Check the type of the value found in the frame
	switch val := value.(type) {
	case OpNode:
		// If it's already an OpNode (e.g., from a let statement), return it directly.
		return val, nil
	case ComponentRuntime:
		// If it's a ComponentRuntime instance, return an InstanceRefNode.
		// The consumer (e.g., assignment eval) will handle this reference.
		return &InstanceRefNode{InstanceName: name}, nil
	default:
		// Found something unexpected in the frame for this identifier.
		return nil, fmt.Errorf("%w: identifier '%s' resolved to unexpected type %T", ErrInvalidType, name, value)
	}
}

func evalLetStmt(stmt *LetStmt, frame *Frame, v *VM) (OpNode, error) {
	varName := stmt.Variable.Name
	valueOpNode, err := Eval(stmt.Value, frame, v)
	if err != nil {
		return nil, fmt.Errorf("evaluating value for let statement '%s': %w", varName, err)
	}

	// Store the resulting OpNode in the current frame
	frame.Set(varName, valueOpNode)

	// 'let' itself doesn't produce a value for subsequent sequence steps
	return theNilNode, nil
}

/** Evaluate a Call and return its value */
func evalCallExpr(expr *CallExpr, frame *Frame, v *VM) (val OpNode, err error) {
	var runtimeInstance ComponentRuntime
	var methodName string

	// 1. Evaluate the Function part to determine what is being called.
	//    Most common case: MemberAccessExpr (instance.method)
	if memberAccess, ok := expr.Function.(*MemberAccessExpr); ok {
		// Let's re-evaluate the receiver identifier directly to get the runtime instance
		receiverIdent, okIdent := memberAccess.Receiver.(*IdentifierExpr)
		if !okIdent {
			// If the receiver isn't a simple identifier (e.g., nested call result),
			// this scenario is more complex and might require the Tree Evaluator.
			// For Stage 1, let's assume simple instance.method calls.
			return nil, fmt.Errorf("method call receiver must be a simple identifier, found %T", memberAccess.Receiver)
		}

		instanceAny, found := frame.Get(receiverIdent.Name)
		if !found {
			return nil, fmt.Errorf("instance '%s' not found for method call '%s'", receiverIdent.Name, memberAccess.Member.Name)
		}

		runtimeInstance, ok = instanceAny.(ComponentRuntime)
		if !ok {
			// This indicates an error - something other than a ComponentRuntime was stored for this identifier.
			return nil, fmt.Errorf("identifier '%s' does not represent a component instance (found type %T)", receiverIdent.Name, instanceAny)
		}

		methodName = memberAccess.Member.Name // Get method name from the AST

	} else if identFunc, ok := expr.Function.(*IdentifierExpr); ok {
		// Case: Calling a potential global/builtin function (less common for components)
		// Look up in VM's internal funcs? Defer implementation for now.
		return nil, fmt.Errorf("calling standalone functions ('%s') not implemented yet", identFunc.Name)
	} else {
		// The function part is some other expression - likely invalid DSL structure
		// or requires evaluation first (Stage 2 Tree Evaluator needed).
		return nil, fmt.Errorf("invalid function/method expression type %T in call", expr.Function)
	}

	// 2. Evaluate Arguments -> []OpNode
	argOpNodes := make([]OpNode, len(expr.Args))
	for i, argExpr := range expr.Args {
		argNode, err := Eval(argExpr, frame, v)
		if err != nil {
			// TODO: Improve error reporting (arg index, method name)
			return nil, fmt.Errorf("evaluating argument %d for method '%s': %w", i, methodName, err)
		}
		argOpNodes[i] = argNode
	}

	// 3. Invoke the method on the ComponentRuntime instance
	//    Pass the current frame (callFrame) for context.
	resultOpNode, err := runtimeInstance.InvokeMethod(methodName, argOpNodes, v, frame)
	if err != nil {
		// Error could be method not found, arg mismatch (checked inside InvokeMethod),
		// or error during execution (native reflection call fail, DSL body eval fail).
		return nil, fmt.Errorf("error calling method '%s' on instance '%s': %w", methodName, runtimeInstance.GetInstanceName(), err)
	}

	// 4. Return the resulting OpNode
	return resultOpNode, nil
}

// evalMemberAccess - Placeholder implementation
// This is called if a MemberAccessExpr is evaluated *outside* a CallExpr context.
// For now, it's unclear if this is needed or how it should behave in Stage 1.
// A simple member access like `myInstance.param` might need the Tree Evaluator
// to extract the parameter value *during execution*. Stage 1 focuses on structure.
// Let's return an error indicating it's not directly evaluatable this way yet.
func evalMemberAccess(expr *MemberAccessExpr, frame *Frame, v *VM) (OpNode, error) {
	// Evaluating receiver to ensure it exists might be useful
	receiverIdent, okIdent := expr.Receiver.(*IdentifierExpr)
	if !okIdent {
		return nil, fmt.Errorf("member access receiver must be a simple identifier, found %T", expr.Receiver)
	}
	instanceAny, found := frame.Get(receiverIdent.Name)
	if !found {
		return nil, fmt.Errorf("identifier '%s' not found for member access '%s'", receiverIdent.Name, expr.Member.Name)
	}
	_ /*runtimeInstance*/, okRuntime := instanceAny.(ComponentRuntime)
	if !okRuntime {
		return nil, fmt.Errorf("identifier '%s' does not represent a component instance (found type %T)", receiverIdent.Name, instanceAny)
	}

	// However, *getting* the parameter might require the Tree Evaluator.
	// We *could* try calling runtimeInstance.GetParam() here, but GetParam returns an OpNode.
	// What does it mean to "evaluate" `instance.param` in Stage 1? Maybe it should
	// return a specific OpNode type representing the parameter access?
	// For now, let's return an error, assuming direct member access isn't handled in Stage 1.
	// It's primarily used within method calls.
	return nil, fmt.Errorf("direct evaluation of member access '%s.%s' not supported in Stage 1; use within method calls or assignments", receiverIdent.Name, expr.Member.Name)
	// Alternative: Could return runtimeInstance.GetParam(expr.Member.Name) if GetParam is robust.
}

/** Evaluate a Block and return its value */
func evalBlockStmt(stmt *BlockStmt, frame *Frame, v *VM) (OpNode, error) {
	blockFrame := NewFrame(frame) // Create block scope
	steps := make([]OpNode, 0, len(stmt.Statements))

	for _, statement := range stmt.Statements {
		resultNode, err := Eval(statement, blockFrame, v)
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
func evalIfStmt(stmt *IfStmt, frame *Frame, v *VM) (val OpNode, err error) {
	// Evaluate the condition expression to get its OpNode representation
	conditionNode, err := Eval(stmt.Condition, frame, v)
	if err != nil {
		// TODO: Improve error reporting
		return nil, fmt.Errorf("evaluating condition for if statement: %w", err)
	}

	// Evaluate the 'then' block to get its OpNode representation
	// Note: Use the *same* environment level as the if statement itself.
	// Scoping for variables *inside* the block is handled by evalBlockStmt.
	thenNode, err := Eval(stmt.Then, frame, v)
	if err != nil {
		// TODO: Improve error reporting
		return nil, fmt.Errorf("evaluating 'then' block for if statement: %w", err)
	}

	// Evaluate the 'else' block/statement, if it exists
	var elseNode OpNode = theNilNode // Default to NilNode if no else
	if stmt.Else != nil {
		elseNode, err = Eval(stmt.Else, frame, v)
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
func evalSwitchStmt(stmt *SwitchStmt, frame *Frame, v *VM) (val Value, err error) {
	return
}

/** Evaluate a Expr as a statement and return its value */
func evalExprStmt(stmt *ExprStmt, frame *Frame, v *VM) (OpNode, error) {
	// Evaluate the expression and return its OpNode result
	return Eval(stmt.Expression, frame, v)
}

/** Evaluate a Assignment as a statement and return its value */
func evalAssignmentStmt(stmt *AssignmentStmt, frame *Frame, v *VM) (val Value, err error) {
	return
}

func evalBinaryExpr(expr *BinaryExpr, frame *Frame, v *VM) (OpNode, error) {
	// Recursively evaluate left and right operands
	leftNode, err := Eval(expr.Left, frame, v)
	if err != nil {
		// TODO: Improve error reporting with position info
		return nil, fmt.Errorf("evaluating left operand for '%s': %w", expr.Operator, err)
	}

	rightNode, err := Eval(expr.Right, frame, v)
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

// --- evalSystemDecl (Processes system body) ---
func evalSystemDecl(stmt *SystemDecl, frame *Frame, v *VM) (OpNode, error) {
	// Systems define a scope, but for now, let's use the passed frame.
	// A system run might eventually need its own top-level frame.
	// systemFrame := NewFrame(frame) // Option for later

	for _, item := range stmt.Body {
		// Evaluate each item within the system's context
		// For now, InstanceDecl modifies the passed frame.
		_, err := Eval(item, frame, v) // Use passed frame
		if err != nil {
			return nil, fmt.Errorf("error evaluating item in system '%s': %w", stmt.Name, err)
		}
		// Ignore the OpNode returned by body items (e.g., InstanceDecl returns NilNode)
	}

	// System declaration itself doesn't produce a value OpNode
	return theNilNode, nil
}

// --- evalInstanceDecl (Instantiates Native or DSL component) ---
func evalInstanceDecl(stmt *InstanceDecl, frame *Frame, v *VM) (OpNode, error) {
	instanceName := stmt.Name
	componentTypeName := stmt.ComponentType.Name

	// Check if instance name already exists in the current scope
	if _, exists := frame.Get(instanceName); exists {
		return nil, fmt.Errorf("identifier '%s' already exists in the current scope", instanceName)
	}

	// --- Call the Factory Method ---
	runtimeInstance, err := v.CreateInstance(componentTypeName, instanceName, stmt.Overrides, frame)
	if err != nil {
		// Wrap error with position info from the InstanceDecl statement
		return nil, fmt.Errorf("failed to create instance '%s' (type %s) at pos %d: %w", instanceName, componentTypeName, stmt.Pos(), err)
	}

	// Store the resulting ComponentRuntime in the current frame's locals
	frame.Set(instanceName, runtimeInstance)

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
	case *core.Outcomes[bool]:
		rawValue, extractOk = outcome.GetValue()
	case *core.Outcomes[string]:
		rawValue, extractOk = outcome.GetValue()
	// case *core.Outcomes[float64]:
	case *core.Outcomes[core.Duration]: // Added Duration case
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
			continue // Skip if no matching field found (could be error?)
		}
		if !field.CanSet() {
			return fmt.Errorf("cannot set field '%s' in target %T", fieldName, targetInstance)
		}

		// Get the actual underlying value (Go instance or *UDComponent)
		var depValueToInject any
		if adapter, ok := depRuntime.(*NativeComponent); ok {
			depValueToInject = adapter.GoInstance
		} else if dslInst, ok := depRuntime.(*UDComponent); ok {
			depValueToInject = dslInst
		} else {
			return fmt.Errorf("dependency %s has unknown ComponentRuntime type %T", name, depRuntime)
		}

		depVal := reflect.ValueOf(depValueToInject)

		if !depVal.IsValid() { // Handle nil dependency value if necessary
			// Check if field type is pointer or interface, if so, setting nil is okay
			if field.Type().Kind() == reflect.Ptr || field.Type().Kind() == reflect.Interface || field.Type().Kind() == reflect.Map || field.Type().Kind() == reflect.Slice {
				// Set the zero value for the field type (which is nil for pointers/interfaces/maps/slices)
				field.Set(reflect.Zero(field.Type()))
				continue
			} else {
				return fmt.Errorf("dependency '%s' value is nil, cannot assign to non-pointer/non-interface field '%s'", name, fieldName)
			}
		}
		if !depVal.Type().AssignableTo(field.Type()) {
			if field.Type().Kind() == reflect.Interface {
				if !depVal.Type().Implements(field.Type()) {
					if field.Type().NumMethod() != 0 { // Allow assignment to empty interface{}
						return fmt.Errorf("type mismatch: cannot assign dependency '%s' type %T to interface field '%s' type %s (does not implement)", name, depValueToInject, fieldName, field.Type())
					}
				}
			} else {
				return fmt.Errorf("type mismatch: cannot assign dependency '%s' type %T to field '%s' type %s", name, depValueToInject, fieldName, field.Type())
			}
		}
		field.Set(depVal)
		// log.Printf("    Successfully injected '%s' into field '%s'", name, fieldName)
	}
	return nil
}
