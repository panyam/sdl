package bitly

import (
	"testing"

	sdlc "github.com/panyam/leetcoach/sdl/components"
	sdl "github.com/panyam/leetcoach/sdl/core"
)

// Helper function to create a default BitlyService setup for testing
func setupBitlyService(t *testing.T) *BitlyService {
	// Use reliable components with SSD for baseline
	idGen := (&IDGenerator{}).Init("TestIDGen")
	cache := sdlc.NewCache() // Use default cache settings (80% hit, fast)
	// Use HashIndex on SSD for DB - generally fast point lookups/inserts
	db := (&DatabaseComponent{}).Init("TestDB")

	// Configure components slightly for clearer test results if needed
	cache.HitRate = 0.90                                                         // 90% Cache Hit Rate
	cache.HitLatency = (&sdl.Outcomes[sdl.Duration]{}).Add(1.0, sdl.Nanos(500))  // 0.5us hit
	cache.MissLatency = (&sdl.Outcomes[sdl.Duration]{}).Add(1.0, sdl.Nanos(800)) // 0.8us miss detect
	cache.FailureProb = 0.0001                                                   // Very reliable cache
	cache.Init()                                                                 // Recalculate after changing params

	db.PrimaryIndex.NumRecords = 500_000_000 // 500M records (affects hash probs)
	db.MaxOutcomeLen = 15                    // Allow more detail in DB outcomes
	idGen.FailureProb = 0.000001             // Make ID gen extremely reliable

	bs := (&BitlyService{}).Init("TestBitly", idGen, cache, db)
	return bs
}

func TestBitlyService_Redirect_Metrics(t *testing.T) {
	bs := setupBitlyService(t)

	// Simulate the Redirect operation
	redirectOutcomes := bs.Redirect("shortcode1", 0) // lambda unused for now

	if redirectOutcomes == nil || redirectOutcomes.Len() == 0 {
		t.Fatal("Redirect() returned nil or empty outcomes")
	}

	// --- DEBUG: Inspect combined outcomes BEFORE reduction ---
	successesSplitDebug, failuresSplitDebug := redirectOutcomes.Split(sdl.AccessResult.IsSuccess)
	t.Logf("Combined Outcomes (Before Reduction): TotalBuckets=%d, SuccessBuckets=%d, FailureBuckets=%d", redirectOutcomes.Len(), successesSplitDebug.Len(), failuresSplitDebug.Len())
	if successesSplitDebug != nil {
		// Log first few success buckets (sorted by latency by PercentileLatency later)
		sdl.PercentileLatency(successesSplitDebug, 0.0) // Call this to force sort for logging
		for i := 0; i < 5 && i < successesSplitDebug.Len(); i++ {
			b := successesSplitDebug.Buckets[i]
			t.Logf("  Success Bucket %d: Weight=%.6f, Latency=%.9fs, Success=%v", i, b.Weight, b.Value.Latency, b.Value.Success)
		}
	}
	// --- END DEBUG ---

	// Calculate overall metrics
	avail := sdl.Availability(redirectOutcomes)
	mean := sdl.MeanLatency(redirectOutcomes)
	p50 := sdl.PercentileLatency(redirectOutcomes, 0.50)
	p99 := sdl.PercentileLatency(redirectOutcomes, 0.99)
	p999 := sdl.PercentileLatency(redirectOutcomes, 0.999)

	t.Logf("Bitly Redirect: Avail=%.6f, Mean=%.6fs, P50=%.6fs, P99=%.6fs, P999=%.6fs (Buckets: %d)",
		avail, mean, p50, p99, p999, redirectOutcomes.Len())

	// --- Plausibility Checks ---

	// 1. Availability: Should be very high, dominated by DB/Cache reliability
	// DB Index Find Avail * Cache Avail (approx)
	// SSD failure is low, Cache failure low, IDGen failure low
	expectedMinAvail := 0.998 // Should be much better than 3 nines with SSD/low fail probs
	if avail < expectedMinAvail {
		t.Errorf("Redirect Availability %.6f is unexpectedly low (expected > %.4f)", avail, expectedMinAvail)
	}

	// 2. P50 Latency: Should be dominated by Cache Hit latency (since HitRate = 90%)
	cacheHitLat, _ := bs.Cache.HitLatency.GetValue() // Assumes single value for simplicity here
	// Add miss latency for comparison
	cacheMissLat, _ := bs.Cache.MissLatency.GetValue()
	dbReadMean := sdl.MeanLatency(bs.DB.GetLongURL("")) // Get avg DB read time

	t.Logf("Component Latencies: CacheHit=%.9fs, CacheMiss=%.9fs, DBReadMean=%.9fs", cacheHitLat, cacheMissLat, dbReadMean)

	// P50 should be very close to CacheHitLatency + CacheMissLatency (weighted by hit/miss rate for P50 point)
	// Since HitRate is 90%, P50 should land squarely in the hit latency bucket.
	// Expected P50 ~= cacheMissLat (latency of check) + cacheHitLat (latency of hit data return) -> No, logic is If Hit -> Hit Latency, If Miss -> Miss Latency + DB Latency
	// Expected P50 ~= Cache Hit Latency
	if !approxEqualTest(p50, cacheHitLat, cacheHitLat*0.5+sdl.Micros(1)) { // Allow 50% or 1us tolerance
		t.Errorf("Redirect P50 Latency (%.6fs) seems too far from Cache Hit Latency (%.6fs)", p50, cacheHitLat)
	}

	// 3. P99 Latency: Should be dominated by the Cache Miss -> DB Read path
	// Expected P99 ~= CacheMissLatency + P99 DB Read Latency
	dbReadP99 := sdl.PercentileLatency(bs.DB.GetLongURL(""), 0.99)
	// Use avg miss latency for simplicity calculation
	expectedP99_ballpark := cacheMissLat + dbReadP99 // cacheMissLat is latency to *detect* miss
	t.Logf("Expected P99 ballpark ~= CacheMissLat(%.6f) + DBReadP99(%.6f) = %.6f", cacheMissLat, dbReadP99, expectedP99_ballpark)
	// Check if calculated P99 is reasonably close to this ballpark figure
	// It will be lower because only 10% of requests take the DB path, P99 overall might still land in cache hit path depending on DB P99 value
	if p99 < cacheHitLat {
		t.Errorf("Redirect P99 latency (%.6fs) cannot be lower than cache hit latency (%.6fs)", p99, cacheHitLat)
	}
	// If DB P99 is much larger than Cache Hit, overall P99 should be influenced by DB P99
	if dbReadP99 > cacheHitLat*10 && p99 < dbReadP99*0.5 {
		t.Errorf("Redirect P99 latency (%.6fs) seems too low given DB P99 latency (%.6fs) on miss path", p99, dbReadP99)
	}
	// Check requirement: P99 < 100ms
	targetP99 := sdl.Millis(100)
	if p99 >= targetP99 {
		t.Errorf("Requirement FAILED: Redirect P99 Latency (%.6fs) is >= target %.6fs", p99, targetP99)
	} else {
		t.Logf("Requirement PASSED: Redirect P99 Latency (%.6fs) is < target %.6fs", p99, targetP99)
	}

}

func TestBitlyService_ShortenURL_Metrics(t *testing.T) {
	bs := setupBitlyService(t)

	// Simulate the ShortenURL operation
	shortenOutcomes := bs.ShortenURL("http://example.com/very/long/url", 0) // lambda unused

	if shortenOutcomes == nil || shortenOutcomes.Len() == 0 {
		t.Fatal("ShortenURL() returned nil or empty outcomes")
	}

	// Calculate overall metrics
	avail := sdl.Availability(shortenOutcomes)
	mean := sdl.MeanLatency(shortenOutcomes)
	p99 := sdl.PercentileLatency(shortenOutcomes, 0.99)

	t.Logf("Bitly ShortenURL: Avail=%.6f, Mean=%.6fs, P99=%.6fs (Buckets: %d)",
		avail, mean, p99, shortenOutcomes.Len())

	// --- Plausibility Checks ---

	// 1. Availability: Should be high, limited by IDGen and DB Write reliability
	idGenAvail := sdl.Availability(bs.IDGen.GenerateID())
	// Use Insert as proxy for DB write reliability
	dbWriteAvail := sdl.Availability(bs.DB.SaveMapping("", ""))
	expectedMinAvail := idGenAvail * dbWriteAvail * 0.999 // Allow slight reduction for composition
	if avail < expectedMinAvail {
		t.Errorf("Shorten Availability %.6f is unexpectedly low (expected > %.4f)", avail, expectedMinAvail)
	}

	// 2. Mean Latency: Should be sum of IDGen latency + DB Insert latency (approx)
	idGenMean := sdl.MeanLatency(bs.IDGen.GenerateID())
	dbWriteMean := sdl.MeanLatency(bs.DB.SaveMapping("", ""))
	expectedMean := idGenMean + dbWriteMean
	t.Logf("Expected Shorten Mean ballpark ~= IDGenMean(%.6f) + DBWriteMean(%.6f) = %.6f", idGenMean, dbWriteMean, expectedMean)
	if !approxEqualTest(mean, expectedMean, expectedMean*0.3) { // Allow tolerance
		t.Errorf("Shorten Mean Latency %.6f differs significantly from expected ballpark %.6f", mean, expectedMean)
	}
}

// Reuse approxEqualTest if needed
// func approxEqualTest(a, b, tolerance float64) bool { ... }
