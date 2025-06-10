package runtime

import (
	"fmt"
	"math"

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
	// Placeholder implementation for simulation:
	// 1. Evaluate all the future thunks.
	// 2. Find the maximum latency among them (makespan).
	// 3. Assume success if all futures succeed (or simply return the desired success code for now).
	// 4. Update total time and return.

	maxLatency := 0.0
	allFuturesSucceeded := true

	for _, futureVal := range futures {
		if futureVal.Type.Tag != TypeTagFuture {
			// This should be caught by type checker, but good to have a runtime check
			panic(fmt.Sprintf("wait expected a future, but got %s", futureVal.Type.String()))
		}
		fval := futureVal.Value.(*FutureValue)

		// A very simplified evaluation of the "gobatch" block.
		// It just evaluates the body once to get a representative latency and result.
		// A full implementation would need to handle the "N" runs and their distribution.
		var futureLatency core.Duration
		res, ret := eval.Eval(fval.Body.Stmt, fval.Body.SavedEnv, &futureLatency)
		if ret && res.IsNil() { // If it's a return from a block without a value
			allFuturesSucceeded = false // Or handle based on some logic
		}

		maxLatency = math.Max(maxLatency, futureLatency)
	}

	// For now, let's just assume the aggregation returns the first success code provided.
	if len(t.SuccessResultCodes) > 0 {
		result = t.SuccessResultCodes[0]
	} else {
		// Fallback if no success code was provided
		result = BoolValue(allFuturesSucceeded)
	}

	// The latency of the wait is the makespan of the parallel operations.
	result.Time = maxLatency
	*currTime += maxLatency

	return
}

type WaitAny struct {
	TimeoutValue       core.Duration
	SuccessResultCodes []Value
}

func (t *WaitAny) Eval(eval *SimpleEval, env *Env[Value], currTime *core.Duration, futures []Value) (result Value, returned bool) {
	// TODO: Implement WaitAny. For now, it can behave like WaitAll for placeholder purposes.
	wa := &WaitAll{SuccessResultCodes: t.SuccessResultCodes, TimeoutValue: t.TimeoutValue}
	return wa.Eval(eval, env, currTime, futures)
}

func (r *Runtime) CreateAggregator(name string, aggParams []Value) Aggregator {
	if name == "WaitAll" {
		return &WaitAll{SuccessResultCodes: aggParams}
	}
	if name == "WaitAny" {
		return &WaitAny{SuccessResultCodes: aggParams}
	}
	panic(fmt.Sprintf("Native aggregator not registered: %s", name))
}
