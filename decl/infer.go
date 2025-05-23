package decl

import (
	"fmt"
	// "log" // Uncomment for debugging
)

// InferExprType recursively infers the type of an expression and sets its InferredType field.
func InferExprType(expr Expr, scope *TypeScope) (*Type, error) {
	if expr == nil {
		return nil, fmt.Errorf("cannot infer type for nil expression")
	}

	if expr.InferredType() != nil {
		return expr.InferredType(), nil
	}

	var inferred *Type
	var err error

	switch e := expr.(type) {
	case *LiteralExpr:
		inferred, err = InferLiteralExprType(e, scope)
	case *IdentifierExpr:
		inferred, err = InferIdentifierExprType(e, scope)
	case *BinaryExpr:
		inferred, err = InferBinaryExprType(e, scope)
	case *UnaryExpr:
		inferred, err = InferUnaryExprType(e, scope)
	case *MemberAccessExpr:
		inferred, err = InferMemberAccessExprType(e, scope)
	case *CallExpr:
		inferred, err = InferCallExprType(e, scope)
	case *TupleExpr:
		inferred, err = InferTupleExprType(e, scope)
	case *DistributeExpr:
		inferred, err = InferDistributeExprType(e, scope)
	case *SampleExpr:
		inferred, err = InferSampleExprType(e, scope)
	case *CaseExpr:
		if e.Body == nil {
			return nil, fmt.Errorf("CaseExpr at pos %s has no body", e.Pos().LineColStr())
		}
		inferred, err = InferExprType(e.Body, scope)
	default:
		return nil, fmt.Errorf("type inference not implemented for expression type %T at pos %s", expr, expr.Pos().LineColStr())
	}

	if err != nil {
		return nil, err
	}

	if inferred == nil {
		return nil, fmt.Errorf("type inference failed for %T at pos %s, but no error reported (inferred type is nil)", expr, expr.Pos().LineColStr())
	}

	expr.SetInferredType(inferred)

	if expr.DeclaredType() != nil && !expr.DeclaredType().Equals(inferred) {
		isIntToFloatPromotion := expr.DeclaredType().Equals(FloatType) && inferred.Equals(IntType)
		if !isIntToFloatPromotion {
			return inferred, fmt.Errorf("pos %s: type mismatch for '%s': inferred type %s, but declared type is %s",
				expr.Pos().LineColStr(), expr.String(), inferred.String(), expr.DeclaredType().String())
		}
	}
	return inferred, nil
}

func InferLiteralExprType(expr *LiteralExpr, scope *TypeScope) (*Type, error) {
	if expr.Value == nil || expr.Value.Type == nil {
		return nil, fmt.Errorf("pos %s: literal expression has invalid internal RuntimeValue or Type", expr.Pos().LineColStr())
	}
	return expr.Value.Type, nil
}

func InferIdentifierExprType(expr *IdentifierExpr, scope *TypeScope) (*Type, error) {
	t, ok := scope.Get(expr.Name)
	if !ok {
		return nil, fmt.Errorf("pos %s: identifier '%s' not found", expr.Pos().LineColStr(), expr.Name)
	}
	if t == nil {
		return nil, fmt.Errorf("pos %s: identifier '%s' resolved but its type is nil (internal error)", expr.Pos().LineColStr(), expr.Name)
	}
	return t, nil
}

func InferMemberAccessExprType(expr *MemberAccessExpr, scope *TypeScope) (*Type, error) {
	receiverType, err := InferExprType(expr.Receiver, scope)
	if err != nil {
		return nil, fmt.Errorf("pos %s: error inferring receiver type for member access '.%s': %w", expr.Receiver.Pos().LineColStr(), expr.Member.Name, err)
	}
	if receiverType == nil {
		return nil, fmt.Errorf("pos %s: could not determine receiver type for member access '.%s'", expr.Pos().LineColStr(), expr.Member.Name)
	}
	memberName := expr.Member.Name

	if receiverType.OriginalDecl != nil {
		switch decl := receiverType.OriginalDecl.(type) {
		case *EnumDecl:
			if !receiverType.IsEnum {
				return nil, fmt.Errorf("pos %s: internal error: receiver type for '%s' (member '%s') has EnumDecl but IsEnum is false", expr.Pos().LineColStr(), receiverType.Name, memberName)
			}
			for _, valNode := range decl.ValuesNode {
				if valNode.Name == memberName {
					return receiverType, nil
				}
			}
			return nil, fmt.Errorf("pos %s: value '%s' not found in enum '%s'", expr.Pos().LineColStr(), memberName, decl.NameNode.Name)

		case *ComponentDecl:
			if paramDecl, _ := decl.GetParam(memberName); paramDecl != nil {
				if paramDecl.Type == nil {
					return nil, fmt.Errorf("pos %s: parameter '%s' in component '%s' lacks a type declaration", paramDecl.Pos().LineColStr(), memberName, decl.NameNode.Name)
				}
				paramType := paramDecl.Type.Type() // This should use TypeUsingScope if paramDecl.Type itself needs scope.
				if paramType == nil {
					// paramDecl.Type.Type() needs to be robust or TypeUsingScope if it can resolve complex types.
					// If paramDecl.Type refers to e.g. an imported enum, Type.Type() must handle it.
					// Let's assume TypeDecl.Type() can resolve simple names or uses a pre-resolved *Type.
					paramType = paramDecl.Type.TypeUsingScope(scope) // Use scope to resolve complex types in TypeDecl
					if paramType == nil {
						return nil, fmt.Errorf("pos %s: parameter '%s' in component '%s' has an unresolved TypeDecl '%s'", paramDecl.Type.Pos().LineColStr(), memberName, decl.NameNode.Name, paramDecl.Type.Name)
					}
				}
				return paramType, nil
			}
			if usesDecl, _ := decl.GetDependency(memberName); usesDecl != nil {
				if scope.env == nil {
					return nil, fmt.Errorf("pos %s: internal error: TypeScope.env is nil when resolving 'uses' dependency '%s' in component '%s'", usesDecl.Pos().LineColStr(), memberName, decl.NameNode.Name)
				}
				depCompName := usesDecl.ComponentNode.Name
				depCompDeclNode, found := scope.env.Get(depCompName)
				if !found {
					return nil, fmt.Errorf("pos %s: 'uses' dependency '%s' in component '%s' refers to unknown component type '%s'", usesDecl.Pos().LineColStr(), memberName, decl.NameNode.Name, depCompName)
				}
				if depCompDecl, ok := depCompDeclNode.(*ComponentDecl); ok {
					return ComponentTypeInstance(depCompDecl), nil
				}
				return nil, fmt.Errorf("pos %s: 'uses' dependency '%s' in component '%s' resolved to a non-component type %T for '%s'", usesDecl.Pos().LineColStr(), memberName, decl.NameNode.Name, depCompDeclNode, depCompName)
			}
			if methodDecl, _ := decl.GetMethod(memberName); methodDecl != nil {
				return &Type{Name: "MethodReference", OriginalDecl: methodDecl}, nil
			}
			return nil, fmt.Errorf("pos %s: member '%s' not found in component '%s' (type %s)", expr.Pos().LineColStr(), memberName, decl.NameNode.Name, receiverType.Name)
		default:
			break
		}
	}

	if receiverType.Name == "List" && memberName == "Len" {
		return IntType, nil
	}

	return nil, fmt.Errorf("pos %s: cannot access member '%s' on type %s; receiver is not an enum, component, or known type with this member",
		expr.Pos().LineColStr(), memberName, receiverType.String())
}

func InferBinaryExprType(expr *BinaryExpr, scope *TypeScope) (*Type, error) {
	leftType, err := InferExprType(expr.Left, scope)
	if err != nil {
		return nil, fmt.Errorf("pos %s: error inferring type for left operand of binary expr ('%s'): %w", expr.Left.Pos().LineColStr(), expr.Operator, err)
	}
	rightType, err := InferExprType(expr.Right, scope)
	if err != nil {
		return nil, fmt.Errorf("pos %s: error inferring type for right operand of binary expr ('%s'): %w", expr.Right.Pos().LineColStr(), expr.Operator, err)
	}
	if leftType == nil || rightType == nil {
		return nil, fmt.Errorf("pos %s: could not determine type for one or both operands for binary expr ('%s')", expr.Pos().LineColStr(), expr.Operator)
	}

	switch expr.Operator {
	case "+", "-", "*", "/":
		if leftType.Equals(IntType) && rightType.Equals(IntType) {
			return IntType, nil
		}
		if (leftType.Equals(IntType) || leftType.Equals(FloatType)) &&
			(rightType.Equals(IntType) || rightType.Equals(FloatType)) {
			return FloatType, nil
		}
		if expr.Operator == "+" && leftType.Equals(StrType) && rightType.Equals(StrType) {
			return StrType, nil
		}
		return nil, fmt.Errorf("pos %s: type mismatch for operator '%s': cannot apply to %s and %s",
			expr.Pos().LineColStr(), expr.Operator, leftType.String(), rightType.String())
	case "%":
		if leftType.Equals(IntType) && rightType.Equals(IntType) {
			return IntType, nil
		}
		return nil, fmt.Errorf("pos %s: type mismatch for operator '%%': requires two integers, got %s and %s",
			expr.Pos().LineColStr(), leftType.String(), rightType.String())
	case "==", "!=", "<", "<=", ">", ">=":
		isLeftNumeric := leftType.Equals(IntType) || leftType.Equals(FloatType)
		isRightNumeric := rightType.Equals(IntType) || rightType.Equals(FloatType)

		if isLeftNumeric && isRightNumeric {
			return BoolType, nil
		}
		if leftType.Equals(rightType) {
			if leftType.IsEnum {
				return BoolType, nil
			}
			if leftType.Name == "List" || leftType.Name == "Tuple" || leftType.Name == "Outcomes" ||
				(leftType.OriginalDecl != nil && leftType.IsComponentType()) ||
				leftType.Name == "OpNode" || leftType.Name == "MethodReference" {
				return nil, fmt.Errorf("pos %s: type mismatch for comparison operator '%s': cannot compare complex type %s", expr.Pos().LineColStr(), expr.Operator, leftType.String())
			}
			return BoolType, nil
		}
		return nil, fmt.Errorf("pos %s: type mismatch for comparison operator '%s': cannot compare %s and %s",
			expr.Pos().LineColStr(), expr.Operator, leftType.String(), rightType.String())
	case "&&", "||":
		if leftType.Equals(BoolType) && rightType.Equals(BoolType) {
			return BoolType, nil
		}
		return nil, fmt.Errorf("pos %s: type mismatch for logical operator '%s': requires two booleans, got %s and %s",
			expr.Pos().LineColStr(), expr.Operator, leftType.String(), rightType.String())
	default:
		return nil, fmt.Errorf("pos %s: unsupported binary operator '%s'", expr.Pos().LineColStr(), expr.Operator)
	}
}

func InferUnaryExprType(expr *UnaryExpr, scope *TypeScope) (*Type, error) {
	rightType, err := InferExprType(expr.Right, scope)
	if err != nil {
		return nil, fmt.Errorf("pos %s: error inferring type for operand of unary expr ('%s'): %w", expr.Right.Pos().LineColStr(), expr.Operator, err)
	}
	if rightType == nil {
		return nil, fmt.Errorf("pos %s: could not determine type for operand of unary expr ('%s')", expr.Pos().LineColStr(), expr.Operator)
	}

	switch expr.Operator {
	case "!":
		if !rightType.Equals(BoolType) {
			return nil, fmt.Errorf("pos %s: type mismatch for operator '!': requires boolean, got %s",
				expr.Pos().LineColStr(), rightType.String())
		}
		return BoolType, nil
	case "-":
		if rightType.Equals(IntType) || rightType.Equals(FloatType) {
			return rightType, nil
		}
		return nil, fmt.Errorf("pos %s: type mismatch for operator '-': requires integer or float, got %s",
			expr.Pos().LineColStr(), rightType.String())
	default:
		return nil, fmt.Errorf("pos %s: unsupported unary operator '%s'", expr.Pos().LineColStr(), expr.Operator)
	}
}

func InferCallExprType(expr *CallExpr, scope *TypeScope) (*Type, error) {
	funcType, err := InferExprType(expr.Function, scope)
	if err != nil {
		return nil, fmt.Errorf("pos %s: error inferring type of function/method being called ('%s'): %w", expr.Function.Pos().LineColStr(), expr.Function.String(), err)
	}
	if funcType == nil {
		return nil, fmt.Errorf("pos %s: could not determine type of function/method being called ('%s')", expr.Function.Pos().LineColStr(), expr.Function.String())
	}

	var returnType *Type = NilType
	var expectedParamTypes []*Type
	var funcNameForError string = expr.Function.String()

	if funcType.Name == "MethodReference" {
		methodDecl, ok := funcType.OriginalDecl.(*MethodDecl)
		if !ok || methodDecl == nil {
			return nil, fmt.Errorf("pos %s: internal error: 'MethodReference' type for '%s' did not contain a valid MethodDecl", expr.Function.Pos().LineColStr(), funcNameForError)
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
				return nil, fmt.Errorf("pos %s: method '%s' return TypeDecl ('%s') did not resolve to a valid Type", methodDecl.ReturnType.Pos().LineColStr(), funcNameForError, methodDecl.ReturnType.Name)
			}
			returnType = resolvedReturnType
		}

		for _, paramDecl := range methodDecl.Parameters {
			if paramDecl.Type == nil {
				return nil, fmt.Errorf("pos %s: parameter '%s' of method '%s' has no type declaration", paramDecl.Pos().LineColStr(), paramDecl.Name.Name, funcNameForError)
			}
			paramSDLType := paramDecl.Type.TypeUsingScope(scope) // Pass current scope
			if paramSDLType == nil {
				return nil, fmt.Errorf("pos %s: parameter '%s' of method '%s' has invalid TypeDecl ('%s')", paramDecl.Type.Pos().LineColStr(), paramDecl.Name.Name, funcNameForError, paramDecl.Type.Name)
			}
			expectedParamTypes = append(expectedParamTypes, paramSDLType)
		}
	} else {
		return nil, fmt.Errorf("pos %s: calling non-method type '%s' as function is not supported or function not found", expr.Function.Pos().LineColStr(), funcType.String())
	}

	if len(expr.Args) != len(expectedParamTypes) {
		return nil, fmt.Errorf("pos %s: argument count mismatch for call to '%s': expected %d, got %d",
			expr.Pos().LineColStr(), funcNameForError, len(expectedParamTypes), len(expr.Args))
	}

	for i, argExpr := range expr.Args {
		argType, err_arg := InferExprType(argExpr, scope)
		if err_arg != nil {
			return nil, fmt.Errorf("pos %s: error inferring type for argument %d of call to '%s': %w",
				argExpr.Pos().LineColStr(), i+1, funcNameForError, err_arg)
		}
		if argType == nil {
			return nil, fmt.Errorf("pos %s: could not determine type for argument %d of call to '%s'", argExpr.Pos().LineColStr(), i+1, funcNameForError)
		}
		if !argType.Equals(expectedParamTypes[i]) {
			isIntToFloat := argType.Equals(IntType) && expectedParamTypes[i].Equals(FloatType)
			if !isIntToFloat {
				return nil, fmt.Errorf("pos %s: type mismatch for argument %d of call to '%s': expected %s, got %s",
					argExpr.Pos().LineColStr(), i+1, funcNameForError, expectedParamTypes[i].String(), argType.String())
			}
		}
	}
	return returnType, nil
}

func InferTupleExprType(expr *TupleExpr, scope *TypeScope) (*Type, error) {
	if len(expr.Children) == 0 {
		return nil, fmt.Errorf("pos %s: tuple expression must have at least one child (empty tuples not supported)", expr.Pos().LineColStr())
	}
	childTypes, err := InferTypesForExprList(expr.Children, scope)
	if err == nil {
		return TupleType(childTypes...), nil
	}
	return nil, err
}

func InferTypesForExprList(exprlist []Expr, scope *TypeScope) ([]*Type, error) {
	childTypes := make([]*Type, len(exprlist))
	for i, childExpr := range exprlist {
		childType, err := InferExprType(childExpr, scope)
		if err != nil {
			return nil, fmt.Errorf("pos %s: error inferring type for tuple element %d: %w", childExpr.Pos().LineColStr(), i+1, err)
		}
		if childType == nil {
			return nil, fmt.Errorf("pos %s: could not determine type for tuple element %d", childExpr.Pos().LineColStr(), i+1)
		}
		childTypes[i] = childType
	}
	return childTypes, nil
}

func InferDistributeExprType(expr *DistributeExpr, scope *TypeScope) (*Type, error) {
	var commonBodyType *Type
	if len(expr.Cases) == 0 && expr.Default == nil {
		return nil, fmt.Errorf("pos %s: distribute expression must have at least one case or a default", expr.Pos().LineColStr())
	}
	if expr.TotalProb != nil {
		totalProbType, err := InferExprType(expr.TotalProb, scope)
		if err != nil {
			return nil, fmt.Errorf("pos %s: error inferring type for total probability of distribute expr: %w", expr.TotalProb.Pos().LineColStr(), err)
		}
		if totalProbType != nil && !(totalProbType.Equals(IntType) || totalProbType.Equals(FloatType)) {
			return nil, fmt.Errorf("pos %s: total probability of distribute expr must be numeric, got %s", expr.TotalProb.Pos().LineColStr(), totalProbType.String())
		}
	}

	for i, caseExpr := range expr.Cases {
		if caseExpr.Condition == nil { // Should be caught by parser
			return nil, fmt.Errorf("pos %s: DistributeExpr case %d has no condition", caseExpr.Pos().LineColStr(), i)
		}
		condType, err := InferExprType(caseExpr.Condition, scope)
		if err != nil {
			return nil, fmt.Errorf("pos %s: error inferring type for condition of case %d in distribute expr: %w", caseExpr.Condition.Pos().LineColStr(), i, err)
		}
		if !(condType.Equals(FloatType) || condType.Equals(IntType)) {
			return nil, fmt.Errorf("pos %s: condition of distribute case %d must be numeric (for weight), got %s", caseExpr.Condition.Pos().LineColStr(), i, condType.String())
		}
		if caseExpr.Body == nil { // Should be caught by parser
			return nil, fmt.Errorf("pos %s: DistributeExpr case %d has no body", caseExpr.Pos().LineColStr(), i)
		}
		bodyType, err := InferExprType(caseExpr.Body, scope)
		if err != nil {
			return nil, fmt.Errorf("pos %s: error inferring type for body of case %d in distribute expr: %w", caseExpr.Body.Pos().LineColStr(), i, err)
		}
		if bodyType == nil {
			return nil, fmt.Errorf("pos %s: could not determine type for body of case %d in distribute expr", caseExpr.Pos().LineColStr(), i)
		}
		if commonBodyType == nil {
			commonBodyType = bodyType
		} else if !commonBodyType.Equals(bodyType) {
			return nil, fmt.Errorf("pos %s: type mismatch in distribute expr cases: expected %s (from case 0), got %s for case %d",
				expr.Pos().LineColStr(), commonBodyType.String(), bodyType.String(), i)
		}
	}

	if expr.Default != nil {
		defaultType, err := InferExprType(expr.Default, scope)
		if err != nil {
			return nil, fmt.Errorf("pos %s: error inferring type for default case of distribute expr: %w", expr.Default.Pos().LineColStr(), err)
		}
		if defaultType == nil {
			return nil, fmt.Errorf("pos %s: could not determine type for default case of distribute expr", expr.Default.Pos().LineColStr())
		}
		if commonBodyType == nil {
			commonBodyType = defaultType
		} else if !commonBodyType.Equals(defaultType) {
			return nil, fmt.Errorf("pos %s: type mismatch between distribute expr cases and default: expected %s, got %s for default",
				expr.Pos().LineColStr(), commonBodyType.String(), defaultType.String())
		}
	}
	if commonBodyType == nil {
		return nil, fmt.Errorf("pos %s: distribute expr has no effective common type", expr.Pos().LineColStr())
	}
	return OutcomesType(commonBodyType), nil
}

func InferSampleExprType(expr *SampleExpr, scope *TypeScope) (*Type, error) {
	fromType, err := InferExprType(expr.FromExpr, scope)
	if err != nil {
		return nil, fmt.Errorf("pos %s: error inferring type for 'from' expression of sample expr: %w", expr.FromExpr.Pos().LineColStr(), err)
	}
	if fromType == nil {
		return nil, fmt.Errorf("pos %s: could not determine type for 'from' expression of sample expr", expr.Pos().LineColStr())
	}
	if fromType.Name != "Outcomes" || len(fromType.ChildTypes) != 1 {
		return nil, fmt.Errorf("pos %s: type mismatch for sample expression: 'from' expression must be Outcomes[T], got %s",
			expr.Pos().LineColStr(), fromType.String())
	}
	return fromType.ChildTypes[0], nil
}

// --- Statement Type Inference ---

func InferTypesForBlockStmt(block *BlockStmt, parentScope *TypeScope) []error {
	var errors []error
	blockScope := parentScope.Push(parentScope.currentComponent, parentScope.currentMethod)
	for _, stmt := range block.Statements {
		errs := InferTypesForStmt(stmt, blockScope)
		errors = append(errors, errs...)
	}
	return errors
}

func InferTypesForStmt(stmt Stmt, scope *TypeScope) []error {
	var errors []error
	switch s := stmt.(type) {
	case *LetStmt:
		valType, err := InferExprType(s.Value, scope)
		if err != nil {
			errors = append(errors, fmt.Errorf("pos %s: error inferring type for value of let statement variable(s): %w", s.Pos().LineColStr(), err))
		} else if valType != nil {
			if len(s.Variables) == 1 {
				varIdent := s.Variables[0]
				if errSet := scope.Set(varIdent.Name, varIdent, valType); errSet != nil {
					errors = append(errors, fmt.Errorf("pos %s: %w", varIdent.Pos().LineColStr(), errSet))
				}
			} else if len(s.Variables) > 1 {
				if valType.Name == "Tuple" && len(valType.ChildTypes) == len(s.Variables) {
					for i, varIdent := range s.Variables {
						elemType := valType.ChildTypes[i]
						if errSet := scope.Set(varIdent.Name, varIdent, elemType); errSet != nil {
							errors = append(errors, fmt.Errorf("pos %s: %w", varIdent.Pos().LineColStr(), errSet))
						}
					}
				} else {
					errors = append(errors, fmt.Errorf("pos %s: let statement assigns to %d variables, but value type %s is not a matching tuple of %d elements", s.Pos().LineColStr(), len(s.Variables), valType.String(), len(s.Variables)))
				}
			}
		}
	case *ExprStmt:
		_, err := InferExprType(s.Expression, scope)
		if err != nil {
			errors = append(errors, fmt.Errorf("pos %s: error in expression statement: %w", s.Expression.Pos().LineColStr(), err))
		}
	case *ReturnStmt:
		var actualReturnType *Type = NilType
		if s.ReturnValue != nil {
			infRetType, err := InferExprType(s.ReturnValue, scope)
			if err != nil {
				errors = append(errors, fmt.Errorf("pos %s: error inferring type for return value: %w", s.ReturnValue.Pos().LineColStr(), err))
			} else if infRetType != nil {
				actualReturnType = infRetType
			}
		}
		if scope.currentMethod != nil {
			var expectedReturnType *Type = NilType
			if scope.currentMethod.ReturnType != nil {
				// Ensure the method's ReturnType (a TypeDecl) is resolved to a *Type
				resolvedExpectedType := scope.currentMethod.ReturnType.ResolvedType() // Or .TypeUsingScope(scope)
				if resolvedExpectedType == nil {
					errors = append(errors, fmt.Errorf("pos %s: method '%s' has an unresolvable return TypeDecl", scope.currentMethod.ReturnType.Pos().LineColStr(), scope.currentMethod.NameNode.Name))
					return errors
				}
				expectedReturnType = resolvedExpectedType
			}
			if !actualReturnType.Equals(expectedReturnType) {
				isPromotion := actualReturnType.Equals(IntType) && expectedReturnType.Equals(FloatType)
				if !isPromotion {
					errors = append(errors, fmt.Errorf("pos %s: return type mismatch for method '%s': expected %s, got %s",
						s.Pos().LineColStr(), scope.currentMethod.NameNode.Name, expectedReturnType.String(), actualReturnType.String()))
				}
			}
		} else {
			errors = append(errors, fmt.Errorf("pos %s: return statement found outside of a method definition", s.Pos().LineColStr()))
		}
	case *IfStmt:
		condType, err := InferExprType(s.Condition, scope)
		if err != nil {
			errors = append(errors, fmt.Errorf("pos %s: error inferring type for if condition: %w", s.Condition.Pos().LineColStr(), err))
		} else if condType != nil && !condType.Equals(BoolType) {
			errors = append(errors, fmt.Errorf("pos %s: if condition must be boolean, got %s", s.Condition.Pos().LineColStr(), condType.String()))
		}
		if s.Then != nil {
			errsThen := InferTypesForBlockStmt(s.Then, scope)
			errors = append(errors, errsThen...)
		}
		if s.Else != nil {
			errsElse := InferTypesForStmt(s.Else, scope)
			errors = append(errors, errsElse...)
		}
	case *BlockStmt:
		errs := InferTypesForBlockStmt(s, scope)
		errors = append(errors, errs...)
	case *AssignmentStmt: // General assignment, var must exist.
		existingVarType, varFound := scope.Get(s.Var.Name)
		if !varFound {
			errors = append(errors, fmt.Errorf("pos %s: assignment to undeclared variable '%s'", s.Var.Pos().LineColStr(), s.Var.Name))
			break
		}
		s.Var.SetInferredType(existingVarType)
		valType, err := InferExprType(s.Value, scope)
		if err != nil {
			errors = append(errors, fmt.Errorf("pos %s: error inferring type for assignment value to '%s': %w", s.Value.Pos().LineColStr(), s.Var.Name, err))
		} else if valType != nil {
			if !valType.Equals(existingVarType) {
				isIntToFloat := valType.Equals(IntType) && existingVarType.Equals(FloatType)
				if !isIntToFloat {
					errors = append(errors, fmt.Errorf("pos %s: type mismatch in assignment to '%s': variable is %s, value is %s", s.Pos().LineColStr(), s.Var.Name, existingVarType.String(), valType.String()))
				}
			}
		}
	case *ExpectStmt:
		errors = append(errors, fmt.Errorf("pos %s: ExpectStmt found outside of an analyze block's expectations; type checking skipped here", s.Pos().LineColStr()))
	default:
		// errors = append(errors, fmt.Errorf("pos %s: type inference for statement type %T not implemented yet", stmt.Pos().LineColStr(), stmt))
	}
	return errors
}

func InferTypesForSystemDeclBodyItem(item SystemDeclBodyItem, systemScope *TypeScope) []error {
	var errors []error
	switch i := item.(type) {
	case *InstanceDecl:
		compTypeNode, foundCompType := systemScope.env.Get(i.ComponentType.Name)
		if !foundCompType {
			errors = append(errors, fmt.Errorf("pos %s: component type '%s' not found for instance '%s'", i.ComponentType.Pos().LineColStr(), i.ComponentType.Name, i.NameNode.Name))
			return errors
		}
		compDefinition, ok := compTypeNode.(*ComponentDecl)
		if !ok {
			errors = append(errors, fmt.Errorf("pos %s: identifier '%s' used as component type for instance '%s' is not a component declaration (got %T)", i.ComponentType.Pos().LineColStr(), i.ComponentType.Name, i.NameNode.Name, compTypeNode))
			return errors
		}
		instanceType := ComponentTypeInstance(compDefinition)
		systemScope.env.Set(i.NameNode.Name, i) // Store InstanceDecl node in env by its name
		i.NameNode.SetInferredType(instanceType)

		for _, assign := range i.Overrides {
			valType, err := InferExprType(assign.Value, systemScope)
			if err != nil {
				errors = append(errors, fmt.Errorf("pos %s: error inferring type for override value of '%s' in instance '%s': %w", assign.Value.Pos().LineColStr(), assign.Var.Name, i.NameNode.Name, err))
				continue
			}
			if valType == nil {
				continue
			}

			paramDecl, _ := compDefinition.GetParam(assign.Var.Name)
			usesDecl, _ := compDefinition.GetDependency(assign.Var.Name)

			if paramDecl != nil {
				if paramDecl.Type == nil {
					errors = append(errors, fmt.Errorf("pos %s: param '%s' of component '%s' has no type", paramDecl.Pos().LineColStr(), paramDecl.Name.Name, compDefinition.NameNode.Name))
					continue
				}
				expectedType := paramDecl.Type.TypeUsingScope(systemScope) // Resolve TypeDecl in the current system scope
				if expectedType == nil {
					errors = append(errors, fmt.Errorf("pos %s: param '%s' of component '%s' has an unresolved TypeDecl '%s'", paramDecl.Type.Pos().LineColStr(), paramDecl.Name.Name, compDefinition.NameNode.Name, paramDecl.Type.Name))
					continue
				}
				if !valType.Equals(expectedType) {
					if !(valType.Equals(IntType) && expectedType.Equals(FloatType)) {
						errors = append(errors, fmt.Errorf("pos %s: type mismatch for override param '%s' in instance '%s': expected %s, got %s",
							assign.Value.Pos().LineColStr(), assign.Var.Name, i.NameNode.Name, expectedType.String(), valType.String()))
					}
				}
			} else if usesDecl != nil {
				expectedDepCompName := usesDecl.ComponentNode.Name
				var assignedInstanceType *Type
				if assignedIdent, isIdent := assign.Value.(*IdentifierExpr); isIdent {
					assignedNodeFromEnv, foundInstance := systemScope.env.Get(assignedIdent.Name)
					if foundInstance {
						if assignedInstDecl, isInst := assignedNodeFromEnv.(*InstanceDecl); isInst {
							// The type of an instance is its component type.
							// Retrieve the ComponentDecl for the assigned instance's component type.
							compTypeOfAssignedInstance, foundComp := systemScope.env.Get(assignedInstDecl.ComponentType.Name)
							if foundComp {
								if actualCompDecl, isActualComp := compTypeOfAssignedInstance.(*ComponentDecl); isActualComp {
									assignedInstanceType = ComponentTypeInstance(actualCompDecl)
								} else {
									errors = append(errors, fmt.Errorf("pos %s: assigned instance '%s' for dependency '%s' has non-component type %T for its ComponentType ('%s')", assign.Value.Pos().LineColStr(), assignedIdent.Name, assign.Var.Name, compTypeOfAssignedInstance, assignedInstDecl.ComponentType.Name))
								}
							} else {
								errors = append(errors, fmt.Errorf("pos %s: could not find component type '%s' for assigned instance '%s' (dependency '%s')", assign.Value.Pos().LineColStr(), assignedInstDecl.ComponentType.Name, assignedIdent.Name, assign.Var.Name))
							}
						} else {
							errors = append(errors, fmt.Errorf("pos %s: assigned value '%s' for dependency '%s' is not an instance declaration (got %T)", assign.Value.Pos().LineColStr(), assignedIdent.Name, assign.Var.Name, assignedNodeFromEnv))
						}
					} else {
						errors = append(errors, fmt.Errorf("pos %s: assigned instance identifier '%s' for dependency '%s' not found in system scope", assign.Value.Pos().LineColStr(), assignedIdent.Name, assign.Var.Name))
					}
				}

				if assignedInstanceType != nil {
					if assignedInstanceType.Name != expectedDepCompName {
						errors = append(errors, fmt.Errorf("pos %s: type mismatch for override dependency '%s' in instance '%s': expected instance of component type %s, got instance of %s",
							assign.Value.Pos().LineColStr(), assign.Var.Name, i.NameNode.Name, expectedDepCompName, assignedInstanceType.Name))
					}
				} else if valType.Name != expectedDepCompName || valType.OriginalDecl == nil || !valType.IsComponentType() {
					// This fallback is if assigned value was not an identifier resolving to an InstanceDecl
					errors = append(errors, fmt.Errorf("pos %s: type mismatch for override dependency '%s' in instance '%s': expected component type %s, got %s (and value was not a known instance of the correct type)",
						assign.Value.Pos().LineColStr(), assign.Var.Name, i.NameNode.Name, expectedDepCompName, valType.String()))
				}
			} else {
				errors = append(errors, fmt.Errorf("pos %s: override target '%s' in instance '%s' is not a known parameter or dependency of component '%s'",
					assign.Var.Pos().LineColStr(), assign.Var.Name, i.NameNode.Name, compDefinition.NameNode.Name))
			}
		}
		/*
			case *AnalyzeDecl:
				targetType, err := InferExprType(i.Target, systemScope)
				if err != nil {
					errors = append(errors, fmt.Errorf("pos %s: error inferring type for analyze target '%s': %w", i.Target.Pos().LineColStr(), i.Name.Name, err))
				}
				if targetType != nil {
					i.Name.SetInferredType(targetType)
					if i.Expectations != nil {
						expectScope := systemScope.Push(nil, nil)
						if errSet := expectScope.Set(i.Name.Name, i.Name, targetType); errSet != nil {
							errors = append(errors, fmt.Errorf("internal error setting analyze result name '%s' in expect scope: %w", i.Name.Name, errSet))
						}

						for _, expectStmt := range i.Expectations.Expects {
							if expectStmt.Target == nil {
								errors = append(errors, fmt.Errorf("pos %s: expect statement is missing a target metric expression", expectStmt.Pos().LineColStr()))
								continue
							}

							// Temporarily ensure the receiver of the expectStmt.Target (e.g., "Result" in "Result.P99")
							// is correctly typed with targetType before inferring the full expectStmt.Target.
							if mae, ok := expectStmt.Target.(*MemberAccessExpr); ok {
								if recvIdent, isIdent := mae.Receiver.(*IdentifierExpr); isIdent {
									if recvIdent.Name == i.Name.Name {
										// The Get method in TypeScope during InferExprType(mae.Receiver, expectScope)
										// should find i.Name in expectScope.env and return targetType.
										// So, explicit SetInferredType might not be needed here if Get works.
									} else {
										errors = append(errors, fmt.Errorf("pos %s: expect metric target receiver '%s' must match analyze block name '%s'", recvIdent.Pos().LineColStr(), recvIdent.Name, i.Name.Name))
									}
								} else {
									errors = append(errors, fmt.Errorf("pos %s: expect metric target receiver must be a simple identifier matching analyze block name ('%s')", mae.Receiver.Pos().LineColStr(), i.Name.Name))
								}
							}

							metricType, errMetric := InferExprType(expectStmt.Target, expectScope)
							if errMetric != nil {
								errors = append(errors, fmt.Errorf("pos %s: error inferring type for expect metric '%s' in analyze block '%s': %w", expectStmt.Target.Pos().LineColStr(), expectStmt.Target.String(), i.Name.Name, errMetric))
							}
							if expectStmt.Threshold == nil {
								errors = append(errors, fmt.Errorf("pos %s: expect statement is missing a threshold expression", expectStmt.Pos().LineColStr()))
								continue
							}
							thresholdType, errThreshold := InferExprType(expectStmt.Threshold, expectScope)
							if errThreshold != nil {
								errors = append(errors, fmt.Errorf("pos %s: error inferring type for expect threshold in analyze block '%s': %w", expectStmt.Threshold.Pos().LineColStr(), i.Name.Name, errThreshold))
							}

							if metricType != nil && thresholdType != nil {
								isMetricNumeric := metricType.Equals(IntType) || metricType.Equals(FloatType)
								isThresholdNumeric := thresholdType.Equals(IntType) || thresholdType.Equals(FloatType)
								isMetricBool := metricType.Equals(BoolType)
								isThresholdBool := thresholdType.Equals(BoolType)
								validComparison := false
								if isMetricNumeric && isThresholdNumeric {
									validComparison = true
								}
								if isMetricBool && isThresholdBool {
									validComparison = true
								}
								if !validComparison {
									errors = append(errors, fmt.Errorf("pos %s: type mismatch in expect statement for analyze block '%s': cannot compare metric type %s with threshold type %s using operator '%s'",
										expectStmt.Pos().LineColStr(), i.Name.Name, metricType.String(), thresholdType.String(), expectStmt.Operator))
								}
							}
						}
					}
				}
		*/
	case *LetStmt:
		errs := InferTypesForStmt(i, systemScope)
		errors = append(errors, errs...)
	case *OptionsDecl:
		if i.Body != nil {
			optionsScope := systemScope.Push(nil, nil)
			errs := InferTypesForBlockStmt(i.Body, optionsScope)
			errors = append(errors, errs...)
		}
	default:
		// errors = append(errors, fmt.Errorf("pos %s: type inference for system body item type %T not implemented yet", item.Pos().LineColStr(), item))
	}
	return errors
}

// InferTypesForFile is the main entry point.
// It now takes the rootEnv populated by the loader.
func InferTypesForFile(file *FileDecl, rootEnv *Env[Node]) []error {
	var errors []error
	if file == nil {
		errors = append(errors, fmt.Errorf("cannot infer types for nil FileDecl"))
		return errors
	}

	if !file.resolved { // Use the 'resolved' field
		if err := file.Resolve(); err != nil {
			errors = append(errors, fmt.Errorf("error resolving file before type inference (pos %s): %w", file.Pos().LineColStr(), err))
			return errors
		}
	}

	rootScope := NewRootTypeScope(rootEnv)

	components, err := file.GetComponents()
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting components for inference (pos %s): %w", file.Pos().LineColStr(), err))
		return errors
	}
	systems, err := file.GetSystems()
	if err != nil {
		errors = append(errors, fmt.Errorf("error getting systems for inference (pos %s): %w", file.Pos().LineColStr(), err))
		return errors
	}

	// First pass: Resolve TypeDecls in component parameter defaults, method parameters, and method return types.
	for _, compDecl := range components {
		// Parameter defaults
		errs := InferTypesForComponent(file, rootScope, compDecl)
		errors = append(errors, errs...)
	}

	// Second pass: Infer types for component method bodies (now that signatures are resolved)
	for _, compDecl := range components {
		for _, method := range compDecl.methods {
			// methodScope needs to see 'self', method params, component params/uses, and globals/imports via rootEnv
			// Parameters are added to scope implicitly by TypeScope.Get looking at currentMethod.
			// 'uses' are also resolved via 'self' access within TypeScope.Get if needed.
			if method.Body != nil {
				methodScope := rootScope.Push(compDecl, method)
				blockErrs := InferTypesForBlockStmt(method.Body, methodScope)
				errors = append(errors, blockErrs...)
			}
		}
	}

	// Third pass: Infer types for system declarations
	for _, sysDecl := range systems {
		systemScope := rootScope.Push(nil, nil) // System scope can see globals/imports from rootEnv
		for _, item := range sysDecl.Body {
			itemErrs := InferTypesForSystemDeclBodyItem(item, systemScope)
			errors = append(errors, itemErrs...)
		}
	}
	return errors
}

func InferTypesForComponent(file *FileDecl, rootScope *TypeScope, compDecl *ComponentDecl) (errors []error) {
	for _, paramDecl := range compDecl.params { // Assuming direct field access or appropriate getter
		errs := InferTypesForParamDecl(file, rootScope, compDecl, paramDecl)
		errors = append(errors, errs...)
	}

	// Method signatures
	for _, method := range compDecl.methods { // Assuming direct field access or appropriate getter
		errs := InferTypesForMethodSignature(file, rootScope, compDecl, method)
		errors = append(errors, errs...)
	}
	return
}

func InferTypesForParamDecl(file *FileDecl, rootScope *TypeScope, compDecl *ComponentDecl, paramDecl *ParamDecl) (errors []error) {
	var resolvedParamType *Type

	if paramDecl.Type != nil { // Type is explicitly declared
		resolvedParamType = paramDecl.Type.TypeUsingScope(rootScope)
		if resolvedParamType == nil {
			errors = append(errors, fmt.Errorf("pos %s: unresolved type '%s' for parameter '%s' in component '%s'", paramDecl.Type.Pos().LineColStr(), paramDecl.Type.Name, paramDecl.Name.Name, compDecl.NameNode.Name))
			// Even if unresolved, continue to check default value if present, but this param is problematic.
		} else {
			paramDecl.Type.SetResolvedType(resolvedParamType)
		}
	} else if paramDecl.DefaultValue != nil { // No explicit type, but has default value
		// Infer type from default value
		defaultValueType, err_dv := InferExprType(paramDecl.DefaultValue, rootScope)
		if err_dv != nil {
			errors = append(errors, fmt.Errorf("pos %s: error inferring type from default value for parameter '%s' in component '%s': %w", paramDecl.DefaultValue.Pos().LineColStr(), paramDecl.Name.Name, compDecl.NameNode.Name, err_dv))
		} else if defaultValueType != nil {
			resolvedParamType = defaultValueType
			// We can create a new TypeDecl node here representing the inferred type,
			// or simply store the *Type on the paramDecl if it has a field for that.
			// For now, let's assume we want to ensure paramDecl.Type is populated if possible.
			// This is a bit tricky as TypeDecl usually comes from parsing.
			// A simpler approach for now: the effective type is known (defaultValueType).
			// If the paramDecl.Type field *must* be a TypeDecl, we might need to synthesize one.
			// Let's assume for now that the primary goal is that `resolvedParamType` is set.
			// And if we needed to set it back on paramDecl.Type, we'd need a way to create a TypeDecl from a Type.
			// To keep it simple and consistent with SetResolvedType on TypeDecl:
			// We could add a field `InferredOrDeclaredType *Type` to ParamDecl itself.
			// Or, if paramDecl.Type *must* be non-nil for later stages, this is an issue.

			// For now, we have `resolvedParamType`. If `paramDecl.Type` was nil, it remains nil,
			// but the inference process now knows the type from the default value.
			// Let's ensure the IdentifierExpr for the param name gets typed.
			paramDecl.Name.SetInferredType(resolvedParamType)

		} else {
			// Default value's type could not be inferred.
			errors = append(errors, fmt.Errorf("pos %s: could not infer type from default value for parameter '%s' in component '%s'", paramDecl.DefaultValue.Pos().LineColStr(), paramDecl.Name.Name, compDecl.NameNode.Name))
		}
	} else { // No explicit type AND no default value
		errors = append(errors, fmt.Errorf("pos %s: parameter '%s' in component '%s' has no type declaration and no default value", paramDecl.Pos().LineColStr(), paramDecl.Name.Name, compDecl.NameNode.Name))
		return // Cannot proceed with this parameter
	}

	// If a default value exists, check its type against the (now hopefully resolved) parameter type.
	if paramDecl.DefaultValue != nil {
		defaultValueActualType, err_dv_check := InferExprType(paramDecl.DefaultValue, rootScope)
		if err_dv_check != nil {
			// Error already added if type inference for default value failed earlier.
			// This is a redundant call if it already failed, but harmless.
			// If it succeeded before but fails now, that's an issue.
			// To avoid double error: only add if not already present from initial inference.
			// However, InferExprType caches, so it won't re-infer.
		} else if defaultValueActualType != nil {
			// Now, `resolvedParamType` should hold the type of the parameter,
			// either from its TypeDecl or inferred from the default value itself.
			if resolvedParamType != nil { // If we have an expected type for the param
				if !defaultValueActualType.Equals(resolvedParamType) {
					// Allow int to float promotion for default value
					isPromotion := defaultValueActualType.Equals(IntType) && resolvedParamType.Equals(FloatType)
					if !isPromotion {
						errors = append(errors, fmt.Errorf("pos %s: type mismatch for default value of parameter '%s' in component '%s': parameter type is %s, default value type is %s", paramDecl.DefaultValue.Pos().LineColStr(), paramDecl.Name.Name, compDecl.NameNode.Name, resolvedParamType.String(), defaultValueActualType.String()))
					}
				}
			} else if paramDecl.Type != nil {
				// This means paramDecl.Type was specified but couldn't be resolved earlier,
				// yet we have a default value. This is an inconsistent state.
				// The earlier error about unresolved type for paramDecl.Type should cover this.
			}
		}
	}
	return
}

func InferTypesForMethodSignature(file *FileDecl, rootScope *TypeScope, compDecl *ComponentDecl, method *MethodDecl) (errors []error) {
	for _, param := range method.Parameters {
		if param.Type != nil {
			resolvedParamType := param.Type.TypeUsingScope(rootScope)
			if resolvedParamType == nil {
				errors = append(errors, fmt.Errorf("pos %s: unresolved type '%s' for parameter '%s' in method '%s.%s'", param.Type.Pos().LineColStr(), param.Type.Name, param.Name.Name, compDecl.NameNode.Name, method.NameNode.Name))
			} else {
				param.Type.SetResolvedType(resolvedParamType)
			}
		} else {
			errors = append(errors, fmt.Errorf("pos %s: parameter '%s' of method '%s.%s' has no type declaration", param.Pos().LineColStr(), param.Name.Name, compDecl.NameNode.Name, method.NameNode.Name))
		}
	}
	if method.ReturnType != nil {
		resolvedReturnType := method.ReturnType.TypeUsingScope(rootScope)
		if resolvedReturnType == nil {
			errors = append(errors, fmt.Errorf("pos %s: unresolved return type '%s' for method '%s.%s'", method.ReturnType.Pos().LineColStr(), method.ReturnType.Name, compDecl.NameNode.Name, method.NameNode.Name))
		} else {
			method.ReturnType.SetResolvedType(resolvedReturnType)
		}
	}
	return
}
