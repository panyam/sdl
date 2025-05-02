package decl

import (
	"errors"
	"fmt"

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
