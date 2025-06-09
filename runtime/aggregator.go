package runtime

import (
	"fmt"

	"github.com/panyam/sdl/core"
)

type Aggregator interface {
	Eval(eval *SimpleEval, env *Env[Value], currTime *core.Duration, futures []Value) (result Value, returned bool)
}

type WaitAll struct {
	TimeoutValue       core.Duration
	SuccessResultCodes []Value
}

func (t *WaitAll) Eval(eval *SimpleEval, env *Env[Value], currTime *core.Duration, futures []Value) (result Value, returned bool) {
	return
}

type WaitAny struct {
	TimeoutValue       core.Duration
	SuccessResultCodes []Value
}

func (t *WaitAny) Eval(eval *SimpleEval, env *Env[Value], currTime *core.Duration, futures []Value) (result Value, returned bool) {
	return
}

func (r *Runtime) CreateAggregator(name string, aggParams []Value) Aggregator {
	if name == "WaitAll" {
		return &WaitAll{SuccessResultCodes: aggParams}
	}
	if name == "WaitAny" {
		return &WaitAny{SuccessResultCodes: aggParams}
	}
	panic(fmt.Sprintf("Native aggregative not registered: %s", name))
}
