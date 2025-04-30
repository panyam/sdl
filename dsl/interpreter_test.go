// sdl/dsl/interpreter_test.go
package dsl

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/panyam/leetcoach/sdl/components"
	"github.com/panyam/leetcoach/sdl/core"
	// "fmt" // For debugging
)

func litStr(v string) Expr { return &LiteralExpr{Kind: "STRING", Value: v} }
func litInt(v int) Expr    { return &LiteralExpr{Kind: "INT", Value: strconv.Itoa(v)} }
func litFloat(v float64) Expr {
	return &LiteralExpr{Kind: "FLOAT", Value: strconv.FormatFloat(v, 'f', -1, 64)}
}
func litBool(v bool) Expr            { return &LiteralExpr{Kind: "BOOL", Value: strconv.FormatBool(v)} }
func ident(n string) Expr            { return &IdentifierExpr{Name: n} }
func member(r Expr, m string) Expr   { return &MemberAccessExpr{Receiver: r, Member: m} }
func call(f Expr, args ...Expr) Expr { return &CallExpr{Function: f, Args: args} }
func and(l, r Expr) Expr             { return &AndExpr{Left: l, Right: r} }
func parallel(l, r Expr) Expr        { return &ParallelExpr{Left: l, Right: r} }
func internalCall(fName string, args ...Expr) Expr {
	return &InternalCallExpr{FuncName: fName, Args: args}
}

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

// --- Test AndExpr & Reduction (Phase 3) ---
// --- Test CallExpr (Method Calls - Phase 4) ---

func TestInterpreter_Eval_CallExpr_DiskRead(t *testing.T) {
	interp := NewInterpreter(10)
	// Setup environment with a disk instance
	disk := components.NewDisk("ssd1") // Uses SSD profile
	interp.Env().Set("myDisk", disk)

	// AST for myDisk.Read()
	call := &CallExpr{
		Function: &MemberAccessExpr{
			Receiver: &IdentifierExpr{Name: "myDisk"},
			Member:   "Read",
		},
		Args: []Expr{}, // No arguments for Read()
	}

	_, err := interp.Eval(call)
	if err != nil {
		t.Fatalf("Eval(CallExpr Disk.Read) failed: %v", err)
	}

	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	// Check if the result is the correct outcome object (pointer comparison)
	expectedOutcome := disk.Read()
	if result != expectedOutcome {
		t.Errorf("Expected myDisk.Read() to return the disk's ReadOutcomes object, got different object")
	}
}

func TestInterpreter_Eval_CallExpr_MethodNotFound(t *testing.T) {
	interp := NewInterpreter(10)
	disk := components.NewDisk("ssd1")
	interp.Env().Set("myDisk", disk)

	// AST for myDisk.NoSuchMethod()
	call := &CallExpr{
		Function: &MemberAccessExpr{
			Receiver: &IdentifierExpr{Name: "myDisk"},
			Member:   "NoSuchMethod",
		},
		Args: []Expr{},
	}

	_, err := interp.Eval(call)
	if err == nil {
		t.Fatal("Eval(CallExpr) should have failed for non-existent method")
	}
	if !errors.Is(err, ErrMethodNotFound) {
		t.Errorf("Expected ErrMethodNotFound, got: %v", err)
	}
}

// TODO: Add tests for CallExpr with arguments once argument handling/conversion is implemented.
// TODO: Add tests for CallExpr calling methods that return ast.Node (Phase 5).

// Helper to register SSD Write profile getter
func registerTestDiskWriteFunc(interp *Interpreter) {
	ssd := components.NewDisk("TestSSD")
	interp.RegisterInternalFunc("GetDiskWriteProfile", func(i *Interpreter, args []any) (any, error) {
		return ssd.Write(), nil
	})
}

func TestInterpreter_Eval_AndExpr_AccessResult(t *testing.T) {
	interp := NewInterpreter(10)      // Max 10 buckets
	registerTestDiskFuncs(interp)     // Registers GetDiskReadProfile
	registerTestDiskWriteFunc(interp) // Registers GetDiskWriteProfile

	// AST for GetDiskReadProfile() THEN GetDiskWriteProfile()
	readCall := &InternalCallExpr{FuncName: "GetDiskReadProfile"}
	writeCall := &InternalCallExpr{FuncName: "GetDiskWriteProfile"}
	andExpr := &AndExpr{Left: readCall, Right: writeCall}

	_, err := interp.Eval(andExpr)
	if err != nil {
		t.Fatalf("Eval(AndExpr) failed: %v", err)
	}

	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	outcome, ok := result.(*core.Outcomes[core.AccessResult])
	if !ok {
		t.Fatalf("Expected result type *core.Outcomes[AccessResult], got %T", result)
	}

	// Basic checks on the combined outcome
	if outcome.Len() == 0 {
		t.Error("Combined outcome should not be empty")
	}
	// Allow slightly more tolerance for weight due to potential FP errors during combine/trim
	if !core.ApproxEq(outcome.TotalWeight(), 1.0, 1e-5) {
		t.Errorf("Expected total weight ~1.0, got %f", outcome.TotalWeight())
	}
	// After splitting, trimming each part to maxLen, and appending, the max possible is 2*maxLen
	maxExpectedLen := 2 * interp.maxOutcomeLen
	if outcome.Len() > maxExpectedLen {
		t.Errorf("Expected outcome length <= %d (2 * maxLen) after reduction, got %d", maxExpectedLen, outcome.Len())
	}

	// Check if latency looks reasonable (sum of means)
	readOutcome := components.NewDisk("").Read()
	writeOutcome := components.NewDisk("").Write()
	expectedMean := core.MeanLatency(readOutcome) + core.MeanLatency(writeOutcome)
	actualMean := core.MeanLatency(outcome)
	// Allow larger tolerance due to reduction possibly shifting the mean
	if !core.ApproxEq(actualMean, expectedMean, expectedMean*0.2) {
		t.Errorf("Mean latency %.6f differs significantly from expected sum %.6f", actualMean, expectedMean)
	}
	t.Logf("AndExpr Test: Final Len=%d, Mean Latency=%.6fs (Expected Sum ~%.6fs)", outcome.Len(), actualMean, expectedMean)
}

func TestInterpreter_Eval_AndExpr_TypeMismatch(t *testing.T) {
	interp := NewInterpreter(10)
	registerTestDiskFuncs(interp) // GetDiskReadProfile returns *Outcomes[AccessResult]

	// AST for GetDiskReadProfile() THEN 123
	readCall := &InternalCallExpr{FuncName: "GetDiskReadProfile"}
	literalInt := &LiteralExpr{Kind: "INT", Value: "123"}
	andExpr := &AndExpr{Left: readCall, Right: literalInt}

	_, err := interp.Eval(andExpr)
	if err == nil {
		t.Fatalf("Eval(AndExpr) should have failed for type mismatch")
	}
	// Check the error type
	if !errors.Is(err, ErrTypeMismatch) && !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("Expected ErrTypeMismatch or ErrUnsupportedType, got: %v", err)
	}
}

func TestInterpreter_Eval_CallExpr_RecursiveAST(t *testing.T) {
	interp := NewInterpreter(10)
	registerTestDiskFuncs(interp) // Need GetDiskReadProfile internal func

	// Setup environment with a *declarative* disk instance
	// The actual Go disk component isn't directly needed in the env,
	// as the call resolves to the Disk.Read() method which returns an AST.
	declDisk := NewDisk("ssd_decl", "SSD")
	interp.Env().Set("myDeclDisk", declDisk)

	// AST for myDeclDisk.Read()
	call := &CallExpr{
		Function: &MemberAccessExpr{
			Receiver: &IdentifierExpr{Name: "myDeclDisk"},
			Member:   "Read", // This call returns an *InternalCallExpr AST node
		},
		Args: []Expr{},
	}

	_, err := interp.Eval(call) // This should trigger recursive evaluation
	if err != nil {
		t.Fatalf("Eval(CallExpr Disk.Read) failed: %v", err)
	}

	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed after recursive eval: %v", err)
	}

	// The final result should be the outcome from evaluating the *returned* AST
	// (which was InternalCallExpr("GetDiskReadProfile", "SSD"))
	// So, it should be the same as calling GetDiskReadProfile directly.
	expectedOutcome := components.NewDisk("").Read() // Get the reference SSD read outcome
	if result != expectedOutcome {                   // Compare pointers
		t.Errorf("Expected recursive call to yield SSD Read profile, got different object")
	}
}

// --- Test RepeatExpr (Sequential - Phase 6) ---

func TestInterpreter_Eval_RepeatExpr_Sequential(t *testing.T) {
	interp := NewInterpreter(10)  // Reduce aggressively for testing
	registerTestDiskFuncs(interp) // GetDiskReadProfile returns *Outcomes[AccessResult]

	// AST for repeat(GetDiskReadProfile(), 3, Sequential)
	repeatCount := int64(3)
	repeatExpr := &RepeatExpr{
		Input: &InternalCallExpr{FuncName: "GetDiskReadProfile"},
		Count: &LiteralExpr{Kind: "INT", Value: fmt.Sprintf("%d", repeatCount)}, // Count must be deterministic Outcomes[int64]
		Mode:  Sequential,
	}

	_, err := interp.Eval(repeatExpr)
	if err != nil {
		t.Fatalf("Eval(RepeatExpr) failed: %v", err)
	}

	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	outcome, ok := result.(*core.Outcomes[core.AccessResult])
	if !ok {
		t.Fatalf("Expected result type *core.Outcomes[AccessResult], got %T", result)
	}

	// Check length (should be <= 2 * maxLen due to split/trim)
	maxExpectedLen := 2 * interp.maxOutcomeLen
	if outcome.Len() > maxExpectedLen {
		t.Errorf("Expected outcome length <= %d after reduction, got %d", maxExpectedLen, outcome.Len())
	}
	if outcome.Len() == 0 {
		t.Error("Repeat result outcome should not be empty")
	}

	// Check latency ~ 3 * single read latency
	singleReadOutcome := components.NewDisk("").Read()
	expectedMean := core.MeanLatency(singleReadOutcome) * float64(repeatCount)
	actualMean := core.MeanLatency(outcome)
	// Allow larger tolerance due to potential repeated reduction
	if !core.ApproxEq(actualMean, expectedMean, expectedMean*0.25) {
		t.Errorf("Mean latency %.6f differs significantly from expected sum %.6f (Count: %d)", actualMean, expectedMean, repeatCount)
	}

	// Check availability approx = single avail ^ count
	expectedAvail := math.Pow(core.Availability(singleReadOutcome), float64(repeatCount))
	actualAvail := core.Availability(outcome)
	if !core.ApproxEq(actualAvail, expectedAvail, 0.01) { // Allow 1% tolerance
		t.Errorf("Availability %.6f differs significantly from expected %.6f (Count: %d)", actualAvail, expectedAvail, repeatCount)
	}

	t.Logf("RepeatExpr Test (Count=%d): Final Len=%d, Mean Latency=%.6fs (Expected ~%.6fs), Avail=%.6f (Expected ~%.6f)",
		repeatCount, outcome.Len(), actualMean, expectedMean, actualAvail, expectedAvail)
}

func TestInterpreter_Eval_RepeatExpr_InvalidCount(t *testing.T) {
	interp := NewInterpreter(10)
	// AST for repeat(GetDiskReadProfile(), -1, Sequential)
	repeatExpr := &RepeatExpr{
		Input: &InternalCallExpr{FuncName: "GetDiskReadProfile"},
		Count: &LiteralExpr{Kind: "INT", Value: "-1"},
		Mode:  Sequential,
	}
	_, err := interp.Eval(repeatExpr)
	if err == nil {
		t.Fatal("Eval(RepeatExpr) should fail for negative count")
	}
	if !errors.Is(err, ErrInvalidRepeatCount) {
		t.Errorf("Expected ErrInvalidRepeatCount, got %v", err)
	}
}

// TODO: Add test for RepeatExpr Count=0

// TODO: Add a test that explicitly triggers reduction.
// This might require setting maxOutcomeLen very low (e.g., 2) and using
// inputs known to produce more buckets, or creating a helper internal func
// that returns an outcome with many buckets. Example:
// func TestInterpreter_Eval_AndExpr_ReductionTriggered(t *testing.T) { ... }

// --- Test Statements (Phase 7) ---

func TestInterpreter_Eval_AssignmentStmt(t *testing.T) {
	interp := NewInterpreter(10)
	registerTestDiskFuncs(interp)

	// AST for: myRead = GetDiskReadProfile()
	assignStmt := &AssignmentStmt{
		Variable: &IdentifierExpr{Name: "myRead"},
		Value:    &InternalCallExpr{FuncName: "GetDiskReadProfile"},
	}

	// Eval the assignment (result should be left on stack AND stored in env)
	err := interp.evalAssignmentStmt(assignStmt) // Use specific eval func for test setup ease
	if err != nil {
		t.Fatalf("evalAssignmentStmt failed: %v", err)
	}

	// Check environment
	envVal, ok := interp.Env().Get("myRead")
	if !ok {
		t.Fatal("'myRead' not found in environment after assignment")
	}
	expectedOutcome := components.NewDisk("").Read()
	if envVal != expectedOutcome {
		t.Errorf("Environment value mismatch. Expected SSD Read Profile, got %T", envVal)
	}

	// Check stack (should have the value)
	stackVal, err := interp.pop() // Pop the value evalAssignmentStmt should have left
	if err != nil {
		t.Fatalf("Stack pop failed after assignment: %v", err)
	}
	if stackVal != expectedOutcome {
		t.Errorf("Stack value mismatch. Expected SSD Read Profile, got %T", stackVal)
	}
	if len(interp.stack) != 0 {
		t.Errorf("Stack should be empty after popping assignment result, len=%d", len(interp.stack))
	}
}

func TestInterpreter_Eval_ReturnStmt(t *testing.T) {
	interp := NewInterpreter(10)
	registerTestDiskFuncs(interp)

	// AST for: return GetDiskReadProfile()
	returnStmt := &ReturnStmt{
		ReturnValue: &InternalCallExpr{FuncName: "GetDiskReadProfile"},
	}

	// Eval the return statement
	err := interp.evalReturnStmt(returnStmt)
	if err == nil {
		t.Fatal("evalReturnStmt should have returned a ReturnValue error")
	}

	// Check if the error is the special ReturnValue wrapper
	var retVal *ReturnValue
	if !errors.As(err, &retVal) {
		t.Fatalf("Expected a ReturnValue error, got: %v", err)
	}

	// Check the value *inside* the ReturnValue wrapper
	expectedOutcome := components.NewDisk("").Read()
	if retVal.Value != expectedOutcome {
		t.Errorf("ReturnValue contained unexpected value. Expected SSD Read Profile, got %T", retVal.Value)
	}

	// Check stack (should contain the return value, as Eval was called internally)
	stackVal, stackErr := interp.pop()
	if stackErr != nil || stackVal != expectedOutcome {
		t.Errorf("Stack check failed after return. Err: %v, Val: %T", stackErr, stackVal)
	}
}

func TestInterpreter_Eval_BlockStmt_Sequence(t *testing.T) {
	interp := NewInterpreter(10)
	registerTestDiskFuncs(interp)
	registerTestDiskWriteFunc(interp)

	// AST for: { readRes = GetDiskReadProfile(); GetDiskWriteProfile() }
	// Implicit return of the And(readRes, writeProfile)
	block := &BlockStmt{
		Statements: []Stmt{
			&AssignmentStmt{
				Variable: &IdentifierExpr{Name: "readRes"},
				Value:    &InternalCallExpr{FuncName: "GetDiskReadProfile"},
			},
			&ExprStmt{ // The write call is just an expression statement
				Expression: &InternalCallExpr{FuncName: "GetDiskWriteProfile"},
			},
		},
	}

	// Eval the block
	blockResult, err := interp.evalBlockStmt(block, interp.Env(), nil) // Eval in current env
	if err != nil {
		t.Fatalf("evalBlockStmt failed: %v", err)
	}
	// The result should be the combined outcome from the implicit AND
	// Verify type and basic properties (similar to AndExpr test)
	outcome, ok := blockResult.(*core.Outcomes[core.AccessResult])
	if !ok {
		t.Fatalf("Expected block result type *core.Outcomes[AccessResult], got %T", blockResult)
	}
	if outcome.Len() == 0 || outcome.Len() > 2*interp.maxOutcomeLen {
		t.Errorf("Unexpected outcome length %d for block result", outcome.Len())
	}
	if !core.ApproxEq(outcome.TotalWeight(), 1.0, 1e-5) {
		t.Errorf("Expected total weight ~1.0, got %f", outcome.TotalWeight())
	}

	// Check env to ensure assignment happened
	_, ok = interp.Env().Get("readRes")
	if !ok {
		t.Error("'readRes' not found in environment after block execution")
	}

	// Stack should be empty after block evaluation returns its result directly
	if len(interp.stack) != 0 {
		t.Errorf("Stack should be empty after evalBlockStmt returns, len=%d", len(interp.stack))
	}
}

// TODO: Test block with ReturnStmt in the middle.

// --- Test If Stmt (Phase 8) ---

// Mock component for If test
type MockConditional struct{}

func (m *MockConditional) ProbCheck(probSuccess float64) *core.Outcomes[bool] {
	if probSuccess < 0 {
		probSuccess = 0
	}
	if probSuccess > 1 {
		probSuccess = 1
	}
	o := &core.Outcomes[bool]{And: func(a, b bool) bool { return a && b }}
	if probSuccess > 1e-9 {
		o.Add(probSuccess, true)
	}
	if (1.0 - probSuccess) > 1e-9 {
		o.Add(1.0-probSuccess, false)
	}
	return o
}
func (m *MockConditional) OpSuccess() *core.Outcomes[core.AccessResult] {
	o := (&core.Outcomes[core.AccessResult]{}).Add(1.0, core.AccessResult{Success: true, Latency: core.Millis(10)})
	o.And = core.AndAccessResults
	return o
}
func (m *MockConditional) OpFailure() *core.Outcomes[core.AccessResult] {
	o := (&core.Outcomes[core.AccessResult]{}).Add(1.0, core.AccessResult{Success: false, Latency: core.Millis(5)})
	o.And = core.AndAccessResults
	return o
}

func TestInterpreter_Eval_IfStmt_AccessResultSuccess(t *testing.T) {
	interp := NewInterpreter(10)
	// Setup: condVar = Outcome[AccessResult] (mostly success)
	condOutcome := (&core.Outcomes[core.AccessResult]{And: core.AndAccessResults}).
		Add(0.8, core.AccessResult{Success: true, Latency: core.Millis(1)}).
		Add(0.2, core.AccessResult{Success: false, Latency: core.Millis(2)})
	interp.Env().Set("condVar", condOutcome)
	// Mock component providing Then/Else operations
	mockComp := &MockConditional{}
	interp.Env().Set("mock", mockComp)

	// AST for: if condVar.Success { mock.OpSuccess() } else { mock.OpFailure() }

	ifStmt := &IfStmt{
		Condition: &MemberAccessExpr{Receiver: &IdentifierExpr{Name: "condVar"}, Member: "Success"}, // Use the member access
		Then: &BlockStmt{Statements: []Stmt{
			&ExprStmt{Expression: call(member(ident("mock"), "OpSuccess"))},
		}},
		Else: &BlockStmt{Statements: []Stmt{
			&ExprStmt{Expression: call(member(ident("mock"), "OpFailure"))},
		}},
	}

	// Eval the If statement - result pushed onto stack
	err := interp.evalIfStmt(ifStmt) // Use specific eval for testing
	if err != nil {
		t.Fatalf("evalIfStmt failed: %v", err)
	}

	result, err := interp.GetFinalResult() // Get combined result from stack
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	finalOutcome, ok := result.(*core.Outcomes[core.AccessResult])
	if !ok {
		t.Fatalf("Expected IfStmt result *core.Outcomes[AccessResult], got %T", result)
	}

	// Analyze the combined result
	if finalOutcome.Len() == 0 || finalOutcome.Len() > 2*interp.maxOutcomeLen { // Allow up to 2*maxLen
		t.Errorf("Unexpected outcome length %d", finalOutcome.Len())
	}

	// Expected availability should be close to the condition's success probability (0.8)
	// because OpSuccess always succeeds and OpFailure always fails.
	expectedAvail := 0.8
	actualAvail := core.Availability(finalOutcome)
	if !core.ApproxEq(actualAvail, expectedAvail, 0.01) {
		t.Errorf("Expected availability ~%.2f, got %.6f", expectedAvail, actualAvail)
	}

	// Check for presence of both latencies
	foundThenLat := false
	foundElseLat := false
	for _, b := range finalOutcome.Buckets {
		if core.ApproxEq(b.Value.Latency, core.Millis(10), 1e-9) {
			foundThenLat = true
		}
		if core.ApproxEq(b.Value.Latency, core.Millis(5), 1e-9) {
			foundElseLat = true
		}
	}
	if !foundThenLat {
		t.Error("Did not find expected 'then' latency (10ms) in combined outcome")
	}
	if !foundElseLat {
		t.Error("Did not find expected 'else' latency (5ms) in combined outcome")
	}

	t.Logf("IfStmt Test: Final Len=%d, Avail=%.6f", finalOutcome.Len(), actualAvail)
}

// --- Test CallExpr with Arguments (Phase 9 / now) ---

func TestInterpreter_Eval_CallExpr_DiskReadProcessWrite(t *testing.T) {
	interp := NewInterpreter(15)
	// Env setup
	disk := components.NewDisk("ssd_rpw")
	interp.Env().Set("myDiskRPW", disk)
	// Processing time outcome (must be deterministic for arg)
	procTimeVal := core.Millis(2) // 2ms
	procTimeOutcome := (&core.Outcomes[core.Duration]{}).Add(1.0, procTimeVal)
	interp.Env().Set("processingDuration", procTimeOutcome) // Store the outcome

	// AST for myDiskRPW.ReadProcessWrite(processingDuration)
	call := &CallExpr{
		Function: &MemberAccessExpr{
			Receiver: &IdentifierExpr{Name: "myDiskRPW"},
			Member:   "ReadProcessWrite",
		},
		Args: []Expr{
			&IdentifierExpr{Name: "processingDuration"}, // Pass the outcome variable
		},
	}

	_, err := interp.Eval(call)
	if err != nil {
		t.Fatalf("Eval(CallExpr Disk.ReadProcessWrite) failed: %v", err)
	}

	result, err := interp.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	outcome, ok := result.(*core.Outcomes[core.AccessResult])
	if !ok {
		t.Fatalf("Expected result type *core.Outcomes[AccessResult], got %T", result)
	}

	// Check combined result (approx)
	baseRead := disk.Read()
	baseWrite := disk.Write()
	expectedMean := core.MeanLatency(baseRead) + core.MeanLatency(baseWrite) + procTimeVal
	actualMean := core.MeanLatency(outcome)
	if !core.ApproxEq(actualMean, expectedMean, expectedMean*0.15) { // Allow tolerance
		t.Errorf("RPW Mean Latency %.6f differs from expected %.6f", actualMean, expectedMean)
	}
	expectedAvail := core.Availability(baseRead) * core.Availability(baseWrite)
	actualAvail := core.Availability(outcome)
	if !core.ApproxEq(actualAvail, expectedAvail, 0.001) {
		t.Errorf("RPW Availability %.6f differs from expected %.6f", actualAvail, expectedAvail)
	}
}

func TestInterpreter_Eval_CallExpr_NonDeterministicArg(t *testing.T) {
	interp := NewInterpreter(15)
	disk := components.NewDisk("ssd_rpw_err")
	interp.Env().Set("myDiskErr", disk)
	// Non-deterministic outcome for argument
	nonDetOutcome := (&core.Outcomes[core.Duration]{}).Add(0.5, core.Millis(1)).Add(0.5, core.Millis(3))
	interp.Env().Set("nonDetDuration", nonDetOutcome)

	// AST for myDiskErr.ReadProcessWrite(nonDetDuration)
	call := &CallExpr{
		Function: &MemberAccessExpr{Receiver: &IdentifierExpr{Name: "myDiskErr"}, Member: "ReadProcessWrite"},
		Args:     []Expr{&IdentifierExpr{Name: "nonDetDuration"}},
	}
	_, err := interp.Eval(call)
	if err == nil {
		t.Fatal("Eval call with non-deterministic argument should have failed")
	}
	if !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("Expected ErrInvalidArgument for non-deterministic arg, got: %v", err)
	}
}

// TODO: Add test for IfStmt without Else branch.
// TODO: Add test for IfStmt where condition is myVar.Success (requires evalMemberAccess)

// --- Test Top Level Driver (Phase X / now) ---

func TestInterpreter_RunDSL_SimpleSystem(t *testing.T) {
	// Manually construct AST for:
	// system "TestSys" {
	//   instance d1: Disk = { ProfileName = "SSD" };
	//   analyze readPerf = d1.Read();
	//   analyze writePerf = d1.Write();
	// }

	systemAST := &SystemDecl{
		Name: "TestSys",
		Body: []Node{ // Use Node interface
			&InstanceDecl{
				Name:          "d1",
				ComponentType: "Disk", // Matches name registered in registerBuiltinComponents
				Params: []*ParamAssignment{ // Use ParamAssignment slice
					{Name: "ProfileName", Value: &LiteralExpr{Kind: "STRING", Value: "SSD"}},
				},
			},
			&AnalyzeDecl{
				Name: "readPerf",
				Target: &CallExpr{
					Function: &MemberAccessExpr{Receiver: &IdentifierExpr{Name: "d1"}, Member: "Read"},
					Args:     []Expr{},
				},
			},
			&AnalyzeDecl{
				Name: "writePerf",
				Target: &CallExpr{
					Function: &MemberAccessExpr{Receiver: &IdentifierExpr{Name: "d1"}, Member: "Write"},
					Args:     []Expr{},
				},
			},
		},
	}

	results, err := RunDSL(systemAST)
	if err != nil {
		t.Fatalf("RunDSL failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 analysis results, got %d", len(results))
	}

	// Check Read Result
	readResult, okRead := results["readPerf"]
	if !okRead {
		t.Fatal("Analysis result wrapper 'readPerf' not found")
	}
	if readResult.Error != nil {
		t.Errorf("readPerf analysis failed: %v", readResult.Error)
	}
	if !readResult.AnalysisPerformed {
		t.Error("readPerf analysis was not performed (or metrics failed)")
	}
	if _, ok := readResult.Outcome.(*core.Outcomes[core.AccessResult]); !ok {
		t.Errorf("readPerf outcome has wrong type: %T", readResult.Outcome)
	}
	if len(readResult.Metrics) < 4 {
		t.Errorf("Expected at least 4 metrics for readPerf, got %d", len(readResult.Metrics))
	}
	// Use core.MetricType constants as keys
	if readResult.Metrics[core.AvailabilityMetric] <= 0.9 || readResult.Metrics[core.AvailabilityMetric] > 1.0 {
		t.Errorf("readPerf availability %.4f seems wrong", readResult.Metrics[core.AvailabilityMetric])
	}

	// Check Write Result (basic checks using wrapper fields)
	writeResult, okWrite := results["writePerf"]
	if !okWrite {
		t.Fatal("Analysis result wrapper 'writePerf' not found")
	}
	if writeResult.Error != nil {
		t.Errorf("writePerf analysis failed: %v", writeResult.Error)
	}
	if !writeResult.AnalysisPerformed {
		t.Error("writePerf analysis was not performed")
	}
	// Check outcome type
	if _, ok := writeResult.Outcome.(*core.Outcomes[core.AccessResult]); !ok {
		t.Errorf("writePerf outcome has wrong type: %T", writeResult.Outcome)
	}
	if len(writeResult.Metrics) < 4 {
		t.Errorf("Expected at least 4 metrics for writePerf, got %d", len(writeResult.Metrics))
	}

	t.Logf("RunDSL Test Read Metrics: %v", readResult.Metrics)
	t.Logf("RunDSL Test Write Metrics: %v", writeResult.Metrics)
}
