package decl

import (
	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
)

func Ar2Value(ar core.AccessResult) Value {
	out, err := decl.NewValue(decl.BoolType, ar.Success)
	if err != nil {
		panic(err)
	}
	out.Time = ar.Latency
	return out
}

func Dur2Value(ar core.Duration) Value {
	out, err := decl.NewValue(decl.BoolType, true)
	if err != nil {
		panic(err)
	}
	out.Time = ar
	return out
}

func OutcomesToValue(outcomes *core.Outcomes[core.AccessResult]) (v decl.Value) {
	out := core.Map(outcomes, Ar2Value)
	outType := decl.OutcomesType(decl.BoolType)
	v, e := decl.NewValue(outType, out)
	if e != nil {
		panic(e)
	}
	return v
}

func OutcomesOfDurationsToValue(outcomes *core.Outcomes[core.Duration]) (v decl.Value) {
	out := core.Map(outcomes, Dur2Value)
	outType := decl.OutcomesType(decl.FloatType)
	v, e := decl.NewValue(outType, out)
	if e != nil {
		panic(e)
	}
	return v
}
