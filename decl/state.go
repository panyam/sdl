package decl

import "github.com/panyam/sdl/core"

var (
	// Lazily initialize these? Or create here? Creating here is simpler.
	nilState   *VarState
	trueState  *VarState
	falseState *VarState
)

func init() {
	// Initialize predefined states used by OpNil, OpTrue, OpFalse
	nilState = &VarState{ValueOutcome: nil, LatencyOutcome: ZeroLatencyOutcome()} // No specific value type for nil
	trueState = &VarState{
		ValueOutcome:   (&core.Outcomes[bool]{}).Add(1.0, true),
		LatencyOutcome: ZeroLatencyOutcome(),
	}
	falseState = &VarState{
		ValueOutcome:   (&core.Outcomes[bool]{}).Add(1.0, false),
		LatencyOutcome: ZeroLatencyOutcome(),
	}
}

// VarState holds the dual-track outcomes for a variable or expression result.
type VarState struct {
	ValueOutcome   any // *core.Outcomes[V] (V is discrete type)
	LatencyOutcome any // *core.Outcomes[Duration]
}

// Helper to create a zero latency outcome
func ZeroLatencyOutcome() *core.Outcomes[core.Duration] {
	o := (&core.Outcomes[core.Duration]{}).Add(1.0, 0.0)
	// Should have the correct Adder func registered
	// o.And = func(a, b Duration) Duration { return a + b }
	return o
}

// Helper to create an identity value outcome (e.g., for delay)
// We need a concrete type, maybe bool=true?
func IdentityValueOutcome() *core.Outcomes[bool] {
	o := (&core.Outcomes[bool]{}).Add(1.0, true)
	// o.And = func(a, b bool) bool { return a && b } // Example
	return o
}

// createIdentityState provides a neutral starting point for composition.
// Value: bool=true (neutral for AND-like success combination)
// Latency: 0ms
func createIdentityState() *VarState {
	// Ensure these helpers return non-nil outcomes
	val := IdentityValueOutcome() // core.Outcomes[bool]{1.0=>true}
	lat := ZeroLatencyOutcome()   // core.Outcomes[Duration]{1.0=>0}
	if val == nil || lat == nil {
		panic("internal error: failed to create identity/zero outcomes")
	}
	return &VarState{ValueOutcome: val, LatencyOutcome: lat}
}

// Add Safe Accessor helpers to VarState to handle nil pointers gracefully
func (vs *VarState) SafeValueOutcome() any {
	if vs == nil {
		return nil
	}
	return vs.ValueOutcome
}

func (vs *VarState) SafeLatencyOutcome() any {
	if vs == nil {
		return nil
	}
	return vs.LatencyOutcome
}

func createNilState() *VarState { return nilState }
func createBoolState(b bool) *VarState {
	if b {
		return trueState
	}
	return falseState
}
