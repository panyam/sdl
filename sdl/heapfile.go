package sdl

import "time"

type HeapFile struct {
	// How many entries are already in this heapfile
	// This would determine latencies on certain operations like scan etc
	NumEntries int

	// Size of each page that is loaded at at iem when doing a disk io
	PageSize int

	// Number of pages in the heapfile
	NumPages int

	// How long does it take to process a record in an operation
	RecordProcessingTime Outcomes[time.Duration]

	// The disk on which the heap file exists
	Disk Disk
}

func (h *HeapFile) Init() *HeapFile {
	h.PageSize = 1024 * 1024
	h.NumPages = 1
	h.RecordProcessingTime.Add(100, time.Millisecond*10)
	return h
}

// Now for the methods
func (h *HeapFile) Insert() (out *Outcomes[AccessResult]) {
	d1 := h.Disk.Read()
	successes, failures := d1.Partition(func(value AccessResult) bool {
		return value.Success
	})

	// Apply the processing delay to
	// Option 2 - Then is a helper method taking 2 outcomes and merging them
	out = Then(successes, &h.RecordProcessingTime, func(this AccessResult, that time.Duration) (out AccessResult) {
		return AccessResult{this.Success, this.Latency + that}
	})

	out = Then(out, h.Disk.Write(), func(this AccessResult, that AccessResult) AccessResult {
		return AccessResult{this.Success && that.Success, this.Latency + that.Latency}
	})

	// Now merge success and failues
	for _, wv := range failures.Buckets {
		out.Add(wv.Weight, AccessResult{wv.Value.Success, wv.Value.Latency})
	}
	return out
}

// Find/Searches for a record by equality
// Similar to a scan but the allows the possibility that there is a 1 / NumPages
// probability that an entry would be within a page
func (h *HeapFile) Find() (out *Outcomes[AccessResult]) {

	// We can do this in a couple of ways -
	// Option 1 - create a tree that is NPages deep - one for each call
	// to disk.Read()
	// However this could "Explode" our tree - even with a distribution with 3 buckets,
	// at NumPages = 10 (eg 1Mb page with 10Mb disk) we have a 10 level deep tree with about 60k nodes
	// Realistically at 1Mb pages - with a 1GB disk - we are talking baout 1000 pages, 3^1000 = crazy big!!!
	//
	// Option 2 - Calculate disk read prob ones - and then calculate after N repetition mathematically
	// TODO - Need to refresh math here
	//
	// Option 3 - Prune/Reduce after each (or few) levels into a few buckets

	// Get the disk read's outcomes and we can reuse them each time
	d1 := h.Disk.Read()
	successes, failures := Convert(d1, func(dar AccessResult) AccessResult {
		return AccessResult{dar.Success, dar.Latency}
	}).Partition(func(value AccessResult) bool {
		return value.Success
	})

	for range h.NumPages {
		// Do another read and append
		// TODO - Right now we are not taking into account the 1/NumPages chance that we can stop
		s2, f2 := Then(successes, d1, func(this AccessResult, that AccessResult) (out AccessResult) {
			return AccessResult{this.Success && that.Success, this.Latency + that.Latency}
		}).Partition(func(value AccessResult) bool {
			return value.Success
		})

		successes = Then(s2, &h.RecordProcessingTime, func(this AccessResult, that time.Duration) (out AccessResult) {
			return AccessResult{this.Success, this.Latency + (that * time.Duration(h.NumEntries))}
		})
		failures.Append(f2)
	}

	// Now Add failres to the failure outcomes
	successes.Append(failures)
	return successes
}

func (h *HeapFile) Delete() (out *Outcomes[AccessResult]) {
	// Has the same complexity as a Find
	return h.Find()
}

// Visits every page - for a scanning through all entries
func (h *HeapFile) Scan() (out *Outcomes[AccessResult]) {
	// We can do this in a couple of ways -
	// Option 1 - create a tree that is NPages deep - one for each call
	// to disk.Read()
	// However this could "Explode" our tree - even with a distribution with 3 buckets,
	// at NumPages = 10 (eg 1Mb page with 10Mb disk) we have a 10 level deep tree with about 60k nodes
	// Realistically at 1Mb pages - with a 1GB disk - we are talking baout 1000 pages, 3^1000 = crazy big!!!
	//
	// Option 2 - Calculate disk read prob ones - and then calculate after N repetition mathematically
	// TODO - Need to refresh math here
	//
	// Option 3 - Prune/Reduce after each (or few) levels into a few buckets

	// Get the disk read's outcomes and we can reuse them each time
	d1 := h.Disk.Read()
	successes, failures := Convert(d1, func(dar AccessResult) AccessResult {
		return AccessResult{dar.Success, dar.Latency}
	}).Partition(func(value AccessResult) bool {
		return value.Success
	})

	for range h.NumPages {
		// Do another read and append
		// TODO - Right now we are not taking into account the 1/NumPages chance that we can stop
		s2, f2 := Then(successes, d1, func(this AccessResult, that AccessResult) (out AccessResult) {
			return
		}).Partition(func(value AccessResult) bool {
			return value.Success
		})

		successes = Then(s2, &h.RecordProcessingTime, func(this AccessResult, that time.Duration) (out AccessResult) {
			return AccessResult{this.Success, this.Latency + (that * time.Duration(h.NumEntries))}
		})
		failures.Append(f2)
	}

	// Now Add failres to the failure outcomes
	successes.Append(failures)
	return successes
}
