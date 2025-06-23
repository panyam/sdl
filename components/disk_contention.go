package components

import (
	"fmt"

	sc "github.com/panyam/sdl/core"
)

// DiskWithContention represents a storage device that models contention
// using either ResourcePool (for parallel I/O) or MM1Queue (for serialized I/O).
type DiskWithContention struct {
	// Contention modeling - choose one based on disk type
	pool  *ResourcePool // For SSDs with parallel I/O
	queue *MM1Queue     // For HDDs with serialized access

	// Base disk for raw performance profiles
	baseDisk *Disk

	// Arrival rates per method
	arrivalRates map[string]float64
}

// Init initializes a DiskWithContention based on the ProfileName.
func (d *DiskWithContention) Init() {
	// Initialize base disk
	d.baseDisk = NewDisk()
	d.arrivalRates = make(map[string]float64)

	// Configure contention model based on disk type
	// HDD uses MM1Queue for serialized access
	d.queue = &MM1Queue{
		Name:           fmt.Sprintf("disk-queue"),
		ArrivalRate:    1e-9, // Will be updated via SetArrivalRate
		AvgServiceTime: 0.01, // 10ms average (from HDD profile)
	}
	d.queue.Init()

	// SSD uses ResourcePool for parallel I/O
	d.pool = &ResourcePool{
		Name:        fmt.Sprintf("disk-pool"),
		Size:        32,     // SSDs can handle multiple parallel I/Os
		ArrivalRate: 1e-9,   // Will be updated via SetArrivalRate
		AvgHoldTime: 0.0005, // 0.5ms average (from SSD profile)
	}
	d.pool.Init()
}

// NewDiskWithContention creates a contention-aware disk component.
func NewDiskWithContention() *DiskWithContention {
	d := &DiskWithContention{}
	d.Init()
	return d
}

// SetArrivalRate sets the arrival rate for a specific method.
func (d *DiskWithContention) SetArrivalRate(method string, rate float64) error {
	d.arrivalRates[method] = rate

	// Update total arrival rate in the contention model
	totalRate := d.GetTotalArrivalRate()

	if d.pool != nil {
		d.pool.ArrivalRate = totalRate
	} else if d.queue != nil {
		d.queue.ArrivalRate = totalRate
	}

	return nil
}

// GetArrivalRate returns the arrival rate for a specific method.
func (d *DiskWithContention) GetArrivalRate(method string) float64 {
	if rate, ok := d.arrivalRates[method]; ok {
		return rate
	}
	return 0
}

// GetTotalArrivalRate returns the sum of all method arrival rates.
func (d *DiskWithContention) GetTotalArrivalRate() float64 {
	total := 0.0
	for _, rate := range d.arrivalRates {
		total += rate
	}
	return total
}

// Read performs a read operation with contention modeling.
func (d *DiskWithContention) Read() *Outcomes[sc.AccessResult] {
	// For now, we'll return the Outcomes and let the caller sample
	// This matches the pattern used by other components

	if d.pool != nil {
		// SSD with parallel I/O - combine pool acquisition with disk read
		poolOutcomes := d.pool.Acquire()
		diskOutcomes := d.baseDisk.Read()

		// Combine outcomes: pool acquisition followed by disk read
		return sc.And(poolOutcomes, diskOutcomes, func(poolResult, diskResult sc.AccessResult) sc.AccessResult {
			if !poolResult.Success {
				// Pool acquisition failed
				return sc.AccessResult{Success: false, Latency: poolResult.Latency}
			}
			// Both pool and disk latencies
			return sc.AccessResult{
				Success: diskResult.Success,
				Latency: poolResult.Latency + diskResult.Latency,
			}
		})
	} else if d.queue != nil {
		// HDD with serialized access - combine queue wait with disk read
		queueOutcomes := d.queue.Dequeue()
		diskOutcomes := d.baseDisk.Read()

		// Combine outcomes: queue wait followed by disk read
		combinedOutcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
		for _, queueBucket := range queueOutcomes.Buckets {
			queueDelay := queueBucket.Value
			for _, diskBucket := range diskOutcomes.Buckets {
				diskResult := diskBucket.Value
				combinedOutcomes.Add(
					queueBucket.Weight*diskBucket.Weight,
					sc.AccessResult{
						Success: diskResult.Success,
						Latency: queueDelay + diskResult.Latency,
					},
				)
			}
		}
		return combinedOutcomes
	} else {
		// No contention modeling
		return d.baseDisk.Read()
	}
}

// Write performs a write operation with contention modeling.
func (d *DiskWithContention) Write() *Outcomes[sc.AccessResult] {
	// Similar to Read, but using Write outcomes

	if d.pool != nil {
		// SSD with parallel I/O - combine pool acquisition with disk write
		poolOutcomes := d.pool.Acquire()
		diskOutcomes := d.baseDisk.Write()

		// Combine outcomes: pool acquisition followed by disk write
		return sc.And(poolOutcomes, diskOutcomes, func(poolResult, diskResult sc.AccessResult) sc.AccessResult {
			if !poolResult.Success {
				// Pool acquisition failed
				return sc.AccessResult{Success: false, Latency: poolResult.Latency}
			}
			// Both pool and disk latencies
			return sc.AccessResult{
				Success: diskResult.Success,
				Latency: poolResult.Latency + diskResult.Latency,
			}
		})
	} else if d.queue != nil {
		// HDD with serialized access - combine queue wait with disk write
		queueOutcomes := d.queue.Dequeue()
		diskOutcomes := d.baseDisk.Write()

		// Combine outcomes: queue wait followed by disk write
		combinedOutcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}
		for _, queueBucket := range queueOutcomes.Buckets {
			queueDelay := queueBucket.Value
			for _, diskBucket := range diskOutcomes.Buckets {
				diskResult := diskBucket.Value
				combinedOutcomes.Add(
					queueBucket.Weight*diskBucket.Weight,
					sc.AccessResult{
						Success: diskResult.Success,
						Latency: queueDelay + diskResult.Latency,
					},
				)
			}
		}
		return combinedOutcomes
	} else {
		// No contention modeling
		return d.baseDisk.Write()
	}
}
