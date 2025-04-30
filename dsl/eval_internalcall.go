package dsl

import "fmt"

// evalInternalCall evaluates the arguments of an InternalCallExpr,
// looks up the function in the interpreter's registry, calls the function
// with the evaluated arguments, and pushes the result onto the stack.
func (i *Interpreter) evalInternalCall(expr *InternalCallExpr) error {
	// 1. Evaluate arguments recursively
	evaluatedArgs := make([]interface{}, len(expr.Args))
	for _, argExpr := range expr.Args {
		_, err := i.Eval(argExpr) // Result is pushed onto the stack
		if err != nil {
			// Potential issue: stack might have partial results.
			// Need careful error handling/stack cleanup if this happens mid-args.
			// For now, assume Eval cleans up its own errors or we handle it higher up.
			return fmt.Errorf("error evaluating arg for %s: %w", expr.FuncName, err)
		}
	}

	// 2. Pop evaluated arguments (in reverse order of evaluation)
	for j := len(expr.Args) - 1; j >= 0; j-- {
		argVal, err := i.pop()
		if err != nil {
			// This indicates a deeper issue, stack should have had enough elements
			return fmt.Errorf("stack underflow retrieving arg %d for %s: %w", j, expr.FuncName, err)
		}
		evaluatedArgs[j] = argVal
	}

	// 3. Look up the internal function
	internalFn, ok := i.internalFuncs[expr.FuncName]
	if !ok {
		return fmt.Errorf("%w: '%s'", ErrInternalFuncNotFound, expr.FuncName)
	}

	// 4. Call the internal function
	// The internal function receives the interpreter and the evaluated args slice
	result, err := internalFn(i, evaluatedArgs)
	if err != nil {
		return fmt.Errorf("error executing internal function '%s': %w", expr.FuncName, err)
	}

	// 5. Push the result onto the stack
	i.push(result)

	return nil
}
