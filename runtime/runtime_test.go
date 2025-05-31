package runtime

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
)

func TestBitly(t *testing.T) {
	l := loader.NewLoader(nil, nil, 10) // Max depth 10
	rt := NewRuntime(l)
	fi := rt.LoadFile("../examples/bitly.sdl")
	sys := fi.NewSystem("Bitly")
	stmts, err := sys.Compile()
	if err != nil {
		panic(err)
	}

	log.Println("Compiled statements:")
	decl.PPrint(stmts)

	se := SimpleEval{
		fi,
		rand.New(rand.NewSource(time.Now().UnixMicro())),
	}
	var currTime core.Duration
	env := fi.Env.Push()
	for _, item := range stmts.Statements {
		se.Eval(item, env, &currTime)
	}

	mae := &MemberAccessExpr{Receiver: &IdentifierExpr{Value: "app"}, Member: &IdentifierExpr{Value: "Shorten"}}
	ce := &CallExpr{Function: mae}
	res2, ret2 := se.Eval(ce, env, &currTime) // reuse env to continue
	log.Println("Now Runnign System.App.Shorten(), ", res2, ret2, currTime)
}
