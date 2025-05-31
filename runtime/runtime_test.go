package runtime

import (
	"log"
	"testing"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/loader"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
)

func TestBitly(t *testing.T) {
	l := loader.NewLoader(nil, nil, 10) // Max depth 10
	rt := NewRuntime(l)
	fi := rt.LoadFile("../examples/bitly.sdl")
	sys := fi.NewSystem("Bitly")

	se := NewSimpleEval(fi)
	env := fi.Env.Push()
	var currTime core.Duration

	se.EvalInitSystem(sys, env, &currTime)

	ce := &CallExpr{Function: &MemberAccessExpr{Receiver: &IdentifierExpr{Value: "app"}, Member: &IdentifierExpr{Value: "Shorten"}}}
	res2, ret2 := se.Eval(ce, env, &currTime) // reuse env to continue
	log.Println("Now Running System.App.Shorten(), ", res2, ret2, currTime)
}
