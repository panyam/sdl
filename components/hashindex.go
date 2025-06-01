package components

import (
	"math"

	sc "github.com/panyam/sdl/core"
)

// HashIndex represents a hash-based index structure (e.g., static, extendible, linear)
type HashIndex struct {
	Index // Embed base Index

	// Hash Specific Parameters
	// LoadFactorThreshold float64 // Load factor above which resizes become likely
	AvgOverflowReads float64 // Average extra reads on collision/overflow
	ResizeCostFactor float64 // Multiplier for disk R/W during a resize operation

	// Internal state / derived values
	// rng *rand.Rand
	// numBuckets uint // Could explicitly track number of buckets if needed
}

// NewHashIndex creates and initializes a new HashIndex component.
func NewHashIndex() (out *HashIndex) {
	out = &HashIndex{}
	out.Init()
	return
}

// Init initializes the HashIndex with defaults.
func (h *HashIndex) Init() {
	h.Index.Init()
	// h.LoadFactorThreshold = 0.75 // Typical threshold for resizing
	h.AvgOverflowReads = 0.2 // Default: 20% chance of needing 1 extra read on avg
	h.ResizeCostFactor = 1.5 // Default: Resize costs 1.5x a full scan/write
	// h.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Probability of collision needing overflow access
func (h *HashIndex) collisionProbability() float64 {
	// Heuristic: Increases slowly with log of NumRecords.
	// Adjust constants based on desired sensitivity.
	if h.NumRecords == 0 {
		return 0
	}
	logRecords := math.Log10(float64(h.NumRecords)) // Base 10 log
	// Example scaling: Starts low, reaches ~20% around 1M records?
	// prob = (logRecords / 6.0) * h.AvgOverflowReads
	// Let's make it increase faster initially
	prob := (logRecords / 7.0) * (logRecords / 7.0) * h.AvgOverflowReads // Quadratic scaling with log(N)

	if prob > 1.0 {
		prob = 1.0
	}
	if prob < 0 {
		prob = 0
	}
	return prob
}

// Calculates current approximate load factor
func (h *HashIndex) currentLoadFactor() float64 {
	// Estimate number of "slots" based on NumRecords at typical occupancy (e.g., 70%)
	// This is approximate as it doesn't know the actual bucket count.
	// A better model might track numBuckets explicitly.
	numSlots := float64(h.NumRecords) / 0.70 // Estimate total slots needed
	if numSlots < 1 {
		numSlots = 1
	}
	// Estimate pages needed for these slots
	// slotsPerPage := h.PageSize / uint64(h.RecordSize) // Simplified: assuming record size dictates slot size
	// numPagesRequired := numSlots / float64(slotsPerPage)
	// Use NumPages() as proxy for number of buckets/pages currently allocated
	numCurrentPages := float64(h.NumPages())
	if numCurrentPages == 0 {
		return 0
	}

	// Load Factor = Records / (Pages * RecordsPerPage) - very approximate
	// Let's use a simpler proxy: ratio of records to pages? Doesn't account for page size well.
	// Alternative: Define NumBuckets explicitly in the struct?

	// --- Simpler approach for now: ---
	// Probability of collision/overflow increases as NumRecords grows relative to some base capacity.
	// Let's use NumPages as a proxy for capacity.
	// Assume overflow/resize probability scales somehow with NumRecords / NumPages.
	// This is very heuristic.
	recordsPerPageEstimate := 1.0
	if h.RecordSize > 0 {
		recordsPerPageEstimate = float64(h.PageSize / uint64(h.RecordSize))
	}
	if recordsPerPageEstimate < 1 {
		recordsPerPageEstimate = 1
	}

	numSlotsEstimate := numCurrentPages * recordsPerPageEstimate
	if numSlotsEstimate == 0 {
		return 1.0
	} // Avoid division by zero, assume full if no slots
	return float64(h.NumRecords) / numSlotsEstimate
}

// Probability of needing a resize on insert
func (h *HashIndex) resizeProbability() float64 {
	// Heuristic: Negligible chance for small N, then increases slowly with log(N),
	// representing the *chance* that an insert pushes it over a threshold.
	minResizeRecords := 10000.0 // Don't consider resize below this many records

	if float64(h.NumRecords) < minResizeRecords {
		return 0.0001 // Very small base chance even for small tables? Or 0?
	}

	// Scale probability based on log10 of records *beyond* the minimum
	logScale := math.Log10(float64(h.NumRecords) / minResizeRecords) // How many orders of magnitude bigger are we?

	// Example scaling: Max probability of ~10% resize chance for a given insert?
	// Let's say prob reaches 10% when logScale = 3 (1000x min = 10M records)
	maxProb := 0.10
	scaleFactor := 3.0
	prob := (logScale / scaleFactor) * maxProb

	// Add a small base probability
	prob += 0.001

	if prob > maxProb {
		prob = maxProb
	} // Cap the max probability
	if prob < 0 {
		prob = 0
	}

	return prob
}

// --- Refined Find ---
// Find searches for a key in the hash index.
func (h *HashIndex) Find() *Outcomes[sc.AccessResult] {
	// 1. Hash calculation (CPU)
	hashCpuCost := sc.Map(&h.RecordProcessingTime, func(p Duration) sc.AccessResult {
		return sc.AccessResult{true, p * 0.5} // Assume hashing is faster than general processing
	})

	// 2. Read primary bucket page
	readPrimaryCost := h.Disk.Read()

	// 3. Cost for potential overflow reads
	pCollision := h.collisionProbability()
	// Cost = 1 Disk Read per overflow read, scaled by AvgOverflowReads
	overflowReadCostBase := h.Disk.Read()
	overflowReadCost := sc.Map(overflowReadCostBase, func(ar sc.AccessResult) sc.AccessResult {
		if ar.Success {
			ar.Latency *= h.AvgOverflowReads
		} else {
			ar.Latency = 0
		} // Assume failure happens quickly
		return ar
	})

	// --- Combine ---
	// Base cost: Hash CPU -> Read Primary Page
	baseCost := sc.And(hashCpuCost, readPrimaryCost, sc.AndAccessResults)

	// Add overflow cost probabilistically
	finalCost := &Outcomes[sc.AccessResult]{And: baseCost.And}
	for _, bucket := range baseCost.Buckets {
		// Path without overflow read
		noOverflowProb := bucket.Weight * (1.0 - pCollision)
		if noOverflowProb > 1e-9 {
			finalCost.Add(noOverflowProb, bucket.Value)
		}
		// Path with overflow read
		overflowProbBase := bucket.Weight * pCollision
		if overflowProbBase > 1e-9 {
			// Combine baseCost outcome with overflowReadCost outcome
			combinedOverflow := sc.And((&Outcomes[sc.AccessResult]{}).Add(1.0, bucket.Value), overflowReadCost, sc.AndAccessResults)
			// Scale weights and add
			for _, ovfBucket := range combinedOverflow.Buckets {
				finalCost.Add(overflowProbBase*ovfBucket.Weight, ovfBucket.Value)
			}
		}
	}

	// Apply Reduction
	successes, failures := finalCost.Split(sc.AccessResult.IsSuccess)
	maxLen := h.MaxOutcomeLen
	if maxLen <= 0 {
		maxLen = 5
	}
	trimmer := sc.TrimToSize(100, maxLen)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalOutcomes := (&Outcomes[sc.AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalOutcomes
}

// --- Refined Insert ---
// Insert adds a new key to the hash index.
func (h *HashIndex) Insert() *Outcomes[sc.AccessResult] {
	// ... (Calculate findTotalReadCost as before using collisionProbability) ...
	findTotalReadCost := h.Find() // Or recalculate parts if needed

	// ... (ModifyCPU and WriteCost as before) ...
	modifyCpuCost := sc.Map(&h.RecordProcessingTime, func(p Duration) sc.AccessResult { return sc.AccessResult{true, p} })
	writeCost := h.Disk.Write()

	// --- Cost of potential resize operation ---
	pResize := h.resizeProbability() // Uses the new heuristic
	numPages := float64(h.NumPages())
	if numPages == 0 {
		numPages = 1
	}
	resizeRwCostBase := sc.And(h.Disk.Read(), h.Disk.Write(), sc.AndAccessResults)
	resizeCost := sc.Map(resizeRwCostBase, func(ar sc.AccessResult) sc.AccessResult {
		if ar.Success {
			ar.Latency *= numPages * h.ResizeCostFactor
		} else {
			ar.Latency = 0
		}
		return ar
	})

	// --- Combine ---
	costPath1 := sc.And(findTotalReadCost, modifyCpuCost, sc.AndAccessResults)
	costPath1 = sc.And(costPath1, writeCost, sc.AndAccessResults)

	// Add resize cost probabilistically (logic remains the same)
	finalCost := &Outcomes[sc.AccessResult]{And: costPath1.And}
	for _, bucket := range costPath1.Buckets {
		noResizeProb := bucket.Weight * (1.0 - pResize)
		if noResizeProb > 1e-9 {
			finalCost.Add(noResizeProb, bucket.Value)
		}
		resizeProbBase := bucket.Weight * pResize
		if resizeProbBase > 1e-9 {
			combinedResize := sc.And((&Outcomes[sc.AccessResult]{}).Add(1.0, bucket.Value), resizeCost, sc.AndAccessResults)
			for _, resizeBucket := range combinedResize.Buckets {
				finalCost.Add(resizeProbBase*resizeBucket.Weight, resizeBucket.Value)
			}
		}
	}

	// --- Reduction --- (Remains the same)
	successes, failures := finalCost.Split(sc.AccessResult.IsSuccess)
	maxLen := h.MaxOutcomeLen
	if maxLen <= 0 {
		maxLen = 5
	}
	trimmer := sc.TrimToSize(100, maxLen)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalOutcomes := (&Outcomes[sc.AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalOutcomes
}

// --- Refined Delete ---
// Delete removes a key from the hash index.
func (h *HashIndex) Delete() *Outcomes[sc.AccessResult] {
	// Cost is typically: Find -> Modify CPU -> Write
	// Similar to Insert, but without resize probability.

	// 1. Find cost (includes potential overflow reads)
	findCost := h.Find() // Use the Find method directly

	// 2. Modify page/slot (CPU cost)
	modifyCpuCost := sc.Map(&h.RecordProcessingTime, func(p Duration) sc.AccessResult {
		return sc.AccessResult{true, p}
	})

	// 3. Write data to the page
	writeCost := h.Disk.Write()

	// Combine: Find -> Modify CPU -> Write
	step1 := findCost
	step2 := sc.And(step1, modifyCpuCost, sc.AndAccessResults)
	finalCost := sc.And(step2, writeCost, sc.AndAccessResults)

	// Apply Reduction
	successes, failures := finalCost.Split(sc.AccessResult.IsSuccess)
	maxLen := h.MaxOutcomeLen
	if maxLen <= 0 {
		maxLen = 5
	}
	trimmer := sc.TrimToSize(100, maxLen)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalOutcomes := (&Outcomes[sc.AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalOutcomes
}
