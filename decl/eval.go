package decl

import (
	"errors"
	"fmt"

	"github.com/panyam/sdl/core"
)

var (
	ErrNotImplemented       = errors.New("evaluation for this node type not implemented")
	ErrNotFound             = errors.New("identifier not found")
	ErrInternalFuncNotFound = errors.New("internal function not found")
	ErrUnsupportedType      = errors.New("unsupported type for operation")
	ErrInvalidType          = errors.New("invalid type")
)

// Starts the execution of a single expression
func Eval(node Node, frame *Frame, v *VM) (val Value, err error) {
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
	case *SystemDecl:
		return evalSystemDecl(n, frame, v)
	case *InstanceDecl:
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
		return val, fmt.Errorf("Eval not implemented for node type %T", node)
	}
}

/** Evaluate a literal and return its value */
func evalLiteral(expr *LiteralExpr, frame *Frame, v *VM) (val Value, err error) {
	return expr.Value, nil
	/*
		var valueOutcome any
		latencyOutcome := ZeroLatencyOutcome()

		switch expr.Kind {
		case "INT":
			v, ok := rawValue.(int64)
			if !ok {
				return nil, fmt.Errorf("internal error: parsed INT literal '%s' resulted in %T, expected int64", expr.Value, rawValue)
			}
			valueOutcome = (&core.Outcomes[int64]{}).Add(1.0, v)
		case "FLOAT":
			v, ok := rawValue.(float64)
			if !ok {
				return nil, fmt.Errorf("internal error: parsed FLOAT literal '%s' resulted in %T, expected float64", expr.Value, rawValue)
			}
			valueOutcome = (&core.Outcomes[float64]{}).Add(1.0, v)
		case "BOOL":
			v, ok := rawValue.(bool)
			if !ok {
				return nil, fmt.Errorf("internal error: parsed BOOL literal '%s' resulted in %T, expected bool", expr.Value, rawValue)
			}
			valueOutcome = (&core.Outcomes[bool]{}).Add(1.0, v)
		case "STRING":
			v, ok := rawValue.(string)
			if !ok {
				return nil, fmt.Errorf("internal error: parsed STRING literal '%s' resulted in %T, expected string", expr.Value, rawValue)
			}
			valueOutcome = (&core.Outcomes[string]{}).Add(1.0, v)
		// TODO: case "DURATION": Handle duration parsing and create *core.Outcomes[core.Duration] value outcome
		default:
			return nil, fmt.Errorf("unsupported literal kind '%s' in evalLiteral", expr.Kind)
		}

		state := &VarState{
			ValueOutcome:   valueOutcome,
			LatencyOutcome: latencyOutcome,
		}
		val.Type = OpNodeType
		val.Value = &LeafNode{State: state}
		return
	*/
}

/** Evaluate a Identifier and return its value */
func evalIdentifier(expr *IdentifierExpr, frame *Frame, v *VM) (val Value, err error) {
	name := expr.Name
	value, ok := frame.Get(name)
	if !ok {
		return value, fmt.Errorf("%w: identifier '%s'", ErrNotFound, name)
	}
	return value, nil
}

func evalLetStmt(stmt *LetStmt, frame *Frame, v *VM) (val Value, err error) {
	varName := stmt.Variables[0].Name
	val, err = Eval(stmt.Value, frame, v)
	if err != nil {
		return val, fmt.Errorf("evaluating value for let statement '%s': %w", varName, err)
	}

	// Store the resulting OpNode in the current frame
	frame.Set(varName, val)

	// 'let' itself doesn't produce a value for subsequent sequence steps
	return
}

/** Evaluate a Call and return its value */
func evalCallExpr(expr *CallExpr, frame *Frame, v *VM) (val Value, err error) {
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
			return val, fmt.Errorf("method call receiver must be a simple identifier, found %T", memberAccess.Receiver)
		}

		instanceAny, found := frame.Get(receiverIdent.Name)
		if !found {
			return val, fmt.Errorf("instance '%s' not found for method call '%s'", receiverIdent.Name, memberAccess.Member.Name)
		}

		runtimeInstance, ok = instanceAny.Value.(ComponentRuntime)
		if !ok {
			// This indicates an error - something other than a ComponentRuntime was stored for this identifier.
			return val, fmt.Errorf("identifier '%s' does not represent a component instance (found type %T)", receiverIdent.Name, instanceAny)
		}

		methodName = memberAccess.Member.Name // Get method name from the AST

	} else if identFunc, ok := expr.Function.(*IdentifierExpr); ok {
		// Case: Calling a potential global/builtin function (less common for components)
		// Look up in VM's internal funcs? Defer implementation for now.
		return val, fmt.Errorf("calling standalone functions ('%s') not implemented yet", identFunc.Name)
	} else {
		// The function part is some other expression - likely invalid DSL structure
		// or requires evaluation first (Stage 2 Tree Evaluator needed).
		return val, fmt.Errorf("invalid function/method expression type %T in call", expr.Function)
	}

	// 2. Evaluate Arguments -> []OpNode
	argOpNodes := make([]Value, len(expr.Args))
	for i, argExpr := range expr.Args {
		argOpNodes[i], err = Eval(argExpr, frame, v)
		if err != nil {
			// TODO: Improve error reporting (arg index, method name)
			return val, fmt.Errorf("evaluating argument %d for method '%s': %w", i, methodName, err)
		}
	}

	// 3. Invoke the method on the ComponentRuntime instance
	//    Pass the current frame (callFrame) for context.
	resultOpNode, err := runtimeInstance.InvokeMethod(methodName, argOpNodes, v, frame)
	if err != nil {
		// Error could be method not found, arg mismatch (checked inside InvokeMethod),
		// or error during execution (native reflection call fail, DSL body eval fail).
		return val, fmt.Errorf("error calling method '%s' on instance '%s': %w", methodName, runtimeInstance.GetInstanceName(), err)
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
func evalMemberAccess(expr *MemberAccessExpr, frame *Frame, v *VM) (val Value, err error) {
	// Evaluating receiver to ensure it exists might be useful
	receiverIdent, okIdent := expr.Receiver.(*IdentifierExpr)
	if !okIdent {
		return val, fmt.Errorf("member access receiver must be a simple identifier, found %T", expr.Receiver)
	}
	instanceAny, found := frame.Get(receiverIdent.Name)
	if !found {
		return val, fmt.Errorf("identifier '%s' not found for member access '%s'", receiverIdent.Name, expr.Member.Name)
	}
	_ /*runtimeInstance*/, okRuntime := instanceAny.Value.(ComponentRuntime)
	if !okRuntime {
		return val, fmt.Errorf("identifier '%s' does not represent a component instance (found type %T)", receiverIdent.Name, instanceAny)
	}

	// However, *getting* the parameter might require the Tree Evaluator.
	// We *could* try calling runtimeInstance.GetParam() here, but GetParam returns an OpNode.
	// What does it mean to "evaluate" `instance.param` in Stage 1? Maybe it should
	// return a specific OpNode type representing the parameter access?
	// For now, let's return an error, assuming direct member access isn't handled in Stage 1.
	// It's primarily used within method calls.
	return val, fmt.Errorf("direct evaluation of member access '%s.%s' not supported in Stage 1; use within method calls or assignments", receiverIdent.Name, expr.Member.Name)
	// Alternative: Could return runtimeInstance.GetParam(expr.Member.Name) if GetParam is robust.
}

/** Evaluate a Block and return its value */
func evalBlockStmt(stmt *BlockStmt, frame *Frame, v *VM) (val Value, err error) {
	blockFrame := NewFrame(frame) // Create block scope
	steps := make([]Value, 0, len(stmt.Statements))

	for _, statement := range stmt.Statements {
		resultNode, err := Eval(statement, blockFrame, v)
		if err != nil {
			// TODO: Improve error reporting with position info from statement
			return val, fmt.Errorf("error in block statement: %w", err)
		}

		// Only include non-nil nodes in the sequence
		if resultNode.Type != nil {
			steps = append(steps, resultNode)
		}
	}

	// Determine return value based on collected steps
	if len(steps) == 0 {
		return val, nil // Empty block or only let statements
	}
	if len(steps) == 1 {
		return steps[0], nil // Single effective statement, return its node
	}
	return NewValue(OpNodeType, &SequenceNode{Steps: steps})
}

/** Evaluate a If and return its value */
func evalIfStmt(stmt *IfStmt, frame *Frame, v *VM) (val Value, err error) {
	// Evaluate the condition expression to get its OpNode representation
	conditionNode, err := Eval(stmt.Condition, frame, v)
	if err != nil {
		// TODO: Improve error reporting
		return val, fmt.Errorf("evaluating condition for if statement: %w", err)
	}

	// Evaluate the 'then' block to get its OpNode representation
	// Note: Use the *same* environment level as the if statement itself.
	// Scoping for variables *inside* the block is handled by evalBlockStmt.
	thenNode, err := Eval(stmt.Then, frame, v)
	if err != nil {
		// TODO: Improve error reporting
		return val, fmt.Errorf("evaluating 'then' block for if statement: %w", err)
	}

	// Evaluate the 'else' block/statement, if it exists
	var elseNode Value
	if stmt.Else != nil {
		elseNode, err = Eval(stmt.Else, frame, v)
		if err != nil {
			// TODO: Improve error reporting
			return val, fmt.Errorf("evaluating 'else' block for if statement: %w", err)
		}
	}

	// Construct and return the IfChoiceNode representing the structure
	return NewValue(OpNodeType, &IfChoiceNode{
		Condition: conditionNode,
		Then:      thenNode,
		Else:      elseNode,
	})
}

/** Evaluate a Expr as a statement and return its value */
func evalExprStmt(stmt *ExprStmt, frame *Frame, v *VM) (val Value, err error) {
	// Evaluate the expression and return its OpNode result
	return Eval(stmt.Expression, frame, v)
}

/** Evaluate a Assignment as a statement and return its value */
func evalAssignmentStmt(stmt *AssignmentStmt, frame *Frame, v *VM) (val Value, err error) {
	return
}

func evalBinaryExpr(expr *BinaryExpr, frame *Frame, v *VM) (val Value, err error) {
	// Recursively evaluate left and right operands
	leftNode, err := Eval(expr.Left, frame, v)
	if err != nil {
		// TODO: Improve error reporting with position info
		return val, fmt.Errorf("evaluating left operand for '%s': %w", expr.Operator, err)
	}

	rightNode, err := Eval(expr.Right, frame, v)
	if err != nil {
		// TODO: Improve error reporting with position info
		return val, fmt.Errorf("evaluating right operand for '%s': %w", expr.Operator, err)
	}

	// Check operator validity if needed (parser should handle this mostly)
	// switch expr.Operator {
	// case "+", "-", "*", "/", "%", "&&", "||", "==", "!=", "<", "<=", ">", ">=":
	// 	// Valid operator
	// default:
	// 	return nil, fmt.Errorf("unsupported binary operator '%s'", expr.Operator)
	// }

	// Construct and return the BinaryOpNode representing the operation
	return NewValue(OpNodeType, &BinaryOpNode{
		Op:    expr.Operator,
		Left:  leftNode,
		Right: rightNode,
	})
}

// --- evalSystemDecl (Processes system body) ---
func evalSystemDecl(stmt *SystemDecl, frame *Frame, v *VM) (val Value, err error) {
	// Systems define a scope, but for now, let's use the passed frame.
	// A system run might eventually need its own top-level frame.
	// systemFrame := NewFrame(frame) // Option for later

	for _, item := range stmt.Body {
		// Evaluate each item within the system's context
		// For now, InstanceDecl modifies the passed frame.
		_, err := Eval(item, frame, v) // Use passed frame
		if err != nil {
			return val, fmt.Errorf("error evaluating item in system '%s': %w", stmt.NameNode.Name, err)
		}
		// Ignore the OpNode returned by body items (e.g., InstanceDecl returns NilNode)
	}

	// System declaration itself doesn't produce a value OpNode
	return val, nil
}

// --- evalInstanceDecl (Instantiates Native or DSL component) ---
func evalInstanceDecl(stmt *InstanceDecl, frame *Frame, v *VM) (val Value, err error) {
	instanceName := stmt.NameNode.Name
	componentTypeName := stmt.ComponentType.Name

	// Check if instance name already exists in the current scope
	if _, exists := frame.Get(instanceName); exists {
		return val, fmt.Errorf("identifier '%s' already exists in the current scope", instanceName)
	}

	// First evaluate the value of the overrides before passing them to the instance creator

	// --- Call the Factory Method ---
	runtimeInstance, err := v.CreateInstance(componentTypeName, instanceName, stmt.Overrides, frame)
	if err != nil {
		// Wrap error with position info from the InstanceDecl statement
		return val, fmt.Errorf("failed to create instance '%s' (type %s) at pos %d: %w", instanceName, componentTypeName, stmt.Pos(), err)
	}

	// Store the resulting ComponentRuntime in the current frame's locals
	val, err = NewValue(ComponentType, runtimeInstance)
	if err != nil {
		return
	}
	frame.Set(instanceName, val)
	return
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
