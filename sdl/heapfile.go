package sdl

type HeapFile struct {
	// How many entries are already in this heapfile
	// This would determine latencies on certain operations like scan etc
	NumEntries int

	// Size of each page that is loaded at at iem when doing a disk io
	PageSize int

	// How long does it take to process a record in an operation
	RecordProcessingTime Outcomes[TimeValue]

	// The disk on which the heap file exists
	Disk *Disk
}

type HeapFileAccessResult struct {
	Success bool
	Latency TimeValue
}

func (h *HeapFile) Init() *HeapFile {
	h.RecordProcessingTime.Add(100, Val(10, MilliSeconds))
	return h
}

func (h *HeapFile) NumPages() int {
	return max(1, h.NumEntries/h.PageSize)
}

// Now for the methods
func (h *HeapFile) Insert() (out Outcomes[HeapFileAccessResult]) {
	d1 := h.Disk.Read()
	successes, failures := d1.Partition(func(value DiskAccessResult) bool {
		return value.Success
	})

	// We have 2 sets of outcomes - success and failures
	// what we need to do is somehow do a cross product of the S and F lists into the "rest" of the lists
	/* Option 1 - If Then was a member of the input Outcomes
	out.To(func(d DiskAccessResult) HeapFileAccessResult {
		return HeapFileAccessResult{d.Success, d.Latency}
	})

	out = out.Then(&h.RecordProcessingTime, func(this HeapFileAccessResult, that any) (out HeapFileAccessResult) {
		return HeapFileAccessResult{this.Success, this.Latency.Add(that.(TimeValue))}
	})
	*/

	// Apply the processing delay to
	// Option 2 - Then is a helper method taking 2 outcomes and merging them
	out = Then(successes, h.RecordProcessingTime, func(this DiskAccessResult, that TimeValue) (out HeapFileAccessResult) {
		return HeapFileAccessResult{this.Success, this.Latency.Add(that)}
	})

	out = Then(out, h.Disk.Write(), func(this HeapFileAccessResult, that DiskAccessResult) HeapFileAccessResult {
		return HeapFileAccessResult{this.Success && that.Success, this.Latency.Add(that.Latency)}
	})

	// Now merge success and failues
	for _, wv := range failures.Values {
		out.Add(wv.Weight, HeapFileAccessResult{wv.Value.Success, wv.Value.Latency})
	}
	return out
}

// Find/Searches for a record by equality
// Similar to a scan but the allows the possibility that there is a 1 / NumPages
// probability that an entry would be within a page
func (h *HeapFile) Find() (out Outcomes[HeapFileAccessResult]) {

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
	successes, failures := Convert(d1, func(dar DiskAccessResult) HeapFileAccessResult {
		return HeapFileAccessResult{dar.Success, dar.Latency}
	}).Partition(func(value HeapFileAccessResult) bool {
		return value.Success
	})

	for range h.NumPages() {
		// Do another read and append
		// TODO - Right now we are not taking into account the 1/NumPages chance that we can stop
		s2, f2 := Then(successes, d1, func(this HeapFileAccessResult, that DiskAccessResult) (out HeapFileAccessResult) {
			return
		}).Partition(func(value HeapFileAccessResult) bool {
			return value.Success
		})

		successes = Then(s2, h.RecordProcessingTime, func(this HeapFileAccessResult, that TimeValue) (out HeapFileAccessResult) {
			return HeapFileAccessResult{this.Success, this.Latency.Add(that.TimesN(int64(h.NumEntries)))}
		})
		failures.Append(&f2)
	}

	// Now Add failres to the failure outcomes
	successes.Append(&failures)
	return successes
}

func (h *HeapFile) Delete() (out Outcomes[HeapFileAccessResult]) {
	// Has the same complexity as a Find
	return h.Find()
}

// Visits every page - for a scanning through all entries
func (h *HeapFile) Scan() (out Outcomes[HeapFileAccessResult]) {
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
	successes, failures := Convert(d1, func(dar DiskAccessResult) HeapFileAccessResult {
		return HeapFileAccessResult{dar.Success, dar.Latency}
	}).Partition(func(value HeapFileAccessResult) bool {
		return value.Success
	})

	for range h.NumPages() {
		// Do another read and append
		// TODO - Right now we are not taking into account the 1/NumPages chance that we can stop
		s2, f2 := Then(successes, d1, func(this HeapFileAccessResult, that DiskAccessResult) (out HeapFileAccessResult) {
			return
		}).Partition(func(value HeapFileAccessResult) bool {
			return value.Success
		})

		successes = Then(s2, h.RecordProcessingTime, func(this HeapFileAccessResult, that TimeValue) (out HeapFileAccessResult) {
			return HeapFileAccessResult{this.Success, this.Latency.Add(that.TimesN(int64(h.NumEntries)))}
		})
		failures.Append(&f2)
	}

	// Now Add failres to the failure outcomes
	successes.Append(&failures)
	return successes
}
