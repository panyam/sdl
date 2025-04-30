// sdl/dsl/interpreter_test.go
package dsl

import (
	"errors"
	"testing"

	"github.com/panyam/leetcoach/sdl/components"
	"github.com/panyam/leetcoach/sdl/core"
	// "fmt" // For debugging
)

func TestInterpreter_NewInterpreter(t *testing.T) {
	interp := NewInterpreter(20)
	if interp == nil {
		t.Fatal("NewInterpreter returned nil")
	}
	if interp.stack == nil {
		t.Error("Interpreter stack is nil")
	}
	if len(interp.stack) != 0 {
		t.Errorf("Expected initial stack length 0, got %d", len(interp.stack))
	}
	if interp.env == nil {
		t.Error("Interpreter environment is nil")
	}
	if len(interp.env.store) != 0 {
		t.Error("Expected initial environment store to be empty")
	}
	if interp.internalFuncs == nil {
		t.Error("Interpreter internalFuncs map is nil")
	}
	if len(interp.internalFuncs) != 0 {
		t.Error("Expected initial internalFuncs map to be empty")
	}
	if interp.maxOutcomeLen != 20 {
		t.Errorf("Expected maxOutcomeLen 20, got %d", interp.maxOutcomeLen)
	}
}

// --- Test Literals (Phase 1) ---

func TestInterpreter_Eval_LiteralInt(t *testing.T) {
	interp := NewInterpreter(10)
	literal := &LiteralExpr{Kind: "INT", Value: "42"}

	_, err := interp.Eval(literal)
	if err != nil {
		t.Fatalf("Eval(LiteralInt) failed: %v", err)
	}

	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	outcome, ok := result.(*core.Outcomes[int64]) // Expecting int64
	if !ok {
		t.Fatalf("Expected result type *core.Outcomes[int64], got %T", result)
	}

	val, ok := outcome.GetValue() // Use GetValue for deterministic outcomes
	if !ok {
		t.Fatalf("GetValue failed for deterministic int outcome")
	}
	if val != 42 {
		t.Errorf("Expected literal value 42, got %d", val)
	}
}

func TestInterpreter_Eval_LiteralFloat(t *testing.T) {
	interp := NewInterpreter(10)
	literal := &LiteralExpr{Kind: "FLOAT", Value: "3.14"}

	_, err := interp.Eval(literal)
	if err != nil {
		t.Fatalf("Eval(LiteralFloat) failed: %v", err)
	}
	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}
	outcome, ok := result.(*core.Outcomes[float64])
	if !ok {
		t.Fatalf("Expected result type *core.Outcomes[float64], got %T", result)
	}
	val, ok := outcome.GetValue()
	if !ok {
		t.Fatalf("GetValue failed for deterministic float outcome")
	}
	if val != 3.14 {
		t.Errorf("Expected literal value 3.14, got %f", val)
	}
}

func TestInterpreter_Eval_LiteralString(t *testing.T) {
	interp := NewInterpreter(10)
	literal := &LiteralExpr{Kind: "STRING", Value: "hello"} // Note: No quotes in Value field

	_, err := interp.Eval(literal)
	if err != nil {
		t.Fatalf("Eval(LiteralString) failed: %v", err)
	}
	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}
	outcome, ok := result.(*core.Outcomes[string])
	if !ok {
		t.Fatalf("Expected result type *core.Outcomes[string], got %T", result)
	}
	val, ok := outcome.GetValue()
	if !ok {
		t.Fatalf("GetValue failed for deterministic string outcome")
	}
	if val != "hello" {
		t.Errorf("Expected literal value 'hello', got '%s'", val)
	}
}

func TestInterpreter_Eval_LiteralBool(t *testing.T) {
	interp := NewInterpreter(10)
	literal := &LiteralExpr{Kind: "BOOL", Value: "true"}

	_, err := interp.Eval(literal)
	if err != nil {
		t.Fatalf("Eval(LiteralBool) failed: %v", err)
	}
	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}
	outcome, ok := result.(*core.Outcomes[bool])
	if !ok {
		t.Fatalf("Expected result type *core.Outcomes[bool], got %T", result)
	}
	val, ok := outcome.GetValue()
	if !ok {
		t.Fatalf("GetValue failed for deterministic bool outcome")
	}
	if val != true {
		t.Errorf("Expected literal value true, got %v", val)
	}
}

func TestEnvironment_SetGet(t *testing.T) {
	env := NewEnvironment()
	env.Set("myVar", 123)
	val, ok := env.Get("myVar")
	if !ok {
		t.Fatal("Failed to get 'myVar' from environment")
	}
	if val != 123 {
		t.Errorf("Expected value 123, got %v", val)
	}

	_, ok = env.Get("notFound")
	if ok {
		t.Error("Expected 'notFound' to not be found")
	}
}

// --- Test Identifier & Internal Calls (Phase 2) ---

func TestInterpreter_Eval_Identifier(t *testing.T) {
	interp := NewInterpreter(10)
	// Setup environment
	testOutcome := (&core.Outcomes[int64]{}).Add(1.0, 99)
	interp.Env().Set("myVar", testOutcome)

	ident := &IdentifierExpr{Name: "myVar"}
	_, err := interp.Eval(ident)
	if err != nil {
		t.Fatalf("Eval(IdentifierExpr) failed: %v", err)
	}

	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	if result != testOutcome { // Compare pointers for this test
		t.Errorf("Expected identifier to resolve to the set outcome object, got different object")
	}
}

func TestInterpreter_Eval_Identifier_NotFound(t *testing.T) {
	interp := NewInterpreter(10)
	ident := &IdentifierExpr{Name: "missingVar"}

	_, err := interp.Eval(ident)
	if err == nil {
		t.Fatal("Eval(IdentifierExpr) should have failed for missing variable")
	}

	// Check if the error is the expected ErrNotFound type (or wraps it)
	if !errors.Is(err, ErrNotFound) { // Use errors.Is for potential wrapping
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

// Helper function to register a dummy SSD read profile getter
func registerTestDiskFuncs(interp *Interpreter) {
	// Ensure Disk component is initialized to provide the outcomes
	ssd := components.NewDisk("TestSSD") // Uses SSD profile by default

	interp.RegisterInternalFunc("GetDiskReadProfile", func(i *Interpreter, args []any) (any, error) {
		// Expects one string arg: profile name (we ignore it here for simplicity)
		// if len(args) != 1 { return nil, fmt.Errorf("GetDiskReadProfile expects 1 arg") }
		// _, ok := args[0].(string); if !ok { return nil, fmt.Errorf("arg 0 must be string profile name") }

		return ssd.Read(), nil // Return the pre-calculated outcome object
	})
}

func TestInterpreter_Eval_InternalCall(t *testing.T) {
	interp := NewInterpreter(10)
	registerTestDiskFuncs(interp)

	// AST for Internal.GetDiskReadProfile("SSD") - Args aren't used yet by dummy func
	call := &InternalCallExpr{FuncName: "GetDiskReadProfile", Args: []Expr{&LiteralExpr{Kind: "STRING", Value: "SSD"}}}

	_, err := interp.Eval(call)
	if err != nil {
		t.Fatalf("Eval(InternalCallExpr) failed: %v", err)
	}

	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	// Check if the result is the expected outcome object
	expectedOutcome := components.NewDisk("").Read() // Get the reference SSD read outcome
	if result != expectedOutcome {                   // Compare pointers
		t.Errorf("Expected internal call to return SSD Read profile, got different object")
	}
}

func TestInterpreter_Eval_InternalCall_NotFound(t *testing.T) {
	interp := NewInterpreter(10)
	call := &InternalCallExpr{FuncName: "NoSuchFunction"}
	_, err := interp.Eval(call)
	if err == nil {
		t.Fatal("Eval(InternalCallExpr) should fail for unregistered function")
	}
	if !errors.Is(err, ErrInternalFuncNotFound) {
		t.Errorf("Expected ErrInternalFuncNotFound, got: %v", err)
	}
}

func TestEnvironment_Scoping(t *testing.T) {
	global := NewEnvironment()
	global.Set("globalVar", "hello")

	local := NewEnclosedEnvironment(global)
	local.Set("localVar", 123)

	// Get local
	valLocal, okLocal := local.Get("localVar")
	if !okLocal || valLocal != 123 {
		t.Errorf("Failed to get local var 'localVar', got %v (%t)", valLocal, okLocal)
	}

	// Get global from local
	valGlobal, okGlobal := local.Get("globalVar")
	if !okGlobal || valGlobal != "hello" {
		t.Errorf("Failed to get global var 'globalVar' from local scope, got %v (%t)", valGlobal, okGlobal)
	}

	// Try to get local from global (should fail)
	_, okLocalFromGlobal := global.Get("localVar")
	if okLocalFromGlobal {
		t.Error("Should not find local var in global scope")
	}

	// Shadowing
	local.Set("globalVar", 999)
	valShadowed, okShadowed := local.Get("globalVar")
	if !okShadowed || valShadowed != 999 {
		t.Errorf("Failed to get shadowed var 'globalVar' from local scope, got %v (%t)", valShadowed, okShadowed)
	}

	// Original global should be unchanged
	valOrigGlobal, okOrigGlobal := global.Get("globalVar")
	if !okOrigGlobal || valOrigGlobal != "hello" {
		t.Errorf("Global var was unexpectedly changed by local scope, got %v (%t)", valOrigGlobal, okOrigGlobal)
	}
}

func TestInterpreter_Stack(t *testing.T) {
	interp := NewInterpreter(10)

	_, err := interp.pop()
	if err == nil || err != ErrStackUnderflow {
		t.Errorf("Expected ErrStackUnderflow on empty pop, got %v", err)
	}

	interp.push(1)
	interp.push("two")

	if len(interp.stack) != 2 {
		t.Fatalf("Expected stack length 2, got %d", len(interp.stack))
	}

	top, ok := interp.peek()
	if !ok || top != "two" {
		t.Errorf("Peek failed: expected 'two', got %v (%t)", top, ok)
	}

	val2, err2 := interp.pop()
	if err2 != nil {
		t.Fatalf("Error popping 'two': %v", err2)
	}
	if val2 != "two" {
		t.Errorf("Expected popped value 'two', got %v", val2)
	}
	if len(interp.stack) != 1 {
		t.Errorf("Expected stack length 1 after pop, got %d", len(interp.stack))
	}

	val1, err1 := interp.pop()
	if err1 != nil {
		t.Fatalf("Error popping 1: %v", err1)
	}
	if val1 != 1 {
		t.Errorf("Expected popped value 1, got %v", val1)
	}
	if len(interp.stack) != 0 {
		t.Errorf("Expected stack length 0 after popping all, got %d", len(interp.stack))
	}

	_, errEmpty := interp.pop()
	if errEmpty == nil || errEmpty != ErrStackUnderflow {
		t.Errorf("Expected ErrStackUnderflow on final empty pop, got %v", errEmpty)
	}
}

func TestInterpreter_RegisterInternalFunc(t *testing.T) {
	interp := NewInterpreter(10)
	testFunc := func(i *Interpreter, args []any) (any, error) {
		return "test success", nil
	}

	interp.RegisterInternalFunc("myTestFunc", testFunc)

	fn, ok := interp.internalFuncs["myTestFunc"]
	if !ok {
		t.Fatal("Failed to register internal function")
	}
	if fn == nil { // Check if function pointer is actually stored
		t.Fatal("Registered function is nil")
	}
	// Note: Comparing function pointers directly can be tricky,
	// but checking presence and non-nil is a good start.
}

func TestInterpreter_Eval_Stub(t *testing.T) {
	interp := NewInterpreter(10)
	// Use an AST node for which Eval is not implemented yet
	node := &MemberAccessExpr{}
	_, err := interp.Eval(node)
	if err == nil {
		t.Error("Expected an error for unimplemented node type")
	}
	// Check if the error indicates non-implementation (the exact message might change)
	// if !strings.Contains(err.Error(), "not implemented") {
	//  t.Errorf("Expected 'not implemented' error, got: %v", err)
	// }
	// More robust: check if the error wraps ErrNotImplemented if we implement it that way later
}
