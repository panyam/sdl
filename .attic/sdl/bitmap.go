package sdl

/**
 * Bitmap Index Concepts:
 *
 * * Use Case: Best suited for columns with low cardinality (few distinct values, e.g., Gender, Country, Status). Not good for high-cardinality columns (like UserID, timestamp).
 * * Structure: Typically uses one bitmap per distinct value in the indexed column. Each bitmap has one bit per row in the table. If the bit is set (1), the corresponding row has that value; otherwise, it's 0.
 * * Querying:
 * 	- Equality (WHERE Status = 'Active'): Very fast. Retrieve the single bitmap for 'Active'. Iterate through set bits to find matching row IDs.
 * 	- Range (WHERE Age BETWEEN 20 AND 30): Can be supported (e.g., bit-sliced indexes or hierarchical bitmaps), often involving bitwise operations (OR) on multiple bitmaps.
 * 	- Multiple Conditions (WHERE Status = 'Active' AND Country = 'USA'): Very fast. Perform bitwise AND between the 'Active' bitmap and the 'USA' bitmap. The resulting bitmap has bits set only for rows matching both conditions. OR and NOT operations are also efficient.
 * * Updates/Inserts/Deletes: Can be expensive. Changing a value for a row requires clearing the bit in the old value's bitmap and setting the bit in the new value's bitmap. Inserts require extending all bitmaps. Deletes might involve clearing bits or using a separate "existence bitmap". Updates are often handled as delete + insert. This cost depends heavily on whether the bitmaps are compressed.
 * * Compression: Bitmap indexes are often compressed (e.g., Run-Length Encoding like WAH - Word Aligned Hybrid) to reduce storage space, especially for sparse bitmaps. Compression affects query and update performance.
 *
 * Simplified Bitmap Model for SDL:
 *
 * Focus on the core performance trade-offs: fast queries (especially multi-condition AND/OR), potentially slower updates.
 */

import (
	"math/rand"
	"time"
	// "math"
)

// BitmapIndex represents a bitmap index structure.
type BitmapIndex struct {
	Index // Embed base index properties

	// Bitmap Specific Parameters
	Cardinality      uint    // Number of distinct values in the indexed column
	IsCompressed     bool    // Whether compression (e.g., RLE/WAH) is assumed
	UpdateCostFactor float64 // Multiplier for update/insert/delete overhead vs base R/W

	// Assumed average selectivity for queries impacting result processing cost
	QuerySelectivity float64

	// Internal state
	rng *rand.Rand
}

// Init initializes the BitmapIndex with default or provided parameters.
func (bmi *BitmapIndex) Init() *BitmapIndex {
	bmi.Index.Init() // Initialize base Index

	// --- Set Default Bitmap Parameters ---
	bmi.Cardinality = 100       // Default: Moderate cardinality (e.g., product category)
	bmi.IsCompressed = true     // Default: Assume modern compression is used
	bmi.UpdateCostFactor = 3.0  // Default: Updates cost 3x base R/W (read/modify/write, higher if compressed)
	bmi.QuerySelectivity = 0.01 // Default: Queries return ~1% of rows matching bitmap

	// Adjust UpdateCostFactor based on compression
	if !bmi.IsCompressed {
		// Uncompressed updates might involve writing more data if cardinality is high
		// but decoding/encoding might be faster. Simplified model: lower factor.
		bmi.UpdateCostFactor = 2.0
	}

	bmi.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Ensure RecordProcessingTime is initialized
	if bmi.Index.RecordProcessingTime.Len() == 0 {
		bmi.Index.RecordProcessingTime.Add(100, Nanos(100))
	}

	return bmi
}

// NewBitmapIndex creates and initializes a new BitmapIndex component.
func NewBitmapIndex() *BitmapIndex {
	bmi := &BitmapIndex{}
	return bmi.Init()
}

func (bmi *BitmapIndex) Find() *Outcomes[AccessResult] {

	// 1. Cost to load initial index structure / relevant bitmap(s)
	// Assume this takes roughly one disk read regardless of query complexity (due to bitwise ops)
	// unless cardinality is extremely high AND uncompressed (maybe refine later).
	loadIndexCost := bmi.Disk.Read()

	// 2. Cost of CPU for bitwise operations (AND, OR, NOT)
	// Model as a small, fixed CPU cost, slightly higher if compressed (decode cost).
	// TODO - May be these can be parameters
	cpuCostPerOp := Nanos(50) // Base cost for bitwise ops
	if bmi.IsCompressed {
		cpuCostPerOp = Nanos(150) // Higher CPU if compressed
	}
	// Assume average query involves a few bitwise ops (e.g., 3)
	numOps := 3
	bitwiseOpDurationOutcomes := (&Outcomes[Duration]{}).Add(100, cpuCostPerOp*float64(numOps))
	bitwiseOpCost := Map(bitwiseOpDurationOutcomes, func(d Duration) AccessResult {
		return AccessResult{true, d}
	})

	// 3. Cost to process resulting bitmap based on selectivity
	// Number of records = Total NumRecords * Selectivity
	numResultRecords := uint64(float64(bmi.NumRecords) * bmi.QuerySelectivity)
	if numResultRecords == 0 && bmi.QuerySelectivity > 0 && bmi.NumRecords > 0 {
		numResultRecords = 1 // Process at least one record if selectivity > 0
	}
	// Need to map from RecordProcessingTime (Outcomes[Duration]) to total time
	resultProcessingDurationOutcomes := Map(&bmi.RecordProcessingTime, func(p Duration) Duration {
		return float64(numResultRecords) * p
	})
	// Ensure And func is set if needed later (though not strictly needed for Map here)
	// resultProcessingDurationOutcomes.And = func(a,b Duration) Duration { return a+b }
	// Now Map the total duration Outcomes to AccessResult Outcomes
	resultProcessingCost := Map(resultProcessingDurationOutcomes, func(totalProcTime Duration) AccessResult {
		return AccessResult{true, totalProcTime}
	})

	// 4. Combine costs sequentially: Load -> Bitwise Ops -> Result Processing
	combinedCost := And(loadIndexCost, bitwiseOpCost, AndAccessResults)
	combinedCost = And(combinedCost, resultProcessingCost, AndAccessResults)

	// 5. Apply Reduction
	successes, failures := combinedCost.Split(AccessResult.IsSuccess)
	maxLen := bmi.MaxOutcomeLen
	if maxLen <= 0 {
		maxLen = 5
	} // Default if not set
	trimmer := TrimToSize(100, maxLen)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalOutcomes := (&Outcomes[AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalOutcomes
}

func (bmi *BitmapIndex) modifyBitmapCost() *Outcomes[AccessResult] {
	// 1. Cost to read relevant bitmap(s) - assume 1 read typically
	readCost := bmi.Disk.Read()

	// 2. Cost to modify in memory (CPU) - related to RecordProcessingTime
	// --- FIX: Create pointer for temporary outcomes ---
	modifyCpuDurationOutcomes := Map(&bmi.RecordProcessingTime, func(p Duration) Duration {
		return p * 3.0 // Assume modification involves a few steps
	})
	// modifyCpuDurationOutcomes.And = func(a,b Duration) Duration { return a+b } // If needed
	modifyCpuCost := Map(modifyCpuDurationOutcomes, func(d Duration) AccessResult {
		return AccessResult{true, d}
	})

	// 3. Cost to write back modified bitmap(s) - assume 1 write typically
	writeCost := bmi.Disk.Write()

	// 4. Combine: Read -> Modify -> Write
	baseModificationCost := And(readCost, modifyCpuCost, AndAccessResults)
	baseModificationCost = And(baseModificationCost, writeCost, AndAccessResults)

	// 5. Apply UpdateCostFactor to latency (simplification of increased I/O or CPU for compression etc.)
	amplifiedCost := Map(baseModificationCost, func(ar AccessResult) AccessResult {
		if ar.Success {
			ar.Latency *= bmi.UpdateCostFactor
		}
		return ar
	})

	// 6. Reduction
	successes, failures := amplifiedCost.Split(AccessResult.IsSuccess)
	maxLen := bmi.MaxOutcomeLen
	if maxLen <= 0 {
		maxLen = 5
	}
	trimmer := TrimToSize(100, maxLen)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalOutcomes := (&Outcomes[AccessResult]{}).Append(trimmedSuccesses, trimmedFailures)

	return finalOutcomes
}

// Insert simulates adding a new row/value. High cost due to bitmap modifications.
func (bmi *BitmapIndex) Insert() *Outcomes[AccessResult] {
	// For simplicity, assume Insert cost is similar to a generic modification.
	// Real cost depends on whether it's a new value (new bitmap) or existing (set bit).
	return bmi.modifyBitmapCost()
}

// Delete simulates removing a row/value. High cost due to bitmap modifications.
func (bmi *BitmapIndex) Delete() *Outcomes[AccessResult] {
	// Assume Delete cost is similar to a generic modification (clearing bits/tombstones).
	return bmi.modifyBitmapCost()
}

// Update simulates changing a value for a row. Often highest cost (clear old bit, set new bit).
func (bmi *BitmapIndex) Update() *Outcomes[AccessResult] {
	// Assume Update cost is similar to a generic modification, potentially slightly higher implicitly
	// due to the UpdateCostFactor being applied to the R-M-W sequence.
	return bmi.modifyBitmapCost()
}
