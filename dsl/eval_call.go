package dsl

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/panyam/leetcoach/sdl/core" // Need outcomes for type checking return value
)

var (
	ErrInvalidReceiver = errors.New("invalid receiver for method call")
	ErrMethodNotFound  = errors.New("method not found")
	ErrNotFunction     = errors.New("expression is not callable")
	ErrInvalidArgument = errors.New("invalid argument type or count")
	ErrInvalidReturn   = errors.New("invalid return type from method call")
)

// evalCallExpr handles evaluating function/method calls.
// Currently focuses on method calls like `instance.Method(args...)`.
func (i *Interpreter) evalCallExpr(expr *CallExpr) error {
	// --- Evaluate Function Expression ---
	// For now, we primarily support MemberAccessExpr as the function (method calls)
	// e.g., myDisk.Read() -> Function is MemberAccessExpr{Receiver: Ident{myDisk}, Member: "Read"}
	memExpr, ok := expr.Function.(*MemberAccessExpr)
	if !ok {
		// TODO: Later support calling functions stored in variables?
		return fmt.Errorf("%w: expected method call (receiver.Member), got %T", ErrNotFunction, expr.Function)
	}

	// --- Evaluate Receiver ---
	// Evaluate the expression that yields the object instance (e.g., the IdentifierExpr "myDisk")
	_, err := i.Eval(memExpr.Receiver)
	if err != nil {
		return fmt.Errorf("error evaluating receiver for method call '%s': %w", memExpr.Member, err)
	}
	receiverObjRaw, err := i.pop() // Pop the evaluated receiver object
	if err != nil {
		return fmt.Errorf("stack error retrieving receiver for method call '%s': %w", memExpr.Member, err)
	}

	// --- Evaluate Arguments ---
	evaluatedArgs := make([]any, len(expr.Args))
	argVals := make([]reflect.Value, len(expr.Args)) // For reflect.Call

	for idx, argExpr := range expr.Args {
		_, err := i.Eval(argExpr) // Results pushed onto stack
		if err != nil {
			// Clean up stack from receiver + previous args before returning
			// i.stack = i.stack[:len(i.stack)-idx] // Remove args evaluated so far
			return fmt.Errorf("error evaluating arg %d for method call '%s': %w", idx, memExpr.Member, err)
		}
	}
	// Pop evaluated args (in reverse)
	for j := len(expr.Args) - 1; j >= 0; j-- {
		argValRaw, err := i.pop()
		if err != nil {
			return fmt.Errorf("stack underflow retrieving arg %d for method call '%s': %w", j, memExpr.Member, err)
		}
		// TODO: Type checking/conversion needed here if methods expect specific Go types
		// Store the raw any value for now. We will convert/check below.
		// Wrap in reflect.Value later after potential conversion.
		evaluatedArgs[j] = argValRaw
		// Wrap in reflect.Value for calling via reflection
		argVals[j] = reflect.ValueOf(argValRaw)
	}

	// --- Method Invocation using Reflection ---
	receiverVal := reflect.ValueOf(receiverObjRaw)
	methodName := memExpr.Member

	// Find the method on the receiver's type
	methodVal := receiverVal.MethodByName(methodName)
	if !methodVal.IsValid() {
		// Check if receiver is a pointer, try on the element type if so
		if receiverVal.Kind() == reflect.Ptr {
			methodVal = receiverVal.Elem().MethodByName(methodName)
		}
	}

	if !methodVal.IsValid() {
		return fmt.Errorf("%w: receiver type %T has no method '%s'", ErrMethodNotFound, receiverObjRaw, methodName)
	}

	// --- Argument Count/Type Check & Conversion ---
	methodType := methodVal.Type()
	if methodType.NumIn() != len(argVals) {
		return fmt.Errorf("%w: method '%s' expects %d arguments, got %d", ErrInvalidArgument, methodName, methodType.NumIn(), len(argVals))
	}

	// Prepare converted reflect.Value arguments
	convertedArgVals := make([]reflect.Value, len(evaluatedArgs))

	for j, argRaw := range evaluatedArgs {
		expectedGoType := methodType.In(j) // The Go type the method expects

		// Argument MUST be a deterministic outcome
		// We need a way to call GetValue generically. Use reflection.
		argOutcomeVal := reflect.ValueOf(argRaw) // Value of the *core.Outcomes[T]
		if argOutcomeVal.Kind() != reflect.Ptr || !argOutcomeVal.MethodByName("GetValue").IsValid() {
			return fmt.Errorf("%w: argument %d for method '%s' is not a valid Outcome object (%T)", ErrInvalidArgument, j, methodName, argRaw)
		}

		// Call GetValue() using reflection -> returns []interface{ Value, bool }
		getValueResults := argOutcomeVal.MethodByName("GetValue").Call(nil)
		if len(getValueResults) != 2 || !getValueResults[1].Bool() {
			return fmt.Errorf("%w: argument %d for method '%s' must be deterministic (single outcome bucket)", ErrInvalidArgument, j, methodName)
		}
		deterministicVal := getValueResults[0].Interface() // The actual inner value (e.g., int64, float64)

		// --- Type Conversion ---
		actualValReflect := reflect.ValueOf(deterministicVal)
		if !actualValReflect.Type().AssignableTo(expectedGoType) {
			// Attempt basic conversions (e.g., int64 -> uint, float64 -> core.Duration)
			if actualValReflect.Type().ConvertibleTo(expectedGoType) {
				convertedValReflect := actualValReflect.Convert(expectedGoType)
				convertedArgVals[j] = convertedValReflect
			} else {
				return fmt.Errorf("%w: cannot convert argument %d for method '%s' from type %T to expected type %s", ErrInvalidArgument, j, methodName, deterministicVal, expectedGoType.String())
			}
		} else {
			// Types are assignable, use the value directly
			convertedArgVals[j] = actualValReflect
		}
	}
	// TODO: Add more detailed argument type checking if needed. For now, rely on reflect.Call panicking or method handling.

	// --- Call the Method ---
	// Use reflect.Call - this can panic if arg types are wrong!
	// Error handling around Call is limited.
	var returnVals []reflect.Value
	// Add basic panic recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic during method call '%s': %v", methodName, r)
			}
		}()
		returnVals = methodVal.Call(convertedArgVals) // Use converted arguments
	}()
	if err != nil { // Check if panic occurred
		return err
	}

	// --- Process Return Value ---
	// Expecting a single return value which is *core.Outcomes[V] or ast.Node
	if len(returnVals) != 1 {
		return fmt.Errorf("%w: method '%s' did not return exactly one value (got %d)", ErrInvalidReturn, methodName, len(returnVals))
	}

	returnValue := returnVals[0].Interface() // Get the any value back from reflect.Value

	// Check if the return value is an AST node to be evaluated further (Phase 5)
	// or an Outcome object to be pushed (this phase).
	switch ret := returnValue.(type) {
	case Node: // Check if it implements our AST Node interface
		// --- Phase 5 Logic ---
		// Recursively evaluate the AST node returned by the Go method
		_, evalErr := i.Eval(ret)
		if evalErr != nil {
			return fmt.Errorf("error evaluating AST returned by method '%s': %w", methodName, evalErr)
		}
		// Result of evaluating the returned AST is now on the stack, nothing more to push here.
	case *core.Outcomes[core.AccessResult], // Add known outcome types explicitly
		*core.Outcomes[core.Duration],
		*core.Outcomes[int],   // Add types used by literals etc.
		*core.Outcomes[int64], // Add types used by literals etc.
		*core.Outcomes[string],
		*core.Outcomes[bool]:
		// This is a final outcome object, push it directly
		i.push(ret)
	default:
		// Check dynamically if it's *any* *core.Outcomes[T] using reflection
		retType := reflect.TypeOf(ret)
		isOutcome := retType.Kind() == reflect.Ptr &&
			retType.Elem().Kind() == reflect.Struct && // Check if it's a pointer to a struct
			retType.Elem().Name() == "Outcomes" && // Check the struct name
			retType.Elem().PkgPath() == "github.com/panyam/leetcoach/sdl/core" // Check package path

		if isOutcome {
			i.push(ret) // Push the generic *core.Outcomes[?]
		} else {
			return fmt.Errorf("%w: method '%s' returned unexpected type %T (expected *core.Outcomes[V] or ast.Node)", ErrInvalidReturn, methodName, ret)
		}
	}

	return nil
}
