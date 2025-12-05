package runtime

import (
	sc "github.com/panyam/sdl/lib/core"
	"github.com/panyam/sdl/lib/decl"
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
	val.Time = sc.Millis(50)
	return
}

func (t *TestNative) ReadInt() (val Value) {
	val = IntValue(42)
	val.Time = sc.Millis(100)
	return
}

func (t *TestNative) ReadInt32() (val Value) {
	val = IntValue(42)
	val.Time = sc.Millis(150)
	return
}

func (t *TestNative) ReadFloat32() (val Value) {
	val = FloatValue(42)
	val.Time = sc.Millis(200)
	return
}

func (t *TestNative) ReadInt64() (val Value) {
	val.Time = sc.Millis(250)
	return
}

func (t *TestNative) ReadFloat64() (val Value) {
	val = FloatValue(42)
	val.Time = sc.Millis(300)
	return
}

func (t *TestNative) ReadString() (val Value) {
	val = StringValue("Hello World")
	val.Time = sc.Millis(350)
	return
}

// This will test if AccessResult can be converted to a Value - and if so how
func (t *TestNative) ReadOutcomes() *sc.Outcomes[sc.AccessResult] {
	return (&sc.Outcomes[sc.AccessResult]{And: sc.AndAccessResults}).
		Add(0.95, sc.AccessResult{Success: true, Latency: sc.Micros(100)}). // 95% very fast read (0.1 ms)
		Add(0.04, sc.AccessResult{Success: true, Latency: sc.Micros(500)}). // 4% slightly slower (0.5 ms)
		Add(0.008, sc.AccessResult{Success: true, Latency: sc.Millis(2)}).  // 0.8% slower (2 ms)
		Add(0.001, sc.AccessResult{Success: false, Latency: sc.Millis(1)}). // 0.1% fast failure
		Add(0.001, sc.AccessResult{Success: false, Latency: sc.Millis(5)})  // 0.1% slower failure
}

func AccessResultToValue(ar sc.AccessResult) (val Value) {
	val, err := NewValue(nil, nil)
	if err != nil {
		panic(err)
	}
	return
}

func (t *TestNative) ReadOutcomesAsTupleValues() *sc.Outcomes[Value] {
	return (&sc.Outcomes[Value]{}).
		Add(0.95, AccessResultToValue(sc.AccessResult{Success: true, Latency: sc.Micros(100)})). // 95% very fast read (0.1 ms)
		Add(0.04, AccessResultToValue(sc.AccessResult{Success: true, Latency: sc.Micros(500)})). // 4% slightly slower (0.5 ms)
		Add(0.008, AccessResultToValue(sc.AccessResult{Success: true, Latency: sc.Millis(2)})).  // 0.8% slower (2 ms)
		Add(0.001, AccessResultToValue(sc.AccessResult{Success: false, Latency: sc.Millis(1)})). // 0.1% fast failure
		Add(0.001, AccessResultToValue(sc.AccessResult{Success: false, Latency: sc.Millis(5)}))  // 0.1% slower failure
}
