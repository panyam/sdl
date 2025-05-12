package components

import (
	"math/rand"
	"time"

	sc "github.com/panyam/sdl/core"
	// "math" // If needed later
)

/*
Core Concepts:

1. Writes (Ingest): Typically very fast. Data is written sequentially to an in-memory structure (memtable) and often appended to a Write-Ahead Log (WAL) on persistent storage (disk).
2. Memtable Flush: When the memtable fills up, it's flushed to disk as an immutable "Sorted String Table" (SSTable) file at Level 0 (L0). This involves a potentially larger disk write.
3. Reads: More complex. To find a key, the query must check:
	* The current memtable (in memory - fast).
	* Potentially multiple SSTables at L0 (in memory via bloom filters/index blocks, or requiring disk reads). L0 files can overlap in key ranges.
	* SSTables at deeper levels (L1, L2, ... Ln). Files at these levels typically do not overlap within the level. Reading might involve checking bloom filters/index blocks per level and potentially reading data blocks from disk for the relevant SSTable. Reads might hit data cached in OS page cache or block cache.
	* Deletes are often handled via "tombstone" markers, adding complexity to reads.
4. Compactions: Background processes that merge SSTables from one level to the next (e.g., L0 -> L1, L1 -> L2).
	* Reads SSTables from level i.
	* Writes new, larger, non-overlapping SSTables to level i+1.
	* Crucially, compactions consume disk I/O and CPU, potentially impacting foreground read/write latencies (Write/Read Amplification).

We will use a simplified model for this and wont implement SSTables or memtable flushes precisely but capture the probabilisitc performance impact of these operations.
*/

// LSMTree represents a Log-Structured Merge Tree index structure.
type LSMTree struct {
	Index // Embed base index properties (Disk, PageSize, RecordSize etc.)

	// LSM Specific Parameters (Probabilities 0.0 to 1.0)
	MemtableHitProb float64 // Probability read finds data in memtable
	Level0HitProb   float64 // Probability read finds data in L0 (given miss in memtable)
	Levels          int     // Number of levels modelled beyond L0 (e.g., L1, L2...)

	// Amplification Factors (Simplified Averages)
	// Average number of "blocks" or "checks" needed per read operation
	// after missing memtable and L0. Represents checks at deeper levels.
	ReadAmpFactor float64
	// Factor representing write overhead beyond simple memtable/WAL write,
	// amortizing flush/compaction costs impacting foreground writes.
	WriteAmpFactor float64

	// Compaction Interference
	CompactionImpactProb float64            // Probability foreground op is slowed by compaction
	CompactionSlowdown   Outcomes[Duration] // Extra latency when impacted

	// Internal state for calculations (can be precomputed)
	rng *rand.Rand
	// Precomputed outcomes if needed (or calculate on the fly)
}

// Init initializes the LSMTree with default or provided parameters.
func (lsm *LSMTree) Init() *LSMTree {
	lsm.Index.Init() // Initialize base Index (Disk, sizes etc.)

	// --- Set Default LSM Parameters ---
	lsm.MemtableHitProb = 0.10                 // 10% chance read hits memtable (configurable based on size/workload)
	lsm.Level0HitProb = 0.30                   // 30% chance L0 hit (given memtable miss)
	lsm.Levels = 4                             // Model L1, L2, L3, L4
	lsm.ReadAmpFactor = 3.0                    // Avg 3 extra block checks/reads for deeper levels
	lsm.WriteAmpFactor = 1.5                   // Avg write cost is 1.5x the base WAL/memtable op cost
	lsm.CompactionImpactProb = 0.05            // 5% chance of interference
	lsm.CompactionSlowdown.Add(100, Millis(5)) // Default 5ms slowdown when impacted

	// Initialize RNG
	lsm.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Ensure RecordProcessingTime (from Index) is initialized if not already
	if lsm.Index.RecordProcessingTime.Len() == 0 {
		lsm.Index.RecordProcessingTime.Add(100, Nanos(100)) // Default from Index.Init
	}
	// Ensure CompactionSlowdown has an And func if needed for composition later
	// lsm.CompactionSlowdown.And = func(a,b Duration) Duration { return a+b } // If needed

	return lsm
}

// NewLSMTree creates and initializes a new LSMTree component.
func NewLSMTree() *LSMTree {
	lsm := &LSMTree{}
	return lsm.Init()
}

// Write simulates inserting or updating data in the LSM tree.
func (lsm *LSMTree) Write() *Outcomes[sc.AccessResult] {

	// 1. Base cost: Memtable update (CPU) + WAL write (Disk write)
	// We approximate this as RecordProcessingTime + one Disk Write.
	baseWriteCost := sc.And(
		sc.Map(&lsm.RecordProcessingTime, func(p Duration) sc.AccessResult { return sc.AccessResult{true, p} }), // Map processing time to sc.AccessResult
		lsm.Disk.Write(),    // Assumes WAL write has similar profile to general disk write
		sc.AndAccessResults, // Combine CPU + Disk Write
	)

	// 2. Add Write Amplification Factor
	// Multiply the latency of successful base writes by WriteAmpFactor.
	// This is a simplification - real WA adds IOs, not just scales latency.
	amplifiedWriteCost := sc.Map(baseWriteCost, func(ar sc.AccessResult) sc.AccessResult {
		if ar.Success {
			ar.Latency *= lsm.WriteAmpFactor
		}
		return ar
	})

	// 3. Factor in Compaction Interference
	// With CompactionImpactProb, add CompactionSlowdown latency to the outcome.
	finalOutcomes := &Outcomes[sc.AccessResult]{And: amplifiedWriteCost.And} // Use AndAccessResults
	compSlowdownOutcomes := &lsm.CompactionSlowdown                          // Pointer for use in And

	for _, bucket := range amplifiedWriteCost.Buckets {
		// Probability this bucket is NOT impacted by compaction
		noImpactProb := bucket.Weight * (1.0 - lsm.CompactionImpactProb)
		if noImpactProb > 1e-9 {
			finalOutcomes.Add(noImpactProb, bucket.Value) // Add outcome as is
		}

		// Probability this bucket IS impacted by compaction
		impactProbBase := bucket.Weight * lsm.CompactionImpactProb
		if impactProbBase > 1e-9 {
			// For each possible compaction slowdown, calculate the combined outcome
			for _, slowBucket := range compSlowdownOutcomes.Buckets {
				impactProb := impactProbBase * (slowBucket.Weight / compSlowdownOutcomes.TotalWeight()) // Scale by slowdown probability
				if impactProb > 1e-9 {
					// Combine the amplified write result with the slowdown
					finalValue := bucket.Value // Copy original value
					if finalValue.Success {    // Only add slowdown to successful writes
						finalValue.Latency += slowBucket.Value // Add slowdown duration
					}
					finalOutcomes.Add(impactProb, finalValue)
				}
			}
		}
	}

	// 4. Apply Reduction (optional, but recommended as compaction step increases buckets)
	// Use the standard TrimToSize defined for sc.AccessResult
	// Need to split -> trim -> append
	successes, failures := finalOutcomes.Split(sc.AccessResult.IsSuccess)
	trimmer := sc.TrimToSize(100, lsm.MaxOutcomeLen) // Use MaxOutcomeLen from Index
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalTrimmedOutcomes := (&Outcomes[sc.AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalTrimmedOutcomes
}

// Read simulates finding a key in the LSM tree.
func (lsm *LSMTree) Read() *Outcomes[sc.AccessResult] {
	// --- Define costs for different read steps ---

	// Cost of checking Memtable (pure CPU)
	memtableCheckCost := sc.Map(&lsm.RecordProcessingTime, func(p Duration) sc.AccessResult {
		// Assume checking memtable always 'succeeds' in the sense the check completes.
		// Whether the key is *found* is handled by probabilities later.
		return sc.AccessResult{true, p}
	})

	// Cost of checking L0 (Bloom filter/index check (CPU) + potential Disk Read)
	// Simplified: CPU check + 1 Disk Read (could refine later)
	l0CheckCost := sc.And(
		sc.Map(&lsm.RecordProcessingTime, func(p Duration) sc.AccessResult { return sc.AccessResult{true, p} }),
		lsm.Disk.Read(), // Assume need one read for L0 check if memtable misses
		sc.AndAccessResults,
	)

	// Cost of checking deeper levels (CPU checks + ReadAmpFactor * Disk Reads)
	// Create base cost: CPU + single disk read
	deepLevelBaseCheckCost := sc.And(
		sc.Map(&lsm.RecordProcessingTime, func(p Duration) sc.AccessResult { return sc.AccessResult{true, p} }),
		lsm.Disk.Read(),
		sc.AndAccessResults,
	)
	// Amplify the latency part for successes
	deepLevelCheckCost := sc.Map(deepLevelBaseCheckCost, func(ar sc.AccessResult) sc.AccessResult {
		if ar.Success {
			// We approximate ReadAmpFactor * DiskReads by scaling latency.
			// More accurate: model N reads sequentially/parallelly.
			ar.Latency *= lsm.ReadAmpFactor
		}
		return ar
	})

	// --- Compose the read path probabilistically ---

	// Start with Memtable check
	currentOutcomes := memtableCheckCost

	// If Memtable MISSES (prob 1 - MemtableHitProb), add cost of L0 check.
	// Note: We assume the *check* cost is incurred regardless of hit/miss,
	// but the *path* continues probabilistically.

	// Prob(Hit Memtable) = MemtableHitProb
	// Prob(Miss Memtable) = 1 - MemtableHitProb

	// Prob(Hit L0) = Prob(Miss Memtable) * Level0HitProb
	// Prob(Miss L0) = Prob(Miss Memtable) * (1 - Level0HitProb)

	// Prob(Hit Deep) = Prob(Miss L0)
	// Prob(Miss Deep) = 0 (assume key must exist if searched)

	// Path 1: Hit Memtable (Cost = Memtable Check) -> Final Weight = MemtableHitProb
	memtableHits := sc.Map(currentOutcomes, func(ar sc.AccessResult) sc.AccessResult { return ar }) // Keep as is
	memtableHits.ScaleWeights(lsm.MemtableHitProb)                                                  // Adjust probability

	// Path 2: Miss Memtable, Check L0
	memtableMissProb := 1.0 - lsm.MemtableHitProb
	if memtableMissProb < 1e-9 {
		memtableMissProb = 0
	} // Avoid negligible paths

	l0CheckPath := sc.And(currentOutcomes, l0CheckCost, sc.AndAccessResults) // Cost if we check L0

	// Path 2a: Hit L0 (Cost = Memtable Check + L0 Check) -> Final Weight = memtableMissProb * Level0HitProb
	l0Hits := sc.Map(l0CheckPath, func(ar sc.AccessResult) sc.AccessResult { return ar })
	l0Hits.ScaleWeights(memtableMissProb * lsm.Level0HitProb)

	// Path 2b: Miss L0, Check Deep Levels
	l0MissProb := memtableMissProb * (1.0 - lsm.Level0HitProb)
	if l0MissProb < 1e-9 {
		l0MissProb = 0
	}

	deepCheckPath := sc.And(l0CheckPath, deepLevelCheckCost, sc.AndAccessResults) // Cost if we check deep

	// Path 3: Hit Deep Levels (Cost = Memtable Check + L0 Check + Deep Check) -> Final Weight = l0MissProb
	deepHits := sc.Map(deepCheckPath, func(ar sc.AccessResult) sc.AccessResult { return ar })
	deepHits.ScaleWeights(l0MissProb)

	// Combine all paths
	combinedReadOutcomes := (&Outcomes[sc.AccessResult]{}).Append(memtableHits, l0Hits, deepHits)

	// Factor in Compaction Interference (same logic as in Write)
	finalOutcomes := &Outcomes[sc.AccessResult]{And: combinedReadOutcomes.And}
	compSlowdownOutcomes := &lsm.CompactionSlowdown

	for _, bucket := range combinedReadOutcomes.Buckets {
		noImpactProb := bucket.Weight * (1.0 - lsm.CompactionImpactProb)
		if noImpactProb > 1e-9 {
			finalOutcomes.Add(noImpactProb, bucket.Value)
		}

		impactProbBase := bucket.Weight * lsm.CompactionImpactProb
		if impactProbBase > 1e-9 {
			for _, slowBucket := range compSlowdownOutcomes.Buckets {
				impactProb := impactProbBase * (slowBucket.Weight / compSlowdownOutcomes.TotalWeight())
				if impactProb > 1e-9 {
					finalValue := bucket.Value
					if finalValue.Success {
						finalValue.Latency += slowBucket.Value
					}
					finalOutcomes.Add(impactProb, finalValue)
				}
			}
		}
	}

	// Apply Reduction
	successes, failures := finalOutcomes.Split(sc.AccessResult.IsSuccess)
	trimmer := sc.TrimToSize(100, lsm.MaxOutcomeLen) // Use MaxOutcomeLen from Index
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalTrimmedOutcomes := (&Outcomes[sc.AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalTrimmedOutcomes
}
