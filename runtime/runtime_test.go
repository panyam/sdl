package runtime

import (
	"testing"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/loader"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
)

func TestNativeAndBitly(t *testing.T) {
	l := loader.NewLoader(nil, nil, 10) // Max depth 10
	rt := NewRuntime(l)
	fi := rt.LoadFile("../examples/bitly/mvp.sdl")

	var currTime core.Duration
	se := NewSimpleEval(fi)
	env := fi.Env.Push()

	sysbitly := fi.NewSystem("Bitly")
	se.EvalInitSystem(sysbitly, env, &currTime)

	RunTestCall(sysbitly, env, "app", "Shorten", 1000)
	/*
		RunTestCall(sys, env, "test", "ReadBool")
			ce := &CallExpr{Function: &MemberAccessExpr{Receiver: &IdentifierExpr{Value: "app"}, Member: &IdentifierExpr{Value: "Shorten"}}}
			res2, ret2 := se.Eval(ce, env, &currTime) // reuse env to continue
	*/
}
