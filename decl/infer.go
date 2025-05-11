package decl

import (
	"fmt"
	"strings"
	// "log" // For debugging
)

// TypeScope manages type information for identifiers within a scope.
type TypeScope struct {
	store            map[string]*Type
	outer            *TypeScope
	file             *FileDecl // Access to global declarations (enums, components)
	currentComponent *ComponentDecl
	currentMethod    *MethodDecl // Useful for 'self' and method params/return
}

// NewRootTypeScope creates a top-level scope for a file.
func NewRootTypeScope(file *FileDecl) *TypeScope {
	return &TypeScope{
		store: make(map[string]*Type),
		outer: nil,
		file:  file,
	}
}

// Push creates a new nested scope.
func (ts *TypeScope) Push(currentComponent *ComponentDecl, currentMethod *MethodDecl) *TypeScope {
	// Inherit component/method context if not overridden
	effectiveCurrentComponent := ts.currentComponent
	if currentComponent != nil {
		effectiveCurrentComponent = currentComponent
	}
	effectiveCurrentMethod := ts.currentMethod
	if currentMethod != nil {
		effectiveCurrentMethod = currentMethod
	}

	return &TypeScope{
		store:            make(map[string]*Type),
		outer:            ts,
		file:             ts.file, // Propagate file
		currentComponent: effectiveCurrentComponent,
		currentMethod:    effectiveCurrentMethod,
	}
}

// Get retrieves the type of an identifier.
// Order: local store -> 'self' -> method params -> outer scopes -> file globals.
func (ts *TypeScope) Get(name string) (*Type, bool) {
	// 1. Local store
	if t, ok := ts.store[name]; ok {
		return t, true
	}

	// 2. 'self' keyword (if in a method context of a component)
	if ts.currentComponent != nil && name == "self" {
		return &Type{Name: ts.currentComponent.NameNode.Name /* IsComponent: true */}, true
	}

	// 3. Method parameters (if in a method context)
	if ts.currentMethod != nil {
		for _, param := range ts.currentMethod.Parameters {
			if param.Name.Name == name {
				if param.Type == nil { // Should be caught by an earlier validation pass
					return nil, false // Or return an error type marker
				}
				return param.Type.Type(), true
			}
		}
	}
	// Note: Component-level params and 'uses' deps are typically accessed via 'self.member',
	// which is handled by MemberAccessExpr inference, not direct lookup here.

	// 4. Outer scopes
	if ts.outer != nil {
		return ts.outer.Get(name)
	}

	// 5. File-level declarations (Enums, Components as type names, Instances in SystemDecl)
	if ts.file != nil {
		if enumDecl, _ := ts.file.GetEnum(name); enumDecl != nil {
			// This identifier refers to an enum type itself
			return &Type{Name: enumDecl.NameNode.Name, IsEnum: true}, true
		}
		// An identifier might refer to a component type (e.g. in `instance x: MyComponent;`)
		// if compDecl, _ := ts.file.GetComponent(name); compDecl != nil {
		//    return &Type{Name: compDecl.NameNode.Name /* IsComponentType: true */}, true
		// }
		// Note: Instance names declared in SystemDecl are added to the scope by inferTypesForSystemDeclBodyItem.
	}

	return nil, false
}

// Set defines the type of an identifier in the current local scope.
func (ts *TypeScope) Set(name string, t *Type) {
	// log.Printf("Scope (comp: %s, meth: %s): Setting type for '%s' to %s", ts.currentComponent != nil, ts.currentMethod != nil, name, t)
	ts.store[name] = t
}

// InferExprType recursively infers the type of an expression and sets its InferredType field.
func InferExprType(expr Expr, scope *TypeScope) (*Type, error) {
	if expr == nil {
		return nil, fmt.Errorf("cannot infer type for nil expression")
	}

	// Prevent re-inference if already done
	if expr.InferredType() != nil {
		return expr.InferredType(), nil
	}

	var inferred *Type
	var err error
	// var err error // already declared

	switch e := expr.(type) {
	case *LiteralExpr:
		inferred, err = inferLiteralExprType(e, scope)
	case *IdentifierExpr:
		inferred, err = inferIdentifierExprType(e, scope)
	case *BinaryExpr:
		inferred, err = inferBinaryExprType(e, scope)
	case *UnaryExpr:
		inferred, err = inferUnaryExprType(e, scope)
	case *MemberAccessExpr:
		inferred, err = inferMemberAccessExprType(e, scope)
	case *CallExpr:
		inferred, err = inferCallExprType(e, scope)
	case *TupleExpr:
		inferred, err = inferTupleExprType(e, scope)
	case *ChainedExpr:
		inferred, err = inferChainedExprType(e, scope)
	case *DistributeExpr:
		inferred, err = inferDistributeExprType(e, scope)
	case *SampleExpr:
		inferred, err = inferSampleExprType(e, scope)
	case *CaseExpr:
		if e.Body == nil {
			return nil, fmt.Errorf("CaseExpr at pos %d has no body", e.Pos())
		}
		inferred, err = InferExprType(e.Body, scope)
	default:
		// This case should have been caught by getExprBase, but as a safeguard:
		return nil, fmt.Errorf("type inference not implemented for expression type %T at pos %d", expr, expr.Pos())
	}

	if err != nil {
		return nil, err
	}

	if inferred == nil { // Should not happen if logic is correct (either type or error)
		return nil, fmt.Errorf("type inference failed to determine type for %T at pos %d, but no error reported (inferred is nil)", expr, expr.Pos())
	}

	expr.SetInferredType(inferred)

	// Optional: Check against DeclaredType if present and populated
	if expr.DeclaredType() != nil && !expr.DeclaredType().Equals(inferred) {
		// Allow int to float promotion if declared is float and inferred is int
		isIntToFloatPromotion := expr.DeclaredType().Equals(FloatType) && inferred.Equals(IntType)
		if !isIntToFloatPromotion {
			return inferred, fmt.Errorf("type mismatch at pos %d for '%s': inferred type %s, but declared type is %s",
				expr.Pos(), expr.String(), inferred.String(), expr.DeclaredType().String())
		}
	}
	// log.Printf("Expr: %s (pos %d), InferredType: %s", expr.String(), expr.Pos(), inferred)
	return inferred, nil
}

func inferLiteralExprType(expr *LiteralExpr, scope *TypeScope) (*Type, error) {
	if expr.Value == nil || expr.Value.Type == nil {
		return nil, fmt.Errorf("literal expression at pos %d has invalid internal RuntimeValue or Type", expr.Pos())
	}
	return expr.Value.Type, nil
}

func inferIdentifierExprType(expr *IdentifierExpr, scope *TypeScope) (*Type, error) {
	t, ok := scope.Get(expr.Name)
	if !ok {
		return nil, fmt.Errorf("identifier '%s' not found at pos %d", expr.Name, expr.Pos())
	}
	return t, nil
}

func inferBinaryExprType(expr *BinaryExpr, scope *TypeScope) (*Type, error) {
	leftType, err := InferExprType(expr.Left, scope)
	if err != nil {
		return nil, fmt.Errorf("error inferring type for left operand of binary expr ('%s') at pos %d: %w", expr.Operator, expr.Left.Pos(), err)
	}
	rightType, err := InferExprType(expr.Right, scope)
	if err != nil {
		return nil, fmt.Errorf("error inferring type for right operand of binary expr ('%s') at pos %d: %w", expr.Operator, expr.Right.Pos(), err)
	}
	if leftType == nil || rightType == nil {
		return nil, fmt.Errorf("could not determine type for one or both operands for binary expr ('%s') at pos %d", expr.Operator, expr.Pos())
	}

	switch expr.Operator {
	case "+", "-", "*", "/":
		if leftType.Equals(IntType) && rightType.Equals(IntType) {
			return IntType, nil
		}
		if (leftType.Equals(IntType) || leftType.Equals(FloatType)) &&
			(rightType.Equals(IntType) || rightType.Equals(FloatType)) {
			return FloatType, nil // Promote to float
		}
		// String concatenation
		if expr.Operator == "+" && leftType.Equals(StrType) && rightType.Equals(StrType) {
			return StrType, nil
		}
		return nil, fmt.Errorf("type mismatch for operator '%s' at pos %d: cannot apply to %s and %s",
			expr.Operator, expr.Pos(), leftType.String(), rightType.String())
	case "%":
		if leftType.Equals(IntType) && rightType.Equals(IntType) {
			return IntType, nil
		}
		return nil, fmt.Errorf("type mismatch for operator '%%' at pos %d: requires two integers, got %s and %s",
			expr.Pos(), leftType.String(), rightType.String())
	case "==", "!=", "<", "<=", ">", ">=":
		isLeftNumeric := leftType.Equals(IntType) || leftType.Equals(FloatType)
		isRightNumeric := rightType.Equals(IntType) || rightType.Equals(FloatType)

		if isLeftNumeric && isRightNumeric { // Comparing two numbers
			return BoolType, nil
		}
		if leftType.Equals(rightType) { // Comparing two values of the same type
			// Disallow comparison for complex types unless specifically defined
			if leftType.Name == "List" || leftType.Name == "Tuple" || leftType.Name == "Outcomes" ||
				leftType.Name == "Component" || leftType.Name == "OpNode" { // Extend OpNode if necessary
				return nil, fmt.Errorf("type mismatch for comparison operator '%s' at pos %d: cannot compare complex type %s", expr.Operator, expr.Pos(), leftType.String())
			}
			return BoolType, nil
		}
		return nil, fmt.Errorf("type mismatch for comparison operator '%s' at pos %d: cannot compare %s and %s",
			expr.Operator, expr.Pos(), leftType.String(), rightType.String())
	case "&&", "||":
		if leftType.Equals(BoolType) && rightType.Equals(BoolType) {
			return BoolType, nil
		}
		return nil, fmt.Errorf("type mismatch for logical operator '%s' at pos %d: requires two booleans, got %s and %s",
			expr.Operator, expr.Pos(), leftType.String(), rightType.String())
	default:
		return nil, fmt.Errorf("unsupported binary operator '%s' at pos %d", expr.Operator, expr.Pos())
	}
}

func inferUnaryExprType(expr *UnaryExpr, scope *TypeScope) (*Type, error) {
	rightType, err := InferExprType(expr.Right, scope)
	if err != nil {
		return nil, fmt.Errorf("error inferring type for operand of unary expr ('%s') at pos %d: %w", expr.Operator, expr.Right.Pos(), err)
	}
	if rightType == nil {
		return nil, fmt.Errorf("could not determine type for operand of unary expr ('%s') at pos %d", expr.Operator, expr.Pos())
	}

	switch expr.Operator {
	case "!":
		if !rightType.Equals(BoolType) {
			return nil, fmt.Errorf("type mismatch for operator '!' at pos %d: requires boolean, got %s",
				expr.Pos(), rightType.String())
		}
		return BoolType, nil
	case "-":
		if rightType.Equals(IntType) || rightType.Equals(FloatType) {
			return rightType, nil
		}
		return nil, fmt.Errorf("type mismatch for operator '-' at pos %d: requires integer or float, got %s",
			expr.Pos(), rightType.String())
	default:
		return nil, fmt.Errorf("unsupported unary operator '%s' at pos %d", expr.Operator, expr.Pos())
	}
}

func inferMemberAccessExprType(expr *MemberAccessExpr, scope *TypeScope) (*Type, error) {
	receiverType, err := InferExprType(expr.Receiver, scope)
	if err != nil {
		return nil, fmt.Errorf("error inferring type for receiver of member access '%s' at pos %d: %w", expr.Member.Name, expr.Receiver.Pos(), err)
	}
	if receiverType == nil {
		return nil, fmt.Errorf("could not determine type for receiver of member access '%s' at pos %d", expr.Member.Name, expr.Pos())
	}
	memberName := expr.Member.Name

	if scope.file != nil {
		// Case 1: Receiver is a Component instance
		compDecl, _ := scope.file.GetComponent(receiverType.Name) // Assumes receiverType.Name is component name
		if compDecl != nil {
			// Check parameters
			if paramDecl, _ := compDecl.GetParam(memberName); paramDecl != nil {
				if paramDecl.Type == nil {
					return nil, fmt.Errorf("parameter '%s' in component '%s' at pos %d has no type declaration", memberName, compDecl.NameNode.Name, paramDecl.Pos())
				}
				return paramDecl.Type.Type(), nil
			}
			// Check methods. If member access refers to a method name itself (not a call).
			// This means the expression `obj.method` evaluates to a function/method type.
			if methodDecl, _ := compDecl.GetMethod(memberName); methodDecl != nil {
				// TODO: Define a FunctionType in types.go if methods can be first-class.
				// For now, returning a placeholder or error if not immediately called.
				// Placeholder: Type{Name: "Function", IsMethod: true, MethodSignature: methodDecl}
				// For simplicity, let's say direct access to a method name like this is for CallExpr to consume.
				// If this MemberAccessExpr is the Function part of a CallExpr, CallExpr handles it.
				// If used elsewhere, its type is more complex (function type).
				// For now, let's assume this is an error or requires a proper FunctionType.
				// Returning a generic "MethodReference" type might be an option.
				// For this exercise, we'll assume it's primarily for param access or within CallExpr.
				// If it's *not* within a CallExpr, this path might mean it's an error or needs specific FunctionType.
				return nil, fmt.Errorf("direct access to method '%s' on component '%s' at pos %d yields a function; its type is complex or it must be called", memberName, compDecl.NameNode.Name, expr.Pos())
			}
			return nil, fmt.Errorf("member '%s' not found in component '%s' at pos %d", memberName, compDecl.NameNode.Name, expr.Pos())
		}

		// Case 2: Receiver is an Enum type
		if receiverType.IsEnum {
			enumDecl, _ := scope.file.GetEnum(receiverType.Name)
			if enumDecl != nil {
				for _, valNode := range enumDecl.ValuesNode {
					if valNode.Name == memberName {
						return receiverType, nil // Accessing an enum value results in the enum type itself
					}
				}
				return nil, fmt.Errorf("value '%s' not found in enum '%s' at pos %d", memberName, enumDecl.NameNode.Name, expr.Pos())
			}
		}
	}

	// Case 3: Accessing metrics on an analysis result (heuristic)
	// e.g., if receiverType is what an AnalyzeDecl's target evaluates to.
	// Typically, AnalyzeDecl.Target is a CallExpr. Its type might be Outcomes[AccessResult] or Outcomes[RangedResult].
	if strings.HasPrefix(receiverType.Name, "Outcomes") {
		// Example: check for ".Availability", ".MeanLatency", etc.
		// These correspond to keys in AnalysisResultWrapper.Metrics, which are float64.
		// This mapping relies on core.MetricType definitions and their string representations.
		// For simplicity, we list common metric names. A more robust solution would inspect Metricable.
		knownMetrics := map[string]bool{
			"AvailabilityMetric": true, "MeanLatencyMetric": true, "P50LatencyMetric": true,
			"P99LatencyMetric": true, "P999LatencyMetric": true,
			// Short names if your DSL uses them, e.g., from core.MetricType.String()
			"availability": true, "mean_latency": true, "p50_latency": true, "p99_latency": true, "p999_latency": true,
		}
		if _, ok := knownMetrics[memberName]; ok {
			return FloatType, nil
		}
		// Potentially `.Len()` on Outcomes if it acts like a collection
		if memberName == "Len" {
			return IntType, nil
		}

	}
	// Case 4: .Len on List
	if receiverType.Name == "List" && memberName == "Len" {
		return IntType, nil
	}

	return nil, fmt.Errorf("cannot access member '%s' on type %s at pos %d; not a known component, enum, or collection type with this member",
		memberName, receiverType.String(), expr.Pos())
}

func inferCallExprType(expr *CallExpr, scope *TypeScope) (*Type, error) {
	var returnType *Type = NilType // Default to NilType if not void
	var expectedParamTypes []*Type
	var funcNameForError string

	switch fn := expr.Function.(type) {
	case *MemberAccessExpr:
		funcNameForError = fmt.Sprintf("%s.%s", fn.Receiver.String(), fn.Member.Name)
		receiverType, err := InferExprType(fn.Receiver, scope)
		if err != nil {
			return nil, fmt.Errorf("error inferring receiver type for method call '%s' at pos %d: %w", funcNameForError, fn.Receiver.Pos(), err)
		}
		if receiverType == nil {
			return nil, fmt.Errorf("could not determine receiver type for method call '%s' at pos %d", funcNameForError, fn.Pos())
		}

		compDecl, _ := scope.file.GetComponent(receiverType.Name)
		if compDecl != nil { // Calling a method on a component
			methodDecl, err := compDecl.GetMethod(fn.Member.Name)
			if err != nil {
				return nil, fmt.Errorf("method '%s' not found in component '%s' at pos %d: %w", fn.Member.Name, compDecl.NameNode.Name, fn.Member.Pos(), err)
			}
			if methodDecl.ReturnType != nil {
				if methodDecl.ReturnType.Type() == nil { // Should not happen if TypeDecl is valid
					return nil, fmt.Errorf("method '%s.%s' return TypeDecl at pos %d did not resolve to a Type", compDecl.NameNode.Name, methodDecl.NameNode.Name, methodDecl.ReturnType.Pos())
				}
				returnType = methodDecl.ReturnType.Type()
			} // else void, so NilType is correct default

			for _, paramDecl := range methodDecl.Parameters {
				if paramDecl.Type == nil || paramDecl.Type.Type() == nil {
					return nil, fmt.Errorf("parameter '%s' of method '%s.%s' has invalid type declaration at pos %d", paramDecl.Name.Name, compDecl.NameNode.Name, methodDecl.NameNode.Name, paramDecl.Pos())
				}
				expectedParamTypes = append(expectedParamTypes, paramDecl.Type.Type())
			}
		} else if strings.HasPrefix(receiverType.Name, "Outcomes") || receiverType.Name == "List" {
			// Built-in methods on Outcomes/List e.g. .Len() (already handled by MemberAccess)
			// If .Filter(lambda) or .Map(lambda) existed, their types would be complex.
			// For now, assume .Len() is the main one, and it's 0-arity, already returns IntType via MemberAccess.
			// If CallExpr encounters .Len(), it means MemberAccessExpr for .Len should have returned FunctionType.
			// This section needs more fleshing out if such methods exist.
			// For now, if MemberAccess didn't error and Call is on it, it implies a function type was returned.
			// This means MemberAccess needs to be enhanced for this. Let's assume it's not yet.
			if fn.Member.Name == "Len" { // If .Len() call
				if len(expr.Args) != 0 {
					return nil, fmt.Errorf("method '%s.Len' at pos %d expects 0 arguments, got %d", receiverType.Name, expr.Pos(), len(expr.Args))
				}
				return IntType, nil // Len() returns int
			}
			return nil, fmt.Errorf("calling method '%s' on non-component type %s at pos %d is not supported or method unknown", fn.Member.Name, receiverType.String(), fn.Pos())

		} else {
			return nil, fmt.Errorf("receiver for method call '%s' is not a known component type (%s) at pos %d", fn.Member.Name, receiverType.String(), fn.Pos())
		}

	case *IdentifierExpr:
		funcNameForError = fn.Name
		// TODO: Handle built-in global functions (need a registry)
		// Example:
		// if sig, ok := builtInGlobalFunctions[fn.Name]; ok {
		//    returnType = sig.ReturnType
		//    expectedParamTypes = sig.ParamTypes
		// } else {
		return nil, fmt.Errorf("type inference for standalone function call '%s' at pos %d not implemented yet", fn.Name, fn.Pos())
		// }
	default:
		return nil, fmt.Errorf("invalid function/method expression type %T in call at pos %d", expr.Function, expr.Function.Pos())
	}

	// Check arguments
	if len(expr.Args) != len(expectedParamTypes) {
		return nil, fmt.Errorf("argument count mismatch for call to '%s' at pos %d: expected %d, got %d",
			funcNameForError, expr.Pos(), len(expectedParamTypes), len(expr.Args))
	}

	for i, argExpr := range expr.Args {
		argType, err := InferExprType(argExpr, scope)
		if err != nil {
			return nil, fmt.Errorf("error inferring type for argument %d of call to '%s' at pos %d: %w",
				i, funcNameForError, argExpr.Pos(), err)
		}
		if argType == nil {
			return nil, fmt.Errorf("could not determine type for argument %d of call to '%s' at pos %d", i, funcNameForError, argExpr.Pos())
		}
		if !argType.Equals(expectedParamTypes[i]) {
			// Allow int to float promotion for args
			isIntToFloat := argType.Equals(IntType) && expectedParamTypes[i].Equals(FloatType)
			if !isIntToFloat {
				return nil, fmt.Errorf("type mismatch for argument %d of call to '%s' at pos %d: expected %s, got %s",
					i, funcNameForError, argExpr.Pos(), expectedParamTypes[i].String(), argType.String())
			}
		}
	}

	return returnType, nil
}

func inferTupleExprType(expr *TupleExpr, scope *TypeScope) (*Type, error) {
	if len(expr.Children) == 0 {
		// Or should this be an error? An empty tuple type might be valid.
		// Let's assume TupleType requires at least one element as per types.go panic.
		return nil, fmt.Errorf("tuple expression at pos %d must have at least one child", expr.Pos())
	}
	childTypes := make([]*Type, len(expr.Children))
	for i, childExpr := range expr.Children {
		childType, err := InferExprType(childExpr, scope)
		if err != nil {
			return nil, fmt.Errorf("error inferring type for tuple element %d at pos %d: %w", i, childExpr.Pos(), err)
		}
		if childType == nil {
			return nil, fmt.Errorf("could not determine type for tuple element %d at pos %d", i, childExpr.Pos())
		}
		childTypes[i] = childType
	}
	return TupleType(childTypes...), nil
}

func inferChainedExprType(expr *ChainedExpr, scope *TypeScope) (*Type, error) {
	if len(expr.Children) == 0 {
		return nil, fmt.Errorf("chained expression at pos %d has no children", expr.Pos())
	}
	if len(expr.Children) != len(expr.Operators)+1 {
		return nil, fmt.Errorf("malformed chained expression at pos %d: %d children, %d operators", expr.Pos(), len(expr.Children), len(expr.Operators))
	}

	currentType, err := InferExprType(expr.Children[0], scope)
	if err != nil {
		return nil, fmt.Errorf("error inferring type for first element of chained expr at pos %d: %w", expr.Children[0].Pos(), err)
	}

	for i, operator := range expr.Operators {
		rightOperandExpr := expr.Children[i+1]
		rightType, err := InferExprType(rightOperandExpr, scope)
		if err != nil {
			return nil, fmt.Errorf("error inferring type for operand %d of chained expr (op '%s') at pos %d: %w", i+1, operator, rightOperandExpr.Pos(), err)
		}
		if currentType == nil || rightType == nil {
			return nil, fmt.Errorf("could not determine type for one or both operands for operator '%s' in chained expr at pos %d", operator, expr.Pos())
		}

		// Simulate a binary expression for type checking currentType op rightType
		// This reuses the logic from inferBinaryExprType's switch statement.
		// This is a simplified model assuming left-associativity.
		tempBinaryExpr := &BinaryExpr{Left: nil, Operator: operator, Right: nil} // Dummy for operator classification
		switch tempBinaryExpr.Operator {                                         // Use operator directly
		case "+", "-", "*", "/":
			if currentType.Equals(IntType) && rightType.Equals(IntType) {
				currentType = IntType
			} else if (currentType.Equals(IntType) || currentType.Equals(FloatType)) &&
				(rightType.Equals(IntType) || rightType.Equals(FloatType)) {
				currentType = FloatType
			} else if operator == "+" && currentType.Equals(StrType) && rightType.Equals(StrType) {
				currentType = StrType
			} else {
				return nil, fmt.Errorf("type mismatch for operator '%s' in chained expr at pos %d: cannot apply to %s and %s",
					operator, expr.Pos(), currentType.String(), rightType.String())
			}
		case "%":
			if currentType.Equals(IntType) && rightType.Equals(IntType) {
				currentType = IntType
			} else {
				return nil, fmt.Errorf("type mismatch for operator '%%' in chained expr at pos %d: requires two integers, got %s and %s",
					expr.Pos(), currentType.String(), rightType.String())
			}
		case "==", "!=", "<", "<=", ">", ">=":
			isLeftNumeric := currentType.Equals(IntType) || currentType.Equals(FloatType)
			isRightNumeric := rightType.Equals(IntType) || rightType.Equals(FloatType)
			if isLeftNumeric && isRightNumeric {
				currentType = BoolType
			} else if currentType.Equals(rightType) {
				if currentType.Name == "List" || currentType.Name == "Tuple" || currentType.Name == "Outcomes" || currentType.Name == "Component" || currentType.Name == "OpNode" {
					return nil, fmt.Errorf("type mismatch for comparison operator '%s' in chained expr at pos %d: cannot compare complex type %s", operator, expr.Pos(), currentType.String())
				}
				currentType = BoolType
			} else {
				return nil, fmt.Errorf("type mismatch for comparison operator '%s' in chained expr at pos %d: cannot compare %s and %s",
					operator, expr.Pos(), currentType.String(), rightType.String())
			}
		case "&&", "||":
			if currentType.Equals(BoolType) && rightType.Equals(BoolType) {
				currentType = BoolType
			} else {
				return nil, fmt.Errorf("type mismatch for logical operator '%s' in chained expr at pos %d: requires two booleans, got %s and %s",
					operator, expr.Pos(), currentType.String(), rightType.String())
			}
		default:
			return nil, fmt.Errorf("unsupported operator '%s' in chained expr at pos %d", operator, expr.Pos())
		}
	}
	return currentType, nil
}

func inferDistributeExprType(expr *DistributeExpr, scope *TypeScope) (*Type, error) {
	var commonBodyType *Type

	if len(expr.Cases) == 0 && expr.Default == nil {
		return nil, fmt.Errorf("distribute expression at pos %d must have at least one case or a default", expr.Pos())
	}

	for i, caseExpr := range expr.Cases {
		if caseExpr.Condition == nil {
			return nil, fmt.Errorf("DistributeExpr case %d at pos %d has no condition", i, caseExpr.Pos())
		}
		// Condition type for distribute cases is often numeric (probability/weight).
		condType, err := InferExprType(caseExpr.Condition, scope)
		if err != nil {
			return nil, fmt.Errorf("error inferring type for condition of case %d in distribute expr at pos %d: %w", i, caseExpr.Condition.Pos(), err)
		}
		if !(condType.Equals(FloatType) || condType.Equals(IntType)) { // Assuming weights are numeric
			return nil, fmt.Errorf("condition of distribute case %d at pos %d must be numeric (for weight), got %s", i, caseExpr.Condition.Pos(), condType.String())
		}

		if caseExpr.Body == nil {
			return nil, fmt.Errorf("DistributeExpr case %d at pos %d has no body", i, caseExpr.Pos())
		}
		bodyType, err := InferExprType(caseExpr.Body, scope)
		if err != nil {
			return nil, fmt.Errorf("error inferring type for body of case %d in distribute expr at pos %d: %w", i, caseExpr.Body.Pos(), err)
		}
		if bodyType == nil { // Should be caught by InferExprType returning error
			return nil, fmt.Errorf("could not determine type for body of case %d in distribute expr at pos %d", i, caseExpr.Pos())
		}

		if commonBodyType == nil {
			commonBodyType = bodyType
		} else if !commonBodyType.Equals(bodyType) {
			// TODO: Type compatibility/promotion rules? For now, require exact match.
			return nil, fmt.Errorf("type mismatch in distribute expr cases at pos %d: expected %s (from case 0), got %s for case %d",
				expr.Pos(), commonBodyType.String(), bodyType.String(), i)
		}
	}

	if expr.Default != nil {
		defaultType, err := InferExprType(expr.Default, scope)
		if err != nil {
			return nil, fmt.Errorf("error inferring type for default case of distribute expr at pos %d: %w", expr.Default.Pos(), err)
		}
		if defaultType == nil {
			return nil, fmt.Errorf("could not determine type for default case of distribute expr at pos %d", expr.Pos())
		}

		if commonBodyType == nil {
			commonBodyType = defaultType
		} else if !commonBodyType.Equals(defaultType) {
			return nil, fmt.Errorf("type mismatch between distribute expr cases and default at pos %d: expected %s, got %s for default",
				expr.Pos(), commonBodyType.String(), defaultType.String())
		}
	}

	if commonBodyType == nil {
		// This implies there were no cases and no default, caught earlier.
		// Or all bodies are NilType. If so, commonBodyType is NilType.
		// The issue might be if some are Nil and some are not.
		// For now, if everything was nil and commonBodyType remained nil, this is an issue.
		return nil, fmt.Errorf("distribute expr at pos %d has no effective common type (all branches might be nil or types mismatched implicitly)", expr.Pos())
	}

	return OutcomesType(commonBodyType), nil
}

func inferSampleExprType(expr *SampleExpr, scope *TypeScope) (*Type, error) {
	fromType, err := InferExprType(expr.FromExpr, scope)
	if err != nil {
		return nil, fmt.Errorf("error inferring type for 'from' expression of sample expr at pos %d: %w", expr.FromExpr.Pos(), err)
	}
	if fromType == nil {
		return nil, fmt.Errorf("could not determine type for 'from' expression of sample expr at pos %d", expr.Pos())
	}

	if fromType.Name != "Outcomes" || len(fromType.ChildTypes) != 1 {
		return nil, fmt.Errorf("type mismatch for sample expression at pos %d: 'from' expression must be Outcomes[T], got %s",
			expr.Pos(), fromType.String())
	}
	return fromType.ChildTypes[0], nil // Returns T from Outcomes[T]
}

// --- Top-level and Statement Helpers ---

// InferTypesForFile initiates type inference for all relevant parts of a FileDecl.
func InferTypesForFile(file *FileDecl) []error {
	var errors []error
	if file == nil {
		errors = append(errors, fmt.Errorf("cannot infer types for nil FileDecl"))
		return errors
	}

	if err := file.ensureResolved(); err != nil {
		errors = append(errors, fmt.Errorf("error resolving file before type inference: %w", err))
		return errors // Stop if resolution fails
	}

	rootScope := NewRootTypeScope(file)

	for _, decl := range file.Declarations {
		switch d := decl.(type) {
		case *ComponentDecl:
			// Infer types for default values of parameters first
			compScope := rootScope.Push(d, nil) // Scope for component itself
			for _, paramDecl := range d.params {
				if paramDecl.DefaultValue != nil {
					valType, err := InferExprType(paramDecl.DefaultValue, compScope)
					if err != nil {
						errors = append(errors, err)
					} else if valType != nil && paramDecl.Type != nil {
						expectedType := paramDecl.Type.Type()
						if !valType.Equals(expectedType) {
							if !(valType.Equals(IntType) && expectedType.Equals(FloatType)) { // allow int to float promotion
								errors = append(errors, fmt.Errorf("type mismatch for default value of param '%s' in component '%s' at pos %d: expected %s, got %s", paramDecl.Name.Name, d.NameNode.Name, paramDecl.DefaultValue.Pos(), expectedType.String(), valType.String()))
							}
						}
					}
				}
			}

			// Infer types within component methods
			for _, method := range d.methods {
				methodScope := rootScope.Push(d, method) // New scope for each method
				// Add parameters to methodScope
				for _, param := range method.Parameters {
					if param.Type == nil {
						errors = append(errors, fmt.Errorf("parameter '%s' of method '%s.%s' at pos %d has no type declaration", param.Name.Name, d.NameNode.Name, method.NameNode.Name, param.Pos()))
						continue
					}
					paramType := param.Type.Type()
					if paramType == nil {
						errors = append(errors, fmt.Errorf("parameter '%s' of method '%s.%s' at pos %d has invalid TypeDecl", param.Name.Name, d.NameNode.Name, method.NameNode.Name, param.Pos()))
						continue
					}
					methodScope.Set(param.Name.Name, paramType)
					param.Name.SetInferredType(paramType) // Also type the identifier node itself
				}
				// Add 'uses' dependencies to scope
				for depName, usesDecl := range d.uses {
					depCompName := usesDecl.ComponentNode.Name
					// The type of a dependency is the component type itself.
					methodScope.Set(depName, &Type{Name: depCompName})
				}

				if method.Body != nil {
					errs := inferTypesForBlockStmt(method.Body, methodScope)
					errors = append(errors, errs...)
				}
			}
		case *SystemDecl:
			systemScope := rootScope.Push(nil, nil)
			for _, item := range d.Body {
				errs := inferTypesForSystemDeclBodyItem(item, systemScope)
				errors = append(errors, errs...)
			}
		}
	}
	return errors
}

func inferTypesForBlockStmt(block *BlockStmt, parentScope *TypeScope) []error {
	var errors []error
	// Create a new scope for the block, inheriting from parentScope
	// The current component/method context is inherited from parentScope
	blockScope := parentScope.Push(parentScope.currentComponent, parentScope.currentMethod)
	for _, stmt := range block.Statements {
		errs := inferTypesForStmt(stmt, blockScope)
		errors = append(errors, errs...)
	}
	return errors
}

func inferTypesForStmt(stmt Stmt, scope *TypeScope) []error {
	var errors []error
	switch s := stmt.(type) {
	case *LetStmt:
		valType, err := InferExprType(s.Value, scope)
		if err != nil {
			errors = append(errors, fmt.Errorf("error inferring type for value of let statement variable(s) near pos %d: %w", s.Pos(), err))
		} else if valType != nil {
			// For `let x = val;`, x gets type of val.
			// For `let x, y = val;` (if supported), val must be tuple, x and y get corresponding tuple element types.
			// Current LetStmt has Variables []*IdentifierExpr. Assume single var for now or all get same type.
			if len(s.Variables) == 1 {
				varIdent := s.Variables[0]
				scope.Set(varIdent.Name, valType)
				varIdent.SetInferredType(valType) // Set type on the IdentifierExpr node
			} else if len(s.Variables) > 1 {
				// Destructuring assignment: valType must be a TupleType
				if valType.Name == "Tuple" && len(valType.ChildTypes) == len(s.Variables) {
					for i, varIdent := range s.Variables {
						elemType := valType.ChildTypes[i]
						scope.Set(varIdent.Name, elemType)
						varIdent.SetInferredType(elemType)
					}
				} else {
					errors = append(errors, fmt.Errorf("let statement at pos %d assigns to multiple variables, but value type %s is not a matching tuple", s.Pos(), valType.String()))
				}
			}
			// TODO: Check against declared type on let var if syntax allows `let x: T = val;`
		}
	case *ExprStmt:
		_, err := InferExprType(s.Expression, scope)
		if err != nil {
			errors = append(errors, fmt.Errorf("error inferring type for expression statement at pos %d: %w", s.Expression.Pos(), err))
		}
	case *ReturnStmt:
		var actualReturnType *Type = NilType
		if s.ReturnValue != nil {
			infRetType, err := InferExprType(s.ReturnValue, scope)
			if err != nil {
				errors = append(errors, fmt.Errorf("error inferring type for return value at pos %d: %w", s.ReturnValue.Pos(), err))
			} else if infRetType != nil {
				actualReturnType = infRetType
			}
		}

		if scope.currentMethod != nil {
			var expectedReturnType *Type = NilType
			if scope.currentMethod.ReturnType != nil {
				expectedReturnType = scope.currentMethod.ReturnType.Type()
				if expectedReturnType == nil { // Should not happen
					errors = append(errors, fmt.Errorf("method '%s' return TypeDecl at pos %d did not resolve to a Type", scope.currentMethod.NameNode.Name, scope.currentMethod.ReturnType.Pos()))
					return errors
				}
			}

			if !actualReturnType.Equals(expectedReturnType) {
				// Allow int to float promotion for return
				isPromotion := actualReturnType.Equals(IntType) && expectedReturnType.Equals(FloatType)
				if !isPromotion {
					errors = append(errors, fmt.Errorf("return type mismatch for method '%s' at pos %d: expected %s, got %s",
						scope.currentMethod.NameNode.Name, s.Pos(), expectedReturnType.String(), actualReturnType.String()))
				}
			}
		} else { // Return outside a method (e.g. top level of a script, if supported)
			// This might be an error depending on DSL semantics.
		}
	case *IfStmt:
		condType, err := InferExprType(s.Condition, scope)
		if err != nil {
			errors = append(errors, fmt.Errorf("error inferring type for if condition at pos %d: %w", s.Condition.Pos(), err))
		} else if condType != nil && !condType.Equals(BoolType) {
			errors = append(errors, fmt.Errorf("if condition at pos %d must be boolean, got %s", s.Condition.Pos(), condType.String()))
		}
		if s.Then != nil {
			errsThen := inferTypesForBlockStmt(s.Then, scope) // Then block gets a new sub-scope from current scope
			errors = append(errors, errsThen...)
		}
		if s.Else != nil {
			// Else statement also operates in a sub-scope of the If's scope.
			// If Else is a BlockStmt, inferTypesForBlockStmt will create its own scope.
			// If Else is another IfStmt, inferTypesForStmt will handle it.
			errsElse := inferTypesForStmt(s.Else, scope)
			errors = append(errors, errsElse...)
		}
	case *BlockStmt: // Nested block
		errs := inferTypesForBlockStmt(s, scope) // Pass current scope; it will push a new one
		errors = append(errors, errs...)
	case *AssignmentStmt: // Typically in InstanceDecl overrides
		valType, err := InferExprType(s.Value, scope)
		if err != nil {
			errors = append(errors, fmt.Errorf("error inferring type for assignment value to '%s' at pos %d: %w", s.Var.Name, s.Value.Pos(), err))
		}
		// The type of s.Var itself (LHS) depends on context (e.g. component param).
		// This check is done within inferTypesForSystemDeclBodyItem for InstanceDecl.
		// Here, we can type s.Var if it's a known identifier, though it's usually a field name.
		if s.Var != nil && s.Var.InferredType() == nil { // If not already typed (e.g. as part of instance override)
			varType, _ := scope.Get(s.Var.Name) // Try to find if it's a re-assignable variable
			if varType != nil {
				s.Var.SetInferredType(varType)
				if valType != nil && !valType.Equals(varType) {
					if !(valType.Equals(IntType) && varType.Equals(FloatType)) {
						errors = append(errors, fmt.Errorf("type mismatch in assignment to '%s' at pos %d: variable is %s, value is %s", s.Var.Name, s.Pos(), varType.String(), valType.String()))
					}
				}
			}
			// If s.Var is not in scope, it's likely an unresolvable assignment target or a field (handled by context)
		}

	// TODO: Other statement types: SwitchStmt, DelayStmt, WaitStmt, GoStmt, LogStmt, ExpectStmt
	case *ExpectStmt: // Occurs inside AnalyzeDecl.Expectations
		// This is typically handled within inferTypesForSystemDeclBodyItem for AnalyzeDecl
		// as it needs the analyze result in scope. If called directly, it's harder.
		// For robustness, can add a basic pass here.
		if s.Target != nil {
			_, err := InferExprType(s.Target, scope)
			if err != nil {
				errors = append(errors, err)
			}
		}
		if s.Threshold != nil {
			_, err := InferExprType(s.Threshold, scope)
			if err != nil {
				errors = append(errors, err)
			}
		}

		// default:
		// log.Printf("Type inference for statement type %T not fully implemented yet at pos %d", stmt, stmt.Pos())
	}
	return errors
}

func inferTypesForSystemDeclBodyItem(item SystemDeclBodyItem, systemScope *TypeScope) []error {
	var errors []error
	switch i := item.(type) {
	case *InstanceDecl:
		compDefinition, err := systemScope.file.GetComponent(i.ComponentType.Name)
		if err != nil || compDefinition == nil {
			errors = append(errors, fmt.Errorf("component type '%s' not found for instance '%s' at pos %d", i.ComponentType.Name, i.NameNode.Name, i.ComponentType.Pos()))
			return errors // Cannot proceed with overrides if component def is missing
		}
		// The instance itself gets the component's type
		instanceType := &Type{Name: compDefinition.NameNode.Name /* IsComponent: true */}
		systemScope.Set(i.NameNode.Name, instanceType)
		i.NameNode.SetInferredType(instanceType) // Type the identifier node

		// Infer types for override values and check against param/dependency types
		for _, assign := range i.Overrides {
			valType, err := InferExprType(assign.Value, systemScope) // Use systemScope for override value expressions
			if err != nil {
				errors = append(errors, fmt.Errorf("error inferring type for override value of '%s' in instance '%s' at pos %d: %w", assign.Var.Name, i.NameNode.Name, assign.Value.Pos(), err))
				continue
			}
			if valType == nil {
				continue // Error already reported by InferExprType or value is untypable
			}

			paramDecl, _ := compDefinition.GetParam(assign.Var.Name)
			usesDecl, _ := compDefinition.GetDependency(assign.Var.Name)

			if paramDecl != nil {
				if paramDecl.Type == nil {
					errors = append(errors, fmt.Errorf("param '%s' of component '%s' has no type at pos %d", paramDecl.Name.Name, compDefinition.NameNode.Name, paramDecl.Pos()))
					continue
				}
				expectedType := paramDecl.Type.Type()
				if !valType.Equals(expectedType) {
					if !(valType.Equals(IntType) && expectedType.Equals(FloatType)) { // Allow int to float promotion
						errors = append(errors, fmt.Errorf("type mismatch for override param '%s' in instance '%s' at pos %d: expected %s, got %s",
							assign.Var.Name, i.NameNode.Name, assign.Value.Pos(), expectedType.String(), valType.String()))
					}
				}
			} else if usesDecl != nil {
				expectedDepCompName := usesDecl.ComponentNode.Name
				// valType should be a component type matching expectedDepCompName
				if !(valType.Name == expectedDepCompName /*&& valType.IsComponent ??? */) {
					errors = append(errors, fmt.Errorf("type mismatch for override dependency '%s' in instance '%s' at pos %d: expected component type %s, got %s",
						assign.Var.Name, i.NameNode.Name, assign.Value.Pos(), expectedDepCompName, valType.String()))
				}
			} else {
				errors = append(errors, fmt.Errorf("override target '%s' in instance '%s' at pos %d is not a known parameter or dependency of component '%s'",
					assign.Var.Name, i.NameNode.Name, assign.Var.Pos(), compDefinition.NameNode.Name))
			}
		}
	case *AnalyzeDecl:
		targetType, err := InferExprType(i.Target, systemScope)
		if err != nil {
			errors = append(errors, fmt.Errorf("error inferring type for analyze target '%s' at pos %d: %w", i.Name.Name, i.Target.Pos(), err))
		}
		// Make the result of the analyze target available in a sub-scope for expectations
		if targetType != nil && i.Expectations != nil {
			// AnalyzeDecl Name (e.g. "Result") goes into scope for the Expectations block.
			expectScope := systemScope.Push(nil, nil) // New scope for expectations
			expectScope.Set(i.Name.Name, targetType)
			i.Name.SetInferredType(targetType) // Type the identifier of the analyze block itself

			for _, expectStmt := range i.Expectations.Expects {
				// expectStmt.Target is MemberAccessExpr, e.g., Result.P99
				// Its Receiver should be an IdentifierExpr matching i.Name.Name
				if expectStmt.Target.Receiver != nil {
					if targetIdent, ok := expectStmt.Target.Receiver.(*IdentifierExpr); ok {
						if targetIdent.Name != i.Name.Name {
							errors = append(errors, fmt.Errorf("expect statement target receiver '%s' at pos %d does not match analyze block name '%s'", targetIdent.Name, targetIdent.Pos(), i.Name.Name))
						}
						// Type the receiver identifier
						targetIdent.SetInferredType(targetType)
					} else {
						errors = append(errors, fmt.Errorf("expect statement target receiver at pos %d must be a simple identifier referring to the analyze block name", expectStmt.Target.Receiver.Pos()))
					}
				}

				metricType, err := InferExprType(expectStmt.Target, expectScope)
				if err != nil {
					errors = append(errors, fmt.Errorf("error inferring type for expect metric '%s' at pos %d: %w", expectStmt.Target.String(), expectStmt.Target.Pos(), err))
				}

				thresholdType, err := InferExprType(expectStmt.Threshold, expectScope)
				if err != nil {
					errors = append(errors, fmt.Errorf("error inferring type for expect threshold at pos %d: %w", expectStmt.Threshold.Pos(), err))
				}

				if metricType != nil && thresholdType != nil {
					isMetricNumeric := metricType.Equals(IntType) || metricType.Equals(FloatType)
					isThresholdNumeric := thresholdType.Equals(IntType) || thresholdType.Equals(FloatType)
					if !((isMetricNumeric && isThresholdNumeric) || metricType.Equals(thresholdType)) {
						errors = append(errors, fmt.Errorf("type mismatch in expect statement at pos %d: cannot compare metric type %s with threshold type %s using operator '%s'",
							expectStmt.Pos(), metricType.String(), thresholdType.String(), expectStmt.Operator))
					}
				}
			}
		}

	case *LetStmt: // LetStmt can also be a SystemDeclBodyItem
		errs := inferTypesForStmt(i, systemScope) // Use systemScope
		errors = append(errors, errs...)
		// OptionsDecl - usually doesn't have expressions that need this kind of complex inference.
	}
	return errors
}
