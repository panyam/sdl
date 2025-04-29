// components/lsm_test.go
package components

import (
	// Added
	"testing"

	sc "github.com/panyam/leetcoach/sdl/core"
)

// Test Init remains the same...
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
}

func TestLSMTree_Write_Read_Metrics(t *testing.T) {
	lsm := NewLSMTree()
	lsm.Disk.Init() // Use default SSD
	lsm.MaxOutcomeLen = 15

	// --- Test Write ---
	writeOutcomes := lsm.Write()
	if writeOutcomes == nil || writeOutcomes.Len() == 0 {
		t.Fatal("Write() returned nil or empty outcomes")
	}
	// Manual
	writeAvail := sc.Availability(writeOutcomes)
	writeMean := sc.MeanLatency(writeOutcomes)
	writeP99 := sc.PercentileLatency(writeOutcomes, 0.99)
	t.Logf("Manual Log - LSM Write: Avail=%.4f, Mean=%.6fs, P99=%.6fs (Buckets: %d)", writeAvail, writeMean, writeP99, writeOutcomes.Len())
	// Analyze
	writeExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, 0.99),                            // Expect high avail
		sc.ExpectMeanLatency(sc.GTE, sc.MeanLatency(lsm.Disk.Write())), // Mean >= base disk write
	}
	writeAnalysis := sc.Analyze("LSM Write", func() *sc.Outcomes[sc.AccessResult] { return writeOutcomes }, writeExpectations...)
	writeAnalysis.LogResults(t)

	// Manual checks
	baseWriteMean := sc.MeanLatency(lsm.Disk.Write())
	if writeMean < baseWriteMean {
		t.Errorf("Manual Check - LSM Write mean latency (%.6fs) should not be less than base disk write mean (%.6fs)", writeMean, baseWriteMean)
	}
	if writeAvail < 0.99 {
		t.Errorf("Manual Check - LSM Write availability (%.4f) seems too low", writeAvail)
	}

	// --- Test Read ---
	readOutcomes := lsm.Read()
	if readOutcomes == nil || readOutcomes.Len() == 0 {
		t.Fatal("Read() returned nil or empty outcomes")
	}
	// Manual
	readAvail := sc.Availability(readOutcomes)
	readMean := sc.MeanLatency(readOutcomes)
	readP50 := sc.PercentileLatency(readOutcomes, 0.50)
	readP99 := sc.PercentileLatency(readOutcomes, 0.99)
	t.Logf("Manual Log - LSM Read : Avail=%.4f, Mean=%.6fs, P50=%.6fs, P99=%.6fs (Buckets: %d)", readAvail, readMean, readP50, readP99, readOutcomes.Len())
	// Analyze
	readExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, 0.99),
		sc.ExpectMeanLatency(sc.GTE, sc.MeanLatency(lsm.Disk.Read())*0.5), // Read can be faster than single disk read if high memtable hit
		sc.ExpectP99(sc.GT, sc.Millis(0)),                                 // P99 should be positive
	}
	readAnalysis := sc.Analyze("LSM Read", func() *sc.Outcomes[sc.AccessResult] { return readOutcomes }, readExpectations...)
	readAnalysis.LogResults(t)

	// Manual checks
	baseReadMean := sc.MeanLatency(lsm.Disk.Read())
	if readMean < baseReadMean && lsm.MemtableHitProb < 0.99 {
		t.Errorf("Manual Check - LSM Read mean latency (%.6fs) is unexpectedly lower than base disk read mean (%.6fs)", readMean, baseReadMean)
	}
	if readAvail < 0.99 {
		t.Errorf("Manual Check - LSM Read availability (%.4f) seems too low", readAvail)
	}
	if approxEqualTest(readP50, readP99, 1e-6) && lsm.MemtableHitProb < 0.9 {
		t.Logf("Manual Check Warning: LSM Read P50 and P99 are very close (P50=%.6f, P99=%.6f)", readP50, readP99)
	}
}
