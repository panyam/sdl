package sdl

import (
	"testing"
	// Ensure metrics helpers are accessible
)

func TestLSMTree_Init(t *testing.T) {
	lsm := NewLSMTree()
	if lsm.Disk.ReadOutcomes == nil { // Check inherited Disk init
		t.Fatal("LSMTree Disk not initialized")
	}
	if lsm.Levels <= 0 {
		t.Error("Default Levels should be > 0")
	}
	if lsm.MemtableHitProb < 0 || lsm.MemtableHitProb > 1.0 {
		t.Error("Invalid MemtableHitProb")
	}
	// Add more checks for other default parameters
}

func TestLSMTree_Write_Read_Metrics(t *testing.T) {
	lsm := NewLSMTree()

	// Configure disk to be SSD for faster base operations
	lsm.Disk.Init(ProfileSSD)
	lsm.MaxOutcomeLen = 15 // Allow more outcomes for testing

	// --- Test Write ---
	writeOutcomes := lsm.Write()
	if writeOutcomes == nil || writeOutcomes.Len() == 0 {
		t.Fatal("Write() returned nil or empty outcomes")
	}

	writeAvail := Availability(writeOutcomes)
	writeMean := MeanLatency(writeOutcomes)
	writeP99 := PercentileLatency(writeOutcomes, 0.99)

	t.Logf("LSM Write: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", writeAvail, writeMean, writeP99, writeOutcomes.Len())

	// Plausibility checks for Write (relative to base SSD write)
	baseWriteMean := MeanLatency(lsm.Disk.Write())
	if writeMean < baseWriteMean {
		t.Errorf("LSM Write mean latency (%.6fs) should not be less than base disk write mean (%.6fs)", writeMean, baseWriteMean)
	}
	// Expected availability should be slightly lower due to compaction etc if modelled perfectly,
	// but current model preserves base success status mostly. Check it's high.
	if writeAvail < 0.99 {
		t.Errorf("LSM Write availability (%.4f) seems too low", writeAvail)
	}

	// --- Test Read ---
	readOutcomes := lsm.Read()
	if readOutcomes == nil || readOutcomes.Len() == 0 {
		t.Fatal("Read() returned nil or empty outcomes")
	}

	readAvail := Availability(readOutcomes)
	readMean := MeanLatency(readOutcomes)
	readP50 := PercentileLatency(readOutcomes, 0.50)
	readP99 := PercentileLatency(readOutcomes, 0.99)

	t.Logf("LSM Read : Avail=%.4f, Mean=%.6fs, P50=%.6fs, P99=%.6fs (Buckets: %d)", readAvail, readMean, readP50, readP99, readOutcomes.Len())

	// Plausibility checks for Read (relative to base SSD read)
	baseReadMean := MeanLatency(lsm.Disk.Read())
	// LSM read involves multiple steps, should be slower than single disk read on average
	if readMean < baseReadMean && lsm.MemtableHitProb < 0.99 { // Only expect faster if memtable hit rate is extremely high
		t.Errorf("LSM Read mean latency (%.6fs) is unexpectedly lower than base disk read mean (%.6fs)", readMean, baseReadMean)
	}
	if readAvail < 0.99 {
		t.Errorf("LSM Read availability (%.4f) seems too low", readAvail)
	}
	// P50 vs P99: Expect some difference due to different read paths
	if approxEqualTest(readP50, readP99, 1e-6) && lsm.MemtableHitProb < 0.9 { // If P50=P99, distribution might be too narrow unless high hit rate
		t.Logf("Warning: LSM Read P50 and P99 are very close (P50=%.6f, P99=%.6f)", readP50, readP99)
	}
}
