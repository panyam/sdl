package sdl

import "sort"

// An index on a disk
type Index struct {
	// How many entries are already in this heapfile
	// This would determine latencies on certain operations like scan etc
	NumRecords uint

	// Size of each record
	RecordSize uint

	// Size of each page (in bytes) that is loaded at at iem when doing a disk io
	PageSize uint64

	// How long does it take to process a record in an operation
	RecordProcessingTime Outcomes[Duration]

	// The disk on which the heap file exists
	Disk Disk

	// Max size of our outcomes
	MaxOutcomeLen int
}

func (i *Index) NumPages() uint {
	return uint(1 + uint64(i.NumRecords*i.RecordSize)/i.PageSize)
}

func (i *Index) Init() {
	i.Disk.Init()
	i.PageSize = 1024 * 1024
	// Number of entries = 1M in this heapfile as a default
	i.NumRecords = 1000000
	i.RecordSize = 1024 // each record is 1kb by default
	i.RecordProcessingTime.Add(100, Nanos(100))
	i.MaxOutcomeLen = 5
}

// func (i *Index) ReadNPages(numPages, numRecordsToProcess int) *Outcomes[V] { return nil }

func TrimToSize(lenTrigger, maxLen int) (out func(*Outcomes[AccessResult]) *Outcomes[AccessResult]) {
	return func(group *Outcomes[AccessResult]) *Outcomes[AccessResult] {
		if group.Len() > lenTrigger {
			// Sort by Latency ONLY - Doesn't group Success/Failure first!
			sort.Slice(group.Buckets, func(i, j int) bool {
				return group.Buckets[i].Value.Latency < group.Buckets[j].Value.Latency
			})
			group = MergeAdjacentAccessResults(group, 0.8)
			group = ReduceAccessResults(group, maxLen)
		}
		return group
	}
}
