package sdl

import (
	"log"
	"sort"
)

type HeapFile struct {
	Index
}

func (h *HeapFile) RecordsPerPage() uint64 {
	return h.PageSize / uint64(h.RecordSize)
}

func (h *HeapFile) NumPages() uint {
	return uint(1 + uint64(h.NumRecords*h.RecordSize)/h.PageSize)
}

func (h *HeapFile) Init() *HeapFile {
	// Say size of a page is 1MB
	h.Index.Init()
	return h
}

// Now for the methods
func (h *HeapFile) Insert() (out *Outcomes[AccessResult]) {
	d1 := h.Disk.Read()
	successes, failures := d1.Split(func(value AccessResult) bool {
		return value.Success
	})

	// Apply the processing delay to
	// Option 2 - And is a helper method taking 2 outcomes and merging them
	return And(successes, &h.RecordProcessingTime,
		func(this AccessResult, that Duration) (out AccessResult) {
			return AccessResult{this.Success, this.Latency + that}
		}).Then(AndAccessResults, h.Disk.Write()).Append(failures)
}

func (h *HeapFile) Reducer(lenTrigger, maxLen int) (out func(*Outcomes[AccessResult]) *Outcomes[AccessResult]) {
	return func(group *Outcomes[AccessResult]) *Outcomes[AccessResult] {
		if group.Len() > h.MaxOutcomeLen {
			sort.Slice(group.Buckets, func(i, j int) bool {
				return group.Buckets[i].Value.Latency < group.Buckets[j].Value.Latency
			})
			group = MergeAdjacentAccessResults(group, 0.8)
			group = ReduceAccessResults(group, h.MaxOutcomeLen)
		}
		return group
	}
}

// Find/Searches for a record by equality
// Similar to a scan but the allows the possibility that there is a 1 / NumPages
// probability that an entry would be within a page
func (h *HeapFile) Find() (out *Outcomes[AccessResult]) {
	// We can do this in a couple of ways -
	if true {
		d1 := h.Disk.Read()
		rpp := float64(h.RecordsPerPage())
		d2 := And(h.Disk.Read(), &h.RecordProcessingTime, func(this AccessResult, that Duration) AccessResult {
			return AccessResult{this.Success, this.Latency + (that * rpp)}
		})

		log.Println("NumPages: ", h.NumPages())
		for range h.NumPages() / 2 { // on average we will need to read half the number of pages
			// Do another disk read (+ record processing on the page) if successful and expand the
			// outcome frontier
			// Reduce to fixed sizes - this is more of an optimization and should
			// be implicit in the DSL
			d1 = d1.If(AccessResult.IsSuccess, d2, nil).
				GroupBy(
					AccessResult.IsSuccess, h.Reducer(50, h.MaxOutcomeLen),
					h.Reducer(10, 3),
				)
			// Reduce Failures
		}

		// Now Add failres to the failure outcomes
		return d1
	} else {
		d1 := h.Disk.Read()
		successes, failures := d1.Split(func(value AccessResult) bool {
			return value.Success
		})

		rpp := float64(h.RecordsPerPage())
		log.Println("NumPages: ", h.NumPages())
		for range h.NumPages() / 2 { // on average we will need to read half the number of pages
			// Do another disk read and expand our outcome space
			s2, f2 := And(successes, d1, AndAccessResults).
				Split(func(value AccessResult) bool { return value.Success })

			// this is the equivalent of:
			// if read().Success {
			//		Delay(processRecords(RecordsPerPage)
			// } else {
			//   	break
			// }
			successes = And(s2, &h.RecordProcessingTime, func(this AccessResult, that Duration) (out AccessResult) {
				return AccessResult{this.Success, this.Latency + (that * rpp)}
			})
			failures = failures.Append(f2)
			// log.Println("Num S, F: ", successes.Len(), failures.Len())

			// Now merge them
			if successes.Len() > h.MaxOutcomeLen {
				sort.Slice(successes.Buckets, func(i, j int) bool {
					return successes.Buckets[i].Value.Latency < successes.Buckets[j].Value.Latency
				})
				successes = MergeAdjacentAccessResults(successes, 0.8)
				successes = ReduceAccessResults(successes, h.MaxOutcomeLen)
			}
			if failures.Len() > 3 {
				sort.Slice(failures.Buckets, func(i, j int) bool {
					return failures.Buckets[i].Value.Latency < failures.Buckets[j].Value.Latency
				})
				failures = MergeAdjacentAccessResults(failures, 0.8)
				failures = ReduceAccessResults(failures, 3)
			}
		}

		// Now Add failres to the failure outcomes
		return successes.Append(failures)
	}
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
	successes, failures := d1.Split(func(value AccessResult) bool {
		return value.Success
	})

	rpp := float64(h.RecordsPerPage())
	for range h.NumPages() {
		// Do another read and append
		// TODO - Right now we are not taking into account the 1/NumPages chance that we can stop
		s2, f2 := And(successes, d1, AndAccessResults).Split(func(value AccessResult) bool {
			return value.Success
		})

		successes = And(s2, &h.RecordProcessingTime, func(this AccessResult, that Duration) (out AccessResult) {
			return AccessResult{this.Success, this.Latency + (that * rpp)}
		})
		failures.Append(f2)
	}

	// Now Add failres to the failure outcomes
	successes.Append(failures)
	return successes
}
