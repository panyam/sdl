package runtime

import (
	"github.com/panyam/sdl/core"
	sc "github.com/panyam/sdl/core"
)

// TestNative represents a hash-based index structure (e.g., static, extendible, linear)
type TestNative struct {
}

// NewTestNative creates and initializes a new TestNative component.
func NewTestNative() *TestNative {
	out := &TestNative{}
	return out.Init()
}

// Init initializes the TestNative with defaults.
func (t *TestNative) Init() *TestNative {
	return t
}

func (t *TestNative) ReadBool() bool {
	return true
}

func (t *TestNative) ReadInt() int {
	return 42
}

func (t *TestNative) ReadInt32() int32 {
	return 42
}

func (t *TestNative) ReadFloat32() float32 {
	return 42
}

func (t *TestNative) ReadInt64() int64 {
	return 42
}

func (t *TestNative) ReadFloat64() float64 {
	return 42
}

func (t *TestNative) ReadString() string {
	return "Hello World"
}

// This will test if AccessResult can be converted to a Value - and if so how
func (t *TestNative) ReadOutcomes() *core.Outcomes[sc.AccessResult] {
	return (&core.Outcomes[sc.AccessResult]{And: sc.AndAccessResults}).
		Add(0.95, sc.AccessResult{Success: true, Latency: core.Micros(100)}). // 95% very fast read (0.1 ms)
		Add(0.04, sc.AccessResult{Success: true, Latency: core.Micros(500)}). // 4% slightly slower (0.5 ms)
		Add(0.008, sc.AccessResult{Success: true, Latency: core.Millis(2)}).  // 0.8% slower (2 ms)
		Add(0.001, sc.AccessResult{Success: false, Latency: core.Millis(1)}). // 0.1% fast failure
		Add(0.001, sc.AccessResult{Success: false, Latency: core.Millis(5)})  // 0.1% slower failure
}

func AccessResultToValue(ar sc.AccessResult) Value {
	out, err := NewValue(nil, nil)
	if err != nil {
		panic(err)
	}
	return out
}

func (t *TestNative) ReadOutcomesAsTupleValues() *core.Outcomes[Value] {
	return (&core.Outcomes[Value]{}).
		Add(0.95, AccessResultToValue(sc.AccessResult{Success: true, Latency: core.Micros(100)})). // 95% very fast read (0.1 ms)
		Add(0.04, AccessResultToValue(sc.AccessResult{Success: true, Latency: core.Micros(500)})). // 4% slightly slower (0.5 ms)
		Add(0.008, AccessResultToValue(sc.AccessResult{Success: true, Latency: core.Millis(2)})).  // 0.8% slower (2 ms)
		Add(0.001, AccessResultToValue(sc.AccessResult{Success: false, Latency: core.Millis(1)})). // 0.1% fast failure
		Add(0.001, AccessResultToValue(sc.AccessResult{Success: false, Latency: core.Millis(5)}))  // 0.1% slower failure
}
