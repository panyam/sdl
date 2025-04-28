package sdl

import (
	"sort"
)

// InvertedIndex represents an inverted index structure for text search
type InvertedIndex struct {
	Index
	
	// Average number of terms per document
	AvgTermsPerDocument float64
	
	// Number of unique terms in the index
	NumUniqueTerms uint
	
	// Average posting list length (number of documents per term)
	AvgPostingListLength float64
	
	// Maximum size of outcomes
	MaxOutcomeLen int
}

func (ii *InvertedIndex) Init() *InvertedIndex {
	ii.Index.Init()
	ii.AvgTermsPerDocument = 100  // Default: average 100 terms per document
	ii.NumUniqueTerms = 100000    // Default: 100K unique terms in corpus
	ii.AvgPostingListLength = float64(ii.NumRecords) / float64(ii.NumUniqueTerms)
	ii.MaxOutcomeLen = 5
	return ii
}

// AvgPostingListSizeBytes calculates the average size of a posting list in bytes
func (ii *InvertedIndex) AvgPostingListSizeBytes() uint64 {
	// Each posting contains docID (4 bytes) and position info (4 bytes)
	return uint64(ii.AvgPostingListLength * 8)
}

// Insert adds a new document to the inverted index
func (ii *InvertedIndex) Insert() (out *Outcomes[AccessResult]) {
	// For inserting a document into an inverted index:
	// 1. Parse the document into terms
	// 2. For each term, read its posting list
	// 3. Update the posting list
	// 4. Write back the posting list
	
	// Add processing time for parsing the document into terms
	parsingTime := NewOutcomes[Duration]().Add(1.0, Micros(100*ii.AvgTermsPerDocument)) 
	
	// Initial success outcome
	initialOutcome := NewOutcomes[AccessResult]().Add(1.0, AccessResult{true, 0})
	
	// Add parsing time
	parseOutcomes := And(initialOutcome, parsingTime, 
		func(this AccessResult, that Duration) AccessResult {
			return AccessResult{this.Success, this.Latency + that}
		})
	
	// Now for each term, simulate read/update/write of posting list
	// For simplicity, we'll model a subset of terms (10% of unique terms in doc)
	termsToProcess := int(ii.AvgTermsPerDocument * 0.1)
	
	insertOutcomes := parseOutcomes
	for i := 0; i < termsToProcess; i++ {
		// Read the posting list
		insertOutcomes = And(insertOutcomes, ii.Disk.Read(), AndAccessResults)
		
		// Add processing time proportional to posting list size
		processingFactor := float64(ii.AvgPostingListSizeBytes()) / 1024.0
		insertOutcomes = And(insertOutcomes, &ii.RecordProcessingTime, 
			func(this AccessResult, that Duration) AccessResult {
				return AccessResult{this.Success, this.Latency + (that * processingFactor)}
			})
		
		// Write the updated posting list
		insertOutcomes = And(insertOutcomes, ii.Disk.Write(), AndAccessResults)
		
		// Reduce outcome space
		if insertOutcomes.Len() > ii.MaxOutcomeLen {
			sort.Slice(insertOutcomes.Buckets, func(i, j int) bool {
				return insertOutcomes.Buckets[i].Value.Latency < insertOutcomes.Buckets[j].Value.Latency
			})
			insertOutcomes = MergeAdjacentAccessResults(insertOutcomes, 0.8)
			insertOutcomes = ReduceAccessResults(insertOutcomes, ii.MaxOutcomeLen)
		}
	}
	
	return insertOutcomes
}

// TermQuery searches for documents containing a specific term
func (ii *InvertedIndex) TermQuery(term string) (out *Outcomes[AccessResult]) {
	// For a term query in an inverted index:
	// 1. Look up the term in the dictionary
	// 2. Read the posting list
	// 3. Process the posting list
	
	// Dictionary lookup
	dictionaryLookup := ii.Disk.Read()
	successes, failures := dictionaryLookup.Split(func(value AccessResult) bool {
		return value.Success
	})
	
	// Add processing time for dictionary lookup
	lookupOutcomes := And(successes, &ii.RecordProcessingTime, 
		func(this AccessResult, that Duration) AccessResult {
			return AccessResult{this.Success, this.Latency + that}
		})
	
	// Read the posting list
	postingListRead := And(lookupOutcomes, ii.Disk.Read(), AndAccessResults)
	
	// Add processing time proportional to posting list size
	processingFactor := float64(ii.AvgPostingListSizeBytes()) / 1024.0
	termQueryOutcomes := And(postingListRead, &ii.RecordProcessingTime, 
		func(this AccessResult, that Duration) AccessResult {
			return AccessResult{this.Success, this.Latency + (that * processingFactor)}
		})
	
	// Combine with failures
	return termQueryOutcomes.Append(failures)
}

// PhraseQuery searches for documents containing a specific phrase
func (ii *InvertedIndex) PhraseQuery(terms []string) (out *Outcomes[AccessResult]) {
	// For a phrase query:
	// 1. Do term queries for each term
	// 2. Intersect the posting lists
	// 3. Check positional information
	
	if len(terms) == 0 {
		return NewOutcomes[AccessResult]().Add(1.0, AccessResult{true, 0})
	}
	
	// Start with the first term
	phraseOutcomes := ii.TermQuery(terms[0])
	
	// For each additional term
	for i := 1; i < len(terms); i++ {
		// Do a term query
		nextTermOutcomes := ii.TermQuery(terms[i])
		
		// Combine results (intersection)
		phraseOutcomes = And(phraseOutcomes, nextTermOutcomes, AndAccessResults)
		
		// Add processing time for intersecting and position checking
		processingFactor := float64(ii.AvgPostingListSizeBytes()) / 1024.0
		phraseOutcomes = And(phraseOutcomes, &ii.RecordProcessingTime, 
			func(this AccessResult, that Duration) AccessResult {
				return AccessResult{this.Success, this.Latency + (that * processingFactor)}
			})
		
		// Reduce outcome space
		if phraseOutcomes.Len() > ii.MaxOutcomeLen {
			sort.Slice(phraseOutcomes.Buckets, func(i, j int) bool {
				return phraseOutcomes.Buckets[i].Value.Latency < phraseOutcomes.Buckets[j].Value.Latency
			})
			phraseOutcomes = MergeAdjacentAccessResults(phraseOutcomes, 0.8)
			phraseOutcomes = ReduceAccessResults(phraseOutcomes, ii.MaxOutcomeLen)
		}
	}
	
	return phraseOutcomes
}

// BooleanQuery executes a boolean query (AND, OR, NOT) over multiple terms
func (ii *InvertedIndex) BooleanQuery(queryTerms []string, isAnd bool) (out *Outcomes[AccessResult]) {
	// For a boolean query:
	// 1. Do term queries for each term
	// 2. Combine the posting lists according to boolean operation
	
	if len(queryTerms) == 0 {
		return NewOutcomes[AccessResult]().Add(1.0, AccessResult{true, 0})
	}
	
	// Start with the first term
	booleanOutcomes := ii.TermQuery(queryTerms[0])
	
	// For each additional term
	for i := 1; i < len(queryTerms); i++ {
		// Do a term query
		nextTermOutcomes := ii.TermQuery(queryTerms[i])
		
		// Combine results based on operation type
		// For AND, both must succeed; for OR, either can succeed
		if isAnd {
			booleanOutcomes = And(booleanOutcomes, nextTermOutcomes, AndAccessResults)
		} else {
			// OR operation - custom merge that preserves success if either succeeds
			booleanOutcomes = And(booleanOutcomes, nextTermOutcomes, 
				func(a, b AccessResult) AccessResult {
					return AccessResult{
						Success: a.Success || b.Success,
						Latency: a.Latency + b.Latency,
					}
				})
		}
		
		// Add processing time for combining posting lists
		processingFactor := float64(ii.AvgPostingListSizeBytes()) / 1024.0
		booleanOutcomes = And(booleanOutcomes, &ii.RecordProcessingTime, 
			func(this AccessResult, that Duration) AccessResult {
				return AccessResult{this.Success, this.Latency + (that * processingFactor)}
			})
		
		// Reduce outcome space
		if booleanOutcomes.Len() > ii.MaxOutcomeLen {
			sort.Slice(booleanOutcomes.Buckets, func(i, j int) bool {
				return booleanOutcomes.Buckets[i].Value.Latency < booleanOutcomes.Buckets[j].Value.Latency
			})
			booleanOutcomes = MergeAdjacentAccessResults(booleanOutcomes, 0.8)
			booleanOutcomes = ReduceAccessResults(booleanOutcomes, ii.MaxOutcomeLen)
		}
	}
	
	return booleanOutcomes
}

// Delete removes a document from the inverted index
func (ii *InvertedIndex) Delete(docId string) (out *Outcomes[AccessResult]) {
	// For deleting a document from an inverted index:
	// 1. For each term in the document, read its posting list
	// 2. Update the posting list to remove the document
	// 3. Write back the posting list
	
	// This is similar to Insert but with different processing
	// For simplicity, we'll model the same I/O pattern as Insert
	return ii.Insert()
}

// Scan is not directly applicable to inverted indices
// Instead we'll model scanning all terms in the dictionary
func (ii *InvertedIndex) Scan() (out *Outcomes[AccessResult]) {
	// Need to read all terms in the dictionary
	d1 := ii.Disk.Read()
	successes, failures := d1.Split(func(value AccessResult) bool {
		return value.Success
	})
	
	// Simulate reading all terms and their posting lists
	// For simplicity, we'll read a subset of terms
	termsToRead := int(ii.NumUniqueTerms / 100) // Read 1% of terms
	if termsToRead < 1 {
		termsToRead = 1
	}
	
	scanOutcomes := successes
	for i := 0; i < termsToRead; i++ {
		// Read each term's posting list
		scanOutcomes = And(scanOutcomes, ii.Disk.Read(), AndAccessResults)
		
		// Add processing time proportional to posting list size
		processingFactor := float64(ii.AvgPostingListSizeBytes()) / 1024.0
		scanOutcomes = And(scanOutcomes, &ii.RecordProcessingTime, 
			func(this AccessResult, that Duration) AccessResult {
				return AccessResult{this.Success, this.Latency + (that * processingFactor)}
			})
		
		// Reduce outcome space
		if scanOutcomes.Len() > ii.MaxOutcomeLen {
			sort.Slice(scanOutcomes.Buckets, func(i, j int) bool {
				return scanOutcomes.Buckets[i].Value.Latency < scanOutcomes.Buckets[j].Value.Latency
			})
			scanOutcomes = MergeAdjacentAccessResults(scanOutcomes, 0.8)
			scanOutcomes = ReduceAccessResults(scanOutcomes, ii.MaxOutcomeLen)
		}
	}
	
	// Combine with failures
	return scanOutcomes.Append(failures)
}