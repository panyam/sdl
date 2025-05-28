package runtime

import (
	"log"
	"testing"

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
	env := fi.Env.Push()
	rt.Eval(stmts, env)
}
