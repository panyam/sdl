package components

import (
	"math"

	sc "github.com/panyam/leetcoach/sdl/core"
)

// BTreeIndex represents a clustered B-Tree index structure
type BTreeIndex struct {
	Index // Embed base Index

	// BTree Specific Parameters
	Occupancy  float64 // Target page occupancy (0 to 1.0)
	NodeFanout int     // Max keys/pointers per node (order)

	// Average number of extra Read/Write pairs incurred due to split propagation on Insert
	AvgSplitPropCost float64
	// Average number of extra Read/Write pairs incurred due to merge/redistribute on Delete
	AvgMergePropCost float64
}

func NewBTreeIndex() *BTreeIndex {
	out := &BTreeIndex{}
	return out.Init()
}

// Init initializes the BTreeIndex with defaults.
func (bt *BTreeIndex) Init() *BTreeIndex {
	bt.Index.Init()
	bt.NodeFanout = 100       // Default fanout (e.g., for 8KB pages, ~100 keys/pointers)
	bt.Occupancy = 0.67       // Default occupancy (B-Trees aim for >50%)
	bt.AvgSplitPropCost = 0.1 // Default: Avg 0.1 extra R/W pairs per insert (low split chance)
	bt.AvgMergePropCost = 0.1 // Default: Avg 0.1 extra R/W pairs per delete (low merge chance)
	return bt
}

// RecordsPerPage estimates how many records fit in a leaf page based on occupancy.
// Note: Inner nodes store keys/pointers, leaf nodes store records. This calculation
// is more relevant for leaf capacity. Fanout determines inner node capacity.
func (bt *BTreeIndex) RecordsPerPage() uint64 {
	if bt.RecordSize == 0 {
		return 0
	}
	// Use PageSize and RecordSize from embedded Index
	recordsFit := float64(bt.PageSize / uint64(bt.RecordSize))
	return uint64(bt.Occupancy * recordsFit)
}

// Height estimates the height of the B-Tree.
func (bt *BTreeIndex) Height() int {
	numPages := float64(bt.NumPages()) // Total pages needed for NumRecords
	fanout := float64(bt.NodeFanout)
	if numPages <= 1.0 || fanout <= 1.0 {
		return 1 // Min height is 1 (root only)
	}
	// Height = log_fanout(NumPages)
	h := math.Log(numPages) / math.Log(fanout)
	return int(math.Ceil(h)) + 1 // Add 1 for leaf level, ceiling for integer height
}

// --- Refined Find ---
// Find searches for a key in the B-Tree. Traverses from root to leaf.
func (bt *BTreeIndex) Find() *Outcomes[sc.AccessResult] {
	height := bt.Height()
	if height <= 0 {
		height = 1
	}

	// Cost of searching within a node (log2 of fanout * processing time)
	log2Fanout := math.Log2(float64(bt.NodeFanout))
	if log2Fanout < 1 {
		log2Fanout = 1
	} // Min 1 comparison
	nodeSearchCpuCost := sc.Map(&bt.RecordProcessingTime, func(p Duration) sc.AccessResult {
		// Assume search within node always succeeds in finding path/record
		return sc.AccessResult{true, p * log2Fanout}
	})

	// Start with zero cost outcome
	currentTotalCost := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
	currentTotalCost.Add(1.0, sc.AccessResult{true, 0}) // Initial state: success, zero latency

	// Traverse levels: For each level (including root and leaf), read page + search node
	for i := 0; i < height; i++ {
		levelCost := sc.And(bt.Disk.Read(), nodeSearchCpuCost, sc.AndAccessResults)
		currentTotalCost = sc.And(currentTotalCost, levelCost, sc.AndAccessResults)

		// Optional: Apply reduction within the loop if height is large
		if i > 0 && i%3 == 0 { // Reduce every few levels
			successes, failures := currentTotalCost.Split(sc.AccessResult.IsSuccess)
			maxLen := bt.MaxOutcomeLen
			if maxLen <= 0 {
				maxLen = 5
			}
			trimmer := sc.TrimToSize(100, maxLen) // Uses Merge+Interpolate
			trimmedSuccesses := trimmer(successes)
			trimmedFailures := trimmer(failures)
			currentTotalCost = (&Outcomes[sc.AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)
		}
	}

	// Final reduction outside loop
	successes, failures := currentTotalCost.Split(sc.AccessResult.IsSuccess)
	maxLen := bt.MaxOutcomeLen
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
// Insert adds a new key to the B-Tree.
func (bt *BTreeIndex) Insert() *Outcomes[sc.AccessResult] {
	// 1. Find the leaf node (cost includes reads + node searches)
	findCost := bt.Find()

	// 2. Modify leaf node (CPU)
	modifyLeafCpuCost := sc.Map(&bt.RecordProcessingTime, func(p Duration) sc.AccessResult {
		return sc.AccessResult{true, p}
	})

	// 3. Write the modified leaf node back to disk
	writeLeafCost := bt.Disk.Write()

	// 4. Cost of potential split propagation (Avg extra R/W pairs)
	// Create cost of one Read+Write pair
	readWritePairCost := sc.And(bt.Disk.Read(), bt.Disk.Write(), sc.AndAccessResults)
	// Scale the latency by the average number of extra pairs needed
	avgPropCostOutcomes := sc.Map(readWritePairCost, func(ar sc.AccessResult) sc.AccessResult {
		if ar.Success {
			ar.Latency *= bt.AvgSplitPropCost
		} else {
			// If the R/W pair failed, the propagation cost is effectively zero latency, but retains failure status?
			// Or assume propagation failure means overall insert failure? Let's assume failure propagates.
			ar.Latency = 0 // Failure happens, latency contribution is negligible
		}
		return ar
	})

	// Combine costs: Find -> Modify CPU -> Write Leaf -> Avg Propagation Cost
	step1 := findCost
	step2 := sc.And(step1, modifyLeafCpuCost, sc.AndAccessResults)
	step3 := sc.And(step2, writeLeafCost, sc.AndAccessResults)
	finalCost := sc.And(step3, avgPropCostOutcomes, sc.AndAccessResults)

	// Apply Reduction
	successes, failures := finalCost.Split(sc.AccessResult.IsSuccess)
	maxLen := bt.MaxOutcomeLen
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
// Delete removes a key from the B-Tree.
func (bt *BTreeIndex) Delete() *Outcomes[sc.AccessResult] {
	// Structure is very similar to Insert, just uses AvgMergePropCost

	// 1. Find the leaf node
	findCost := bt.Find()

	// 2. Modify leaf node (CPU)
	modifyLeafCpuCost := sc.Map(&bt.RecordProcessingTime, func(p Duration) sc.AccessResult {
		return sc.AccessResult{true, p}
	})

	// 3. Write the modified leaf node back to disk
	writeLeafCost := bt.Disk.Write()

	// 4. Cost of potential merge/redistribution propagation (Avg extra R/W pairs)
	readWritePairCost := sc.And(bt.Disk.Read(), bt.Disk.Write(), sc.AndAccessResults)
	avgPropCostOutcomes := sc.Map(readWritePairCost, func(ar sc.AccessResult) sc.AccessResult {
		if ar.Success {
			ar.Latency *= bt.AvgMergePropCost // Use Merge cost factor
		} else {
			ar.Latency = 0 // Failure happens, latency contribution is negligible
		}
		return ar
	})

	// Combine costs: Find -> Modify CPU -> Write Leaf -> Avg Propagation Cost
	step1 := findCost
	step2 := sc.And(step1, modifyLeafCpuCost, sc.AndAccessResults)
	step3 := sc.And(step2, writeLeafCost, sc.AndAccessResults)
	finalCost := sc.And(step3, avgPropCostOutcomes, sc.AndAccessResults)

	// Apply Reduction
	successes, failures := finalCost.Split(sc.AccessResult.IsSuccess)
	maxLen := bt.MaxOutcomeLen
	if maxLen <= 0 {
		maxLen = 5
	}
	trimmer := sc.TrimToSize(100, maxLen)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalOutcomes := (&Outcomes[sc.AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalOutcomes
}

// --- Deprecated Scan Method ---
// Scan is less relevant for BTree primary key index; use Find or range scans (future).
// If needed, it would iterate through leaf pages sequentially.
/*
func (h *BTreeIndex) Scan() (out *Outcomes[sc.AccessResult]) { ... }
*/

// NumLeafPages - estimation might be useful internally but less so for external user.
/*
func (bt *BTreeIndex) NumLeafPages() uint {
	recordsPerPage := bt.RecordsPerPage()
	if recordsPerPage == 0 { return 1 } // Avoid division by zero
    numRequired := float64(bt.NumRecords) / float64(recordsPerPage)
	return uint(math.Ceil(numRequired))
}
*/
