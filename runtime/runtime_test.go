package runtime

import (
	"testing"

	"github.com/panyam/sdl/core"
	"github.com/panyam/sdl/loader"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
)

func RunLoadTest(sdlfile, systemName, componentName, methodName string) {
	l := loader.NewLoader(nil, nil, 10) // Max depth 10
	rt := NewRuntime(l)
	fi, err := rt.LoadFile(sdlfile)
	if err != nil {
		panic(err)
	}
	var currTime core.Duration
	se := NewSimpleEval(fi, nil)
	env := fi.Env()

	sys, _ := fi.NewSystem(systemName, false)
	se.EvalInitSystem(sys, env, &currTime)

	RunTestCall(sys, env, componentName, methodName, 1000)
}

func TestNativeAndBitly(t *testing.T) {
	RunLoadTest("../examples/bitly/mvp.sdl", "Bitly", "app", "Shorten")
}

func TestTwitter(t *testing.T) {
	RunLoadTest("../examples/twitter/services.sdl", "Twitter", "tls", "GetTimeline")
}
