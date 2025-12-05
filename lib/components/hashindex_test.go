package components

import (
	"fmt"
	"testing"

	sc "github.com/panyam/sdl/lib/core"
)

func TestHashIndex_Probabilities_Heuristic(t *testing.T) {
	hi := NewHashIndex()
	hi.PageSize = 4096 // Needed for NumPages -> resize cost calculation
	hi.RecordSize = 100

	// Low number of records (below minResizeRecords default)
	hi.NumRecords = 5000
	pc1 := hi.collisionProbability()
	pr1 := hi.resizeProbability()
	t.Logf("Probs ( 5k records): Collision=%.4f, Resize=%.4f", pc1, pr1)
	// Expect low collision, potentially zero or near-zero resize

	// Medium number of records (above minResizeRecords)
	hi.NumRecords = 100000
	pc2 := hi.collisionProbability()
	pr2 := hi.resizeProbability()
	t.Logf("Probs (100k records): Collision=%.4f, Resize=%.4f", pc2, pr2)
	if !(pc2 > pc1) {
		t.Errorf("Collision probability %.4f should increase vs %.4f", pc2, pc1)
	}
	if !(pr2 > pr1) {
		t.Errorf("Resize probability %.4f should increase vs %.4f", pr2, pr1)
	}
	if pr2 <= 0.001 {
		t.Errorf("Resize probability %.4f should be noticeably > base", pr2)
	}

	// High number of records
	hi.NumRecords = 1000000
	pc3 := hi.collisionProbability()
	pr3 := hi.resizeProbability()
	t.Logf("Probs (1M records): Collision=%.4f, Resize=%.4f", pc3, pr3)
	if !(pc3 > pc2) {
		t.Errorf("Collision probability %.4f should increase further vs %.4f", pc3, pc2)
	}
	if !(pr3 > pr2) {
		t.Errorf("Resize probability %.4f should increase further vs %.4f", pr3, pr2)
	}

	// Very high number of records
	hi.NumRecords = 10000000
	pc4 := hi.collisionProbability()
	pr4 := hi.resizeProbability()
	t.Logf("Probs (10M records): Collision=%.4f, Resize=%.4f", pc4, pr4)
	if !(pc4 > pc3) {
		t.Errorf("Collision probability %.4f should increase further vs %.4f", pc4, pc3)
	}
	if !(pr4 > pr3) {
		t.Errorf("Resize probability %.4f should increase further vs %.4f", pr4, pr3)
	}
	// Check if it's nearing the cap
	maxResizeProb := 0.10 // From heuristic function
	if !approxEqualTest(pr4, maxResizeProb, maxResizeProb*0.1) && pr4 > maxResizeProb*0.8 {
		t.Logf("Resize probability %.4f approaching cap %.4f", pr4, maxResizeProb)
	} else if !approxEqualTest(pr4, maxResizeProb, maxResizeProb*0.1) {
		t.Errorf("Resize probability %.4f hasn't increased enough by 10M records (cap %.4f)", pr4, maxResizeProb)
	}
}

func TestHashIndex_Find_Vs_Modify_Metrics(t *testing.T) {
	hi := NewHashIndex()
	hi.Disk.Init() // Use default SSD
	hi.MaxOutcomeLen = 15
	// Test with different record counts to see impact
	recordCounts := []uint{200000, 900000} // Moderate load, High load (expecting resize)

	for _, rCount := range recordCounts {
		hi.NumRecords = rCount
		t.Logf("\n--- Testing HashIndex with %d records ---", rCount)
		t.Logf("CollisionProb=%.4f, ResizeProb=%.4f", hi.collisionProbability(), hi.resizeProbability())

		// Define some basic expectations for Analyze (can be refined)
		// Note: We won't use Assert() here, just Assert() for comparison
		findExpectations := []sc.Expectation{
			sc.ExpectAvailability(sc.GTE, 0.99), // Expect high availability
			sc.ExpectP99(sc.LT, 1.0),            // Expect P99 < 1 second (loose check)
		}
		insertExpectations := []sc.Expectation{
			sc.ExpectAvailability(sc.GTE, 0.99),
			sc.ExpectP99(sc.LT, 1.5), // Allow higher latency for insert
		}
		deleteExpectations := []sc.Expectation{
			sc.ExpectAvailability(sc.GTE, 0.99),
			sc.ExpectP99(sc.LT, 1.5), // Similar to insert
		}

		// --- Test Find ---
		findOutcomes := hi.Find()
		if findOutcomes == nil || findOutcomes.Len() == 0 {
			t.Fatalf("[%d recs] Find() returned nil or empty outcomes", rCount)
		}
		// Manual calculation + logging (keep for comparison)
		findAvail := sc.Availability(findOutcomes)
		findMean := sc.MeanLatency(findOutcomes)
		findP99 := sc.PercentileLatency(findOutcomes, 0.99)
		t.Logf("[%d recs] Manual Log - Hash Find: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", rCount, findAvail, findMean, findP99, findOutcomes.Len())
		// Use Analyze and Assert
		findAnalysisName := fmt.Sprintf("Hash Find (%d recs)", rCount)
		findAnalysisResult := sc.Analyze(findAnalysisName, func() *sc.Outcomes[sc.AccessResult] { return findOutcomes }, findExpectations...)
		findAnalysisResult.Assert(t) // Log results from Analyze

		// --- Test Insert ---
		insertOutcomes := hi.Insert()
		if insertOutcomes == nil || insertOutcomes.Len() == 0 {
			t.Fatalf("[%d recs] Insert() returned nil or empty outcomes", rCount)
		}
		// Manual calculation + logging
		insertAvail := sc.Availability(insertOutcomes)
		insertMean := sc.MeanLatency(insertOutcomes)
		insertP99 := sc.PercentileLatency(insertOutcomes, 0.99)
		t.Logf("[%d recs] Manual Log - Hash Insert: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", rCount, insertAvail, insertMean, insertP99, insertOutcomes.Len())
		// Use Analyze and Assert
		insertAnalysisName := fmt.Sprintf("Hash Insert (%d recs)", rCount)
		insertAnalysisResult := sc.Analyze(insertAnalysisName, func() *sc.Outcomes[sc.AccessResult] { return insertOutcomes }, insertExpectations...)
		insertAnalysisResult.Assert(t) // Log results from Analyze

		// --- Test Delete ---
		deleteOutcomes := hi.Delete()
		if deleteOutcomes == nil || deleteOutcomes.Len() == 0 {
			t.Fatalf("[%d recs] Delete() returned nil or empty outcomes", rCount)
		}
		// Manual calculation + logging
		deleteAvail := sc.Availability(deleteOutcomes)
		deleteMean := sc.MeanLatency(deleteOutcomes)
		deleteP99 := sc.PercentileLatency(deleteOutcomes, 0.99)
		t.Logf("[%d recs] Manual Log - Hash Delete: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", rCount, deleteAvail, deleteMean, deleteP99, deleteOutcomes.Len())
		// Use Analyze and Assert
		deleteAnalysisName := fmt.Sprintf("Hash Delete (%d recs)", rCount)
		deleteAnalysisResult := sc.Analyze(deleteAnalysisName, func() *sc.Outcomes[sc.AccessResult] { return deleteOutcomes }, deleteExpectations...)
		deleteAnalysisResult.Assert(t) // Log results from Analyze

		// --- Keep Manual Compare Metrics ---
		if findMean >= deleteMean {
			t.Errorf("[%d recs] Manual Check - Hash Find Mean Latency (%.6fs) should typically be less than Delete Mean (%.6fs)", rCount, findMean, deleteMean)
		}
		// Expect Insert to be slower than Delete, especially when resize is probable
		if deleteMean >= insertMean && hi.resizeProbability() < 0.01 {
			t.Logf("[%d recs] Manual Check Warning: Delete mean (%.6fs) unexpectedly higher than/equal to Insert mean (%.6fs) when resize prob is low", rCount, deleteMean, insertMean)
		} else if deleteMean < insertMean {
			t.Logf("[%d recs] Manual Check Info: Insert mean (%.6fs) slower than delete mean (%.6fs) as expected", rCount, insertMean, deleteMean)
		}
		t.Logf("--- Finished %d records ---", rCount)
	}
}
