package runtime

import (
	"fmt"
	"math/rand"
	"slices"
	"strconv"
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
	idgen    SimpleIDGen
	RootFile *FileInstance
	Rand     *rand.Rand
	Tracer   *ExecutionTracer
	Errors   []error
}

func NewSimpleEval(fi *FileInstance, tracer *ExecutionTracer) *SimpleEval {
	out := &SimpleEval{
		RootFile: fi,
		Rand:     rand.New(rand.NewSource(time.Now().UnixMicro())),
		Tracer:   tracer,
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

	// Now check if all things are initialized
	uninited := sys.GetUninitializedComponents(env)
	for _, unin := range uninited {
		out := unin.Attrib
		errorFile := sys.System.ParentFileDecl.FullPath
		errorLoc := unin.Pos
		curr := unin.From
		for curr != nil {
			errorLoc = curr.Pos
			out = curr.CompInst.ComponentDecl.Name.Value + "." + out
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
	// ... (rest of the Eval method remains the same)
	switch n := node.(type) {
	// --- Statement Nodes ---
	case *BlockStmt:
		// With a block statement we usually push an extra context so it can be removed
		// at the end of the block
		return s.evalBlockStmt(n, env.Push(), currTime)
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
		panic(err)
	} else {
		result = value
	}
	return
}

func (s *SimpleEval) evalLiteralExpr(e *LiteralExpr, _ *Env[Value], _ *core.Duration) (result Value, returned bool) {
	result = e.Value
	return
}

func (s *SimpleEval) evalSetStmt(set *SetStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	// evaluate the Expression and unzip and assign to variables in the same environment
	result, _ = s.Eval(set.Value, env, currTime)
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
	if r.ReturnValue != nil {
		result, _ = s.Eval(r.ReturnValue, env, currTime)
	}
	return result, true
}

func (s *SimpleEval) evalForStmt(f *ForStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	var err error
	counter := int64(0)
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
		tupleValues, err := result.GetTuple()
		ensureNoErr(err)
		for i, val := range tupleValues {
			letvar := l.Variables[i].Value
			env.Set(letvar, val)
		}
	}
	return
}

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

func (s *SimpleEval) evalSampleExpr(samp *decl.SampleExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	res, _ := s.Eval(samp.FromExpr, env, currTime)
	outcomes := res.OutcomesVal()
	result, _ = outcomes.Sample(s.Rand)
	return
}

func (s *SimpleEval) evalNewExpr(n *decl.NewExpr, _ *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	compInst, result, err := NewComponentInstance(s.idgen.NextID("comp"), s.RootFile, n.ComponentDecl)
	ensureNoErr(err)
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
	if u.Operator == "not" {
		// Handle both direct bool and Outcomes[Bool] types
		if lr.Type.Equals(BoolType) {
			lr.Value = !lr.Value.(bool)
		} else if lr.Type.Tag == decl.TypeTagOutcomes {
			// Sample from outcomes and apply 'not' to the result
			outcomesVal := lr.OutcomesVal()
			sampledVal, ok := outcomesVal.Sample(s.Rand)
			if !ok {
				panic("failed to sample from outcomes in unary not expression")
			}
			*currTime += sampledVal.Time // Add the sampled time
			if sampledVal.Type.Equals(BoolType) {
				sampledVal.Value = !sampledVal.Value.(bool)
				lr = sampledVal
			} else {
				panic(fmt.Sprintf("Unary operator 'not' expected bool in outcomes, got: %s", sampledVal.Type))
			}
		} else {
			panic(fmt.Sprintf("Unary operator not supported for type: %s", lr.Type))
		}
	} else {
		panic(fmt.Sprintf("Unary operator not supported: %s", u.Operator))
	}
	result = lr
	return
}

func (s *SimpleEval) evalBinaryExpr(b *BinaryExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	_, _ = s.Eval(b.Left, env, currTime)
	_, _ = s.Eval(b.Right, env, currTime)

	// TODO - Evaluate based on the operator
	// log.Println("Left Result: ", lr)
	// log.Println("Right Result: ", rr)
	return
}

func (s *SimpleEval) evalExprStmt(stmt *ExprStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	return s.Eval(stmt.Expression, env, currTime)
}

func (s *SimpleEval) evalIfStmt(stmt *IfStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	condResult, _ := s.Eval(stmt.Condition, env, currTime)
	if condResult.IsTrue() {
		return s.Eval(stmt.Then, env, currTime)
	} else if stmt.Else != nil {
		return s.Eval(stmt.Else, env, currTime)
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
	var traceID int
	loopValue, _ := s.Eval(m.LoopExpr, env, currTime)
	if s.Tracer != nil {
		loopCount := "1"
		if !loopValue.IsNil() {
			if intVal, err := loopValue.GetInt(); err == nil {
				loopCount = strconv.FormatInt(intVal, 10)
			}
		}
		target := fmt.Sprintf("goroutine_for_%s", m.InferredType().String())
		traceID = s.Tracer.Enter(*currTime, EventGo, target, loopCount)
	}

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
		TraceID: traceID,
	})
	ensureNoErr(err)
	return
}

func (s *SimpleEval) evalWaitExpr(expr *WaitExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	var aggName string
	if expr.AggregatorName != nil {
		aggName = expr.AggregatorName.Value
	}
	if s.Tracer != nil {
		s.Tracer.Enter(*currTime, EventWait, "wait", aggName)
		defer s.Tracer.Exit(*currTime, 0, Nil, nil)
	}

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

	aggregator := s.RootFile.Runtime.CreateAggregator(aggName, aggParams)
	result, _ = aggregator.Eval(s, env, currTime, futureValues)
	return
}

func (s *SimpleEval) evalMemberAccessExpr(m *MemberAccessExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	var err error
	if idexpr, ok := m.Receiver.(*IdentifierExpr); ok && idexpr.InferredType() != nil {
		idexprType := idexpr.InferredType()
		if idexprType.Tag == decl.TypeTagEnum {
			enumDecl := idexprType.Info.(*EnumDecl)
			ensureNoErr(err, "Enum value %s not found in enum %s", m.Member.Value, idexpr.Value)
			idx := enumDecl.IndexOfVariant(m.Member.Value)
			result, err = NewValue(idexprType, idx)
			ensureNoErr(err, "Error creating enum value: %v", err)
			return
		}
	}
	maeTarget, _ := s.Eval(m.Receiver, env, currTime)
	var compInst *ComponentInstance
	if maeTarget.Type.Tag == decl.TypeTagRef {
		refVal := maeTarget.Value.(*decl.RefValue)
		compInst = refVal.Receiver.Value.(*ComponentInstance)
		usedInst, _ := compInst.Get(refVal.Attrib)
		if usedInst.IsNil() {
			// TODO - This is a runtime error - but a user one so we should flag instead of panicking
			// This means a "set" needs to be called - for example in DB, the ByShortCode dependency is not
			// set - should we require that these are set manually each time or allow default values somehow for components too?
			err := fmt.Errorf("Dependency %s not set. Either override it or set it", refVal.Attrib)
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
	finalReceiver, err := NewValue(compType, compInst)
	ensureNoErr(err)
	paramDecl, _ := compDecl.GetParam(m.Member.Value)
	if paramDecl != nil {
		// paramType := paramDecl.Name.InferredType()
		// refType := decl.RefType(compDecl, paramType)
		// result, err = NewValue(refType, &decl.RefValue{Receiver: finalReceiver, Attrib: m.Member.Value})
		// ensureNoErr(err)
		result, ok := compInst.Get(m.Member.Value) // NewValue(refType, &decl.RefValue{Receiver: finalReceiver, Attrib: m.Member.Value})
		if !ok {
			panic("Invalid...")
		}
		return result, false
	}
	usesDecl, _ := compDecl.GetDependency(m.Member.Value)
	if usesDecl != nil {
		depType := decl.ComponentType(usesDecl.ResolvedComponent)
		refType := decl.RefType(compDecl, depType)
		result, err = NewValue(refType, &decl.RefValue{Receiver: finalReceiver, Attrib: m.Member.Value})
		ensureNoErr(err)
		return
	}
	methodDecl, _ := compDecl.GetMethod(m.Member.Value)
	if methodDecl != nil {
		methodType := decl.MethodType(methodDecl)
		methodVal := &decl.MethodValue{
			Method: methodDecl, SavedEnv: compInst.InitialEnv.Push(), IsNative: compDecl.IsNative,
		}
		if compInst.IsNative {
			methodVal.BoundInstance = compInst.NativeInstance
		}
		result, err = NewValue(methodType, methodVal)
		ensureNoErr(err)
		return result, false
	}
	err = fmt.Errorf("in file %s at line %d, col %d: member '%s' not found on component of type '%s'",
		s.RootFile.Decl.FullPath, m.Member.Pos().Line, m.Member.Pos().Col, m.Member.Value, compDecl.Name.Value)
	s.AddErrors(err)
	result = decl.Nil
	return result, false
}

func (s *SimpleEval) evalCallExpr(expr *CallExpr, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	receiver, _ := s.Eval(expr.Function, env, currTime)
	methodValue := receiver.Value.(*decl.MethodValue)
	methodDecl := methodValue.Method

	argValues := make([]Value, expr.NumArgs())
	for i, argExpr := range expr.ArgList {
		argValue, _ := s.Eval(argExpr, env, currTime)
		argValues[i] = argValue
	}

	if s.Tracer != nil {
		argStrings := make([]string, len(argValues))
		for i, v := range argValues {
			argStrings[i] = v.String()
		}
		target := expr.Function.String()
		eventID := s.Tracer.Enter(*currTime, EventEnter, target, argStrings...)
		startTime := *currTime
		defer func() {
			duration := *currTime - startTime
			s.Tracer.Exit(*currTime, duration, result, nil)
			_ = eventID // To silence unused error, though it's conceptually used by Exit
		}()
	}

	newenv := methodValue.SavedEnv.Push()
	for idx, param := range methodDecl.Parameters {
		newenv.Set(param.Name.Value, argValues[idx])
	}

	if methodValue.IsNative {
		if methodValue.BoundInstance != nil {
			result, err := InvokeMethod(methodValue.BoundInstance, methodValue.Method.Name.Value, argValues, env, currTime, s.Rand, true)
			ensureNoErr(err, "Error calling method: ", err)
			return result, false
		} else {
			// It's a global native method
			nativeFunc, found := s.RootFile.Runtime.nativeMethods[methodDecl.Name.Value]
			if !found {
				panic(fmt.Sprintf("Global native method '%s' not found in runtime registry", methodDecl.Name.Value))
			}
			return nativeFunc(s, newenv, currTime, argValues...)
		}
	} else {
		result, _ = s.Eval(methodDecl.Body, newenv, currTime)
	}
	return
}

func (s *SimpleEval) evalAssignmentStmt(stmt *AssignmentStmt, env *Env[Value], currTime *core.Duration) (result Value, returned bool) {
	panic("to be implemented")
}
