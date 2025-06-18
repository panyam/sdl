package runtime

import (
	"fmt"
	"math"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
)

type Aggregator interface {
	Eval(eval *SimpleEval, env *Env[Value], currTime *core.Duration, futures []Value) (result Value, returned bool)
}

type WaitAll struct {
	TimeoutValue       core.Duration
	SuccessResultCodes []Value
}

func (t *WaitAll) Eval(eval *SimpleEval, env *Env[Value], currTime *core.Duration, futures []Value) (result Value, returned bool) {
	maxLatency := 0.0
	allFuturesSucceeded := true

	for _, futureVal := range futures {
		if futureVal.Type.Tag != TypeTagFuture {
			panic(fmt.Sprintf("wait expected a future, but got %s", futureVal.Type.String()))
		}
		fval := futureVal.Value.(*FutureValue)

		// Set the tracer's parent context before evaluating the deferred code
		if eval.Tracer != nil && fval.TraceID > 0 {
			eval.Tracer.PushParentID(fval.TraceID)
		}

		// A very simplified evaluation of the "gobatch" block.
		// It just evaluates the body once to get a representative latency and result.
		var futureLatency core.Duration
		res, ret := eval.Eval(fval.Body.Stmt, fval.Body.SavedEnv, &futureLatency)

		// Emit exit event for the future
		if eval.Tracer != nil && fval.TraceID > 0 {
			// For go expressions, we don't have component/method info
			eval.Tracer.Exit(float64(*currTime)/1e9, futureLatency, nil, nil, res, nil)
			eval.Tracer.PopParent()
		}

		if !ret {
			allFuturesSucceeded = false
		} else {
			isSuccess := false
			for _, successCode := range t.SuccessResultCodes {
				if res.Equals(&successCode) {
					isSuccess = true
					break
				}
			}
			if !isSuccess {
				allFuturesSucceeded = false
			}
		}

		maxLatency = math.Max(maxLatency, futureLatency)
	}

	// For now, let's just assume the aggregation returns the first success code provided.
	if allFuturesSucceeded && len(t.SuccessResultCodes) > 0 {
		// If all futures returned a success code, then the aggregator returns that code.
		// This is a simplification; a real aggregator might return a summary.
		result = t.SuccessResultCodes[0]
	} else {
		// Fallback if any future failed or no success code was provided.
		// Here, we should return a sensible failure value. Let's find "InternalError" in the enum.
		// This is still a placeholder for proper error handling.
		if len(t.SuccessResultCodes) > 0 {
			enumType := t.SuccessResultCodes[0].Type
			if enumType.Tag == decl.TypeTagEnum {
				enumDecl := enumType.Info.(*decl.EnumDecl)
				errIndex := enumDecl.IndexOfVariant("InternalError")
				if errIndex >= 0 {
					result, _ = NewValue(enumType, errIndex)
				}
			}
		}
		if result.IsNil() {
			result = BoolValue(false)
		}
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
