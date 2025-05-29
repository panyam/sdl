package runtime

import (
	"fmt"
	"log"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
)

// A simple evaluator
type SimpleEval struct {
	RootFile *FileInstance
}

// The main Eval loop of an expression/statement
func (s *SimpleEval) Eval(node Node, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	if env == nil {
		env = s.RootFile.Env.Push()
	}
	// fmt.Printf("Eval entry: %T - %s\n", node, node) // Debug entry
	switch n := node.(type) {
	// --- Statement Nodes ---
	case *BlockStmt:
		// With a block statement we usually push an extra context so it can be removed
		// at the end of the block
		return s.evalBlockStmt(n, env.Push()) // Pass nil context
	case *LetStmt:
		return s.evalLetStmt(n, env)
	case *SetStmt:
		return s.evalSetStmt(n, env)
	case *ExprStmt:
		return s.evalExprStmt(n, env)
	case *IfStmt:
		return s.evalIfStmt(n, env)
	case *DelayStmt:
		return s.evalDelayStmt(n, env)
	case *AssignmentStmt:
		return s.evalAssignmentStmt(n, env)

	// --- Expression Nodes ---
	case *LiteralExpr:
		return s.evalLiteralExpr(n, env)
	case *IdentifierExpr:
		return s.evalIdentifierExpr(n, env)
	case *BinaryExpr:
		return s.evalBinaryExpr(n, env)
	case *UnaryExpr:
		return s.evalUnaryExpr(n, env)
	case *decl.NewExpr:
		return s.evalNewExpr(n, env)
	case *decl.SampleExpr:
		return s.evalSampleExpr(n, env)
	case *decl.DistributeExpr:
		return s.evalDistributeExpr(n, env)
	case *CallExpr:
		return s.evalCallExpr(n, env)
	/* - TODO
	case *SwitchStmt: // <-- Will be implemented now
		return s.evalSwitchStmt(n, env)
	*/

	default:
		panic(fmt.Errorf("Eval not implemented for node type %T", node))
	}
}

func (s *SimpleEval) evalBlockStmt(b *BlockStmt, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	for _, statement := range b.Statements {
		var timeTaken core.Duration
		result, timeTaken, returned = s.Eval(statement, env)
		duration += timeTaken
		if returned {
			break
		}
	}
	return
}

// Evaluates the value of an Identifier Expression
func (s *SimpleEval) evalIdentifierExpr(i *IdentifierExpr, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	name := i.Name
	value, ok := env.Get(name)
	if !ok {
		err := fmt.Errorf("identifier not found '%s'", name)
		panic(err) // or register and continue?
	} else {
		result = value
	}
	return
}

// Evaluates the value of a Literal Expression
func (s *SimpleEval) evalLiteralExpr(e *LiteralExpr, _ *Env[Value]) (result Value, duration core.Duration, returned bool) {
	result = e.Value
	return
}

func (s *SimpleEval) evalSetStmt(set *SetStmt, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	// evaluate the Expression and unzip and assign to variables in the same environment
	result, duration, _ = s.Eval(set.Value, env)

	// Now find *where* it needs to be set, it can be:
	// 1. A var in the local env
	// 2. A member access expression - of the form a.b.c.d.e where a, b, c, d are components and e is a field/param name
	// or a component instance - either way it should have the same type as the RHS

	switch lhs := set.TargetExpr.(type) {
	case *IdentifierExpr:
		env.Set(lhs.Name, result)
	case *MemberAccessExpr:
		maeTarget, _, _ := s.Eval(lhs.Receiver, env)
		if maeTarget.Type.Tag != decl.TypeTagComponent {
			panic(fmt.Sprintf("Expected mae to be a component, found: %s -> %s", maeTarget, maeTarget.Type))
		}
		maeTarget.Value.(*ComponentInstance).Set(lhs.Member.Name, result)
	default:
		panic(fmt.Sprintf("Expected Identifier or MAE, Expected: %v", lhs))
	}
	// only duration increases - no change in result or returned status
	return
}

func (s *SimpleEval) evalLetStmt(l *LetStmt, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	// evaluate the Expression and unzip and assign to variables in the same environment
	result, duration, returned = s.Eval(l.Value, env)

	tupleValues, err := result.GetTuple()
	if err != nil {
		return
	}
	for i, val := range tupleValues {
		letvar := l.Variables[i].Name
		env.Set(letvar, val)
	}
	return
}

// Evaluates a distrbute expression that returns an Outcomes value type.
func (s *SimpleEval) evalDistributeExpr(d *decl.DistributeExpr, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	return
}

// Evaluate a sample expression that evaluates a random value based on the child
// distribution
func (s *SimpleEval) evalSampleExpr(samp *decl.SampleExpr, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	/*res*/ _, d, _ := s.Eval(samp.FromExpr, env)
	duration += d

	// res should be an Outcome type
	panic("Not sure whats next")
	/*
		outcomes, err := res.GetOutcomes()
		if err != nil {
			panic(err)
		}
		// TODO - Need a "Samplable or Distribution type"
		// Outcomes are essentially weight => Value.  When Values are discrete its not an issue.
		// But when you have Values that could be continuous (eg latencies) we need these to be sampled
				result, err = NewValue(outcomes.Sample())
				if err != nil {
					panic(err)
				}
			return
	*/
}

// Evaluate a component construction expression
func (s *SimpleEval) evalNewExpr(n *decl.NewExpr, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	// New contains the name of the component to instantiate
	// Since exection begins from a single File the File's env should contain the identifer
	compInst, err := s.RootFile.NewComponent(n.ComponentExpr.Name)
	if err != nil {
		panic(err)
	}
	compType := decl.ComponentType(compInst.ComponentDecl)
	result, err = NewValue(compType, compInst)
	if err != nil {
		panic(err)
	}
	return
}

func (s *SimpleEval) evalUnaryExpr(u *UnaryExpr, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	lr, ld, _ := s.Eval(u.Right, env)
	duration += ld

	// TODO - Evaluate based on the operator
	log.Println("Child Result: ", lr)
	return
}

func (s *SimpleEval) evalBinaryExpr(b *BinaryExpr, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	lr, ld, _ := s.Eval(b.Left, env)
	duration += ld
	rr, rd, _ := s.Eval(b.Right, env)
	duration += rd

	// TODO - Evaluate based on the operator
	log.Println("Left Result: ", lr)
	log.Println("Right Result: ", rr)
	return
}

/** Evaluate a Expr as a statement and return its value */
func (s *SimpleEval) evalExprStmt(stmt *ExprStmt, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	return s.Eval(stmt.Expression, env)
}

/** Evaluate a If and return its value */
func (s *SimpleEval) evalIfStmt(stmt *IfStmt, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	// Evaluate the condition expression to get its OpNode representation
	condResult, condDuration, _ := s.Eval(stmt.Condition, env)
	duration += condDuration

	if condResult.IsTrue() {
		thenResult, thenDuration, returned := s.Eval(stmt.Then, env)
		return thenResult, duration + thenDuration, returned
	} else {
		elseResult, elseDuration, returned := s.Eval(stmt.Then, env)
		return elseResult, duration + elseDuration, returned
	}
}

// Delay expressions
func (s *SimpleEval) evalDelayStmt(d *DelayStmt, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	// Evaluate the condition expression to get its OpNode representation
	_, duration, _ = s.Eval(d.Duration, env)
	return
}

///////////////////////////////////////////////////////

// Evaluate a Call and return its value
// Call expression are of the form a.b.c.d(params)
// The a.b.c.d must resolve to a callable (either a component method or a native function)
func (s *SimpleEval) evalCallExpr(expr *CallExpr, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	panic("TBD")
	/*
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
			argOpNodes[i], err = Eval(argExpr, env)
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
	*/
}

/** Evaluate a Assignment as a statement and return its value */
func (s *SimpleEval) evalAssignmentStmt(stmt *AssignmentStmt, env *Env[Value]) (result Value, duration core.Duration, returned bool) {
	panic("to be implemented")
}
