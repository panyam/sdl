// sdl/dsl/vm_test.go
package dsl

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/panyam/leetcoach/sdl/components"
	"github.com/panyam/leetcoach/sdl/core"
	"github.com/panyam/leetcoach/sdl/decl"
	// "fmt" // For debugging
)

func TestVM_NewVM(t *testing.T) {
	v := NewVM(20)
	if v == nil {
		t.Fatal("NewVM returned nil")
	}
	if v.stack == nil {
		t.Error("VM stack is nil")
	}
	if len(v.stack) != 0 {
		t.Errorf("Expected initial stack length 0, got %d", len(v.stack))
	}
	if v.env == nil {
		t.Error("VM environment is nil")
	}
	if len(v.env.store) != 0 {
		t.Error("Expected initial environment store to be empty")
	}
	if v.internalFuncs == nil {
		t.Error("VM internalFuncs map is nil")
	}
	if len(v.internalFuncs) != 0 {
		t.Error("Expected initial internalFuncs map to be empty")
	}
	if v.maxOutcomeLen != 20 {
		t.Errorf("Expected maxOutcomeLen 20, got %d", v.maxOutcomeLen)
	}
}

// --- Test Literals (Phase 1) ---

func TestVM_Eval_LiteralInt(t *testing.T) {
	v := NewVM(10)
	literal := &decl.LiteralExpr{Kind: "INT", Value: "42"}

	_, err := v.Eval(literal)
	if err != nil {
		t.Fatalf("Eval(LiteralInt) failed: %v", err)
	}

	result, err := v.GetFinalResult()
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

func TestVM_Eval_LiteralFloat(t *testing.T) {
	v := NewVM(10)
	literal := &decl.LiteralExpr{Kind: "FLOAT", Value: "3.14"}

	_, err := v.Eval(literal)
	if err != nil {
		t.Fatalf("Eval(LiteralFloat) failed: %v", err)
	}
	result, err := v.GetFinalResult()
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

func TestVM_Eval_LiteralString(t *testing.T) {
	v := NewVM(10)
	literal := &decl.LiteralExpr{Kind: "STRING", Value: "hello"} // Note: No quotes in Value field

	_, err := v.Eval(literal)
	if err != nil {
		t.Fatalf("Eval(LiteralString) failed: %v", err)
	}
	result, err := v.GetFinalResult()
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

func TestVM_Eval_LiteralBool(t *testing.T) {
	v := NewVM(10)
	literal := &decl.LiteralExpr{Kind: "BOOL", Value: "true"}

	_, err := v.Eval(literal)
	if err != nil {
		t.Fatalf("Eval(LiteralBool) failed: %v", err)
	}
	result, err := v.GetFinalResult()
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

func TestVM_Eval_Identifier(t *testing.T) {
	v := NewVM(10)
	// Setup environment
	testOutcome := (&core.Outcomes[int64]{}).Add(1.0, 99)
	v.Env().Set("myVar", testOutcome)

	ident := &IdentifierExpr{Name: "myVar"}
	_, err := v.Eval(ident)
	if err != nil {
		t.Fatalf("Eval(IdentifierExpr) failed: %v", err)
	}

	result, err := v.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	if result != testOutcome { // Compare pointers for this test
		t.Errorf("Expected identifier to resolve to the set outcome object, got different object")
	}
}

func TestVM_Eval_Identifier_NotFound(t *testing.T) {
	v := NewVM(10)
	ident := &IdentifierExpr{Name: "missingVar"}

	_, err := v.Eval(ident)
	if err == nil {
		t.Fatal("Eval(IdentifierExpr) should have failed for missing variable")
	}

	// Check if the error is the expected ErrNotFound type (or wraps it)
	if !errors.Is(err, ErrNotFound) { // Use errors.Is for potential wrapping
		t.Errorf("Expected ErrNotFound, got: %v", err)
	}
}

// Helper function to register a dummy SSD read profile getter
func registerTestDiskFuncs(v *VM) {
	// Ensure Disk component is initialized to provide the outcomes
	ssd := components.NewDisk("TestSSD") // Uses SSD profile by default

	v.RegisterInternalFunc("GetDiskReadProfile", func(v *VM, args []any) (any, error) {
		// Expects one string arg: profile name (we ignore it here for simplicity)
		// if len(args) != 1 { return nil, fmt.Errorf("GetDiskReadProfile expects 1 arg") }
		// _, ok := args[0].(string); if !ok { return nil, fmt.Errorf("arg 0 must be string profile name") }

		return ssd.Read(), nil // Return the pre-calculated outcome object
	})
}

func TestVM_Eval_InternalCall(t *testing.T) {
	v := NewVM(10)
	registerTestDiskFuncs(v)

	// AST for Internal.GetDiskReadProfile("SSD") - Args aren't used yet by dummy func
	call := &Internaldecl.CallExpr{FuncName: "GetDiskReadProfile", Args: []Expr{&decl.LiteralExpr{Kind: "STRING", Value: "SSD"}}}

	_, err := v.Eval(call)
	if err != nil {
		t.Fatalf("Eval(Internaldecl.CallExpr) failed: %v", err)
	}

	result, err := v.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	// Check if the result is the expected outcome object
	expectedOutcome := components.NewDisk("").Read() // Get the reference SSD read outcome
	if result != expectedOutcome {                   // Compare pointers
		t.Errorf("Expected internal call to return SSD Read profile, got different object")
	}
}

func TestVM_Eval_InternalCall_NotFound(t *testing.T) {
	v := NewVM(10)
	call := &Internaldecl.CallExpr{FuncName: "NoSuchFunction"}
	_, err := v.Eval(call)
	if err == nil {
		t.Fatal("Eval(Internaldecl.CallExpr) should fail for unregistered function")
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

func TestVM_Stack(t *testing.T) {
	v := NewVM(10)

	_, err := v.pop()
	if err == nil || err != ErrStackUnderflow {
		t.Errorf("Expected ErrStackUnderflow on empty pop, got %v", err)
	}

	v.push(1)
	v.push("two")

	if len(v.stack) != 2 {
		t.Fatalf("Expected stack length 2, got %d", len(v.stack))
	}

	top, ok := v.peek()
	if !ok || top != "two" {
		t.Errorf("Peek failed: expected 'two', got %v (%t)", top, ok)
	}

	val2, err2 := v.pop()
	if err2 != nil {
		t.Fatalf("Error popping 'two': %v", err2)
	}
	if val2 != "two" {
		t.Errorf("Expected popped value 'two', got %v", val2)
	}
	if len(v.stack) != 1 {
		t.Errorf("Expected stack length 1 after pop, got %d", len(v.stack))
	}

	val1, err1 := v.pop()
	if err1 != nil {
		t.Fatalf("Error popping 1: %v", err1)
	}
	if val1 != 1 {
		t.Errorf("Expected popped value 1, got %v", val1)
	}
	if len(v.stack) != 0 {
		t.Errorf("Expected stack length 0 after popping all, got %d", len(v.stack))
	}

	_, errEmpty := v.pop()
	if errEmpty == nil || errEmpty != ErrStackUnderflow {
		t.Errorf("Expected ErrStackUnderflow on final empty pop, got %v", errEmpty)
	}
}

func TestVM_RegisterInternalFunc(t *testing.T) {
	v := NewVM(10)
	testFunc := func(v *VM, args []any) (any, error) {
		return "test success", nil
	}

	v.RegisterInternalFunc("myTestFunc", testFunc)

	fn, ok := v.internalFuncs["myTestFunc"]
	if !ok {
		t.Fatal("Failed to register internal function")
	}
	if fn == nil { // Check if function pointer is actually stored
		t.Fatal("Registered function is nil")
	}
	// Note: Comparing function pointers directly can be tricky,
	// but checking presence and non-nil is a good start.
}

func TestVM_Eval_Stub(t *testing.T) {
	v := NewVM(10)
	// Use an AST node for which Eval is not implemented yet
	node := &decl.MemberAccessExpr{}
	_, err := v.Eval(node)
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
// --- Test decl.CallExpr (Method Calls - Phase 4) ---

func TestVM_Eval_CallExpr_DiskRead(t *testing.T) {
	v := NewVM(10)
	// Setup environment with a disk instance
	disk := components.NewDisk("ssd1") // Uses SSD profile
	v.Env().Set("myDisk", disk)

	// AST for myDisk.Read()
	call := &decl.CallExpr{
		Function: &decl.MemberAccessExpr{
			Receiver: &IdentifierExpr{Name: "myDisk"},
			Member:   "Read",
		},
		Args: []Expr{}, // No arguments for Read()
	}

	_, err := v.Eval(call)
	if err != nil {
		t.Fatalf("Eval(decl.CallExpr Disk.Read) failed: %v", err)
	}

	result, err := v.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	// Check if the result is the correct outcome object (pointer comparison)
	expectedOutcome := disk.Read()
	if result != expectedOutcome {
		t.Errorf("Expected myDisk.Read() to return the disk's ReadOutcomes object, got different object")
	}
}

func TestVM_Eval_CallExpr_MethodNotFound(t *testing.T) {
	v := NewVM(10)
	disk := components.NewDisk("ssd1")
	v.Env().Set("myDisk", disk)

	// AST for myDisk.NoSuchMethod()
	call := &decl.CallExpr{
		Function: &decl.MemberAccessExpr{
			Receiver: &IdentifierExpr{Name: "myDisk"},
			Member:   "NoSuchMethod",
		},
		Args: []Expr{},
	}

	_, err := v.Eval(call)
	if err == nil {
		t.Fatal("Eval(decl.CallExpr) should have failed for non-existent method")
	}
	if !errors.Is(err, ErrMethodNotFound) {
		t.Errorf("Expected ErrMethodNotFound, got: %v", err)
	}
}

// TODO: Add tests for decl.CallExpr with arguments once argument handling/conversion is implemented.
// TODO: Add tests for decl.CallExpr calling methods that return ast.Node (Phase 5).

// Helper to register SSD Write profile getter
func registerTestDiskWriteFunc(v *VM) {
	ssd := components.NewDisk("TestSSD")
	v.RegisterInternalFunc("GetDiskWriteProfile", func(v *VM, args []any) (any, error) {
		return ssd.Write(), nil
	})
}

func TestVM_Eval_AndExpr_AccessResult(t *testing.T) {
	v := NewVM(10)               // Max 10 buckets
	registerTestDiskFuncs(v)     // Registers GetDiskReadProfile
	registerTestDiskWriteFunc(v) // Registers GetDiskWriteProfile

	// AST for GetDiskReadProfile() THEN GetDiskWriteProfile()
	readCall := &Internaldecl.CallExpr{FuncName: "GetDiskReadProfile"}
	writeCall := &Internaldecl.CallExpr{FuncName: "GetDiskWriteProfile"}
	andExpr := &AndExpr{Left: readCall, Right: writeCall}

	_, err := v.Eval(andExpr)
	if err != nil {
		t.Fatalf("Eval(AndExpr) failed: %v", err)
	}

	result, err := v.GetFinalResult()
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
	maxExpectedLen := 2 * v.maxOutcomeLen
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

func TestVM_Eval_AndExpr_TypeMismatch(t *testing.T) {
	v := NewVM(10)
	registerTestDiskFuncs(v) // GetDiskReadProfile returns *Outcomes[AccessResult]

	// AST for GetDiskReadProfile() THEN 123
	readCall := &Internaldecl.CallExpr{FuncName: "GetDiskReadProfile"}
	literalInt := &decl.LiteralExpr{Kind: "INT", Value: "123"}
	andExpr := &AndExpr{Left: readCall, Right: literalInt}

	_, err := v.Eval(andExpr)
	if err == nil {
		t.Fatalf("Eval(AndExpr) should have failed for type mismatch")
	}
	// Check the error type
	if !errors.Is(err, ErrTypeMismatch) && !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("Expected ErrTypeMismatch or ErrUnsupportedType, got: %v", err)
	}
}

func TestVM_Eval_CallExpr_RecursiveAST(t *testing.T) {
	v := NewVM(10)
	registerTestDiskFuncs(v) // Need GetDiskReadProfile internal func

	// Setup environment with a *declarative* disk instance
	// The actual Go disk component isn't directly needed in the env,
	// as the call resolves to the Disk.Read() method which returns an AST.
	declDisk := NewDisk("ssd_decl", "SSD")
	v.Env().Set("myDisk", declDisk)

	// AST for myDisk.Read()
	call := &decl.CallExpr{
		Function: &decl.MemberAccessExpr{
			Receiver: &IdentifierExpr{Name: "myDisk"},
			Member:   "Read", // This call returns an *Internaldecl.CallExpr AST node
		},
		Args: []Expr{},
	}

	_, err := v.Eval(call) // This should trigger recursive evaluation
	if err != nil {
		t.Fatalf("Eval(decl.CallExpr Disk.Read) failed: %v", err)
	}

	result, err := v.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed after recursive eval: %v", err)
	}

	// The final result should be the outcome from evaluating the *returned* AST
	// (which was Internaldecl.CallExpr("GetDiskReadProfile", "SSD"))
	// So, it should be the same as calling GetDiskReadProfile directly.
	expectedOutcome := components.NewDisk("").Read() // Get the reference SSD read outcome
	if result != expectedOutcome {                   // Compare pointers
		t.Errorf("Expected recursive call to yield SSD Read profile, got different object")
	}
}

// --- Test RepeatExpr (Sequential - Phase 6) ---

func TestVM_Eval_RepeatExpr_Sequential(t *testing.T) {
	v := NewVM(10)           // Reduce aggressively for testing
	registerTestDiskFuncs(v) // GetDiskReadProfile returns *Outcomes[AccessResult]

	// AST for repeat(GetDiskReadProfile(), 3, Sequential)
	repeatCount := int64(3)
	repeatExpr := &RepeatExpr{
		Input: &Internaldecl.CallExpr{FuncName: "GetDiskReadProfile"},
		Count: &decl.LiteralExpr{Kind: "INT", Value: fmt.Sprintf("%d", repeatCount)}, // Count must be deterministic Outcomes[int64]
		Mode:  Sequential,
	}

	_, err := v.Eval(repeatExpr)
	if err != nil {
		t.Fatalf("Eval(RepeatExpr) failed: %v", err)
	}

	result, err := v.GetFinalResult()
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	outcome, ok := result.(*core.Outcomes[core.AccessResult])
	if !ok {
		t.Fatalf("Expected result type *core.Outcomes[AccessResult], got %T", result)
	}

	// Check length (should be <= 2 * maxLen due to split/trim)
	maxExpectedLen := 2 * v.maxOutcomeLen
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

func TestVM_Eval_RepeatExpr_InvalidCount(t *testing.T) {
	v := NewVM(10)
	// AST for repeat(GetDiskReadProfile(), -1, Sequential)
	repeatExpr := &RepeatExpr{
		Input: &Internaldecl.CallExpr{FuncName: "GetDiskReadProfile"},
		Count: &decl.LiteralExpr{Kind: "INT", Value: "-1"},
		Mode:  Sequential,
	}
	_, err := v.Eval(repeatExpr)
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
// func TestVM_Eval_AndExpr_ReductionTriggered(t *testing.T) { ... }

// --- Test Statements (Phase 7) ---

func TestVM_Eval_AssignmentStmt(t *testing.T) {
	v := NewVM(10)
	registerTestDiskFuncs(v)

	// AST for: myRead = GetDiskReadProfile()
	assignStmt := &AssignmentStmt{
		Variable: &IdentifierExpr{Name: "myRead"},
		Value:    &Internaldecl.CallExpr{FuncName: "GetDiskReadProfile"},
	}

	// Eval the assignment (result should be left on stack AND stored in env)
	err := v.evalAssignmentStmt(assignStmt) // Use specific eval func for test setup ease
	if err != nil {
		t.Fatalf("evalAssignmentStmt failed: %v", err)
	}

	// Check environment
	envVal, ok := v.Env().Get("myRead")
	if !ok {
		t.Fatal("'myRead' not found in environment after assignment")
	}
	expectedOutcome := components.NewDisk("").Read()
	if envVal != expectedOutcome {
		t.Errorf("Environment value mismatch. Expected SSD Read Profile, got %T", envVal)
	}

	// Check stack (should have the value)
	stackVal, err := v.pop() // Pop the value evalAssignmentStmt should have left
	if err != nil {
		t.Fatalf("Stack pop failed after assignment: %v", err)
	}
	if stackVal != expectedOutcome {
		t.Errorf("Stack value mismatch. Expected SSD Read Profile, got %T", stackVal)
	}
	if len(v.stack) != 0 {
		t.Errorf("Stack should be empty after popping assignment result, len=%d", len(v.stack))
	}
}

func TestVM_Eval_ReturnStmt(t *testing.T) {
	v := NewVM(10)
	registerTestDiskFuncs(v)

	// AST for: return GetDiskReadProfile()
	returnStmt := &ReturnStmt{
		ReturnValue: &Internaldecl.CallExpr{FuncName: "GetDiskReadProfile"},
	}

	// Eval the return statement
	err := v.evalReturnStmt(returnStmt)
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
	stackVal, stackErr := v.pop()
	if stackErr != nil || stackVal != expectedOutcome {
		t.Errorf("Stack check failed after return. Err: %v, Val: %T", stackErr, stackVal)
	}
}

func TestVM_Eval_BlockStmt_Sequence(t *testing.T) {
	v := NewVM(10)
	registerTestDiskFuncs(v)
	registerTestDiskWriteFunc(v)

	// AST for: { readRes = GetDiskReadProfile(); GetDiskWriteProfile() }
	// Implicit return of the And(readRes, writeProfile)
	block := &BlockStmt{
		Statements: []Stmt{
			&AssignmentStmt{
				Variable: &IdentifierExpr{Name: "readRes"},
				Value:    &Internaldecl.CallExpr{FuncName: "GetDiskReadProfile"},
			},
			&ExprStmt{ // The write call is just an expression statement
				Expression: &Internaldecl.CallExpr{FuncName: "GetDiskWriteProfile"},
			},
		},
	}

	// Eval the block
	blockResult, err := v.evalBlockStmt(block, v.Env(), nil) // Eval in current env
	if err != nil {
		t.Fatalf("evalBlockStmt failed: %v", err)
	}
	// The result should be the combined outcome from the implicit AND
	// Verify type and basic properties (similar to AndExpr test)
	outcome, ok := blockResult.(*core.Outcomes[core.AccessResult])
	if !ok {
		t.Fatalf("Expected block result type *core.Outcomes[AccessResult], got %T", blockResult)
	}
	if outcome.Len() == 0 || outcome.Len() > 2*v.maxOutcomeLen {
		t.Errorf("Unexpected outcome length %d for block result", outcome.Len())
	}
	if !core.ApproxEq(outcome.TotalWeight(), 1.0, 1e-5) {
		t.Errorf("Expected total weight ~1.0, got %f", outcome.TotalWeight())
	}

	// Check env to ensure assignment happened
	_, ok = v.Env().Get("readRes")
	if !ok {
		t.Error("'readRes' not found in environment after block execution")
	}

	// Stack should be empty after block evaluation returns its result directly
	if len(v.stack) != 0 {
		t.Errorf("Stack should be empty after evalBlockStmt returns, len=%d", len(v.stack))
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

func TestVM_Eval_IfStmt_AccessResultSuccess(t *testing.T) {
	v := NewVM(10)
	// Setup: condVar = Outcome[AccessResult] (mostly success)
	condOutcome := (&core.Outcomes[core.AccessResult]{And: core.AndAccessResults}).
		Add(0.8, core.AccessResult{Success: true, Latency: core.Millis(1)}).
		Add(0.2, core.AccessResult{Success: false, Latency: core.Millis(2)})
	v.Env().Set("condVar", condOutcome)
	// Mock component providing Then/Else operations
	mockComp := &MockConditional{}
	v.Env().Set("mock", mockComp)

	// AST for: if condVar.Success { mock.OpSuccess() } else { mock.OpFailure() }

	ifStmt := &IfStmt{
		Condition: &decl.MemberAccessExpr{Receiver: &IdentifierExpr{Name: "condVar"}, Member: "Success"}, // Use the member access
		Then: &BlockStmt{Statements: []Stmt{
			&ExprStmt{Expression: Call(Member(Ident("mock"), "OpSuccess"))},
		}},
		Else: &BlockStmt{Statements: []Stmt{
			&ExprStmt{Expression: Call(Member(Ident("mock"), "OpFailure"))},
		}},
	}

	// Eval the If statement - result pushed onto stack
	err := v.evalIfStmt(ifStmt) // Use specific eval for testing
	if err != nil {
		t.Fatalf("evalIfStmt failed: %v", err)
	}

	result, err := v.GetFinalResult() // Get combined result from stack
	if err != nil {
		t.Fatalf("GetFinalResult failed: %v", err)
	}

	finalOutcome, ok := result.(*core.Outcomes[core.AccessResult])
	if !ok {
		t.Fatalf("Expected IfStmt result *core.Outcomes[AccessResult], got %T", result)
	}

	// decl.Analyze the combined result
	if finalOutcome.Len() == 0 || finalOutcome.Len() > 2*v.maxOutcomeLen { // Allow up to 2*maxLen
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

// --- Test decl.CallExpr with Arguments (Phase 9 / now) ---

func TestVM_Eval_CallExpr_DiskReadProcessWrite(t *testing.T) {
	v := NewVM(15)
	// Env setup
	disk := components.NewDisk("ssd_rpw")
	v.Env().Set("myDiskRPW", disk)
	// Processing time outcome (must be deterministic for arg)
	procTimeVal := core.Millis(2) // 2ms
	procTimeOutcome := (&core.Outcomes[core.Duration]{}).Add(1.0, procTimeVal)
	v.Env().Set("processingDuration", procTimeOutcome) // Store the outcome

	// AST for myDiskRPW.ReadProcessWrite(processingDuration)
	call := &decl.CallExpr{
		Function: &decl.MemberAccessExpr{
			Receiver: &IdentifierExpr{Name: "myDiskRPW"},
			Member:   "ReadProcessWrite",
		},
		Args: []Expr{
			&IdentifierExpr{Name: "processingDuration"}, // Pass the outcome variable
		},
	}

	_, err := v.Eval(call)
	if err != nil {
		t.Fatalf("Eval(decl.CallExpr Disk.ReadProcessWrite) failed: %v", err)
	}

	result, err := v.GetFinalResult()
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

func TestVM_Eval_CallExpr_NonDeterministicArg(t *testing.T) {
	v := NewVM(15)
	disk := components.NewDisk("ssd_rpw_err")
	v.Env().Set("myDiskErr", disk)
	// Non-deterministic outcome for argument
	nonDetOutcome := (&core.Outcomes[core.Duration]{}).Add(0.5, core.Millis(1)).Add(0.5, core.Millis(3))
	v.Env().Set("nonDetDuration", nonDetOutcome)

	// AST for myDiskErr.ReadProcessWrite(nonDetDuration)
	call := &decl.CallExpr{
		Function: &decl.MemberAccessExpr{Receiver: &IdentifierExpr{Name: "myDiskErr"}, Member: "ReadProcessWrite"},
		Args:     []Expr{&IdentifierExpr{Name: "nonDetDuration"}},
	}
	_, err := v.Eval(call)
	if err == nil {
		t.Fatal("Eval call with non-deterministic argument should have failed")
	}
	if !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("Expected ErrInvalidArgument for non-deterministic arg, got: %v", err)
	}
}

// TODO: Add test for IfStmt without Else branch.
// TODO: Add test for IfStmt where condition is myVar.Success (requires evaldecl.MemberAccess)

// --- Test Top Level Driver (Phase X / now) ---

func TestVM_RunDSL_SimpleSystem(t *testing.T) {
	// Manually construct AST for:
	// system "TestSys" {
	//   instance d1: Disk = { ProfileName = "SSD" };
	//   analyze readPerf = d1.Read();
	//   analyze writePerf = d1.Write();
	// }

	systemAST := &decl.System{
		Name: Ident("TestSys"),
		Body: []decl.Node{ // Use Node interface
			&decl.Instance{
				Name:          Ident("d1"),
				ComponentType: Ident("Disk"), // Matches name registered in registerBuiltinComponents
				Params: []*decl.ParamAssignment{ // Use ParamAssignment slice
					{Name: "ProfileName", Value: &decl.LiteralExpr{Kind: "STRING", Value: "SSD"}},
				},
			},
			&decl.decl.Analyze{
				Name: "readPerf",
				Target: &decl.CallExpr{
					Function: &decl.MemberAccessExpr{Receiver: &decl.IdentifierExpr{Name: "d1"}, Member: "Read"},
					Args:     []decl.Expr{},
				},
			},
			&decl.Analyze{
				Name: "writePerf",
				Target: &decl.CallExpr{
					Function: &decl.MemberAccessExpr{Receiver: &decl.IdentifierExpr{Name: "d1"}, Member: "Write"},
					Args:     []decl.Expr{},
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
