package sdl

import (
	"sort"
)

// SecondaryIndex represents a secondary index structure
type SecondaryIndex struct {
	Index
	
	// Primary index type this secondary index is built upon
	// Could be: "btree", "hash", etc.
	UnderlyingIndexType string
	
	// Selectivity of the index (fraction of records that match a typical query)
	Selectivity float64
	
	// Whether the index is dense (entry for each record) or sparse (entry for some records)
	IsDense bool
	
	// Maximum size of outcomes
	MaxOutcomeLen int
	
	// Reference to primary storage
	PrimaryStorage *HeapFile
}

func (si *SecondaryIndex) Init() *SecondaryIndex {
	si.Index.Init()
	si.UnderlyingIndexType = "btree"  // Default: B-tree based secondary index
	si.Selectivity = 0.01             // Default: 1% selectivity
	si.IsDense = true                 // Default: dense index
	si.MaxOutcomeLen = 5
	si.PrimaryStorage = new(HeapFile).Init()
	return si
}

// Insert adds a new entry to the secondary index
func (si *SecondaryIndex) Insert() (out *Outcomes[AccessResult]) {
	// For inserting into a secondary index:
	// 1. Locate the position in the index structure
	// 2. Insert the index entry
	
	// Secondary index insert performance depends on underlying index type
	var indexInsert *Outcomes[AccessResult]
	
	switch si.UnderlyingIndexType {
	case "btree":
		// Use B-tree insert pattern
		bt := new(BTreeIndex)
		bt.Index = si.Index
		bt.Height = 3
		bt.NodeCapacity = 100
		bt.MaxOutcomeLen = si.MaxOutcomeLen
		indexInsert = bt.Insert()
	case "hash":
		// Use Hash insert pattern
		hi := new(HashIndex)
		hi.Index = si.Index
		hi.NumBuckets = 1000
		hi.MaxOutcomeLen = si.MaxOutcomeLen
		indexInsert = hi.Insert()
	default:
		// Default simple insert
		d1 := si.Disk.Read()
		successes, failures := d1.Split(func(value AccessResult) bool {
			return value.Success
		})
		
		writeOutcomes := And(successes, si.Disk.Write(), AndAccessResults)
		indexInsert = writeOutcomes.Append(failures)
	}
	
	return indexInsert
}

// Find searches for entries in the secondary index and retrieves from primary storage
func (si *SecondaryIndex) Find() (out *Outcomes[AccessResult]) {
	// For a find operation in a secondary index:
	// 1. Search the secondary index to locate matching entries
	// 2. For each match, access the primary storage
	
	// Secondary index search depends on underlying index type
	var indexSearch *Outcomes[AccessResult]
	
	switch si.UnderlyingIndexType {
	case "btree":
		// Use B-tree search pattern
		bt := new(BTreeIndex)
		bt.Index = si.Index
		bt.Height = 3
		bt.NodeCapacity = 100
		bt.MaxOutcomeLen = si.MaxOutcomeLen
		indexSearch = bt.Find()
	case "hash":
		// Use Hash search pattern
		hi := new(HashIndex)
		hi.Index = si.Index
		hi.NumBuckets = 1000
		hi.MaxOutcomeLen = si.MaxOutcomeLen
		indexSearch = hi.Find()
	default:
		// Default simple search
		indexSearch = si.Disk.Read()
	}
	
	// Split into successful and failed searches
	successes, failures := indexSearch.Split(func(value AccessResult) bool {
		return value.Success
	})
	
	// For successful index lookups, access the primary storage
	// Estimate number of records to fetch based on selectivity
	numRecordsToFetch := int(float64(si.NumRecords) * si.Selectivity)
	if numRecordsToFetch < 1 {
		numRecordsToFetch = 1
	}
	
	// Limit to a reasonable number for simulation
	if numRecordsToFetch > 100 {
		numRecordsToFetch = 100
	}
	
	// Fetch records from primary storage
	primaryFetches := successes
	for i := 0; i < numRecordsToFetch; i++ {
		// Each matching record requires a primary storage access
		primaryFetches = And(primaryFetches, si.PrimaryStorage.Disk.Read(), AndAccessResults)
		
		// Add processing time for each record
		primaryFetches = And(primaryFetches, &si.RecordProcessingTime, 
			func(this AccessResult, that Duration) AccessResult {
				return AccessResult{this.Success, this.Latency + that}
			})
		
		// Reduce outcome space
		if primaryFetches.Len() > si.MaxOutcomeLen {
			sort.Slice(primaryFetches.Buckets, func(i, j int) bool {
				return primaryFetches.Buckets[i].Value.Latency < primaryFetches.Buckets[j].Value.Latency
			})
			primaryFetches = MergeAdjacentAccessResults(primaryFetches, 0.8)
			primaryFetches = ReduceAccessResults(primaryFetches, si.MaxOutcomeLen)
		}
	}
	
	// Combine with failures
	return primaryFetches.Append(failures)
}

// RangeQuery executes a range query using the secondary index
func (si *SecondaryIndex) RangeQuery(rangeSelectivity float64) (out *Outcomes[AccessResult]) {
	// For a range query:
	// 1. Search the secondary index for the range
	// 2. Access primary storage for each match
	
	// Only B-tree supports efficient range queries
	var indexRangeSearch *Outcomes[AccessResult]
	
	if si.UnderlyingIndexType == "btree" {
		// Use B-tree range query
		bt := new(BTreeIndex)
		bt.Index = si.Index
		bt.Height = 3
		bt.NodeCapacity = 100
		bt.MaxOutcomeLen = si.MaxOutcomeLen
		indexRangeSearch = bt.RangeQuery(rangeSelectivity)
	} else {
		// Other index types need a full scan with filtering
		indexRangeSearch = si.Scan()
	}
	
	// Split into successful and failed searches
	successes, failures := indexRangeSearch.Split(func(value AccessResult) bool {
		return value.Success
	})
	
	// For successful index lookups, access the primary storage
	// Estimate number of records to fetch based on range selectivity
	numRecordsToFetch := int(float64(si.NumRecords) * rangeSelectivity)
	if numRecordsToFetch < 1 {
		numRecordsToFetch = 1
	}
	
	// Limit to a reasonable number for simulation
	if numRecordsToFetch > 100 {
		numRecordsToFetch = 100
	}
	
	// Fetch records from primary storage
	primaryFetches := successes
	for i := 0; i < numRecordsToFetch; i++ {
		// Each matching record requires a primary storage access
		primaryFetches = And(primaryFetches, si.PrimaryStorage.Disk.Read(), AndAccessResults)
		
		// Add processing time for each record
		primaryFetches = And(primaryFetches, &si.RecordProcessingTime, 
			func(this AccessResult, that Duration) AccessResult {
				return AccessResult{this.Success, this.Latency + that}
			})
		
		// Reduce outcome space
		if primaryFetches.Len() > si.MaxOutcomeLen {
			sort.Slice(primaryFetches.Buckets, func(i, j int) bool {
				return primaryFetches.Buckets[i].Value.Latency < primaryFetches.Buckets[j].Value.Latency
			})
			primaryFetches = MergeAdjacentAccessResults(primaryFetches, 0.8)
			primaryFetches = ReduceAccessResults(primaryFetches, si.MaxOutcomeLen)
		}
	}
	
	// Combine with failures
	return primaryFetches.Append(failures)
}

// Delete removes entries from the secondary index
func (si *SecondaryIndex) Delete() (out *Outcomes[AccessResult]) {
	// For delete, first find the entry, then update the index
	
	// Secondary index delete depends on underlying index type
	var indexDelete *Outcomes[AccessResult]
	
	switch si.UnderlyingIndexType {
	case "btree":
		// Use B-tree delete pattern
		bt := new(BTreeIndex)
		bt.Index = si.Index
		bt.Height = 3
		bt.NodeCapacity = 100
		bt.MaxOutcomeLen = si.MaxOutcomeLen
		indexDelete = bt.Delete()
	case "hash":
		// Use Hash delete pattern
		hi := new(HashIndex)
		hi.Index = si.Index
		hi.NumBuckets = 1000
		hi.MaxOutcomeLen = si.MaxOutcomeLen
		indexDelete = hi.Delete()
	default:
		// Default simple delete: find + write
		findResult := si.Disk.Read()
		successes, failures := findResult.Split(func(value AccessResult) bool {
			return value.Success
		})
		
		writeOutcomes := And(successes, si.Disk.Write(), AndAccessResults)
		indexDelete = writeOutcomes.Append(failures)
	}
	
	return indexDelete
}

// Scan visits all entries in the secondary index
func (si *SecondaryIndex) Scan() (out *Outcomes[AccessResult]) {
	// For a scan operation in a secondary index:
	// 1. Scan the entire index structure
	// 2. For dense indices, this is equivalent to scanning the whole table
	
	// Secondary index scan depends on underlying index type
	var indexScan *Outcomes[AccessResult]
	
	switch si.UnderlyingIndexType {
	case "btree":
		// Use B-tree scan pattern
		bt := new(BTreeIndex)
		bt.Index = si.Index
		bt.Height = 3
		bt.NodeCapacity = 100
		bt.MaxOutcomeLen = si.MaxOutcomeLen
		indexScan = bt.RangeQuery(1.0)  // Full range query
	case "hash":
		// Use Hash scan pattern (inefficient)
		hi := new(HashIndex)
		hi.Index = si.Index
		hi.NumBuckets = 1000
		hi.MaxOutcomeLen = si.MaxOutcomeLen
		indexScan = hi.Scan()
	default:
		// Default to primary storage scan
		indexScan = si.PrimaryStorage.Scan()
	}
	
	return indexScan
}