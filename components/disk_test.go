// components/disk_test.go
package components

import (
	// Added
	"testing"

	sc "github.com/panyam/leetcoach/sdl/core"
)

// Tests for Init remain the same...
func TestDiskInit_DefaultSSD(t *testing.T) {
	// Test default initialization (empty profile name)
	d := NewDisk("") // Should default to SSD

	if d.ProfileName != ProfileSSD {
		t.Errorf("Expected ProfileName to be '%s', got '%s'", ProfileSSD, d.ProfileName)
	}
	if d.ReadOutcomes == nil || d.WriteOutcomes == nil {
		t.Fatal("Default disk outcomes are nil")
	}
	// Check if it points to the predefined SSD outcomes
	if d.ReadOutcomes != ssdReadOutcomes {
		t.Error("Default disk ReadOutcomes do not point to predefined SSD read outcomes")
	}
	if d.WriteOutcomes != ssdWriteOutcomes {
		t.Error("Default disk WriteOutcomes do not point to predefined SSD write outcomes")
	}

	// Test initialization with unrecognized profile name
	dUnrecognized := NewDisk("QuantumLeapDrive")
	if dUnrecognized.ProfileName != ProfileSSD {
		t.Errorf("Expected unrecognized ProfileName to default to '%s', got '%s'", ProfileSSD, dUnrecognized.ProfileName)
	}
	if dUnrecognized.ReadOutcomes != ssdReadOutcomes {
		t.Error("Unrecognized disk ReadOutcomes do not point to predefined SSD read outcomes")
	}
}

func TestDiskInit_ExplicitProfiles(t *testing.T) {
	// Test explicit SSD initialization
	ssd := NewDisk(ProfileSSD)
	if ssd.ProfileName != ProfileSSD {
		t.Errorf("Expected ProfileName '%s', got '%s'", ProfileSSD, ssd.ProfileName)
	}
	if ssd.ReadOutcomes != ssdReadOutcomes || ssd.WriteOutcomes != ssdWriteOutcomes {
		t.Error("SSD disk does not point to correct predefined outcomes")
	}

	// Test explicit HDD initialization
	hdd := NewDisk(ProfileHDD)
	if hdd.ProfileName != ProfileHDD {
		t.Errorf("Expected ProfileName '%s', got '%s'", ProfileHDD, hdd.ProfileName)
	}
	if hdd.ReadOutcomes != hddReadOutcomes || hdd.WriteOutcomes != hddWriteOutcomes {
		t.Error("HDD disk does not point to correct predefined outcomes")
	}
}

func TestDisk_PerformanceMetrics(t *testing.T) {
	ssd := NewDisk(ProfileSSD)
	hdd := NewDisk(ProfileHDD)

	// --- Analyze SSD Read ---
	ssdReadOutcomes := ssd.Read()
	// Manual
	ssdReadAvail := sc.Availability(ssdReadOutcomes)
	ssdReadMean := sc.MeanLatency(ssdReadOutcomes)
	ssdReadP99 := sc.PercentileLatency(ssdReadOutcomes, 0.99)
	t.Logf("Manual Log - SSD Read : Avail=%.4f, Mean=%.6fs, P99=%.6fs", ssdReadAvail, ssdReadMean, ssdReadP99)
	// Analyze
	ssdReadExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 0.998),
		sc.ExpectMeanLatency(sc.LT, sc.Millis(0.2)), // Expect very fast mean
		sc.ExpectP99(sc.EQ, sc.Millis(2.0)),         // SSD P99 is 2ms in definition
	}
	ssdReadAnalysis := sc.Analyze("SSD Read", func() *sc.Outcomes[sc.AccessResult] { return ssdReadOutcomes }, ssdReadExpectations...)
	ssdReadAnalysis.LogResults(t)

	// --- Analyze SSD Write ---
	ssdWriteOutcomes := ssd.Write()
	// Manual
	ssdWriteAvail := sc.Availability(ssdWriteOutcomes)
	ssdWriteMean := sc.MeanLatency(ssdWriteOutcomes)
	ssdWriteP99 := sc.PercentileLatency(ssdWriteOutcomes, 0.99)
	t.Logf("Manual Log - SSD Write: Avail=%.4f, Mean=%.6fs, P99=%.6fs", ssdWriteAvail, ssdWriteMean, ssdWriteP99)
	// Analyze
	ssdWriteExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 0.998),
		sc.ExpectMeanLatency(sc.LT, sc.Millis(0.3)),
		sc.ExpectP99(sc.EQ, sc.Millis(5.0)), // SSD P99 Write is 5ms
	}
	ssdWriteAnalysis := sc.Analyze("SSD Write", func() *sc.Outcomes[sc.AccessResult] { return ssdWriteOutcomes }, ssdWriteExpectations...)
	ssdWriteAnalysis.LogResults(t)

	// --- Analyze HDD Read ---
	hddReadOutcomes := hdd.Read()
	// Manual
	hddReadAvail := sc.Availability(hddReadOutcomes)
	hddReadMean := sc.MeanLatency(hddReadOutcomes)
	hddReadP99 := sc.PercentileLatency(hddReadOutcomes, 0.99)
	t.Logf("Manual Log - HDD Read : Avail=%.4f, Mean=%.6fs, P99=%.6fs", hddReadAvail, hddReadMean, hddReadP99)
	// Analyze
	hddReadExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 0.990),
		sc.ExpectMeanLatency(sc.GT, sc.Millis(5)),
		sc.ExpectP99(sc.EQ, sc.Millis(100.0)), // HDD P99 Read is 100ms
	}
	hddReadAnalysis := sc.Analyze("HDD Read", func() *sc.Outcomes[sc.AccessResult] { return hddReadOutcomes }, hddReadExpectations...)
	hddReadAnalysis.LogResults(t)

	// --- Analyze HDD Write ---
	hddWriteOutcomes := hdd.Write()
	// Manual
	hddWriteAvail := sc.Availability(hddWriteOutcomes)
	hddWriteMean := sc.MeanLatency(hddWriteOutcomes)
	hddWriteP99 := sc.PercentileLatency(hddWriteOutcomes, 0.99)
	t.Logf("Manual Log - HDD Write: Avail=%.4f, Mean=%.6fs, P99=%.6fs", hddWriteAvail, hddWriteMean, hddWriteP99)
	// Analyze
	hddWriteExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.EQ, 0.990),
		sc.ExpectMeanLatency(sc.GT, sc.Millis(8)),
		sc.ExpectP99(sc.EQ, sc.Millis(150.0)), // HDD P99 Write is 150ms
	}
	hddWriteAnalysis := sc.Analyze("HDD Write", func() *sc.Outcomes[sc.AccessResult] { return hddWriteOutcomes }, hddWriteExpectations...)
	hddWriteAnalysis.LogResults(t)

	// --- Keep Manual Assertions ---
	if ssdReadMean >= hddReadMean {
		t.Errorf("Manual Check - SSD Read Mean Latency (%.6fs) should be less than HDD (%.6fs)", ssdReadMean, hddReadMean)
	}
	if ssdWriteMean >= hddWriteMean {
		t.Errorf("Manual Check - SSD Write Mean Latency (%.6fs) should be less than HDD (%.6fs)", ssdWriteMean, hddWriteMean)
	}
	if ssdReadP99 >= hddReadP99 {
		t.Errorf("Manual Check - SSD Read P99 Latency (%.6fs) should be less than HDD (%.6fs)", ssdReadP99, hddReadP99)
	}
	// Optional: Check specific expected values
	// expectedSSDP99Read := sc.Millis(2.0) // P99 = 0.998 -> 2ms bucket
	// if !approxEqualTest(ssdReadP99, expectedSSDP99Read, 1e-9) {
	// 	t.Errorf("Manual Check - SSD Read P99 Latency %.6fs doesn't match expected %.6fs", ssdReadP99, expectedSSDP99Read)
	// }
	// expectedHDDP99Read := sc.Millis(100) // P99 = 0.99 -> 100ms bucket
	// if !approxEqualTest(hddReadP99, expectedHDDP99Read, 1e-9) {
	// 	t.Errorf("Manual Check - HDD Read P99 Latency %.6fs doesn't match expected %.6fs", hddReadP99, expectedHDDP99Read)
	// }
}

func TestDisk_ReadProcessWrite(t *testing.T) {
	ssd := NewDisk(ProfileSSD)
	processingTime := sc.Millis(1) // 1ms processing

	rpwOutcomes := ssd.ReadProcessWrite(processingTime)

	if rpwOutcomes == nil {
		t.Fatal("ReadProcessWrite returned nil")
	}

	// Manual calculation + logging
	rpwAvail := sc.Availability(rpwOutcomes)
	rpwMean := sc.MeanLatency(rpwOutcomes)
	rpwP99 := sc.PercentileLatency(rpwOutcomes, 0.99)
	// Expected values calculation (approximation)
	expectedAvail := sc.Availability(ssd.Read()) * sc.Availability(ssd.Write())
	expectedMean := sc.MeanLatency(ssd.Read()) + sc.MeanLatency(ssd.Write()) + processingTime
	t.Logf("Manual Log - RPW SSD: Avail=%.6f (Exp: %.6f), Mean=%.6fs (Exp: %.6fs), P99=%.6fs", rpwAvail, expectedAvail, rpwMean, expectedMean, rpwP99)

	// Analyze call
	rpwExpectations := []sc.Expectation{
		sc.ExpectAvailability(sc.GTE, expectedAvail*0.99), // Expect availability close to product
		sc.ExpectAvailability(sc.LTE, expectedAvail*1.01),
		sc.ExpectMeanLatency(sc.GTE, expectedMean*0.9), // Expect mean close to sum
		sc.ExpectMeanLatency(sc.LTE, expectedMean*1.1),
	}
	rpwAnalysis := sc.Analyze("SSD ReadProcessWrite", func() *sc.Outcomes[sc.AccessResult] { return rpwOutcomes }, rpwExpectations...)
	rpwAnalysis.LogResults(t)

	// Keep Manual Assertions
	if !approxEqualTest(rpwAvail, expectedAvail, 0.001) {
		t.Errorf("Manual Check - ReadProcessWrite Availability mismatch: got %.6f, expected around %.6f", rpwAvail, expectedAvail)
	}
	if !approxEqualTest(rpwMean, expectedMean, expectedMean*0.1) { // Allow 10% tolerance for mean approximation
		t.Errorf("Manual Check - ReadProcessWrite Mean Latency mismatch: got %.6fs, expected around %.6fs", rpwMean, expectedMean)
	}
}
