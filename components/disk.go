package components

import (
	"fmt"

	sc "github.com/panyam/sdl/core"
	// Assuming metrics helpers are available via import or same package
	// "log" // If needed for warnings
)

// Define constants for profile names
const (
	ProfileSSD = "SSD"
	ProfileHDD = "HDD"
)

// Predefined Outcome distributions for different profiles
var (
	ssdReadOutcomes  *Outcomes[sc.AccessResult]
	ssdWriteOutcomes *Outcomes[sc.AccessResult]
	hddReadOutcomes  *Outcomes[sc.AccessResult]
	hddWriteOutcomes *Outcomes[sc.AccessResult]
)

// Initialize the default profiles (called once, e.g., in an init function)
func init() {
	// --- SSD Profile ---
	ssdReadOutcomes = (&Outcomes[sc.AccessResult]{And: sc.AndAccessResults}).
		Add(0.95, sc.AccessResult{true, Micros(100)}). // 95% very fast read (0.1 ms)
		Add(0.04, sc.AccessResult{true, Micros(500)}). // 4% slightly slower (0.5 ms)
		Add(0.008, sc.AccessResult{true, Millis(2)}).  // 0.8% slower (2 ms)
		Add(0.001, sc.AccessResult{false, Millis(1)}). // 0.1% fast failure
		Add(0.001, sc.AccessResult{false, Millis(5)})  // 0.1% slower failure

	ssdWriteOutcomes = (&Outcomes[sc.AccessResult]{And: sc.AndAccessResults}).
		Add(0.96, sc.AccessResult{true, Micros(150)}). // 96% very fast write (0.15 ms)
		Add(0.03, sc.AccessResult{true, Micros(800)}). // 3% slower (0.8 ms)
		Add(0.008, sc.AccessResult{true, Millis(5)}).  // 0.8% much slower (5 ms)
		Add(0.001, sc.AccessResult{false, Millis(1)}). // 0.1% fast failure
		Add(0.001, sc.AccessResult{false, Millis(10)}) // 0.1% slower failure

	// --- HDD Profile ---
	hddReadOutcomes = (&Outcomes[sc.AccessResult]{And: sc.AndAccessResults}).
		Add(0.85, sc.AccessResult{true, Millis(5)}).    // 85% typical read (5 ms - seek + transfer)
		Add(0.10, sc.AccessResult{true, Millis(15)}).   // 10% slower seek/contention (15 ms)
		Add(0.04, sc.AccessResult{true, Millis(100)}).  // 4% very slow (100 ms)
		Add(0.005, sc.AccessResult{false, Millis(10)}). // 0.5% read failure
		Add(0.005, sc.AccessResult{false, Millis(50)})  // 0.5% slower failure

	hddWriteOutcomes = (&Outcomes[sc.AccessResult]{And: sc.AndAccessResults}).
		Add(0.88, sc.AccessResult{true, Millis(8)}).    // 88% typical write (8 ms)
		Add(0.08, sc.AccessResult{true, Millis(25)}).   // 8% slower write (25 ms)
		Add(0.03, sc.AccessResult{true, Millis(150)}).  // 3% very slow (150 ms)
		Add(0.005, sc.AccessResult{false, Millis(10)}). // 0.5% write failure
		Add(0.005, sc.AccessResult{false, Millis(50)})  // 0.5% slower failure
}

// Disk represents a storage device component
type Disk struct {
	// ProfileName identifies the type of disk (e.g., "SSD", "HDD")
	ProfileName string

	// Attributes for a Disk component
	// These will be pointers to the shared predefined outcomes
	ReadOutcomes  *Outcomes[sc.AccessResult]
	WriteOutcomes *Outcomes[sc.AccessResult]
}

// Init initializes a Disk based on the ProfileName.
// Defaults to SSD if profileName is empty or unrecognized.
func (d *Disk) Init() *Disk {
	switch d.ProfileName {
	case ProfileHDD:
		d.ReadOutcomes = hddReadOutcomes
		d.WriteOutcomes = hddWriteOutcomes
		// log.Printf("Initialized HDD Disk Profile")
	case ProfileSSD:
		fallthrough // Explicit fallthrough for SSD as default
	default:
		// Default to SSD
		if d.ProfileName != ProfileSSD && d.ProfileName != "" {
			// log.Printf("Warning: Unrecognized disk profile '%s'. Defaulting to SSD.", profileName)
		}
		d.ProfileName = ProfileSSD // Ensure ProfileName is set correctly for default
		d.ReadOutcomes = ssdReadOutcomes
		d.WriteOutcomes = ssdWriteOutcomes
		// log.Printf("Initialized SSD Disk Profile (or default)")
	}

	// Ensure outcomes are not nil, though init() should handle this. Defensive check.
	if d.ReadOutcomes == nil || d.WriteOutcomes == nil {
		// This should not happen if init() works correctly
		panic(fmt.Sprintf("Disk profile '%s' failed to load outcomes.", d.ProfileName))
	}

	return d
}

// NewDisk creates and initializes a new Disk component with the specified profile.
// Defaults to SSD if profileName is empty or unrecognized.
func NewDisk(profileName string) *Disk {
	d := &Disk{ProfileName: profileName}
	return d.Init()
}

// Read returns the read performance profile for this disk.
// The returned Outcomes should generally not be modified directly by callers,
// treat it as read-only. Use Copy() if modification is needed.
func (d *Disk) Read() *Outcomes[sc.AccessResult] {
	return d.ReadOutcomes
}

// Write returns the write performance profile for this disk.
// The returned Outcomes should generally not be modified directly by callers,
// treat it as read-only. Use Copy() if modification is needed.
func (d *Disk) Write() *Outcomes[sc.AccessResult] {
	return d.WriteOutcomes
}

// A common operation where a page is read, some processing is done and written back
// Note: This operation creates copies internally and does not modify the Disk's base outcomes.
func (d *Disk) ReadProcessWrite(processingTime Duration) *Outcomes[sc.AccessResult] {
	// Important: Use the Read() and Write() methods which return the pointers.
	// The And/If functions work with these pointers but create *new* Outcomes instances,
	// they do not modify the original ReadOutcomes/WriteOutcomes.

	readOutcomes := d.Read()
	writeOutcomes := d.Write()

	if readOutcomes == nil || writeOutcomes == nil {
		// Handle error: disk not initialized?
		return nil
	}

	// Combine Read -> Write outcomes sequentially
	combined := sc.And(readOutcomes, writeOutcomes, func(readRes, writeRes sc.AccessResult) sc.AccessResult {
		// Final success requires both read and write to succeed
		// Final latency is sum of latencies
		return sc.AccessResult{
			Success: readRes.Success && writeRes.Success,
			Latency: readRes.Latency + writeRes.Latency,
		}
	})

	// Add the processing latency *only* to the successful outcomes of the combined Read+Write operation
	// Need to create a new Outcomes list for the result
	finalOutcomes := &Outcomes[sc.AccessResult]{And: combined.And}
	for _, bucket := range combined.Buckets {
		newValue := bucket.Value
		if newValue.Success { // If the combined read+write was successful
			newValue.Latency += processingTime
		}
		finalOutcomes.Buckets = append(finalOutcomes.Buckets, sc.Bucket[sc.AccessResult]{
			Weight: bucket.Weight,
			Value:  newValue,
		})
	}

	return finalOutcomes
}
