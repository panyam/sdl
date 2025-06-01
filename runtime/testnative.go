package runtime

import (
	"github.com/panyam/sdl/core"
	sc "github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
)

// TestNative represents a hash-based index structure (e.g., static, extendible, linear)
type TestNative struct {
	Name string
}

func (n *TestNative) Set(name string, value decl.Value) error {
	return nil
}

func (n *TestNative) Get(name string) (v decl.Value, ok bool) {
	return
}

// NewTestNative creates and initializes a new TestNative component.
func NewTestNative(name string) *TestNative {
	out := &TestNative{Name: name}
	return out.Init()
}

// Init initializes the TestNative with defaults.
func (t *TestNative) Init() *TestNative {
	return t
}

func (t *TestNative) ReadBool() (val Value) {
	val = BoolValue(true)
	val.Time = core.Millis(50)
	return
}

func (t *TestNative) ReadInt() (val Value) {
	val = IntValue(42)
	val.Time = core.Millis(100)
	return
}

func (t *TestNative) ReadInt32() (val Value) {
	val = IntValue(42)
	val.Time = core.Millis(150)
	return
}

func (t *TestNative) ReadFloat32() (val Value) {
	val = FloatValue(42)
	val.Time = core.Millis(200)
	return
}

func (t *TestNative) ReadInt64() (val Value) {
	val.Time = core.Millis(250)
	return
}

func (t *TestNative) ReadFloat64() (val Value) {
	val = FloatValue(42)
	val.Time = core.Millis(300)
	return
}

func (t *TestNative) ReadString() (val Value) {
	val = StringValue("Hello World")
	val.Time = core.Millis(350)
	return
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

func AccessResultToValue(ar sc.AccessResult) (val Value) {
	val, err := NewValue(nil, nil)
	if err != nil {
		panic(err)
	}
	return
}

func (t *TestNative) ReadOutcomesAsTupleValues() *core.Outcomes[Value] {
	return (&core.Outcomes[Value]{}).
		Add(0.95, AccessResultToValue(sc.AccessResult{Success: true, Latency: core.Micros(100)})). // 95% very fast read (0.1 ms)
		Add(0.04, AccessResultToValue(sc.AccessResult{Success: true, Latency: core.Micros(500)})). // 4% slightly slower (0.5 ms)
		Add(0.008, AccessResultToValue(sc.AccessResult{Success: true, Latency: core.Millis(2)})).  // 0.8% slower (2 ms)
		Add(0.001, AccessResultToValue(sc.AccessResult{Success: false, Latency: core.Millis(1)})). // 0.1% fast failure
		Add(0.001, AccessResultToValue(sc.AccessResult{Success: false, Latency: core.Millis(5)}))  // 0.1% slower failure
}
