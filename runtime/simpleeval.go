package runtime

import (
	"fmt"
	"log"
	"math/rand"
	"slices"
	"time"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
)

func ensureTypes(typ *Type, types ...*Type) {
	if !slices.ContainsFunc(types, func(t *Type) bool { return t.Equals(typ) }) {
		panic("Expected given type but failed")
	}
}

// A simple evaluator
type SimpleEval struct {
	ErrorCollector
	RootFile *FileInstance
	Rand     *rand.Rand
	Errors   []error
}

func NewSimpleEval(fi *FileInstance) *SimpleEval {
	out := &SimpleEval{
		RootFile: fi,
		Rand:     rand.New(rand.NewSource(time.Now().UnixMicro())),
	}
	out.MaxErrors = 1
	return out
}

func (s *SimpleEval) EvalInitSystem(sys *SystemInstance, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	stmts, err := sys.Initializer()
	ensureNoErr(err)
	_, returned, timeTaken := s.EvalStatements(stmts.Statements, env)
	*currTime += timeTaken
	result.Time = timeTaken

	log.Println("Compiled statements:")
	decl.PPrint(stmts)

	// Now check if all things are initialized
	uninited := sys.GetUninitializedComponents(env)
	for _, unin := range uninited {
		log.Println("Uninitialized Dependency: ", unin.Attrib)
		curr := unin.From
		out := unin.Attrib
		errorFile := sys.System.ParentFileDecl.FullPath
		errorLoc := unin.Pos
		for curr != nil {
			errorLoc = curr.Pos
			out = curr.CompInst.ComponentDecl.Name.Value + "." + out
			log.Printf("    From %s, Line: %d, Column %d", curr.CompInst.ComponentDecl.ParentFileDecl.FullPath, curr.Pos.Line, curr.Pos.Col)
			curr = curr.From
		}
		s.AddErrors(fmt.Errorf("%s, Line: %d, Col: %d - Uninitialized Dependency: %s", errorFile, errorLoc.Line, errorLoc.Col, out))
	}
	return
}

func (s *SimpleEval) EvalStatements(stmts []Stmt, env *Env[Value]) (result []Value, returned bool, timeTaken core.Duration) {
	for _, item := range stmts {
		res, ret := s.Eval(item, env, &timeTaken)
		result = append(result, res)
		if ret {
			returned = true
			break
		}
	}
	return
}

// The main Eval loop of an expression/statement
func (s *SimpleEval) Eval(node Node, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	// fmt.Printf("Eval entry: %T - %s\n", node, node) // Debug entry
	switch n := node.(type) {
	// --- Statement Nodes ---
	case *BlockStmt:
		// With a block statement we usually push an extra context so it can be removed
		// at the end of the block
		return s.evalBlockStmt(n, env.Push(), currTime) // Pass nil context
	case *LetStmt:
		return s.evalLetStmt(n, env, currTime)
	case *SetStmt:
		return s.evalSetStmt(n, env, currTime)
	case *ForStmt:
		return s.evalForStmt(n, env, currTime)
	case *ReturnStmt:
		return s.evalReturnStmt(n, env, currTime)
	case *ExprStmt:
		return s.evalExprStmt(n, env, currTime)
	case *IfStmt:
		return s.evalIfStmt(n, env, currTime)
	case *DelayStmt:
		return s.evalDelayStmt(n, env, currTime)
	case *AssignmentStmt:
		return s.evalAssignmentStmt(n, env, currTime)

	// --- Expression Nodes ---
	case *LiteralExpr:
		return s.evalLiteralExpr(n, env, currTime)
	case *IdentifierExpr:
		return s.evalIdentifierExpr(n, env, currTime)
	case *BinaryExpr:
		return s.evalBinaryExpr(n, env, currTime)
	case *UnaryExpr:
		return s.evalUnaryExpr(n, env, currTime)
	case *decl.NewExpr:
		return s.evalNewExpr(n, env, currTime)
	case *decl.SampleExpr:
		return s.evalSampleExpr(n, env, currTime)
	case *decl.DistributeExpr:
		return s.evalDistributeExpr(n, env, currTime)
	case *CallExpr:
		return s.evalCallExpr(n, env, currTime)
	case *MemberAccessExpr:
		return s.evalMemberAccessExpr(n, env, currTime)
	case *TupleExpr:
		return s.evalTupleExpr(n, env, currTime)
	case *GoExpr:
		return s.evalGoExpr(n, env, currTime)
	case *WaitExpr:
		return s.evalWaitExpr(n, env, currTime)
	default:
		panic(fmt.Errorf("Eval not implemented for node type %T", node))
	}
}

func (s *SimpleEval) evalBlockStmt(b *BlockStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	for _, statement := range b.Statements {
		result, returned = s.Eval(statement, env, currTime)
		if returned {
			break
		}
	}
	return
}

// Evaluates the value of an Identifier Expression
func (s *SimpleEval) evalIdentifierExpr(i *IdentifierExpr, env *Env[Value], _ *core.Duration) (result Value, returned bool) {
	name := i.Value
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
func (s *SimpleEval) evalLiteralExpr(e *LiteralExpr, _ *Env[Value], _ *core.Duration) (result Value, returned bool) {
	result = e.Value
	return
}

func (s *SimpleEval) evalSetStmt(set *SetStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	// evaluate the Expression and unzip and assign to variables in the same environment
	result, _ = s.Eval(set.Value, env, currTime)

	// Now find *where* it needs to be set, it can be:
	// 1. A var in the local env
	// 2. A member access expression - of the form a.b.c.d.e where a, b, c, d are components and e is a field/param name
	// or a component instance - either way it should have the same type as the RHS

	switch lhs := set.TargetExpr.(type) {
	case *IdentifierExpr:
		env.Set(lhs.Value, result)
	case *MemberAccessExpr:
		maeTarget, _ := s.Eval(lhs.Receiver, env, currTime)
		if maeTarget.Type.Tag != decl.TypeTagComponent {
			panic(fmt.Sprintf("Expected mae to be a component, found: %s -> %s", maeTarget.String(), maeTarget.Type))
		}
		maeTarget.Value.(*ComponentInstance).Set(lhs.Member.Value, result)
	default:
		panic(fmt.Sprintf("Expected Identifier or MAE, Expected: %v", lhs))
	}
	// only duration increases - no change in result or returned status
	return
}

func (s *SimpleEval) evalReturnStmt(r *ReturnStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	result, _ = s.Eval(r.ReturnValue, env, currTime)
	return result, true
}

func (s *SimpleEval) evalForStmt(f *ForStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	var err error
	counter := int64(0)

	// now evaluate the body
	for {
		condVal, _ := s.Eval(f.Condition, env, currTime)
		isCondInt := condVal.Type.Equals(IntType)
		condIntVal := int64(0)
		condBoolVal := true
		if isCondInt {
			condIntVal, err = condVal.GetInt()
		} else {
			condBoolVal, err = condVal.GetBool()
		}
		ensureNoErr(err)
		if isCondInt {
			if condIntVal > 0 && counter >= condIntVal {
				return
			}
		} else if !condBoolVal {
			return
		}

		// Evaluate body and see if it returned
		bodyRes, bodyReturned := s.Eval(f.Body, env, currTime)
		if bodyReturned {
			return bodyRes, bodyReturned
		}
		counter += 1
	}
}

func (s *SimpleEval) evalLetStmt(l *LetStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	// evaluate the Expression and unzip and assign to variables in the same environment
	result, returned = s.Eval(l.Value, env, currTime)

	if len(l.Variables) == 1 {
		letvar := l.Variables[0].Value
		env.Set(letvar, result)
	} else {
		// If there are multiple variables, we expect the result to be a tuple
		tupleValues, err := result.GetTuple()
		ensureNoErr(err)
		for i, val := range tupleValues {
			letvar := l.Variables[i].Value
			env.Set(letvar, val)
		}
	}
	return
}

// Evaluates a distrbute expression that returns an Outcomes value type.
func (s *SimpleEval) evalDistributeExpr(dist *decl.DistributeExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	var totalValue Value
	totalProb := 0.0
	if dist.TotalProb != nil {
		totalValue, _ = s.Eval(dist.TotalProb, env, currTime)
		ensureTypes(totalValue.Type, IntType, FloatType)
		if totalValue.Type.Equals(IntType) {
			totalProb = float64(totalValue.IntVal())
		} else {
			totalProb = totalValue.FloatVal()
		}
	}
	var caseConds []Value
	var caseBodies []Value
	var totalCasesProb = 0.0
	outcomes := &core.Outcomes[Value]{}
	var outcomeType *decl.Type
	for idx, caseExp := range dist.Cases {
		condVal, _ := s.Eval(caseExp.Condition, env, currTime)
		caseConds = append(caseConds, condVal)
		ensureTypes(condVal.Type, IntType, FloatType)
		condProb := 0.0
		if condVal.Type.Equals(IntType) {
			condProb = float64(condVal.IntVal())
		} else {
			condProb = condVal.FloatVal()
		}
		totalCasesProb += condProb

		bodyVal, _ := s.Eval(caseExp.Body, env, currTime)
		if idx == 0 {
			outcomeType = bodyVal.Type
		} else if !bodyVal.Type.Equals(outcomeType) {
			panic("type mismatch - should have been checked by type checker")
		}
		caseBodies = append(caseBodies, bodyVal)
		outcomes.Add(condProb, bodyVal)
	}

	// check default
	if dist.Default != nil {
		if dist.TotalProb == nil {
			panic("Default cannot exist when total prob exists - type checker cannot check this??")
		}
		defaultValue, _ := s.Eval(dist.Default, env, currTime)
		if !defaultValue.Type.Equals(outcomeType) {
			panic("type mismatch - should have been checked by type checker")
		}
		defaultProb := totalProb - totalCasesProb
		if defaultProb > 0 {
			outcomes.Add(defaultProb, defaultValue)
		}
	}
	outVal, err := NewValue(decl.OutcomesType(outcomeType), outcomes)
	ensureNoErr(err, "unexpected error.  should have been caught by validator?")
	return outVal, false
}

// Evaluate a sample expression that evaluates a random value based on the child
// distribution
func (s *SimpleEval) evalSampleExpr(samp *decl.SampleExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	res, _ := s.Eval(samp.FromExpr, env, currTime)

	outcomes := res.OutcomesVal()
	result, _ = outcomes.Sample(s.Rand)
	// log.Println("Sampled from: ", outcomes, result, ok)
	return
	// TODO - Need a "Samplable or Distribution type"
	// Outcomes are essentially weight => Value.  When Values are discrete its not an issue.
	// But when you have Values that could be continuous (eg latencies) we need these to be sampled
}

// Evaluate a component construction expression
func (s *SimpleEval) evalNewExpr(n *decl.NewExpr, _ *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	// New contains the name of the component to instantiate
	// Since exection begins from a single File the File's env should contain the identifer
	compInst, result, err := NewComponentInstance(s.RootFile, n.ComponentDecl)
	ensureNoErr(err)

	// Now for any "instantiated" components set it here
	// For native components we dont need this as they will take care of it themselves -
	// ie initialization is "entire"
	// Later on we can also have native components expose their dependencies so we can take
	// care of it but out of scope for now
	if !compInst.IsNative {
		stmts, err := compInst.Initializer()
		ensureNoErr(err)
		_, _, timeTaken := s.EvalStatements(stmts.Statements, compInst.InitialEnv)
		*currTime += timeTaken
		result.Time += timeTaken
	}
	return
}

func (s *SimpleEval) evalUnaryExpr(u *UnaryExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	lr, _ := s.Eval(u.Right, env, currTime)

	// TODO - Evaluate based on the operator
	if u.Operator == "not" {
		if lr.Type.Equals(BoolType) {
			lr.Value = !lr.Value.(bool)
		} else {
			panic(fmt.Sprintf("Unary operator not supported for type: %s", lr.Type))
		}
	} else {
		panic(fmt.Sprintf("Unary operator not supported: %s", u.Operator))
	}
	return
}

func (s *SimpleEval) evalBinaryExpr(b *BinaryExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	lr, _ := s.Eval(b.Left, env, currTime)
	rr, _ := s.Eval(b.Right, env, currTime)

	// TODO - Evaluate based on the operator
	log.Println("Left Result: ", lr)
	log.Println("Right Result: ", rr)
	return
}

/** Evaluate a Expr as a statement and return its value */
func (s *SimpleEval) evalExprStmt(stmt *ExprStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	return s.Eval(stmt.Expression, env, currTime)
}

/** Evaluate a If and return its value */
func (s *SimpleEval) evalIfStmt(stmt *IfStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	// Evaluate the condition expression to get its OpNode representation
	condResult, _ := s.Eval(stmt.Condition, env, currTime)

	if condResult.IsTrue() {
		thenResult, returned := s.Eval(stmt.Then, env, currTime)
		return thenResult, returned
	} else {
		elseResult, returned := s.Eval(stmt.Then, env, currTime)
		return elseResult, returned
	}
}

// Delay expressions
func (s *SimpleEval) evalDelayStmt(d *DelayStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	// Evaluate the condition expression to get its OpNode representation
	result, _ = s.Eval(d.Duration, env, currTime)
	if i, err := result.GetInt(); err == nil {
		*currTime += core.Duration(i)
		result.Time += core.Duration(i)
	} else if f, err := result.GetFloat(); err == nil {
		*currTime += f
		result.Time += f
	} else {
		panic("delay value should have been int or float.  type checking failed")
	}
	return
}

func (s *SimpleEval) evalTupleExpr(m *TupleExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	var vals []Value
	for _, argExpr := range m.Children {
		argres, _ := s.Eval(argExpr, env, currTime)
		vals = append(vals, argres)
	}
	result = TupleValue(vals...)
	return
}

func (s *SimpleEval) evalGoExpr(m *GoExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	loopValue, _ := s.Eval(m.LoopExpr, env, currTime)
	target := m.Stmt
	if m.Expr != nil {
		target = m.Expr
	}
	futureType := m.InferredType()
	result, err := NewValue(futureType, &FutureValue{
		LoopValue: loopValue,
		StartedAt: *currTime,
		Body: ThunkValue{
			Stmt:     target,
			SavedEnv: env.Push(),
		},
	})
	ensureNoErr(err)

	return
}

func (s *SimpleEval) evalWaitExpr(expr *WaitExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	var futureValues []Value
	for _, fname := range expr.FutureNames {
		futureVal, _ := s.Eval(fname, env, currTime)
		futureValues = append(futureValues, futureVal)
	}

	var aggParams []Value
	for _, aggParam := range expr.AggregatorParams {
		aggVal, _ := s.Eval(aggParam, env, currTime)
		aggParams = append(aggParams, aggVal)
	}

	aggregator := s.RootFile.Runtime.CreateAggregator(expr.AggregatorName.Value, aggParams)
	result, _ = aggregator(s, env, currTime, futureValues)
	return
}

///////////////////////////////////////////////////////

// MemberAccessExprs are used to access fields/params of a component instance
// In most cases they are straightforward - however when used as a set target
// we somehow need to capture a reference to it so the value can be set later.
// Right now it is being solved by special casing SetStmt and CallExpr
func (s *SimpleEval) evalMemberAccessExpr(m *MemberAccessExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	var err error
	// check if receive is an enum
	if idexpr, ok := m.Receiver.(*IdentifierExpr); ok && idexpr.InferredType() != nil {
		idexprType := idexpr.InferredType()
		if idexprType.Tag == decl.TypeTagEnum {
			// This is an enum - so we can return the value directly
			enumDecl := idexprType.Info.(*EnumDecl)
			ensureNoErr(err, "Enum value %s not found in enum %s", m.Member.Value, idexpr.Value)
			// log.Println("Enum Value: ", enumDecl)
			idx := enumDecl.IndexOfVariant(m.Member.Value)
			result, err = NewValue(idexprType, idx)
			ensureNoErr(err, "Error creating enum value: %v", err)
			return
		}
	}

	maeTarget, _ := s.Eval(m.Receiver, env, currTime)
	finalReceiver := maeTarget

	var compInst *ComponentInstance
	if maeTarget.Type.Tag == decl.TypeTagRef {
		refVal := maeTarget.Value.(*decl.RefValue)
		compInst = refVal.Receiver.Value.(*ComponentInstance)
		usedInst, _ := compInst.Get(refVal.Attrib)
		if usedInst.IsNil() {
			// TODO - This is a runtime error - but a user one so we should flag instead of panicking
			// This means a "set" needs to be called - for example in DB, the ByShortCode dependency is not
			// set - should we require that these are set manually each time or allow default values somehow for components too?
			err := fmt.Errorf("Depenendency %s not set.  Either override it or set it", refVal.Attrib)
			s.AddErrors(err)
			panic(err)
		}
		compInst = usedInst.Value.(*ComponentInstance)
	} else if maeTarget.Type.Tag != decl.TypeTagComponent {
		panic(fmt.Sprintf("Expected mae to be a component, found: %s -> %s", maeTarget.String(), maeTarget.Type))
	} else {
		compInst = maeTarget.Value.(*ComponentInstance)
	}

	if compInst == nil {
		panic(fmt.Sprintf("Expected mae to be a component, found: %s -> %s", maeTarget.String(), maeTarget.Type))
	}

	compDecl := compInst.ComponentDecl
	compType := decl.ComponentType(compDecl)
	finalReceiver, err = NewValue(compType, compInst)
	ensureNoErr(err)
	paramDecl, _ := compDecl.GetParam(m.Member.Value)

	// See if we are just dealing with a param - in which case we return a value
	if paramDecl != nil {
		paramType := paramDecl.Name.InferredType()
		refType := decl.RefType(compDecl, paramType)
		result, err = NewValue(refType, &decl.RefValue{Receiver: finalReceiver, Attrib: m.Member.Value})
		ensureNoErr(err)
		return
	}

	// "uses" declarations are same as params
	usesDecl, _ := compDecl.GetDependency(m.Member.Value)
	if usesDecl != nil {
		depType := decl.ComponentType(usesDecl.ResolvedComponent)
		refType := decl.RefType(compDecl, depType)
		result, err = NewValue(refType, &decl.RefValue{Receiver: finalReceiver, Attrib: m.Member.Value})
		ensureNoErr(err)
		return
	}

	// Methods are different - instead of returning a reference, return a bound method
	// so it can be evaluated correctly
	methodDecl, _ := compDecl.GetMethod(m.Member.Value)
	if methodDecl != nil {
		methodType := decl.MethodType(compDecl, methodDecl)
		methodVal := &decl.MethodValue{
			Method:   methodDecl,
			SavedEnv: compInst.InitialEnv.Push(),
			IsNative: compDecl.IsNative,
		}
		if compInst.IsNative {
			methodVal.BoundInstance = compInst.NativeInstance
		} else {
			// log.Println("what do we do here?")
		}
		result, err = NewValue(methodType, methodVal)
		ensureNoErr(err)
		return result, false
	}

	// Otherwise see if it is a uses field
	/*
		usesDecl, _ := compDecl.GetDependency(m.Member.Value)
		if usesDecl != nil {
			refType := decl.RefType(compDecl, usesDecl.Type.ResolvedType())
			result, err = NewValue(refType, &decl.RefValue{Receiver: maeTarget, Attrib: m.Member.Value})
			ensureNoErr(err) {
				panic(err)
			}
			return
		}
	*/

	// Return the reference value here
	panic("Invalid member type")
}

// Evaluate a Call and return its value
// Call expression are of the form a.b.c.d(params)
// The a.b.c.d must resolve to a callable (either a component method or a native function)
func (s *SimpleEval) evalCallExpr(expr *CallExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	// Simplified version - evaluate the Function -
	// it MUST evaluate to a MethodValue with a saved env
	receiver, _ := s.Eval(expr.Function, env, currTime)
	methodValue := receiver.Value.(*decl.MethodValue)
	methodDecl := methodValue.Method // compInstance.ComponentDecl.GetMethod(fexpr.Member.Value)

	// Now we have the target component instance, we can invoke the method
	// Evaluate the arguments
	argValues := make([]Value, len(expr.Args))
	for i, argExpr := range expr.Args {
		argValue, _ := s.Eval(argExpr, env, currTime)
		argValues[i] = argValue
	}
	// No currying for now
	newenv := methodValue.SavedEnv.Push()
	for idx, param := range methodDecl.Parameters {
		newenv.Set(param.Name.Value, argValues[idx])
	}

	// Now evaluate the body using the new env
	if methodValue.IsNative {
		// Native method invocation to be handled differently
		result, err := InvokeMethod(methodValue.BoundInstance, methodValue.Method.Name.Value,
			argValues, env, currTime, s.Rand)
		ensureNoErr(err, "Error calling method: ", err)
		return result, false
	} else {
		result, _ = s.Eval(methodDecl.Body, newenv, currTime)
	}
	return

	/*
		newEnv := compInstance.InitialEnv.Push()
	*/
}

/** Evaluate a Assignment as a statement and return its value */
func (s *SimpleEval) evalAssignmentStmt(stmt *AssignmentStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	panic("to be implemented")
}
