package loader

import (
	"fmt"
	"log"

	"github.com/panyam/sdl/decl"
	// "log" // Uncomment for debugging
)

type InferenceError struct {
	Pos Location
	Msg string
}

func InfErrorf(pos Location, format string, args ...any) *InferenceError {
	return &InferenceError{
		Pos: pos,
		Msg: fmt.Sprintf(format, args...),
	}
}

func (i *InferenceError) Error() string {
	return fmt.Sprintf("%s: %s", i.Pos.LineColStr(), i.Msg)
}

type Inference struct {
	ErrorCollector

	// Path of the file being inferred (if provided)
	filePath string

	// Inference starts the root file
	rootFile *FileDecl
}

func NewInference(fp string, fd *FileDecl) *Inference {
	return &Inference{
		filePath: fp,
		rootFile: fd,
	}
}

// Begins type inference starting at the root file
func (i *Inference) Eval(rootEnv *Env[Node]) bool {
	file := i.rootFile
	if file == nil {
		i.AddErrors(fmt.Errorf("cannot infer types for nil FileDecl"))
		return false
	}

	// Ensure file is resolved
	if err := file.Resolve(); err != nil {
		i.AddErrors(err)
		return false
	}

	rootScope := NewRootTypeScope(rootEnv)

	components, _ := file.GetComponents()
	systems, _ := file.GetSystems()

	// First pass: Resolve TypeDecls in component parameter defaults, method parameters, and method return types.
	for _, compDecl := range components {
		// Parameter defaults
		i.EvalForComponent(compDecl, rootScope)
	}

	// Second pass: Infer types for component method bodies (now that signatures are resolved)
	for _, compDecl := range components {
		methods, _ := compDecl.Methods()
		for _, method := range methods {
			// methodScope needs to see 'self', method params, component params/uses, and globals/imports via rootEnv
			// Parameters are added to scope implicitly by TypeScope.Get looking at currentMethod.
			// 'uses' are also resolved via 'self' access within TypeScope.Get if needed.
			if method.Body != nil {
				methodScope := rootScope.Push(compDecl, method)
				i.EvalForBlockStmt(method.Body, methodScope)
			}
		}
	}

	// Third pass: Infer types for system declarations
	for _, sysDecl := range systems {
		i.EvalForSystemDecl(sysDecl, rootScope.Push(nil, nil)) // System scope can see globals/imports from rootEnv
	}
	return false
}

func (i *Inference) EvalForComponent(compDecl *ComponentDecl, rootScope *TypeScope) (success bool) {
	params, _ := compDecl.Params()

	for _, paramDecl := range params { // Assuming direct field access or appropriate getter
		i.EvalForParamDecl(paramDecl, compDecl, rootScope)
	}

	// Now look at "uses"
	usesDecls, _ := compDecl.Dependencies()
	for _, usesDecl := range usesDecls { // Assuming direct field access or appropriate getter
		compTypeNode, foundCompType := rootScope.env.Get(usesDecl.ComponentName.Value)
		if !foundCompType {
			i.Errorf(usesDecl.ComponentName.Pos(),
				"component type '%s' not found for dependency '%s'",
				usesDecl.ComponentName.Value, usesDecl.Name.Value)
			return false
		}
		compDefinition, ok := compTypeNode.(*ComponentDecl)
		if !ok {
			i.Errorf(usesDecl.ComponentName.Pos(), "identifier '%s' used as component type for instance '%s' is not a component declaration (got %T)", usesDecl.ComponentName.Value, usesDecl.Name.Value, compTypeNode)
			return false
		}
		instanceType := ComponentType(compDefinition)
		rootScope.env.Set(usesDecl.Name.Value, compDefinition) // Store InstanceDecl node in env by its name
		usesDecl.Name.SetInferredType(instanceType)
	}

	// Method signatures
	methods, _ := compDecl.Methods()
	for _, method := range methods {
		i.EvalForMethodSignature(method, compDecl, rootScope)
	}
	return
}

func (i *Inference) EvalForParamDecl(paramDecl *ParamDecl, compDecl *ComponentDecl, rootScope *TypeScope) (success bool) {
	var resolvedParamType *Type

	// Ensure that if all succeeds the type for the param is set in the root scope
	defer func() {
		if success && resolvedParamType != nil {
			rootScope.Set(paramDecl.Name.Value, paramDecl.Name, resolvedParamType) // Register the parameter type in the scope
		}
	}()

	if paramDecl.TypeDecl != nil { // Type is explicitly declared
		resolvedParamType = rootScope.ResolveType(paramDecl.TypeDecl)
		if resolvedParamType == nil {
			i.Errorf(paramDecl.TypeDecl.Pos(), "unresolved type '%s' for parameter '%s' in component '%s'", paramDecl.TypeDecl.Name, paramDecl.Name.Value, compDecl.Name.Value)
			// Even if unresolved, continue to check default value if present, but this param is problematic.
		} else {
			paramDecl.TypeDecl.SetResolvedType(resolvedParamType)
			paramDecl.Name.SetInferredType(resolvedParamType)
		}
	} else if paramDecl.DefaultValue != nil { // No explicit type, but has default value
		// Infer type from default value
		resolvedParamType, success = i.EvalForExprType(paramDecl.DefaultValue, rootScope)
		if success {
			if resolvedParamType != nil {
				paramDecl.Name.SetInferredType(resolvedParamType)
			} else {
				// Default value's type could not be inferred.
				i.Errorf(paramDecl.DefaultValue.Pos(), "could not infer type from default value for parameter '%s' in component '%s'", paramDecl.Name.Value, compDecl.Name.Value)
			}
		}
	} else { // No explicit type AND no default value
		i.Errorf(paramDecl.Pos(), "parameter '%s' in component '%s' has no type declaration and no default value", paramDecl.Name.Value, compDecl.Name.Value)
		return // Cannot proceed with this parameter
	}

	// If a default value exists, check its type against the (now hopefully resolved) parameter type.
	if paramDecl.DefaultValue != nil {
		defaultValueActualType, success := i.EvalForExprType(paramDecl.DefaultValue, rootScope)
		if success {
			if defaultValueActualType != nil {
				// Now, `resolvedParamType` should hold the type of the parameter,
				// either from its TypeDecl or inferred from the default value itself.
				if resolvedParamType != nil { // If we have an expected type for the param
					if !defaultValueActualType.Equals(resolvedParamType) {
						// Allow int to float promotion for default value
						isPromotion := defaultValueActualType.Equals(IntType) && resolvedParamType.Equals(FloatType)
						if !isPromotion {
							i.Errorf(paramDecl.DefaultValue.Pos(), "type mismatch for default value of parameter '%s' in component '%s': parameter type is %s, default value type is %s", paramDecl.Name.Value, compDecl.Name.Value, resolvedParamType.String(), defaultValueActualType.String())
						}
					}
				} else if paramDecl.TypeDecl != nil {
					// This means paramDecl.TypeDecl was specified but couldn't be resolved earlier,
					// yet we have a default value. This is an inconsistent state.
					// The earlier error about unresolved type for paramDecl.TypeDecl should cover this.
				}
			}
		}
	}
	return
}

// Infer/Check types for a method signature.  The body is not evaluated here
func (i *Inference) EvalForMethodSignature(method *MethodDecl, compDecl *ComponentDecl, rootScope *TypeScope) (errors []error) {
	for _, param := range method.Parameters {
		if param.TypeDecl != nil {
			resolvedParamType := rootScope.ResolveType(param.TypeDecl)
			if resolvedParamType == nil {
				i.Errorf(param.TypeDecl.Pos(), "unresolved type '%s' for parameter '%s' in method '%s.%s'", param.TypeDecl.Name, param.Name.Value, compDecl.Name.Value, method.Name.Value)
			} else {
				param.TypeDecl.SetResolvedType(resolvedParamType)
			}
		} else {
			i.Errorf(param.Pos(), "parameter '%s' of method '%s.%s' has no type declaration", param.Name.Value, compDecl.Name.Value, method.Name.Value)
		}
	}
	if method.ReturnType != nil {
		resolvedReturnType := rootScope.ResolveType(method.ReturnType)
		if resolvedReturnType == nil {
			i.Errorf(method.ReturnType.Pos(), "unresolved return type '%s' for method '%s.%s'", method.ReturnType.Name, compDecl.Name.Value, method.Name.Value)
		} else {
			method.ReturnType.SetResolvedType(resolvedReturnType)
		}
	}
	return
}

// i.EvalForExprType recursively infers the type of an expression and sets its InferredType field.
func (i *Inference) EvalForExprType(expr Expr, scope *TypeScope) (inferred *Type, success bool) {
	if expr == nil {
		i.AddErrors(fmt.Errorf("cannot infer type for nil expression"))
		return nil, false
	}

	if expr.InferredType() != nil {
		return expr.InferredType(), true
	}

	switch e := expr.(type) {
	case *LiteralExpr:
		inferred, success = i.EvalForLiteralExpr(e, scope)
	case *IdentifierExpr:
		inferred, success = i.EvalForIdentifierExpr(e, scope)
	case *BinaryExpr:
		inferred, success = i.EvalForBinaryExpr(e, scope)
	case *UnaryExpr:
		inferred, success = i.EvalForUnaryExpr(e, scope)
	case *MemberAccessExpr:
		inferred, success = i.EvalForMemberAccessExpr(e, scope)
	case *CallExpr:
		inferred, success = i.EvalForCallExpr(e, scope)
	case *TupleExpr:
		inferred, success = i.EvalForTupleExpr(e, scope)
	case *DistributeExpr:
		inferred, success = i.EvalForDistributeExpr(e, scope)
	case *SampleExpr:
		inferred, success = i.EvalForSampleExpr(e, scope)
	case *IndexExpr:
		inferred, success = i.EvalForIndexExpr(e, scope)
	case *CaseExpr:
		if e.Body == nil {
			i.Errorf(e.Pos(), "CaseExpr at has no body")
			return nil, false
		}
		inferred, success = i.EvalForExprType(e.Body, scope)
	default:
		panic(fmt.Errorf("type inference not implemented for expression type %T at pos %s", expr, expr.Pos().LineColStr()))
	}

	if !success {
		return
	}

	expr.SetInferredType(inferred)

	if expr.DeclaredType() != nil && !expr.DeclaredType().Equals(inferred) {
		isIntToFloatPromotion := expr.DeclaredType().Equals(FloatType) && inferred.Equals(IntType)
		if !isIntToFloatPromotion {
			i.Errorf(expr.Pos(), "type mismatch for '%s': inferred type %s, but declared type is %s", expr.String(), inferred.String(), expr.DeclaredType().String())
			return nil, false
		}
	}
	return
}

func (i *Inference) EvalForLiteralExpr(expr *LiteralExpr, scope *TypeScope) (*Type, bool) {
	if expr.Value.IsNil() || expr.Value.Type == nil {
		log.Println("Value: ", expr.Value, expr.Value.Value)
		log.Println("Type: ", expr.Value.Type)
		log.Println("Expr: ", expr.Pos(), expr.String())
		i.Errorf(expr.Pos(), "literal expression has invalid internal Value or Type")
		return nil, false
	}
	return expr.Value.Type, true
}

func (i *Inference) EvalForIdentifierExpr(expr *IdentifierExpr, scope *TypeScope) (t *Type, ok bool) {
	t, ok = scope.Get(expr.Value)
	if !ok {
		return nil, i.Errorf(expr.Pos(), "identifier '%s' not found", expr.Value)
	}
	if t == nil {
		return nil, i.Errorf(expr.Pos(), "identifier '%s' resolved but its type is nil (internal error)", expr.Value)
	}
	return
}

func (i *Inference) EvalForMemberAccessExpr(expr *MemberAccessExpr, scope *TypeScope) (t *Type, ok bool) {
	receiverType, ok := i.EvalForExprType(expr.Receiver, scope)
	if !ok {
		return
	}
	if receiverType == nil {
		return nil, i.Errorf(expr.Pos(), "could not determine receiver type for member access '.%s'", expr.Member.Value)
	}
	memberName := expr.Member.Value

	// Enums are easy - they have a fixed set of values.
	if receiverType.Tag == decl.TypeTagEnum {
		decl := receiverType.Info.(*EnumDecl)
		for _, valNode := range decl.Values {
			if valNode.Value == memberName {
				return receiverType, true
			}
		}
		return nil, i.Errorf(expr.Pos(), "value '%s' not found in enum '%s'", memberName, decl.Name.Value)
	}

	// before we drill into specific types, check if the receiver is a RefType
	// In this case resolve the specific component or value and then create a Ref value from this.
	// With memberAccessExpr, we always want to return References so they are resolved by the caller.
	// This will make Set statements work correctly.
	if receiverType.Tag == decl.TypeTagRef {
		refTypeInfo := receiverType.Info.(*decl.RefTypeInfo)
		if refTypeInfo.ParamType == nil {
			return nil, i.Errorf(expr.Pos(), "ref type '%s' has no parameter type declared", receiverType.String())
		}
		receiverType = refTypeInfo.ParamType // Use the parameter type as the receiver type
	}

	// Receiver MUST be a Component now
	if receiverType.Tag != decl.TypeTagComponent {
		return nil, i.Errorf(expr.Pos(), "cannot access member '%s' on type %s; receiver is not an enum, component, or known type with this member", memberName, receiverType.String())
	}

	decl := receiverType.Info.(*ComponentDecl)
	if paramDecl, _ := decl.GetParam(memberName); paramDecl != nil {
		paramType := paramDecl.Name.InferredType()
		if paramType == nil {
			panic("internal error: parameter type for '" + memberName + "' in component '" + decl.Name.Value + "' is nil")
		}
		return RefType(decl, paramType), true
	} else if usesDecl, _ := decl.GetDependency(memberName); usesDecl != nil {
		if scope.env == nil {
			return nil, i.Errorf(usesDecl.Pos(), "internal error: TypeScope.env is nil when resolving 'uses' dependency '%s' in component '%s'", memberName, decl.Name.Value)
		}
		depCompName := usesDecl.ComponentName.Value
		depCompDeclNode, found := scope.env.Get(depCompName)
		if !found {
			return nil, i.Errorf(usesDecl.Pos(), "'uses' dependency '%s' in component '%s' refers to unknown component type '%s'", memberName, decl.Name.Value, depCompName)
		}
		if depCompDecl, ok := depCompDeclNode.(*ComponentDecl); ok {
			usesDecl.ResolvedComponent = depCompDecl
			return RefType(decl, ComponentType(depCompDecl)), true
		}
		return nil, i.Errorf(usesDecl.Pos(), "'uses' dependency '%s' in component '%s' resolved to a non-component type %T for '%s'", memberName, decl.Name.Value, depCompDeclNode, depCompName)
	} else if methodDecl, _ := decl.GetMethod(memberName); methodDecl != nil {
		return MethodType(decl, methodDecl), ok
	}
	return nil, i.Errorf(expr.Pos(), "member '%s' not found in component '%s' (type %s)", memberName, decl.Name.Value, receiverType)
}

func (i *Inference) EvalForBinaryExpr(expr *BinaryExpr, scope *TypeScope) (*Type, bool) {
	leftType, lok := i.EvalForExprType(expr.Left, scope)
	rightType, rok := i.EvalForExprType(expr.Right, scope)
	if !lok || !rok || leftType == nil || rightType == nil {
		return nil, i.Errorf(expr.Pos(), "could not determine type for one or both operands for binary expr ('%s')", expr.Operator)
	}

	switch expr.Operator {
	case "+", "-", "*", "/":
		if leftType.Equals(IntType) && rightType.Equals(IntType) {
			return IntType, true
		}
		if (leftType.Equals(IntType) || leftType.Equals(FloatType)) &&
			(rightType.Equals(IntType) || rightType.Equals(FloatType)) {
			return FloatType, true
		}
		if expr.Operator == "+" && leftType.Equals(StrType) && rightType.Equals(StrType) {
			return StrType, true
		}
		return nil, i.Errorf(expr.Pos(), "type mismatch for operator '%s': cannot apply to %s and %s", expr.Operator, leftType.String(), rightType.String())

	case "%":
		if leftType.Equals(IntType) && rightType.Equals(IntType) {
			return IntType, true
		}
		return nil, i.Errorf(expr.Pos(), "type mismatch for operator '%%': requires two integers, got %s and %s", leftType.String(), rightType.String())

	case "==", "!=", "<", "<=", ">", ">=":
		isLeftNumeric := leftType.Equals(IntType) || leftType.Equals(FloatType)
		isRightNumeric := rightType.Equals(IntType) || rightType.Equals(FloatType)

		if isLeftNumeric && isRightNumeric {
			return BoolType, true
		}
		if leftType.Equals(rightType) {
			/*
				if leftType.Tag == decl. TypeTagEnum {
					return BoolType, true
				}
				if leftType.Name == "List" || leftType.Name == "Tuple" || leftType.Name == "Outcomes" ||
					(leftType.OriginalDecl != nil && leftType.IsComponentType()) ||
					leftType.Name == "OpNode" || leftType.Name == "MethodReference" {
					return nil, i.Errorf(expr.Pos(), "type mismatch for comparison operator '%s': cannot compare complex type %s", expr.Operator, leftType.String())
				}
			*/
			log.Println("TODO - Unresolved case")
			return BoolType, true
		}
		return nil, i.Errorf(expr.Pos(), "type mismatch for comparison operator '%s': cannot compare %s and %s", expr.Operator, leftType.String(), rightType.String())

	case "&&", "||":
		if leftType.Equals(BoolType) && rightType.Equals(BoolType) {
			return BoolType, true
		}
		return nil, i.Errorf(expr.Pos(), "type mismatch for logical operator '%s': requires two booleans, got %s and %s", expr.Operator, leftType.String(), rightType.String())

	default:
		return nil, i.Errorf(expr.Pos(), "unsupported binary operator '%s'", expr.Operator)
	}
}

func (i *Inference) EvalForUnaryExpr(expr *UnaryExpr, scope *TypeScope) (*Type, bool) {
	rightType, ok := i.EvalForExprType(expr.Right, scope)
	if !ok || rightType == nil {
		return nil, i.Errorf(expr.Pos(), "could not determine type for operand of unary expr ('%s')", expr.Operator)
	}

	switch expr.Operator {
	case "!", "not":
		if !rightType.Equals(BoolType) {
			return nil, i.Errorf(expr.Pos(), "type mismatch for operator '!': requires boolean, got %s", rightType.String())
		}
		return BoolType, true
	case "-":
		if rightType.Equals(IntType) || rightType.Equals(FloatType) {
			return rightType, true
		}
		return nil, i.Errorf(expr.Pos(), "type mismatch for operator '-': requires integer or float, got %s", rightType.String())
	default:
		return nil, i.Errorf(expr.Pos(), "unsupported unary operator '%s'", expr.Operator)
	}
}

func (i *Inference) EvalForCallExpr(expr *CallExpr, scope *TypeScope) (*Type, bool) {
	funcType, ok := i.EvalForExprType(expr.Function, scope)
	if !ok || funcType == nil {
		return nil, i.Errorf(expr.Function.Pos(), "could not determine type of function/method being called ('%s')", expr.Function.String())
	}

	var returnType *Type = NilType
	var expectedParamTypes []*Type
	var funcNameForError string = expr.Function.String()

	if funcType.Tag != decl.TypeTagMethod {
		return nil, i.Errorf(expr.Function.Pos(), "calling non-method type '%s' as function is not supported or function not found", funcType.String())
	}

	methodTypeInfo := funcType.Info.(*decl.MethodTypeInfo)
	methodDecl := methodTypeInfo.Method

	// Attempt to construct a more descriptive name for error messages
	if mae, isMae := expr.Function.(*MemberAccessExpr); isMae {
		receiverStr := mae.Receiver.String() // Assuming String() is safe for resolved expressions
		if receiverStr != "" {
			funcNameForError = fmt.Sprintf("%s.%s", receiverStr, methodDecl.Name.Value)
		} else { // Fallback if receiver string is empty (e.g. if it was complex and String() was minimal)
			funcNameForError = methodDecl.Name.Value
		}
	} else { // Not a MemberAccessExpr, use method name directly
		funcNameForError = methodDecl.Name.Value
	}

	if methodDecl.ReturnType != nil {
		// Use ResolveType for resolving TypeDecl within the method's component context if needed
		// However, return/param types are usually resolved using the global/import scope during the first pass.
		// For now, assume methodDecl.ReturnType.Type() is sufficient or TypeDecl.Type() handles resolution.
		resolvedReturnType := scope.ResolveType(methodDecl.ReturnType) // Pass current scope
		if resolvedReturnType == nil {
			return nil, i.Errorf(methodDecl.ReturnType.Pos(), "method '%s' return TypeDecl ('%s') did not resolve to a valid Type", funcNameForError, methodDecl.ReturnType.Name)
		}
		returnType = resolvedReturnType
	}

	for _, paramDecl := range methodDecl.Parameters {
		if paramDecl.TypeDecl == nil {
			return nil, i.Errorf(paramDecl.Pos(), "parameter '%s' of method '%s' has no type declaration", paramDecl.Name.Value, funcNameForError)
		}
		paramSDLType := scope.ResolveType(paramDecl.TypeDecl) // Pass current scope
		if paramSDLType == nil {
			return nil, i.Errorf(paramDecl.TypeDecl.Pos(), "parameter '%s' of method '%s' has invalid TypeDecl ('%s')", paramDecl.Name.Value, funcNameForError, paramDecl.TypeDecl.Name)
		}
		expectedParamTypes = append(expectedParamTypes, paramSDLType)
	}

	if len(expr.Args) != len(expectedParamTypes) {
		return nil, i.Errorf(expr.Pos(), "argument count mismatch for call to '%s': expected %d, got %d", funcNameForError, len(expectedParamTypes), len(expr.Args))
	}

	for idx, argExpr := range expr.Args {
		argType, ok := i.EvalForExprType(argExpr, scope)
		if !ok || argType == nil {
			return nil, i.Errorf(argExpr.Pos(), "could not determine type for argument %d of call to '%s'", idx+1, funcNameForError)
		}
		if !argType.Equals(expectedParamTypes[idx]) {
			isIntToFloat := argType.Equals(IntType) && expectedParamTypes[idx].Equals(FloatType)
			if !isIntToFloat {
				return nil, i.Errorf(argExpr.Pos(), "type mismatch for argument %d of call to '%s': expected %s, got %s", idx, funcNameForError, expectedParamTypes[idx].String(), argType.String())
			}
		}
	}
	return returnType, true
}

func (i *Inference) EvalForTupleExpr(expr *TupleExpr, scope *TypeScope) (*Type, bool) {
	if len(expr.Children) == 0 {
		return nil, i.Errorf(expr.Pos(), "tuple expression must have at least one child (empty tuples not supported)")
	}
	childTypes, ok := i.EvalForExprList(expr.Children, scope)
	if !ok {
		return nil, ok
	}
	return TupleType(childTypes...), ok
}

func (inf *Inference) EvalForExprList(exprlist []Expr, scope *TypeScope) ([]*Type, bool) {
	childTypes := make([]*Type, len(exprlist))
	for i, childExpr := range exprlist {
		childType, ok := inf.EvalForExprType(childExpr, scope)
		if !ok || childType == nil {
			return nil, inf.Errorf(childExpr.Pos(), "could not determine type for tuple element %d", i+1)
		}
		childTypes[i] = childType
	}
	return childTypes, true
}

func (inf *Inference) EvalForDistributeExpr(expr *DistributeExpr, scope *TypeScope) (*Type, bool) {
	var commonBodyType *Type
	if len(expr.Cases) == 0 && expr.Default == nil {
		return nil, inf.Errorf(expr.Pos(), "distribute expression must have at least one case or a default")
	}
	if expr.TotalProb != nil {
		totalProbType, ok := inf.EvalForExprType(expr.TotalProb, scope)
		if !ok || (totalProbType != nil && !(totalProbType.Equals(IntType) || totalProbType.Equals(FloatType))) {
			return nil, inf.Errorf(expr.TotalProb.Pos(), "total probability of distribute expr must be numeric, got %s", totalProbType.String())
		}
	}

	for i, caseExpr := range expr.Cases {
		if caseExpr.Condition == nil { // Should be caught by parser
			return nil, inf.Errorf(caseExpr.Pos(), "DistributeExpr case %d has no condition", i)
		}
		condType, ok := inf.EvalForExprType(caseExpr.Condition, scope)
		if !ok {
			return nil, false
		}
		if !(condType.Equals(FloatType) || condType.Equals(IntType)) {
			return nil, inf.Errorf(caseExpr.Condition.Pos(), "condition of distribute case %d must be numeric (for weight), got %s", i, condType.String())
		}
		if caseExpr.Body == nil { // Should be caught by parser
			return nil, inf.Errorf(caseExpr.Pos(), "DistributeExpr case %d has no body", i)
		}
		bodyType, ok := inf.EvalForExprType(caseExpr.Body, scope)
		if !ok {
			// return nil, inf.Errorf(caseExpr.Body.Pos(), "error inferring type for body of case %d in distribute expr: %w", i, err)
			return nil, ok
		}
		if bodyType == nil {
			return nil, inf.Errorf(caseExpr.Pos(), "could not determine type for body of case %d in distribute expr", i)
		}
		if commonBodyType == nil {
			commonBodyType = bodyType
		} else if !commonBodyType.Equals(bodyType) {
			return nil, inf.Errorf(expr.Pos(), "type mismatch in distribute expr cases: expected %s (from case 0), got %s for case %d", commonBodyType.String(), bodyType.String(), i)
		}
	}

	if expr.Default != nil {
		defaultType, ok := inf.EvalForExprType(expr.Default, scope)
		if !ok || defaultType == nil {
			return nil, inf.Errorf(expr.Default.Pos(), "could not determine type for default case of distribute expr")
		}
		if commonBodyType == nil {
			commonBodyType = defaultType
		} else if !commonBodyType.Equals(defaultType) {
			return nil, inf.Errorf(expr.Pos(), "type mismatch between distribute expr cases and default: expected %s, got %s for default", commonBodyType.String(), defaultType.String())
		}
	}
	if commonBodyType == nil {
		return nil, inf.Errorf(expr.Pos(), "distribute expr has no effective common type")
	}
	return OutcomesType(commonBodyType), true
}

func (i *Inference) EvalForSampleExpr(expr *SampleExpr, scope *TypeScope) (*Type, bool) {
	fromType, ok := i.EvalForExprType(expr.FromExpr, scope)
	if !ok || fromType == nil {
		return nil, i.Errorf(expr.Pos(), "could not determine type for 'from' expression of sample expr")
	}
	if fromType.Tag != decl.TypeTagOutcomes || fromType.Info == nil {
		return nil, i.Errorf(expr.Pos(), "type mismatch for sample expression: 'from' expression must be Outcomes[T], got %s", fromType.String())
	}
	return fromType.Info.(*Type), true
}

// EvalForIndexExpr infers the type of an IndexExpr (e.g., list[0], string[1]).
func (i *Inference) EvalForIndexExpr(expr *IndexExpr, scope *TypeScope) (*Type, bool) {
	receiverType, ok := i.EvalForExprType(expr.Receiver, scope)
	if !ok || receiverType == nil {
		return nil, i.Errorf(expr.Receiver.Pos(), "could not determine type of receiver for index expression")
	}

	keyType, ok := i.EvalForExprType(expr.Key, scope)
	if !ok || keyType == nil {
		return nil, i.Errorf(expr.Key.Pos(), "could not determine type of key for index expression")
	}

	switch receiverType.Tag {
	case decl.TypeTagList:
		if !keyType.Equals(IntType) {
			return nil, i.Errorf(expr.Key.Pos(), "list index must be an integer, got %s", keyType.String())
		}
		// The element type is stored in receiverType.Info for ListType
		elementType, isType := receiverType.Info.(*Type)
		if !isType || elementType == nil {
			return nil, i.Errorf(expr.Receiver.Pos(), "internal error: ListType has invalid element type information")
		}
		return elementType, true

	case decl.TypeTagSimple:
		if !receiverType.Equals(decl.StrType) {
			return nil, i.Errorf(expr.Key.Pos(), "receiver for Simple types must be a string, got %s", receiverType.String())
		}

		if !keyType.Equals(IntType) {
			return nil, i.Errorf(expr.Key.Pos(), "string index must be an integer, got %s", keyType.String())
		}
		// Indexing a string results in a string (character)
		return StrType, true

	case decl.TypeTagTuple:
		if !keyType.Equals(IntType) {
			return nil, i.Errorf(expr.Key.Pos(), "tuple index must be an integer, got %s", keyType.String())
		}
		// For tuples, we can only determine the specific element type if the key is a compile-time integer literal.
		if keyLiteral, isLiteral := expr.Key.(*LiteralExpr); isLiteral && keyLiteral.Value.Type.Equals(IntType) {
			indexVal, err := keyLiteral.Value.GetInt()
			if err != nil {
				// Should not happen if type is IntType, but defensive
				return nil, i.Errorf(expr.Key.Pos(), "internal error: could not get int value from int literal for tuple index")
			}

			tupleElementTypes, isTypeList := receiverType.Info.([]*Type)
			if !isTypeList {
				return nil, i.Errorf(expr.Receiver.Pos(), "internal error: TupleType has invalid element type information")
			}

			if indexVal < 0 || int(indexVal) >= len(tupleElementTypes) {
				return nil, i.Errorf(expr.Key.Pos(), "tuple index %d out of bounds (len %d)", indexVal, len(tupleElementTypes))
			}
			return tupleElementTypes[indexVal], true
		}
		// If the key is not an integer literal, we cannot statically determine which tuple element is accessed.
		// For now, we'll disallow non-literal integer indexing for tuples in type inference.
		// A more advanced system might return a union type or a generic "any_tuple_element" type.
		return nil, i.Errorf(expr.Key.Pos(), "tuple index must be an integer literal for precise type inference")

	case decl.TypeTagOutcomes:
		// It's generally an error to directly index an Outcomes[T] type.
		// The user should use 'sample' first to get a concrete value.
		return nil, i.Errorf(expr.Receiver.Pos(), "cannot directly index an Outcomes type; use 'sample' first to get a concrete value (e.g., 'let concrete_list = sample my_outcomes_list; concrete_list[0]')")

	default:
		return nil, i.Errorf(expr.Receiver.Pos(), "type %s is not indexable", receiverType.String())
	}
}

// --- Statement Type Inference ---

func (i *Inference) EvalForDelayStmt(d *DelayStmt, scope *TypeScope) (ok bool) {
	// Evaluate the type for the duration expr - it should resolve to int or float
	childType, ok := i.EvalForExprType(d.Duration, scope)
	if ok {
		if !childType.Equals(IntType) && !childType.Equals(FloatType) {
			return i.Errorf(d.Pos(), "Delay expression should be int or float, found: %s", childType.String())
		}
	}
	return
}

func (i *Inference) EvalForSetStmt(s *SetStmt, scope *TypeScope) (ok bool) {
	valType, ok := i.EvalForExprType(s.Value, scope)
	if !ok {
		return ok
	}

	if valType == nil {
		return false
	}

	// Evaluate the lhs type
	lhsType, ok := i.EvalForExprType(s.TargetExpr, scope)
	if !ok {
		return ok
	}
	if lhsType == nil {
		return false
	}

	// LHS MUST be a RefType
	if lhsType.Tag != decl.TypeTagRef {
		return i.Errorf(s.Pos(), "Cannot assign to a non ref type lhs.  Found: %s", lhsType.String())
	}

	// Make sure lhs ref type's param type and valtype match
	if !lhsType.Info.(*decl.RefTypeInfo).ParamType.Equals(valType) {
		return i.Errorf(s.Pos(), "LHS Type (%s) != RHS Type (%s)", lhsType.String(), valType.String())
	}
	return
}

func (i *Inference) EvalForLogStmt(l *LogStmt, scope *TypeScope) (ok bool) {
	ok = true
	for _, expr := range l.Args {
		_, ok2 := i.EvalForExprType(expr, scope)
		ok = ok && ok2
	}
	return
}

func (i *Inference) EvalForLetStmt(l *LetStmt, scope *TypeScope) (ok bool) {
	valType, ok := i.EvalForExprType(l.Value, scope)
	if !ok {
		return ok
	}

	if valType == nil {
		return false
	}
	if len(l.Variables) == 1 {
		varIdent := l.Variables[0]
		if errSet := scope.Set(varIdent.Value, varIdent, valType); errSet != nil {
			ok = ok && i.Errorf(varIdent.Pos(), "%v", errSet)
		}
	} else if len(l.Variables) > 1 {
		if valType.Tag == decl.TypeTagTuple && len(valType.Info.([]*Type)) == len(l.Variables) {
			childTypes := valType.Info.([]*Type)
			for idx, varIdent := range l.Variables {
				elemType := childTypes[idx]
				if errSet := scope.Set(varIdent.Value, varIdent, elemType); errSet != nil {
					ok = i.Errorf(varIdent.Pos(), "%v", errSet)
				}
			}
		} else {
			i.Errorf(l.Pos(), "let statement assigns to %d variables, but value type %s is not a matching tuple of %d elements", len(l.Variables), valType.String(), len(l.Variables))
		}
	}
	return
}

func (i *Inference) EvalForForStmt(f *ForStmt, scope *TypeScope) (ok bool) {
	ok = true
	condType, condOk := i.EvalForExprType(f.Condition, scope)
	if !condOk {
		return false
	}
	if !condType.Equals(BoolType) && !condType.Equals(IntType) {
		ok = i.Errorf(f.Pos(), "For loop condition can be bool or int, found: %s", condType.String())
	}

	// Evaluate block
	bodyScope := scope.Push(scope.currentComponent, scope.currentMethod)
	ok = ok && i.EvalForStmt(f.Body, bodyScope)
	return
}

func (i *Inference) EvalForBlockStmt(block *BlockStmt, parentScope *TypeScope) (ok bool) {
	ok = true
	blockScope := parentScope.Push(parentScope.currentComponent, parentScope.currentMethod)
	for _, stmt := range block.Statements {
		ok = ok && i.EvalForStmt(stmt, blockScope)
	}
	return
}

func (i *Inference) EvalForStmt(stmt Stmt, scope *TypeScope) (ok bool) {
	ok = true
	switch s := stmt.(type) {
	case *LetStmt:
		ok = ok && i.EvalForLetStmt(s, scope)
	case *LogStmt:
		ok = ok && i.EvalForLogStmt(s, scope)
	case *SetStmt:
		ok = ok && i.EvalForSetStmt(s, scope)
	case *ExprStmt:
		_, ok2 := i.EvalForExprType(s.Expression, scope)
		ok = ok && ok2
	case *ReturnStmt:
		var actualReturnType *Type = NilType
		if s.ReturnValue != nil {
			infRetType, ok2 := i.EvalForExprType(s.ReturnValue, scope)
			ok = ok && ok2
			if ok2 && infRetType != nil {
				actualReturnType = infRetType
			}
		}
		if scope.currentMethod != nil {
			var expectedReturnType *Type = NilType
			if scope.currentMethod.ReturnType != nil {
				// Ensure the method's ReturnType (a TypeDecl) is resolved to a *Type
				resolvedExpectedType := scope.currentMethod.ReturnType.ResolvedType() // Or .ResolveType(scope)
				if resolvedExpectedType == nil {
					i.Errorf(scope.currentMethod.ReturnType.Pos(), "method '%s' has an unresolvable return TypeDecl", scope.currentMethod.Name.Value)
					return false
				}
				expectedReturnType = resolvedExpectedType
			}
			if !actualReturnType.Equals(expectedReturnType) {
				isPromotion := actualReturnType.Equals(IntType) && expectedReturnType.Equals(FloatType)
				if !isPromotion {
					i.Errorf(s.Pos(), "return type mismatch for method '%s': expected %s, got %s", scope.currentMethod.Name.Value, expectedReturnType.String(), actualReturnType.String())
				}
			}
		} else {
			i.Errorf(s.Pos(), "return statement found outside of a method definition")
		}
	case *IfStmt:
		condType, ok2 := i.EvalForExprType(s.Condition, scope)
		ok = ok && ok2
		if ok2 && condType != nil && !condType.Equals(BoolType) {
			i.Errorf(s.Condition.Pos(), "if condition must be boolean, got %s", condType.String())
		}
		if s.Then != nil {
			ok = ok && i.EvalForBlockStmt(s.Then, scope)
		}
		if s.Else != nil {
			ok = ok && i.EvalForStmt(s.Else, scope)
		}
	case *ForStmt:
		ok = ok && i.EvalForForStmt(s, scope)
	case *DelayStmt:
		ok = ok && i.EvalForDelayStmt(s, scope)
	case *BlockStmt:
		ok = ok && i.EvalForBlockStmt(s, scope)
	case *AssignmentStmt: // General assignment, var must exist.
		existingVarType, varFound := scope.Get(s.Var.Value)
		if !varFound {
			i.Errorf(s.Var.Pos(), "assignment to undeclared variable '%s'", s.Var.Value)
			break
		}
		s.Var.SetInferredType(existingVarType)
		valType, ok := i.EvalForExprType(s.Value, scope)
		if ok && valType != nil && !valType.Equals(existingVarType) {
			isIntToFloat := valType.Equals(IntType) && existingVarType.Equals(FloatType)
			if !isIntToFloat {
				i.Errorf(s.Pos(), "type mismatch in assignment to '%s': variable is %s, value is %s", s.Var.Value, existingVarType.String(), valType.String())
			}
		}
		/*
			case *ExpectStmt:
				i.Errorf(pos, "ExpectStmt found outside of an analyze block's expectations; type checking skipped here", s.Pos().LineColStr())
		*/
	default:
		i.Errorf(stmt.Pos(), "type inference for statement type %T not implemented yet: %v", stmt, stmt)
	}
	return ok
}

func (i *Inference) EvalForSystemDecl(systemDecl *SystemDecl, nodeScope *TypeScope) (ok bool) {
	ok = true
	// Pass 1 - Infer types for all components and instances in the system declaration
	var instanceDecls []*InstanceDecl
	for _, item := range systemDecl.Body {
		switch it := item.(type) {
		case *InstanceDecl:
			compTypeNode, foundCompType := nodeScope.env.Get(it.ComponentName.Value)
			if !foundCompType {
				return i.Errorf(it.ComponentName.Pos(), "component type '%s' not found for instance '%s'", it.ComponentName.Value, it.Name.Value)
			}
			compDefinition, ok2 := compTypeNode.(*ComponentDecl)
			ok = ok && ok2
			if !ok2 {
				i.Errorf(it.ComponentName.Pos(), "identifier '%s' used as component type for instance '%s' is not a component declaration (got %T)", it.ComponentName.Value, it.Name.Value, compTypeNode)
				return false
			}
			instanceType := ComponentType(compDefinition)
			nodeScope.env.Set(it.Name.Value, it) // Store InstanceDecl node in env by its name
			it.Name.SetInferredType(instanceType)
			instanceDecls = append(instanceDecls, it)
		case *LetStmt:
			ok = ok && i.EvalForStmt(it, nodeScope)
		case *SetStmt:
			ok = ok && i.EvalForSetStmt(it, nodeScope)
		case *OptionsDecl:
			if it.Body != nil {
				optionsScope := nodeScope.Push(nil, nil)
				ok = ok && i.EvalForBlockStmt(it.Body, optionsScope)
			}
		default:
			// i.Errorf(item.Pos(), "type inference for system body item type %T not implemented yet", item)
		}
	}

	if !ok {
		return
	}

	// Pass 2 - Infer types for all instances and their overrides
	for _, it := range instanceDecls {
		for _, assign := range it.Overrides {
			valType, ok2 := i.EvalForExprType(assign.Value, nodeScope)
			if !ok2 {
				ok = false
				continue
			} else if valType == nil {
				continue
			}

			compTypeNode, _ := nodeScope.env.Get(it.ComponentName.Value) // no need to check ok here, it was checked in Pass 1
			compDefinition, _ := compTypeNode.(*ComponentDecl)

			paramDecl, _ := compDefinition.GetParam(assign.Var.Value)
			usesDecl, _ := compDefinition.GetDependency(assign.Var.Value)

			if paramDecl != nil {
				if paramDecl.TypeDecl == nil {
					i.Errorf(paramDecl.Pos(), "param '%s' of component '%s' has no type", paramDecl.Name.Value, compDefinition.Name.Value)
					continue
				}
				expectedType := nodeScope.ResolveType(paramDecl.TypeDecl) // Resolve TypeDecl in the current system scope
				if expectedType == nil {
					i.Errorf(paramDecl.TypeDecl.Pos(), "param '%s' of component '%s' has an unresolved TypeDecl '%s'", paramDecl.Name.Value, compDefinition.Name.Value, paramDecl.TypeDecl.Name)
					continue
				}
				if !valType.Equals(expectedType) {
					if !(valType.Equals(IntType) && expectedType.Equals(FloatType)) {
						i.Errorf(assign.Value.Pos(), "type mismatch for override param '%s' in instance '%s': expected %s, got %s", assign.Var.Value, it.Name.Value, expectedType.String(), valType.String())
					}
				}
			} else if usesDecl != nil {
				var assignedInstanceType *Type
				if assignedIdent, isIdent := assign.Value.(*IdentifierExpr); isIdent {
					assignedNodeFromEnv, foundInstance := nodeScope.env.Get(assignedIdent.Value)
					if foundInstance {
						if assignedInstDecl, isInst := assignedNodeFromEnv.(*InstanceDecl); isInst {
							// The type of an instance is its component type.
							// Retrieve the ComponentDecl for the assigned instance's component type.
							compTypeOfAssignedInstance, foundComp := nodeScope.env.Get(assignedInstDecl.ComponentName.Value)
							if foundComp {
								if actualCompDecl, isActualComp := compTypeOfAssignedInstance.(*ComponentDecl); isActualComp {
									assignedInstanceType = ComponentType(actualCompDecl)
								} else {
									i.Errorf(assign.Value.Pos(), "assigned instance '%s' for dependency '%s' has non-component type %T for its ComponentName ('%s')", assignedIdent.Value, assign.Var.Value, compTypeOfAssignedInstance, assignedInstDecl.ComponentName.Value)
								}
							} else {
								i.Errorf(assign.Value.Pos(), "could not find component type '%s' for assigned instance '%s' (dependency '%s')", assignedInstDecl.ComponentName.Value, assignedIdent.Value, assign.Var.Value)
							}
						} else {
							i.Errorf(assign.Value.Pos(), "assigned value '%s' for dependency '%s' is not an instance declaration (got %T)", assignedIdent.Value, assign.Var.Value, assignedNodeFromEnv)
						}
					} else {
						i.Errorf(assign.Value.Pos(), "assigned instance identifier '%s' for dependency '%s' not found in system scope", assignedIdent.Value, assign.Var.Value)
					}
				}

				expectedDepComp := usesDecl.ComponentName
				if assignedInstanceType != nil {
					assignedCompDecl := assignedInstanceType.Info.(*ComponentDecl)
					if assignedCompDecl.Name.Value != expectedDepComp.Value {
						i.Errorf(assign.Value.Pos(), "type mismatch for override dependency '%s' in instance '%s': expected instance of component type %s, got instance of %s", assign.Var.Value, it.Name.Value, expectedDepComp.Value, assignedCompDecl.Name.Value)
					}
				} else if valType.Tag != decl.TypeTagComponent || valType.Info.(*ComponentDecl).Name.Value != expectedDepComp.Value {
					// This fallback is if assigned value was not an identifier resolving to an InstanceDecl
					i.Errorf(assign.Value.Pos(), "type mismatch for override dependency '%s' in instance '%s': expected component type %s, got %s (and value was not a known instance of the correct type)", assign.Var.Value, it.Name.Value, expectedDepComp.Value, valType.String())
				}
			} else {
				i.Errorf(assign.Var.Pos(), "override target '%s' in instance '%s' is not a known parameter or dependency of component '%s'", assign.Var.Value, it.Name.Value, compDefinition.Name.Value)
			}
		}
	}
	return
}
