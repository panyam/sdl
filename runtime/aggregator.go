package runtime

import (
	"fmt"

	"github.com/panyam/sdl/core"
)

type Aggregator = func(eval *SimpleEval, env *Env[Value], currTime *core.Duration, futures []Value) (result Value, returned bool)

func (r *Runtime) CreateAggregator(name string, aggParams []Value) Aggregator {
	// if name == "AllHttpStatusesLessThan" { return cd.NewDisk(name) }
	panic(fmt.Sprintf("Native aggregative not registered: %s", name))
}
