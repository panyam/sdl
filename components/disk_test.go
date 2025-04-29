package components

import (
	"testing"

	sc "github.com/panyam/leetcoach/sdl/core"
	// Ensure metrics package/file is accessible
)

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

	// Calculate metrics (ensure metrics helpers are available)
	ssdReadAvail := sc.Availability(ssd.Read())
	ssdReadMean := sc.MeanLatency(ssd.Read())
	ssdReadP99 := sc.PercentileLatency(ssd.Read(), 0.99) // P99 of read latency
	ssdWriteMean := sc.MeanLatency(ssd.Write())

	hddReadAvail := sc.Availability(hdd.Read())
	hddReadMean := sc.MeanLatency(hdd.Read())
	hddReadP99 := sc.PercentileLatency(hdd.Read(), 0.99) // P99 of read latency
	hddWriteMean := sc.MeanLatency(hdd.Write())

	t.Logf("SSD Read : Avail=%.4f, Mean=%.6fs, P99=%.6fs", ssdReadAvail, ssdReadMean, ssdReadP99)
	t.Logf("SSD Write: Mean=%.6fs", ssdWriteMean)
	t.Logf("HDD Read : Avail=%.4f, Mean=%.6fs, P99=%.6fs", hddReadAvail, hddReadMean, hddReadP99)
	t.Logf("HDD Write: Mean=%.6fs", hddWriteMean)

	// --- Assertions based on expected profile differences ---

	// Availability might be similar by design, but check if reasonable
	if !approxEqualTest(ssdReadAvail, 0.998, 0.0001) { // 1 - 0.001 - 0.001
		t.Errorf("SSD Read Availability %.4f doesn't match expected 0.9980", ssdReadAvail)
	}
	if !approxEqualTest(hddReadAvail, 0.990, 0.0001) { // 1 - 0.005 - 0.005
		t.Errorf("HDD Read Availability %.4f doesn't match expected 0.9900", hddReadAvail)
	}

	// Latency: SSD should be significantly faster than HDD
	if ssdReadMean >= hddReadMean {
		t.Errorf("SSD Read Mean Latency (%.6fs) should be less than HDD (%.6fs)", ssdReadMean, hddReadMean)
	}
	if ssdWriteMean >= hddWriteMean {
		t.Errorf("SSD Write Mean Latency (%.6fs) should be less than HDD (%.6fs)", ssdWriteMean, hddWriteMean)
	}

	// Check P99 - HDD likely has a much higher P99 tail
	if ssdReadP99 >= hddReadP99 {
		t.Errorf("SSD Read P99 Latency (%.6fs) should be less than HDD (%.6fs)", ssdReadP99, hddReadP99)
	}

	// Check specific expected values (optional, but good for verifying profile definition)
	// Example: Expected SSD P99 read latency (0.95 + 0.04 = 0.99 cumulative weight) falls into the 2ms bucket
	expectedSSDP99Read := Micros(500)
	if !approxEqualTest(ssdReadP99, expectedSSDP99Read, 1e-9) {
		t.Errorf("SSD Read P99 Latency %.6fs doesn't match expected %.6fs", ssdReadP99, expectedSSDP99Read)
	}
	// Example: Expected HDD P99 read latency (0.85 + 0.10 = 0.95 cumulative weight) falls into the 15ms bucket? No, need 0.99*0.99 = 0.9801 target weight -> 100ms bucket.
	expectedHDDP99Read := Millis(100)
	if !approxEqualTest(hddReadP99, expectedHDDP99Read, 1e-9) {
		t.Errorf("HDD Read P99 Latency %.6fs doesn't match expected %.6fs", hddReadP99, expectedHDDP99Read)
	}

}

// Optional: Add test for ReadProcessWrite if not covered elsewhere
func TestDisk_ReadProcessWrite(t *testing.T) {
	ssd := NewDisk(ProfileSSD)
	processingTime := Millis(1) // 1ms processing

	rpwOutcomes := ssd.ReadProcessWrite(processingTime)

	if rpwOutcomes == nil {
		t.Fatal("ReadProcessWrite returned nil")
	}

	// Calculate metrics on the result
	rpwAvail := sc.Availability(rpwOutcomes)
	rpwMean := sc.MeanLatency(rpwOutcomes)

	// Expected availability = read_avail * write_avail
	expectedAvail := sc.Availability(ssd.Read()) * sc.Availability(ssd.Write())
	// Expected mean latency = read_mean + write_mean + processing_time (approximation)
	expectedMean := sc.MeanLatency(ssd.Read()) + sc.MeanLatency(ssd.Write()) + processingTime

	t.Logf("RPW SSD: Avail=%.6f (Exp: %.6f), Mean=%.6fs (Exp: %.6fs)", rpwAvail, expectedAvail, rpwMean, expectedMean)

	// Allow some tolerance due to how weights combine vs simple multiplication/addition
	if !approxEqualTest(rpwAvail, expectedAvail, 0.001) {
		t.Errorf("ReadProcessWrite Availability mismatch: got %.6f, expected around %.6f", rpwAvail, expectedAvail)
	}
	if !approxEqualTest(rpwMean, expectedMean, expectedMean*0.1) { // Allow 10% tolerance for mean approximation
		t.Errorf("ReadProcessWrite Mean Latency mismatch: got %.6fs, expected around %.6fs", rpwMean, expectedMean)
	}
}
