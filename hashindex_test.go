package sdl

import (
	"testing"
	// Ensure metrics helpers are accessible
)

func TestHashIndex_Init(t *testing.T) {
	hi := NewHashIndex()
	if hi.Disk.ReadOutcomes == nil {
		t.Fatal("HashIndex Disk not initialized")
	}
	// Add more checks
}

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
	hi.Disk.Init()
	hi.MaxOutcomeLen = 15
	// Test with different record counts to see impact
	recordCounts := []uint{200000, 900000} // Moderate load, High load (expecting resize)

	for _, rCount := range recordCounts {
		hi.NumRecords = rCount
		t.Logf("--- Testing HashIndex with %d records ---", rCount)
		t.Logf("CollisionProb=%.4f, ResizeProb=%.4f", hi.collisionProbability(), hi.resizeProbability())

		// --- Test Find ---
		findOutcomes := hi.Find()
		if findOutcomes == nil || findOutcomes.Len() == 0 {
			t.Fatalf("[%d recs] Find() returned nil or empty outcomes", rCount)
		}
		findAvail := Availability(findOutcomes)
		findMean := MeanLatency(findOutcomes)
		findP99 := PercentileLatency(findOutcomes, 0.99)
		t.Logf("[%d recs] Hash Find: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", rCount, findAvail, findMean, findP99, findOutcomes.Len())

		// --- Test Insert ---
		insertOutcomes := hi.Insert()
		if insertOutcomes == nil || insertOutcomes.Len() == 0 {
			t.Fatalf("[%d recs] Insert() returned nil or empty outcomes", rCount)
		}
		insertAvail := Availability(insertOutcomes)
		insertMean := MeanLatency(insertOutcomes)
		insertP99 := PercentileLatency(insertOutcomes, 0.99)
		t.Logf("[%d recs] Hash Insert: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", rCount, insertAvail, insertMean, insertP99, insertOutcomes.Len())

		// --- Test Delete ---
		deleteOutcomes := hi.Delete()
		if deleteOutcomes == nil || deleteOutcomes.Len() == 0 {
			t.Fatalf("[%d recs] Delete() returned nil or empty outcomes", rCount)
		}
		deleteAvail := Availability(deleteOutcomes)
		deleteMean := MeanLatency(deleteOutcomes)
		deleteP99 := PercentileLatency(deleteOutcomes, 0.99)
		t.Logf("[%d recs] Hash Delete: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", rCount, deleteAvail, deleteMean, deleteP99, deleteOutcomes.Len())

		// --- Compare Metrics ---
		if findMean >= deleteMean {
			t.Errorf("[%d recs] Hash Find Mean Latency (%.6fs) should typically be less than Delete Mean (%.6fs)", rCount, findMean, deleteMean)
		}
		// Expect Insert to be slower than Delete, especially when resize is probable
		if deleteMean >= insertMean && hi.resizeProbability() < 0.01 {
			t.Logf("[%d recs] Warning: Delete mean (%.6fs) unexpectedly higher than/equal to Insert mean (%.6fs) when resize prob is low", rCount, deleteMean, insertMean)
		} else if deleteMean < insertMean {
			t.Logf("[%d recs] Info: Insert mean (%.6fs) slower than delete mean (%.6fs) as expected", rCount, insertMean, deleteMean)
		}
	}
}
