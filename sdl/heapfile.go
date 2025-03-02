package sdl

type HeapFile struct {
	Index
}

func (h *HeapFile) RecordsPerPage() uint64 {
	return h.PageSize / uint64(h.RecordSize)
}

func (h *HeapFile) Init() *HeapFile {
	// Say size of a page is 1MB
	h.Index.Init()
	return h
}

// Visits every page - for a scanning through all entries
func (h *HeapFile) Scan() (out *Outcomes[AccessResult]) {
	// Get the disk read's outcomes and we can reuse them each time
	out = h.Disk.Read()

	// Read all pages
	for range h.NumPages() {
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

// Now for the methods
func (h *HeapFile) Insert() (out *Outcomes[AccessResult]) {
	// Just the effect of a write to the last page + cost of processing a record
	return h.Disk.Read().If(
		AccessResult.IsSuccess,
		Map(&h.RecordProcessingTime, func(val Duration) AccessResult { return AccessResult{true, val} }),
		nil,
		AndAccessResults)
}

// Find/Searches for a record by equality
// Similar to a scan but the allows the possibility that there is a 1 / NumPages
// probability that an entry would be within a page
func (h *HeapFile) Find() (out *Outcomes[AccessResult]) {
	// We can do this in a couple of ways -
	out = h.Disk.Read()

	// Read half the pages
	// A more fun way is to actually put in an out come which has "Found/NotFound/Error" and then make Found something
	// like 1 / NumPages and compare that that is almost as this appropximation
	for range h.NumPages() / 2 {
		// Do another read and append
		// TODO - Right now we are not taking into account the 1/NumPages chance that we can stop
		successes, failures := out.If(AccessResult.IsSuccess,
			And(h.Disk.Read(), &h.RecordProcessingTime, AccessResult.AddLatency), nil,
			AndAccessResults).Split(AccessResult.IsSuccess)
		successes = TrimToSize(20, h.MaxOutcomeLen)(successes)
		failures = TrimToSize(20, 5)(failures)
		out = successes.Append(failures)
	}
	return out
}

func (h *HeapFile) Delete() (out *Outcomes[AccessResult]) {
	// Has the same complexity as a Find
	return h.Find()
}
