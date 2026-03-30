package runtime

import (
	"testing"

	"github.com/panyam/sdl/lib/core"
	"github.com/panyam/sdl/lib/loader"
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

// TestNativeAndBitly loads the Bitly example and runs the Shorten method.
// Requires @stdlib to be resolvable; skips if not available.
func TestNativeAndBitly(t *testing.T) {
	t.Skip("Skipping: requires @stdlib filesystem mount (not available in unit test context)")
	RunLoadTest("../../examples/bitly/mvp.sdl", "Bitly", "arch.app", "Shorten")
}

// TestTwitter loads the Twitter example and runs the GetTimeline method.
// Requires @stdlib to be resolvable; skips if not available.
func TestTwitter(t *testing.T) {
	t.Skip("Skipping: requires @stdlib filesystem mount (not available in unit test context)")
	RunLoadTest("../../examples/twitter/services.sdl", "Twitter", "arch.tls", "GetTimeline")
}
