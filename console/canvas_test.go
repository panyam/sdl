package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutionRecipe(t *testing.T) {
	canvas := NewCanvas()
	assert.NotNil(t, canvas, "Canvas should not be nil")

	// 1. Load the system model
	err := canvas.Load("../examples/bitly/mvp.sdl")
	assert.NoError(t, err, "Loading MVP SDL file should not produce an error")

	// 2. Set the active system
	err = canvas.Use("BitlyWithCache")
	assert.NoError(t, err, "Setting the active system should not produce an error")

	// --- Baseline Scenario ---
	t.Log("--- Running Baseline Scenario (95% Cache Hit Rate) ---")
	err = canvas.Set("app.cache.HitRate", 0.95)
	assert.NoError(t, err, "Setting cache hit rate should not fail")
	// TODO: Add RunOptions for things like --runs=N
	err = canvas.Run("run_baseline", "app.Redirect")
	assert.NoError(t, err, "Running baseline simulation should not fail")

	// --- Contended Scenario ---
	t.Log("--- Running Contended Scenario (60% Cache Hit Rate) ---")
	err = canvas.Set("app.cache.HitRate", 0.60)
	assert.NoError(t, err, "Setting cache hit rate should not fail")
	err = canvas.Run("run_contended", "app.Redirect")
	assert.NoError(t, err, "Running contended simulation should not fail")

	// --- Flaky Cache Scenario ---
	t.Log("--- Running Failure Scenario (5% Cache Failure) ---")
	err = canvas.Set("app.cache.HitRate", 0.95) // Reset hit rate
	assert.NoError(t, err)
	err = canvas.Set("app.cache.FailureProb", 0.05)
	assert.NoError(t, err, "Setting cache failure probability should not fail")
	err = canvas.Run("run_flaky", "app.Redirect")
	assert.NoError(t, err, "Running flaky cache simulation should not fail")

	// --- Plotting ---
	t.Log("--- Generating Latency Comparison Plot ---")
	// TODO: Define PlotOptions for specifying series, from, output etc.
	err = canvas.Plot() // This will need to be fleshed out with options
	assert.NoError(t, err, "Generating the plot should not fail")
}
