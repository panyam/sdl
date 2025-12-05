package runtime

import (
	"testing"

	"github.com/panyam/sdl/lib/components"
)

// TestCacheHitRate verifies that the Cache component properly reports its hit rate
func TestCacheHitRate(t *testing.T) {
	// Create a cache with 80% hit rate
	cache := components.NewCache()
	cache.HitRate = 0.8

	// Get the flow pattern for Read method
	pattern := cache.GetFlowPattern("Read", 100.0)

	t.Logf("Cache Read pattern: SuccessRate=%.2f", pattern.SuccessRate)

	if pattern.SuccessRate != 0.8 {
		t.Errorf("Expected Cache Read success rate to be 0.8, got %.2f", pattern.SuccessRate)
	}
}
