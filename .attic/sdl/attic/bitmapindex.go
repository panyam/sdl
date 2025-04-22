package sdl

import (
	"sort"
)

// BitmapIndex represents a bitmap index structure for low-cardinality attributes
type BitmapIndex struct {
	Index
	
	// Number of distinct values for the indexed attribute
	Cardinality uint
	
	// Size of each bitmap in bytes
	BitmapSize uint64
	
	// Maximum size of outcomes
	MaxOutcomeLen int
}

func (bi *BitmapIndex) Init() *BitmapIndex {
	bi.Index.Init()
	bi.Cardinality = 100  // Default cardinality (e.g., 100 distinct values)
	// Each bitmap uses 1 bit per record
	bi.BitmapSize = uint64(bi.NumRecords+7) / 8  // Round up to nearest byte
	bi.MaxOutcomeLen = 5
	return bi
}

// BitmapsPerPage calculates how many complete bitmaps can fit in a page
func (bi *BitmapIndex) BitmapsPerPage() uint {
	return uint(bi.PageSize / bi.BitmapSize)
}

// TotalPages calculates how many pages are needed to store all bitmaps
func (bi *BitmapIndex) TotalPages() uint {
	return uint((uint64(bi.Cardinality) + uint64(bi.BitmapsPerPage()) - 1) / uint64(bi.BitmapsPerPage()))
}

// Insert updates the bitmap when a new record is added
func (bi *BitmapIndex) Insert() (out *Outcomes[AccessResult]) {
	// For an insert, we need to update the bitmap for the specific value
	// 1. Read the bitmap page
	// 2. Update the bitmap
	// 3. Write the bitmap back
	
	// Read the bitmap page
	d1 := bi.Disk.Read()
	successes, failures := d1.Split(func(value AccessResult) bool {
		return value.Success
	})
	
	// Add processing time for determining which bit to set
	insertOutcomes := And(successes, &bi.RecordProcessingTime, 
		func(this AccessResult, that Duration) AccessResult {
			return AccessResult{this.Success, this.Latency + that}
		})
	
	// Write the updated bitmap
	insertOutcomes = And(insertOutcomes, bi.Disk.Write(), AndAccessResults)
	
	// Combine with failures
	return insertOutcomes.Append(failures)
}

// Find queries the bitmap for records matching a specific value
func (bi *BitmapIndex) Find() (out *Outcomes[AccessResult]) {
	// For a find operation in a bitmap index:
	// 1. Read the bitmap for the specific value
	// 2. Process the bitmap to find matching records
	
	// Read the bitmap
	d1 := bi.Disk.Read()
	successes, failures := d1.Split(func(value AccessResult) bool {
		return value.Success
	})
	
	// Add processing time for scanning the bitmap
	// For a bitmap, processing time is proportional to the size of the bitmap
	bitmapProcessingFactor := float64(bi.BitmapSize) / 1024.0  // Normalize by 1KB
	
	findOutcomes := And(successes, &bi.RecordProcessingTime, 
		func(this AccessResult, that Duration) AccessResult {
			return AccessResult{this.Success, this.Latency + (that * bitmapProcessingFactor)}
		})
	
	// Combine with failures
	return findOutcomes.Append(failures)
}

// RangeQuery performs a range scan using bitmap operations
func (bi *BitmapIndex) RangeQuery(selectivity float64) (out *Outcomes[AccessResult]) {
	// For a range query in a bitmap index:
	// 1. Read multiple bitmaps (one for each value in the range)
	// 2. Perform OR operations on the bitmaps
	// 3. Process the resulting bitmap
	
	// Calculate number of bitmaps to read based on selectivity
	numBitmapsToRead := int(float64(bi.Cardinality) * selectivity)
	if numBitmapsToRead < 1 {
		numBitmapsToRead = 1
	}
	
	// Initial read of first bitmap
	d1 := bi.Disk.Read()
	successes, failures := d1.Split(func(value AccessResult) bool {
		return value.Success
	})
	
	// Read additional bitmaps
	rangeOutcomes := successes
	for i := 1; i < numBitmapsToRead; i++ {
		// Read each bitmap
		rangeOutcomes = And(rangeOutcomes, bi.Disk.Read(), AndAccessResults)
		
		// Reduce outcome space
		if rangeOutcomes.Len() > bi.MaxOutcomeLen {
			sort.Slice(rangeOutcomes.Buckets, func(i, j int) bool {
				return rangeOutcomes.Buckets[i].Value.Latency < rangeOutcomes.Buckets[j].Value.Latency
			})
			rangeOutcomes = MergeAdjacentAccessResults(rangeOutcomes, 0.8)
			rangeOutcomes = ReduceAccessResults(rangeOutcomes, bi.MaxOutcomeLen)
		}
	}
	
	// Add processing time for bitmap operations (OR) and scanning
	// Processing time increases with number of bitmaps and bitmap size
	bitmapProcessingFactor := float64(bi.BitmapSize) / 1024.0 * float64(numBitmapsToRead)
	
	rangeOutcomes = And(rangeOutcomes, &bi.RecordProcessingTime, 
		func(this AccessResult, that Duration) AccessResult {
			return AccessResult{this.Success, this.Latency + (that * bitmapProcessingFactor)}
		})
	
	// Combine with failures
	return rangeOutcomes.Append(failures)
}

// Delete updates the bitmap when a record is deleted
func (bi *BitmapIndex) Delete() (out *Outcomes[AccessResult]) {
	// For delete, similar to insert but clearing a bit instead of setting it
	return bi.Insert()  // Same I/O pattern as insert
}

// Scan requires reading all bitmaps (inefficient for this purpose)
func (bi *BitmapIndex) Scan() (out *Outcomes[AccessResult]) {
	// Need to read all bitmaps
	d1 := bi.Disk.Read()
	successes, failures := d1.Split(func(value AccessResult) bool {
		return value.Success
	})
	
	// Simulate reading all bitmap pages
	scanOutcomes := successes
	totalPages := bi.TotalPages()
	
	for i := uint(0); i < totalPages; i++ {
		// Read each page
		scanOutcomes = And(scanOutcomes, bi.Disk.Read(), AndAccessResults)
		
		// Process all bitmaps on the page
		bitmapsPerPage := bi.BitmapsPerPage()
		bitmapProcessingFactor := float64(bi.BitmapSize) * float64(bitmapsPerPage) / 1024.0
		
		scanOutcomes = And(scanOutcomes, &bi.RecordProcessingTime, 
			func(this AccessResult, that Duration) AccessResult {
				return AccessResult{this.Success, this.Latency + (that * bitmapProcessingFactor)}
			})
		
		// Reduce outcome space
		if scanOutcomes.Len() > bi.MaxOutcomeLen {
			sort.Slice(scanOutcomes.Buckets, func(i, j int) bool {
				return scanOutcomes.Buckets[i].Value.Latency < scanOutcomes.Buckets[j].Value.Latency
			})
			scanOutcomes = MergeAdjacentAccessResults(scanOutcomes, 0.8)
			scanOutcomes = ReduceAccessResults(scanOutcomes, bi.MaxOutcomeLen)
		}
	}
	
	// Combine with failures
	return scanOutcomes.Append(failures)
}

// BitwiseAND performs a bitwise AND operation between two bitmap values
func (bi *BitmapIndex) BitwiseAND(bitmap1Value, bitmap2Value string) (out *Outcomes[AccessResult]) {
	// Simulate reading two bitmaps
	bitmap1Read := bi.Disk.Read()
	bitmap2Read := bi.Disk.Read()
	
	// Combine the reads (both must succeed)
	combinedReads := And(bitmap1Read, bitmap2Read, AndAccessResults)
	
	successes, failures := combinedReads.Split(func(value AccessResult) bool {
		return value.Success
	})
	
	// Add processing time for the AND operation
	bitmapProcessingFactor := float64(bi.BitmapSize) / 1024.0
	andOutcomes := And(successes, &bi.RecordProcessingTime, 
		func(this AccessResult, that Duration) AccessResult {
			return AccessResult{this.Success, this.Latency + (that * bitmapProcessingFactor)}
		})
	
	// Combine with failures
	return andOutcomes.Append(failures)
}

// BitwiseOR performs a bitwise OR operation between two bitmap values
func (bi *BitmapIndex) BitwiseOR(bitmap1Value, bitmap2Value string) (out *Outcomes[AccessResult]) {
	// Uses same pattern as BitwiseAND
	return bi.BitwiseAND(bitmap1Value, bitmap2Value)
}