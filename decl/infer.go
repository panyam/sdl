package decl

import (
	"fmt"
	"os"
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
	// Path of the file being inferred (if provided)
	filePath string

	// Inference starts the root file
	rootFile *FileDecl

	// All the errors collected during inference
	Errors []error
}

func (i *Inference) HasErrors() bool {
	return len(i.Errors) > 0
}

func (i *Inference) PrintErrors() {
	for _, err := range i.Errors {
		fmt.Fprintln(os.Stderr, err)
	}
}

func (i *Inference) Errorf(pos Location, format string, args ...any) bool {
	i.AddErrors(InfErrorf(pos, format, args...))
	return false
}

func (i *Inference) AddErrors(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
		i.Errors = append(i.Errors, err)
	}
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
		for _, method := range compDecl.methods {
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
		compTypeNode, foundCompType := rootScope.env.Get(usesDecl.ComponentNode.Name)
		if !foundCompType {
			i.Errorf(usesDecl.ComponentNode.Pos(),
				"component type '%s' not found for dependency '%s'",
				usesDecl.ComponentNode.Name, usesDecl.NameNode.Name)
			return false
		}
		compDefinition, ok := compTypeNode.(*ComponentDecl)
		if !ok {
			i.Errorf(usesDecl.ComponentNode.Pos(), "identifier '%s' used as component type for instance '%s' is not a component declaration (got %T)", usesDecl.ComponentNode.Name, usesDecl.NameNode.Name, compTypeNode)
			return false
		}
		instanceType := ComponentTypeInstance(compDefinition)
		rootScope.env.Set(usesDecl.NameNode.Name, compDefinition) // Store InstanceDecl node in env by its name
		usesDecl.NameNode.SetInferredType(instanceType)
	}

	// Method signatures
	for _, method := range compDecl.methods { // Assuming direct field access or appropriate getter
		i.EvalForMethodSignature(method, compDecl, rootScope)
	}
	return
}

func (i *Inference) EvalForParamDecl(paramDecl *ParamDecl, compDecl *ComponentDecl, rootScope *TypeScope) (success bool) {
	var resolvedParamType *Type

	// Ensure that if all succeeds the type for the param is set in the root scope
	defer func() {
		if success && resolvedParamType != nil {
			rootScope.Set(paramDecl.Name.Name, paramDecl.Name, resolvedParamType) // Register the parameter type in the scope
		}
	}()

	if paramDecl.Type != nil { // Type is explicitly declared
		resolvedParamType = paramDecl.Type.TypeUsingScope(rootScope)
		if resolvedParamType == nil {
			i.Errorf(paramDecl.Type.Pos(), "unresolved type '%s' for parameter '%s' in component '%s'", paramDecl.Type.Name, paramDecl.Name.Name, compDecl.NameNode.Name)
			// Even if unresolved, continue to check default value if present, but this param is problematic.
		} else {
			paramDecl.Type.SetResolvedType(resolvedParamType)
		}
	} else if paramDecl.DefaultValue != nil { // No explicit type, but has default value
		// Infer type from default value
		resolvedParamType, success = i.EvalForExprType(paramDecl.DefaultValue, rootScope)
		if success {
			if resolvedParamType != nil {
				paramDecl.Name.SetInferredType(resolvedParamType)
			} else {
				// Default value's type could not be inferred.
				i.Errorf(paramDecl.DefaultValue.Pos(), "could not infer type from default value for parameter '%s' in component '%s'", paramDecl.Name.Name, compDecl.NameNode.Name)
			}
		}
	} else { // No explicit type AND no default value
		i.Errorf(paramDecl.Pos(), "parameter '%s' in component '%s' has no type declaration and no default value", paramDecl.Name.Name, compDecl.NameNode.Name)
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
							i.Errorf(paramDecl.DefaultValue.Pos(), "type mismatch for default value of parameter '%s' in component '%s': parameter type is %s, default value type is %s", paramDecl.Name.Name, compDecl.NameNode.Name, resolvedParamType.String(), defaultValueActualType.String())
						}
					}
				} else if paramDecl.Type != nil {
					// This means paramDecl.Type was specified but couldn't be resolved earlier,
					// yet we have a default value. This is an inconsistent state.
					// The earlier error about unresolved type for paramDecl.Type should cover this.
				}
			}
		}
	}
	return
}

// Infer/Check types for a method signature.  The body is not evaluated here
func (i *Inference) EvalForMethodSignature(method *MethodDecl, compDecl *ComponentDecl, rootScope *TypeScope) (errors []error) {
	for _, param := range method.Parameters {
		if param.Type != nil {
			resolvedParamType := param.Type.TypeUsingScope(rootScope)
			if resolvedParamType == nil {
				i.Errorf(param.Type.Pos(), "unresolved type '%s' for parameter '%s' in method '%s.%s'", param.Type.Name, param.Name.Name, compDecl.NameNode.Name, method.NameNode.Name)
			} else {
				param.Type.SetResolvedType(resolvedParamType)
			}
		} else {
			i.Errorf(param.Pos(), "parameter '%s' of method '%s.%s' has no type declaration", param.Name.Name, compDecl.NameNode.Name, method.NameNode.Name)
		}
	}
	if method.ReturnType != nil {
		resolvedReturnType := method.ReturnType.TypeUsingScope(rootScope)
		if resolvedReturnType == nil {
			i.Errorf(method.ReturnType.Pos(), "unresolved return type '%s' for method '%s.%s'", method.ReturnType.Name, compDecl.NameNode.Name, method.NameNode.Name)
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
	if expr.Value == nil || expr.Value.Type == nil {
		i.Errorf(expr.Pos(), "literal expression has invalid internal RuntimeValue or Type")
		return nil, false
	}
	return expr.Value.Type, true
}

func (i *Inference) EvalForIdentifierExpr(expr *IdentifierExpr, scope *TypeScope) (t *Type, ok bool) {
	t, ok = scope.Get(expr.Name)
	if !ok {
		return nil, i.Errorf(expr.Pos(), "identifier '%s' not found", expr.Name)
	}
	if t == nil {
		return nil, i.Errorf(expr.Pos(), "identifier '%s' resolved but its type is nil (internal error)", expr.Name)
	}
	return
}

func (i *Inference) EvalForMemberAccessExpr(expr *MemberAccessExpr, scope *TypeScope) (t *Type, ok bool) {
	receiverType, ok := i.EvalForExprType(expr.Receiver, scope)
	if !ok {
		return
	}
	if receiverType == nil {
		return nil, i.Errorf(expr.Pos(), "could not determine receiver type for member access '.%s'", expr.Member.Name)
	}
	memberName := expr.Member.Name

	if receiverType.OriginalDecl != nil {
		switch decl := receiverType.OriginalDecl.(type) {
		case *EnumDecl:
			if !receiverType.IsEnum {
				return nil, i.Errorf(expr.Pos(), "internal error: receiver type for '%s' (member '%s') has EnumDecl but IsEnum is false", receiverType.Name, memberName)
			}
			for _, valNode := range decl.ValuesNode {
				if valNode.Name == memberName {
					return receiverType, true
				}
			}
			return nil, i.Errorf(expr.Pos(), "value '%s' not found in enum '%s'", memberName, decl.NameNode.Name)

		case *ComponentDecl:
			if paramDecl, _ := decl.GetParam(memberName); paramDecl != nil {
				if paramDecl.Type == nil {
					return nil, i.Errorf(paramDecl.Pos(), "parameter '%s' in component '%s' lacks a type declaration", memberName, decl.NameNode.Name)
				}
				paramType := paramDecl.Type.Type() // This should use TypeUsingScope if paramDecl.Type itself needs scope.
				if paramType == nil {
					// paramDecl.Type.Type() needs to be robust or TypeUsingScope if it can resolve complex types.
					// If paramDecl.Type refers to e.g. an imported enum, Type.Type() must handle it.
					// Let's assume TypeDecl.Type() can resolve simple names or uses a pre-resolved *Type.
					paramType = paramDecl.Type.TypeUsingScope(scope) // Use scope to resolve complex types in TypeDecl
				}
				if paramType == nil {
					return nil, i.Errorf(paramDecl.Type.Pos(), "parameter '%s' in component '%s' has an unresolved TypeDecl '%s'", memberName, decl.NameNode.Name, paramDecl.Type.Name)
				}
				return paramType, true
			}
			if usesDecl, _ := decl.GetDependency(memberName); usesDecl != nil {
				if scope.env == nil {
					return nil, i.Errorf(usesDecl.Pos(), "internal error: TypeScope.env is nil when resolving 'uses' dependency '%s' in component '%s'", memberName, decl.NameNode.Name)
				}
				depCompName := usesDecl.ComponentNode.Name
				depCompDeclNode, found := scope.env.Get(depCompName)
				if !found {
					return nil, i.Errorf(usesDecl.Pos(), "'uses' dependency '%s' in component '%s' refers to unknown component type '%s'", memberName, decl.NameNode.Name, depCompName)
				}
				if depCompDecl, ok := depCompDeclNode.(*ComponentDecl); ok {
					return ComponentTypeInstance(depCompDecl), true
				}
				return nil, i.Errorf(usesDecl.Pos(), "'uses' dependency '%s' in component '%s' resolved to a non-component type %T for '%s'", memberName, decl.NameNode.Name, depCompDeclNode, depCompName)
			}
			if methodDecl, _ := decl.GetMethod(memberName); methodDecl != nil {
				return &Type{Name: "MethodReference", OriginalDecl: methodDecl}, ok
			}
			return nil, i.Errorf(expr.Pos(), "member '%s' not found in component '%s' (type %s)", memberName, decl.NameNode.Name, receiverType.Name)
		default:
			panic("Invalid original decl type - only enums and components are supported for now")
		}
	}

	if receiverType.Name == "List" && memberName == "Len" {
		return IntType, ok
	}

	return nil, i.Errorf(expr.Pos(), "cannot access member '%s' on type %s; receiver is not an enum, component, or known type with this member", memberName, receiverType.String())
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
			if leftType.IsEnum {
				return BoolType, true
			}
			if leftType.Name == "List" || leftType.Name == "Tuple" || leftType.Name == "Outcomes" ||
				(leftType.OriginalDecl != nil && leftType.IsComponentType()) ||
				leftType.Name == "OpNode" || leftType.Name == "MethodReference" {
				return nil, i.Errorf(expr.Pos(), "type mismatch for comparison operator '%s': cannot compare complex type %s", expr.Operator, leftType.String())
			}
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

	if funcType.Name == "MethodReference" {
		methodDecl, ok := funcType.OriginalDecl.(*MethodDecl)
		if !ok || methodDecl == nil {
			return nil, i.Errorf(expr.Function.Pos(), "internal error: 'MethodReference' type for '%s' did not contain a valid MethodDecl", funcNameForError)
		}

		// Attempt to construct a more descriptive name for error messages
		if mae, isMae := expr.Function.(*MemberAccessExpr); isMae {
			receiverStr := mae.Receiver.String() // Assuming String() is safe for resolved expressions
			if receiverStr != "" {
				funcNameForError = fmt.Sprintf("%s.%s", receiverStr, methodDecl.NameNode.Name)
			} else { // Fallback if receiver string is empty (e.g. if it was complex and String() was minimal)
				funcNameForError = methodDecl.NameNode.Name
			}
		} else { // Not a MemberAccessExpr, use method name directly
			funcNameForError = methodDecl.NameNode.Name
		}

		if methodDecl.ReturnType != nil {
			// Use TypeUsingScope for resolving TypeDecl within the method's component context if needed
			// However, return/param types are usually resolved using the global/import scope during the first pass.
			// For now, assume methodDecl.ReturnType.Type() is sufficient or TypeDecl.Type() handles resolution.
			resolvedReturnType := methodDecl.ReturnType.TypeUsingScope(scope) // Pass current scope
			if resolvedReturnType == nil {
				return nil, i.Errorf(methodDecl.ReturnType.Pos(), "method '%s' return TypeDecl ('%s') did not resolve to a valid Type", funcNameForError, methodDecl.ReturnType.Name)
			}
			returnType = resolvedReturnType
		}

		for _, paramDecl := range methodDecl.Parameters {
			if paramDecl.Type == nil {
				return nil, i.Errorf(paramDecl.Pos(), "parameter '%s' of method '%s' has no type declaration", paramDecl.Name.Name, funcNameForError)
			}
			paramSDLType := paramDecl.Type.TypeUsingScope(scope) // Pass current scope
			if paramSDLType == nil {
				return nil, i.Errorf(paramDecl.Type.Pos(), "parameter '%s' of method '%s' has invalid TypeDecl ('%s')", paramDecl.Name.Name, funcNameForError, paramDecl.Type.Name)
			}
			expectedParamTypes = append(expectedParamTypes, paramSDLType)
		}
	} else {
		return nil, i.Errorf(expr.Function.Pos(), "calling non-method type '%s' as function is not supported or function not found", funcType.String())
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
	if fromType.Name != "Outcomes" || len(fromType.ChildTypes) != 1 {
		return nil, i.Errorf(expr.Pos(), "type mismatch for sample expression: 'from' expression must be Outcomes[T], got %s", fromType.String())
	}
	return fromType.ChildTypes[0], true
}

// --- Statement Type Inference ---

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
		if valType, ok := i.EvalForExprType(s.Value, scope); ok {
			if valType != nil {
				if len(s.Variables) == 1 {
					varIdent := s.Variables[0]
					if errSet := scope.Set(varIdent.Name, varIdent, valType); errSet != nil {
						ok = i.Errorf(varIdent.Pos(), "%v", errSet)
					}
				} else if len(s.Variables) > 1 {
					if valType.Name == "Tuple" && len(valType.ChildTypes) == len(s.Variables) {
						for idx, varIdent := range s.Variables {
							elemType := valType.ChildTypes[idx]
							if errSet := scope.Set(varIdent.Name, varIdent, elemType); errSet != nil {
								ok = i.Errorf(varIdent.Pos(), "%v", errSet)
							}
						}
					} else {
						i.Errorf(s.Pos(), "let statement assigns to %d variables, but value type %s is not a matching tuple of %d elements", len(s.Variables), valType.String(), len(s.Variables))
					}
				}
			}
		}
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
				resolvedExpectedType := scope.currentMethod.ReturnType.ResolvedType() // Or .TypeUsingScope(scope)
				if resolvedExpectedType == nil {
					i.Errorf(scope.currentMethod.ReturnType.Pos(), "method '%s' has an unresolvable return TypeDecl", scope.currentMethod.NameNode.Name)
					return false
				}
				expectedReturnType = resolvedExpectedType
			}
			if !actualReturnType.Equals(expectedReturnType) {
				isPromotion := actualReturnType.Equals(IntType) && expectedReturnType.Equals(FloatType)
				if !isPromotion {
					i.Errorf(s.Pos(), "return type mismatch for method '%s': expected %s, got %s", scope.currentMethod.NameNode.Name, expectedReturnType.String(), actualReturnType.String())
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
	case *BlockStmt:
		ok = ok && i.EvalForBlockStmt(s, scope)
	case *AssignmentStmt: // General assignment, var must exist.
		existingVarType, varFound := scope.Get(s.Var.Name)
		if !varFound {
			i.Errorf(s.Var.Pos(), "assignment to undeclared variable '%s'", s.Var.Name)
			break
		}
		s.Var.SetInferredType(existingVarType)
		valType, ok := i.EvalForExprType(s.Value, scope)
		if ok && valType != nil && !valType.Equals(existingVarType) {
			isIntToFloat := valType.Equals(IntType) && existingVarType.Equals(FloatType)
			if !isIntToFloat {
				i.Errorf(s.Pos(), "type mismatch in assignment to '%s': variable is %s, value is %s", s.Var.Name, existingVarType.String(), valType.String())
			}
		}
		/*
			case *ExpectStmt:
				i.Errorf(pos, "ExpectStmt found outside of an analyze block's expectations; type checking skipped here", s.Pos().LineColStr())
		*/
	default:
		// i.Errorf(stmt.Pos(), "type inference for statement type %T not implemented yet", stmt)
	}
	return false
}

func (i *Inference) EvalForSystemDecl(systemDecl *SystemDecl, nodeScope *TypeScope) (ok bool) {
	ok = true
	// Pass 1 - Infer types for all components and instances in the system declaration
	var instanceDecls []*InstanceDecl
	for _, item := range systemDecl.Body {
		switch it := item.(type) {
		case *InstanceDecl:
			compTypeNode, foundCompType := nodeScope.env.Get(it.ComponentType.Name)
			if !foundCompType {
				return i.Errorf(it.ComponentType.Pos(), "component type '%s' not found for instance '%s'", it.ComponentType.Name, it.NameNode.Name)
			}
			compDefinition, ok2 := compTypeNode.(*ComponentDecl)
			ok = ok && ok2
			if !ok2 {
				i.Errorf(it.ComponentType.Pos(), "identifier '%s' used as component type for instance '%s' is not a component declaration (got %T)", it.ComponentType.Name, it.NameNode.Name, compTypeNode)
				return false
			}
			instanceType := ComponentTypeInstance(compDefinition)
			nodeScope.env.Set(it.NameNode.Name, it) // Store InstanceDecl node in env by its name
			it.NameNode.SetInferredType(instanceType)
			instanceDecls = append(instanceDecls, it)
		case *LetStmt:
			ok = ok && i.EvalForStmt(it, nodeScope)
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

			compTypeNode, _ := nodeScope.env.Get(it.ComponentType.Name) // no need to check ok here, it was checked in Pass 1
			compDefinition, _ := compTypeNode.(*ComponentDecl)

			paramDecl, _ := compDefinition.GetParam(assign.Var.Name)
			usesDecl, _ := compDefinition.GetDependency(assign.Var.Name)

			if paramDecl != nil {
				if paramDecl.Type == nil {
					i.Errorf(paramDecl.Pos(), "param '%s' of component '%s' has no type", paramDecl.Name.Name, compDefinition.NameNode.Name)
					continue
				}
				expectedType := paramDecl.Type.TypeUsingScope(nodeScope) // Resolve TypeDecl in the current system scope
				if expectedType == nil {
					i.Errorf(paramDecl.Type.Pos(), "param '%s' of component '%s' has an unresolved TypeDecl '%s'", paramDecl.Name.Name, compDefinition.NameNode.Name, paramDecl.Type.Name)
					continue
				}
				if !valType.Equals(expectedType) {
					if !(valType.Equals(IntType) && expectedType.Equals(FloatType)) {
						i.Errorf(assign.Value.Pos(), "type mismatch for override param '%s' in instance '%s': expected %s, got %s", assign.Var.Name, it.NameNode.Name, expectedType.String(), valType.String())
					}
				}
			} else if usesDecl != nil {
				expectedDepCompName := usesDecl.ComponentNode.Name
				var assignedInstanceType *Type
				if assignedIdent, isIdent := assign.Value.(*IdentifierExpr); isIdent {
					assignedNodeFromEnv, foundInstance := nodeScope.env.Get(assignedIdent.Name)
					if foundInstance {
						if assignedInstDecl, isInst := assignedNodeFromEnv.(*InstanceDecl); isInst {
							// The type of an instance is its component type.
							// Retrieve the ComponentDecl for the assigned instance's component type.
							compTypeOfAssignedInstance, foundComp := nodeScope.env.Get(assignedInstDecl.ComponentType.Name)
							if foundComp {
								if actualCompDecl, isActualComp := compTypeOfAssignedInstance.(*ComponentDecl); isActualComp {
									assignedInstanceType = ComponentTypeInstance(actualCompDecl)
								} else {
									i.Errorf(assign.Value.Pos(), "assigned instance '%s' for dependency '%s' has non-component type %T for its ComponentType ('%s')", assignedIdent.Name, assign.Var.Name, compTypeOfAssignedInstance, assignedInstDecl.ComponentType.Name)
								}
							} else {
								i.Errorf(assign.Value.Pos(), "could not find component type '%s' for assigned instance '%s' (dependency '%s')", assignedInstDecl.ComponentType.Name, assignedIdent.Name, assign.Var.Name)
							}
						} else {
							i.Errorf(assign.Value.Pos(), "assigned value '%s' for dependency '%s' is not an instance declaration (got %T)", assignedIdent.Name, assign.Var.Name, assignedNodeFromEnv)
						}
					} else {
						i.Errorf(assign.Value.Pos(), "assigned instance identifier '%s' for dependency '%s' not found in system scope", assignedIdent.Name, assign.Var.Name)
					}
				}

				if assignedInstanceType != nil {
					if assignedInstanceType.Name != expectedDepCompName {
						i.Errorf(assign.Value.Pos(), "type mismatch for override dependency '%s' in instance '%s': expected instance of component type %s, got instance of %s", assign.Var.Name, it.NameNode.Name, expectedDepCompName, assignedInstanceType.Name)
					}
				} else if valType.Name != expectedDepCompName || valType.OriginalDecl == nil || !valType.IsComponentType() {
					// This fallback is if assigned value was not an identifier resolving to an InstanceDecl
					i.Errorf(assign.Value.Pos(), "type mismatch for override dependency '%s' in instance '%s': expected component type %s, got %s (and value was not a known instance of the correct type)", assign.Var.Name, it.NameNode.Name, expectedDepCompName, valType.String())
				}
			} else {
				i.Errorf(assign.Var.Pos(), "override target '%s' in instance '%s' is not a known parameter or dependency of component '%s'", assign.Var.Name, it.NameNode.Name, compDefinition.NameNode.Name)
			}
		}
	}
	return
}
