package sdl

import (
	"math"
	"math/rand"
	"time"
)

// HashIndex represents a hash-based index structure (e.g., static, extendible, linear)
type HashIndex struct {
	Index // Embed base Index

	// Hash Specific Parameters
	LoadFactorThreshold float64 // Load factor above which resizes become likely
	AvgOverflowReads    float64 // Average extra reads on collision/overflow
	ResizeCostFactor    float64 // Multiplier for disk R/W during a resize operation

	// Internal state / derived values
	rng *rand.Rand
	// numBuckets uint // Could explicitly track number of buckets if needed
}

// NewHashIndex creates and initializes a new HashIndex component.
func NewHashIndex() *HashIndex {
	out := &HashIndex{}
	return out.Init()
}

// Init initializes the HashIndex with defaults.
func (h *HashIndex) Init() *HashIndex {
	h.Index.Init()
	h.LoadFactorThreshold = 0.75 // Typical threshold for resizing
	h.AvgOverflowReads = 0.2     // Default: 20% chance of needing 1 extra read on avg
	h.ResizeCostFactor = 1.5     // Default: Resize costs 1.5x a full scan/write
	h.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	return h
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

// Probability of collision needing overflow access
func (h *HashIndex) collisionProbability() float64 {
	// Heuristic: Increases as load factor approaches 1.0
	loadFactor := h.currentLoadFactor()
	// Simple quadratic scaling - adjust as needed
	prob := math.Pow(loadFactor, 2) * h.AvgOverflowReads
	if prob > 1.0 {
		prob = 1.0
	}
	if prob < 0 {
		prob = 0
	}
	return prob
}

// Probability of needing a resize on insert
func (h *HashIndex) resizeProbability() float64 {
	loadFactor := h.currentLoadFactor()
	if loadFactor > h.LoadFactorThreshold {
		// Probability increases sharply after threshold
		// Simple linear increase from 0% at threshold to 100% at 1.0 load factor?
		if h.LoadFactorThreshold >= 1.0 {
			return 0
		} // Avoid division by zero
		prob := (loadFactor - h.LoadFactorThreshold) / (1.0 - h.LoadFactorThreshold)
		if prob > 1.0 {
			prob = 1.0
		}
		if prob < 0 {
			prob = 0
		}
		return prob * 0.1 // Make resize less frequent (e.g., 10% chance even if LF=1.0)
	}
	return 0.0
}

// --- Refined Find ---
// Find searches for a key in the hash index.
func (h *HashIndex) Find() *Outcomes[AccessResult] {
	// 1. Hash calculation (CPU)
	hashCpuCost := Map(&h.RecordProcessingTime, func(p Duration) AccessResult {
		return AccessResult{true, p * 0.5} // Assume hashing is faster than general processing
	})

	// 2. Read primary bucket page
	readPrimaryCost := h.Disk.Read()

	// 3. Cost for potential overflow reads
	pCollision := h.collisionProbability()
	// Cost = 1 Disk Read per overflow read, scaled by AvgOverflowReads
	overflowReadCostBase := h.Disk.Read()
	overflowReadCost := Map(overflowReadCostBase, func(ar AccessResult) AccessResult {
		if ar.Success {
			ar.Latency *= h.AvgOverflowReads
		} else {
			ar.Latency = 0
		} // Assume failure happens quickly
		return ar
	})

	// --- Combine ---
	// Base cost: Hash CPU -> Read Primary Page
	baseCost := And(hashCpuCost, readPrimaryCost, AndAccessResults)

	// Add overflow cost probabilistically
	finalCost := &Outcomes[AccessResult]{And: baseCost.And}
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
			combinedOverflow := And((&Outcomes[AccessResult]{}).Add(1.0, bucket.Value), overflowReadCost, AndAccessResults)
			// Scale weights and add
			for _, ovfBucket := range combinedOverflow.Buckets {
				finalCost.Add(overflowProbBase*ovfBucket.Weight, ovfBucket.Value)
			}
		}
	}

	// Apply Reduction
	successes, failures := finalCost.Split(AccessResult.IsSuccess)
	maxLen := h.MaxOutcomeLen
	if maxLen <= 0 {
		maxLen = 5
	}
	trimmer := TrimToSize(100, maxLen)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalOutcomes := (&Outcomes[AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalOutcomes
}

// --- Refined Insert ---
// Insert adds a new key to the hash index.
func (h *HashIndex) Insert() *Outcomes[AccessResult] {
	// 1. Find cost (Hash CPU + Read Primary + Prob Overflow Read)
	// Note: Find already includes reduction. Let's calculate base cost without reduction first.
	// Calculate Hash + Read Primary only first
	hashCpuCost := Map(&h.RecordProcessingTime, func(p Duration) AccessResult { return AccessResult{true, p * 0.5} })
	readPrimaryCost := h.Disk.Read()
	findBaseCost := And(hashCpuCost, readPrimaryCost, AndAccessResults)

	// Add cost of reading overflow pages probabilistically (if collision)
	pCollision := h.collisionProbability()
	overflowReadCostBase := h.Disk.Read()
	overflowReadCost := Map(overflowReadCostBase, func(ar AccessResult) AccessResult {
		if ar.Success {
			ar.Latency *= h.AvgOverflowReads
		} else {
			ar.Latency = 0
		}
		return ar
	})
	// Combine base read + potential overflow read cost
	findTotalReadCost := &Outcomes[AccessResult]{And: findBaseCost.And}
	for _, bucket := range findBaseCost.Buckets {
		noOverflowProb := bucket.Weight * (1.0 - pCollision)
		if noOverflowProb > 1e-9 {
			findTotalReadCost.Add(noOverflowProb, bucket.Value)
		}
		overflowProbBase := bucket.Weight * pCollision
		if overflowProbBase > 1e-9 {
			combinedOverflow := And((&Outcomes[AccessResult]{}).Add(1.0, bucket.Value), overflowReadCost, AndAccessResults)
			for _, ovfBucket := range combinedOverflow.Buckets {
				findTotalReadCost.Add(overflowProbBase*ovfBucket.Weight, ovfBucket.Value)
			}
		}
	}

	// 2. Modify page/slot (CPU cost)
	modifyCpuCost := Map(&h.RecordProcessingTime, func(p Duration) AccessResult {
		return AccessResult{true, p}
	})

	// 3. Write data to the page (Primary or Overflow)
	// Assume 1 write covers the necessary update.
	writeCost := h.Disk.Write()

	// 4. Cost of potential resize operation
	pResize := h.resizeProbability()
	// Simplified resize cost: NumPages * Factor * (Read+Write) - very expensive
	numPages := float64(h.NumPages())
	if numPages == 0 {
		numPages = 1
	}
	resizeRwCostBase := And(h.Disk.Read(), h.Disk.Write(), AndAccessResults)
	resizeCost := Map(resizeRwCostBase, func(ar AccessResult) AccessResult {
		if ar.Success {
			ar.Latency *= numPages * h.ResizeCostFactor
		} else {
			ar.Latency = 0
		} // Assume resize failure is catastrophic but quick? Or part of main path failure?
		return ar
	})

	// --- Combine ---
	// Cost without resize: FindRead -> ModifyCPU -> WriteData
	costPath1 := And(findTotalReadCost, modifyCpuCost, AndAccessResults)
	costPath1 = And(costPath1, writeCost, AndAccessResults)

	// Add resize cost probabilistically
	finalCost := &Outcomes[AccessResult]{And: costPath1.And}
	for _, bucket := range costPath1.Buckets {
		// Path without resize
		noResizeProb := bucket.Weight * (1.0 - pResize)
		if noResizeProb > 1e-9 {
			finalCost.Add(noResizeProb, bucket.Value)
		}
		// Path with resize
		resizeProbBase := bucket.Weight * pResize
		if resizeProbBase > 1e-9 {
			// Combine non-resize path outcome with resize cost outcome
			combinedResize := And((&Outcomes[AccessResult]{}).Add(1.0, bucket.Value), resizeCost, AndAccessResults)
			for _, resizeBucket := range combinedResize.Buckets {
				finalCost.Add(resizeProbBase*resizeBucket.Weight, resizeBucket.Value)
			}
		}
	}

	// Apply Reduction
	successes, failures := finalCost.Split(AccessResult.IsSuccess)
	maxLen := h.MaxOutcomeLen
	if maxLen <= 0 {
		maxLen = 5
	}
	trimmer := TrimToSize(100, maxLen)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalOutcomes := (&Outcomes[AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalOutcomes
}

// --- Refined Delete ---
// Delete removes a key from the hash index.
func (h *HashIndex) Delete() *Outcomes[AccessResult] {
	// Cost is typically: Find -> Modify CPU -> Write
	// Similar to Insert, but without resize probability.

	// 1. Find cost (includes potential overflow reads)
	findCost := h.Find() // Use the Find method directly

	// 2. Modify page/slot (CPU cost)
	modifyCpuCost := Map(&h.RecordProcessingTime, func(p Duration) AccessResult {
		return AccessResult{true, p}
	})

	// 3. Write data to the page
	writeCost := h.Disk.Write()

	// Combine: Find -> Modify CPU -> Write
	step1 := findCost
	step2 := And(step1, modifyCpuCost, AndAccessResults)
	finalCost := And(step2, writeCost, AndAccessResults)

	// Apply Reduction
	successes, failures := finalCost.Split(AccessResult.IsSuccess)
	maxLen := h.MaxOutcomeLen
	if maxLen <= 0 {
		maxLen = 5
	}
	trimmer := TrimToSize(100, maxLen)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalOutcomes := (&Outcomes[AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalOutcomes
}
