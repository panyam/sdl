package components

import (
	"math"

	sc "github.com/panyam/sdl/core"
)

type SortedFile struct {
	Index

	// Max size of our outcomes
	MaxOutcomeLen int
}

func (h *SortedFile) RecordsPerPage() uint64 {
	return h.PageSize / uint64(h.RecordSize)
}

func (h *SortedFile) NumPages() uint {
	return uint(1 + uint64(h.NumRecords*h.RecordSize)/h.PageSize)
}

func NewSortedFile() *SortedFile {
	s := &SortedFile{}
	s.Init()
	return s
}

func (s *SortedFile) Init() {
	s.Index.Init()
	// Say size of a page is 1MB
	s.MaxOutcomeLen = 5
}

// Visits every page - for a scanning through all entries
func (h *SortedFile) Scan() (out *Outcomes[sc.AccessResult]) {
	// Get the disk read's outcomes and we can reuse them each time
	out = h.Disk.Read()

	// t := time.Now() defer func() { log.Println("Scan Took: ", time.Now().Sub(t)) }()

	// Read all pages
	for range h.NumPages() {
		// Do another read and append
		// TODO - Right now we are not taking into account the 1/NumPages chance that we can stop
		out = out.If(sc.AccessResult.IsSuccess,
			sc.And(h.Disk.Read(), &h.RecordProcessingTime, sc.AccessResult.AddLatency), nil,
			sc.AndAccessResults)

		// Trim to prevent exponential growth
		if out.Len() > h.MaxOutcomeLen {
			out = sc.TrimToSize(100, h.MaxOutcomeLen)(out)
		}
		// log.Println("Scan Iter: ", i, out.Len())
	}
	return out
}

// Assuming the sorting is on the key being searched for
// This means we can do a bisection to find the item
//
//	Find() {
//		for i := range Log2(NumPages()) {
//			if Read()  == "Failure" {
//				return
//			}
//			rec = SearchPage()		// needs RecordProcessingTime * Log2(RecordsPerPage)
//			if rec.Found {
//				return
//			}
//		}
//	}
func (h *SortedFile) Find() (out *Outcomes[sc.AccessResult]) {
	// t := time.Now() defer func() { log.Println("Find Took: ", time.Now().Sub(t)) }()

	out = h.Disk.Read()
	pagesLeft := h.NumPages()
	log2R := math.Log2(float64(h.RecordsPerPage()))
	for pagesLeft > 0 {
		pagesLeft /= 2

		// Read the page
		successes, failures := out.If(sc.AccessResult.IsSuccess,
			sc.And(h.Disk.Read(), &h.RecordProcessingTime, func(a sc.AccessResult, latency Duration) sc.AccessResult {
				return sc.AccessResult{a.Success, a.Latency + (log2R * latency)}
			}),
			nil,
			sc.AndAccessResults).Split(sc.AccessResult.IsSuccess)

		// Trim if need be
		successes = sc.TrimToSize(100, h.MaxOutcomeLen)(successes)
		failures = sc.TrimToSize(100, h.MaxOutcomeLen)(failures)
		out = successes.Append(failures)
	}
	return
}

// Insert an entry into a sorted file.
func (h *SortedFile) Insert() (out *Outcomes[sc.AccessResult]) {
	// Even though "finding" the right place can happen logarithmically
	// after a find, all records after it have to be shifted
	return h.Find()
}

func (h *SortedFile) Delete() (out *Outcomes[sc.AccessResult]) {
	// Same as Insert due to shifting records
	return h.Insert()
}

// SetArrivalRate sets the arrival rate for a specific method.
// This forwards the rate to the underlying disk for contention modeling.
func (s *SortedFile) SetArrivalRate(method string, rate float64) error {
	return s.Disk.SetArrivalRate(method, rate)
}

// GetArrivalRate returns the arrival rate for a specific method.
func (s *SortedFile) GetArrivalRate(method string) float64 {
	return s.Disk.GetArrivalRate(method)
}

// GetTotalArrivalRate returns the sum of all method arrival rates.
func (s *SortedFile) GetTotalArrivalRate() float64 {
	return s.Disk.GetTotalArrivalRate()
}
