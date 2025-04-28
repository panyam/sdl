package sdl

import (
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"
)

// Helper to create a large AccessResult distribution by chaining 'And'
func createLargeAccessResultOutcomes(base *Outcomes[AccessResult], depth int, maxBucketsBeforeTrim int, maxLenAfterTrim int) *Outcomes[AccessResult] {
	out := base
	trimmer := TrimToSize(maxBucketsBeforeTrim, maxLenAfterTrim) // Use the existing trimmer

	for i := 0; i < depth; i++ {
		// Combine with self - simulates repeated operations
		out = And(out, base, AndAccessResults)
		// Apply trimming within the loop if specified
		if maxBucketsBeforeTrim > 0 {
			// Need to split/trim/append as TrimToSize expects AccessResult
			successes, failures := out.Split(AccessResult.IsSuccess)
			successes = trimmer(successes)
			failures = trimmer(failures)
			out = successes.Append(failures)
		}
		// Optional: Check length to prevent excessive memory use in tests
		// if out.Len() > 20000 { // Arbitrary limit for safety
		//     fmt.Printf("Warning: Outcome length exceeded limit (%d) at depth %d. Stopping generation.\n", out.Len(), i)
		//     break
		// }
	}
	return out
}

// Helper to create a large RangedResult distribution
func createLargeRangedResultOutcomes(base *Outcomes[RangedResult], depth int) *Outcomes[RangedResult] {
	// Note: Trimming RangedResult within the loop is trickier as we don't have a
	// standard TrimToSize equivalent yet. For now, just do the And.
	// In practice, reduction would likely be applied *after* the full composition.
	out := base
	for i := 0; i < depth; i++ {
		out = And(out, base, AndRangedResults)
		// Optional: Check length
		// if out.Len() > 20000 {
		//     fmt.Printf("Warning: Ranged Outcome length exceeded limit (%d) at depth %d. Stopping generation.\n", out.Len(), i)
		//     break
		// }
	}
	return out
}

// --- Benchmarks for AccessResult Reduction ---

var benchmarkAccessResultOutcomes *Outcomes[AccessResult] // Store pre-generated outcomes

func setupAccessResultBenchmark(b *testing.B) {
	if benchmarkAccessResultOutcomes == nil {
		b.Logf("Setting up large AccessResult outcomes...")
		base := &Outcomes[AccessResult]{}
		base.Add(80, AccessResult{true, Millis(1)})
		base.Add(15, AccessResult{true, Millis(10)})
		base.Add(3, AccessResult{false, Millis(5)})
		base.Add(2, AccessResult{true, Millis(100)})

		// Create a reasonably large distribution (e.g., 5 levels deep without trimming)
		// Adjust depth as needed based on performance. 5 levels => 4^5 = 1024 buckets
		// 6 levels => 4^6 = 4096 buckets
		// 7 levels => 4^7 = 16384 buckets
		depth := 7
		benchmarkAccessResultOutcomes = createLargeAccessResultOutcomes(base, depth, 0, 0) // No trimming during generation
		b.Logf("Setup complete. Outcome size: %d buckets", benchmarkAccessResultOutcomes.Len())
	}
	// Reset timer after setup
	b.ResetTimer()
}

func BenchmarkReduceAccessResults_Adaptive(b *testing.B) {
	setupAccessResultBenchmark(b)
	targetBuckets := 10 // Typical target size

	for i := 0; i < b.N; i++ {
		// Run the reduction function under test
		reduced := ReduceAccessResults(benchmarkAccessResultOutcomes, targetBuckets)
		// Prevent compiler optimizing away the call
		if reduced == nil {
			b.Fatal("Reduction resulted in nil")
		}
	}
}

func BenchmarkReduceAccessResults_MergeAdjacent(b *testing.B) {
	setupAccessResultBenchmark(b)
	mergeThreshold := 0.8 // Example threshold

	// Sort needed for MergeAdjacent
	sort.SliceStable(benchmarkAccessResultOutcomes.Buckets, func(i, j int) bool {
		// Basic sort: failures first, then by latency
		bi, bj := benchmarkAccessResultOutcomes.Buckets[i].Value, benchmarkAccessResultOutcomes.Buckets[j].Value
		if bi.Success != bj.Success {
			return !bi.Success // false (failure) comes before true (success)
		}
		return bi.Latency < bj.Latency
	})

	for i := 0; i < b.N; i++ {
		merged := MergeAdjacentAccessResults(benchmarkAccessResultOutcomes, mergeThreshold)
		if merged == nil {
			b.Fatal("Merge resulted in nil")
		}
	}
}

// Benchmark TrimToSize (combines MergeAdjacent + AdaptiveReduce)
func BenchmarkReduceAccessResults_TrimToSize(b *testing.B) {
	setupAccessResultBenchmark(b)
	lenTrigger := 100 // Trigger reduction if > 100 buckets
	maxLen := 10      // Target max length after reduction
	trimmer := TrimToSize(lenTrigger, maxLen)

	for i := 0; i < b.N; i++ {
		// Need to split -> trim -> append because TrimToSize works on success/failure groups
		successes, failures := benchmarkAccessResultOutcomes.Split(AccessResult.IsSuccess)
		trimmedSuccesses := trimmer(successes)
		trimmedFailures := trimmer(failures)
		final := trimmedSuccesses.Append(trimmedFailures)

		if final == nil && benchmarkAccessResultOutcomes.Len() > 0 { // Check if original wasn't empty
			b.Fatal("TrimToSize resulted in nil")
		}
	}
}

// --- Benchmarks for RangedResult Reduction ---

var benchmarkRangedResultOutcomes *Outcomes[RangedResult] // Store pre-generated outcomes

func setupRangedResultBenchmark(b *testing.B) {
	if benchmarkRangedResultOutcomes == nil {
		b.Logf("Setting up large RangedResult outcomes...")
		base := &Outcomes[RangedResult]{}
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		// Create a slightly more varied base case for ranges
		for i := 0; i < 5; i++ {
			success := rng.Float64() < 0.95              // 95% success base
			minLat := Millis(rng.Float64() * 5)          // 0-5ms
			modeLat := minLat + Millis(rng.Float64()*10) // mode up to 10ms higher
			maxLat := modeLat + Millis(rng.Float64()*15) // max up to 15ms higher
			weight := rng.Float64()*20 + 1               // weight 1-21
			base.Add(weight, RangedResult{success, minLat, modeLat, maxLat})
		}

		// Normalize weights for cleaner composition (optional but good practice)
		// totalW := base.TotalWeight()
		// for k := range base.Buckets { base.Buckets[k].Weight /= totalW }

		// 7 levels -> 5^7 = 78125 buckets (potentially large!) Adjust depth if needed.
		depth := 6 // 5^6 = 15625
		benchmarkRangedResultOutcomes = createLargeRangedResultOutcomes(base, depth)
		b.Logf("Setup complete. Ranged Outcome size: %d buckets", benchmarkRangedResultOutcomes.Len())
	}
	b.ResetTimer()
}

func BenchmarkReduceRangedResults_Adaptive(b *testing.B) {
	setupRangedResultBenchmark(b)
	targetBuckets := 15

	for i := 0; i < b.N; i++ {
		reduced := ReduceRangedResults(benchmarkRangedResultOutcomes, targetBuckets)
		if reduced == nil && benchmarkRangedResultOutcomes.Len() > 0 {
			b.Fatal("Reduction resulted in nil")
		}
	}
}

func BenchmarkReduceRangedResults_MergeOverlapping(b *testing.B) {
	setupRangedResultBenchmark(b)
	overlapThreshold := 0.5 // Example threshold

	for i := 0; i < b.N; i++ {
		merged := MergeOverlappingRangedResults(benchmarkRangedResultOutcomes, overlapThreshold)
		if merged == nil && benchmarkRangedResultOutcomes.Len() > 0 {
			b.Fatal("Merge resulted in nil")
		}
	}
}

// --- Accuracy Check (Example - run as a test, not benchmark) ---

func TestReductionAccuracy_AccessResult(t *testing.T) {
	t.Logf("Setting up large AccessResult outcomes for accuracy check...")
	base := &Outcomes[AccessResult]{}
	base.Add(80, AccessResult{true, Millis(1)})
	base.Add(15, AccessResult{true, Millis(10)})
	base.Add(3, AccessResult{false, Millis(5)})
	base.Add(2, AccessResult{true, Millis(100)})
	depth := 7
	largeOutcomes := createLargeAccessResultOutcomes(base, depth, 0, 0)
	t.Logf("Setup complete. Outcome size: %d buckets", largeOutcomes.Len())

	if largeOutcomes.Len() == 0 {
		t.Skip("Skipping accuracy test as generated outcomes are empty.")
		return
	}

	// --- Calculate Original Metrics ---
	origAvail := Availability(largeOutcomes)
	origMean := MeanLatency(largeOutcomes)
	origP50 := PercentileLatency(largeOutcomes, 0.50)
	origP99 := PercentileLatency(largeOutcomes, 0.99)
	t.Logf("Original Metrics: Avail=%.4f, Mean=%.6fs, P50=%.6fs, P99=%.6fs", origAvail, origMean, origP50, origP99)

	// --- Apply Reductions ---
	targetBuckets := 10
	reducedAdaptive := ReduceAccessResults(largeOutcomes, targetBuckets)

	// Need to sort for MergeAdjacent
	sort.SliceStable(largeOutcomes.Buckets, func(i, j int) bool {
		bi, bj := largeOutcomes.Buckets[i].Value, largeOutcomes.Buckets[j].Value
		if bi.Success != bj.Success {
			return !bi.Success
		}
		return bi.Latency < bj.Latency
	})
	reducedMerged := MergeAdjacentAccessResults(largeOutcomes, 0.8)
	// Typically Merge is followed by AdaptiveReduce for size control
	reducedMergedThenAdaptive := ReduceAccessResults(reducedMerged, targetBuckets)

	// --- Calculate Reduced Metrics ---
	if reducedAdaptive != nil {
		adapAvail := Availability(reducedAdaptive)
		adapMean := MeanLatency(reducedAdaptive)
		adapP50 := PercentileLatency(reducedAdaptive, 0.50)
		adapP99 := PercentileLatency(reducedAdaptive, 0.99)
		t.Logf("Adaptive Reduced (%d buckets): Avail=%.4f (%.2f%%), Mean=%.6fs (%.2f%%), P50=%.6fs (%.2f%%), P99=%.6fs (%.2f%%)",
			reducedAdaptive.Len(),
			adapAvail, percentDiff(origAvail, adapAvail),
			adapMean, percentDiff(origMean, adapMean),
			adapP50, percentDiff(origP50, adapP50),
			adapP99, percentDiff(origP99, adapP99))
	}

	if reducedMerged != nil {
		mergedAvail := Availability(reducedMerged)
		mergedMean := MeanLatency(reducedMerged)
		mergedP50 := PercentileLatency(reducedMerged, 0.50)
		mergedP99 := PercentileLatency(reducedMerged, 0.99)
		t.Logf("Merged Adjacent (%d buckets): Avail=%.4f (%.2f%%), Mean=%.6fs (%.2f%%), P50=%.6fs (%.2f%%), P99=%.6fs (%.2f%%)",
			reducedMerged.Len(),
			mergedAvail, percentDiff(origAvail, mergedAvail),
			mergedMean, percentDiff(origMean, mergedMean),
			mergedP50, percentDiff(origP50, mergedP50),
			mergedP99, percentDiff(origP99, mergedP99))
	}

	if reducedMergedThenAdaptive != nil {
		combAvail := Availability(reducedMergedThenAdaptive)
		combMean := MeanLatency(reducedMergedThenAdaptive)
		combP50 := PercentileLatency(reducedMergedThenAdaptive, 0.50)
		combP99 := PercentileLatency(reducedMergedThenAdaptive, 0.99)
		t.Logf("Merged+Adaptive (%d buckets): Avail=%.4f (%.2f%%), Mean=%.6fs (%.2f%%), P50=%.6fs (%.2f%%), P99=%.6fs (%.2f%%)",
			reducedMergedThenAdaptive.Len(),
			combAvail, percentDiff(origAvail, combAvail),
			combMean, percentDiff(origMean, combMean),
			combP50, percentDiff(origP50, combP50),
			combP99, percentDiff(origP99, combP99))
	}
}

// Helper for percentage difference
func percentDiff(original, current float64) float64 {
	if original == 0 {
		if current == 0 {
			return 0.0
		}
		return math.Inf(1) // Or NaN? Indicate undefined change from zero
	}
	return math.Abs((current-original)/original) * 100.0
}
