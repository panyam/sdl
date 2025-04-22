package sdl

import (
	"math"
)

// BTreeIndex represents a clustered B-Tree index structure
type BTreeIndex struct {
	Index

	// Occupancy (between 0 and 1) - usually leave room in leaf pages
	// so that inserts/deletes Do not need complete shifts
	Occupancy float64

	// Fanout factor at each node - determines tree height (eg Log NumPages base F)
	NodeFanout int
}

func (bt *BTreeIndex) Init() *BTreeIndex {
	bt.Index.Init()
	bt.NodeFanout = 5 // Default capacity
	bt.Occupancy = 0.67
	return bt
}

// RecordsPerPage returns how many records fit in a page
func (bt *BTreeIndex) RecordsPerPage() uint64 {
	return uint64(bt.Occupancy * float64(bt.PageSize/uint64(bt.RecordSize)))
}

func (bt *BTreeIndex) Height() int {
	// Fanout factor at each node - determines tree height (eg Log NumPages base F)
	return int(math.Log(float64(bt.NumPages())) / math.Log(float64(bt.NodeFanout)))
}

func (bt *BTreeIndex) NumLeafPages() uint {
	return uint(float64(bt.NumPages()) / bt.Occupancy)
}

// Visits every page - for a scanning through all entries
func (h *BTreeIndex) Scan() (out *Outcomes[AccessResult]) {
	// Get the disk read's outcomes and we can reuse them each time
	out = h.Disk.Read()

	// Read all leaf pages but only process available records
	for range h.NumLeafPages() {
		// Do another read and append
		// TODO - Right now we are not taking into account the 1/NumPages chance that we can stop
		successes, failures := out.If(AccessResult.IsSuccess,
			And(h.Disk.Read(), &h.RecordProcessingTime, func(a AccessResult, latency Duration) AccessResult {
				return AccessResult{a.Success, a.Latency + (Duration(h.RecordsPerPage()) * latency)}
			}),
			nil,
			AndAccessResults).Split(AccessResult.IsSuccess)

		// Trim if need be
		successes = TrimToSize(100, h.MaxOutcomeLen)(successes)
		failures = TrimToSize(100, h.MaxOutcomeLen)(failures)
		out = successes.Append(failures)
	}
	return out
}

// Find searches for a key in the B-Tree
func (bt *BTreeIndex) Find() (out *Outcomes[AccessResult]) {
	// For a find operation, we need log_m(N) disk reads in the worst case
	// where m is the node capacity and N is the number of keys

	// Need to
	pagesLeft := bt.NumLeafPages()
	log2R := math.Log(float64(bt.RecordsPerPage()))
	for pagesLeft > 0 {
		pagesLeft /= uint(bt.NodeFanout)

		// Read the page, and process records in the page
		out = out.If(AccessResult.IsSuccess,
			And(bt.Disk.Read(), &bt.RecordProcessingTime, func(a AccessResult, latency Duration) AccessResult {
				return AccessResult{a.Success, a.Latency + (log2R * latency)}
			}),
			nil,
			AndAccessResults)

		// Trim if need be
		successes, failures := out.Split(AccessResult.IsSuccess)
		successes = TrimToSize(100, bt.MaxOutcomeLen)(successes)
		failures = TrimToSize(100, bt.MaxOutcomeLen)(failures)

		out = successes.Append(failures)
	}

	// Combine with failures
	return out
}

// Insert adds a new key to the B-Tree
func (bt *BTreeIndex) Insert() (out *Outcomes[AccessResult]) {
	// For an insert, we need to:
	// 1. Traverse to the leaf node (log_m(N) disk reads)
	// 2. Insert the key
	// 3. Potentially split and propagate up (worst case: height additional writes)

	// 1. First write the record to the final page
	out = bt.Disk.Read()

	out = out.If(AccessResult.IsSuccess,
		And(bt.Disk.Read(), &bt.RecordProcessingTime, AccessResult.AddLatency),
		nil,
		AndAccessResults)

	// Then update index pages - same complexity as Find
	out = out.If(AccessResult.IsSuccess, bt.Find(), nil, AndAccessResults)

	// And write back
	out = out.If(AccessResult.IsSuccess, bt.Disk.Write(), nil, AndAccessResults)

	// Trim if need be
	/*
		successes, failures := out.Split(AccessResult.IsSuccess)
		successes = TrimToSize(100, bt.MaxOutcomeLen)(successes)
		failures = TrimToSize(100, bt.MaxOutcomeLen)(failures)
	*/

	// Combine with failures
	return out
}

// Delete removes a key from the B-Tree
func (bt *BTreeIndex) Delete() (out *Outcomes[AccessResult]) {
	// Similar to Insert
	// TODO - Not quite but will finetune later
	return bt.Insert()
}
