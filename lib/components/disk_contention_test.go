package components

import (
	"testing"
)

func TestDiskWithContention(t *testing.T) {
	t.Run("SSD with parallel I/O", func(t *testing.T) {
		disk := NewDiskWithContention()

		// Set arrival rates
		disk.SetArrivalRate("Read", 1000.0) // 1000 RPS for reads
		disk.SetArrivalRate("Write", 500.0) // 500 RPS for writes

		// Verify rates are set
		if rate := disk.GetArrivalRate("Read"); rate != 1000.0 {
			t.Errorf("Expected Read arrival rate 1000, got %f", rate)
		}
		if rate := disk.GetArrivalRate("Write"); rate != 500.0 {
			t.Errorf("Expected Write arrival rate 500, got %f", rate)
		}
		if total := disk.GetTotalArrivalRate(); total != 1500.0 {
			t.Errorf("Expected total arrival rate 1500, got %f", total)
		}

		// Verify pool is configured with total rate
		if disk.pool == nil {
			t.Fatal("Expected SSD to use ResourcePool")
		}
		if disk.pool.ArrivalRate != 1500.0 {
			t.Errorf("Expected pool arrival rate 1500, got %f", disk.pool.ArrivalRate)
		}

		// Test read operation
		readOutcomes := disk.Read()
		if readOutcomes == nil || len(readOutcomes.Buckets) == 0 {
			t.Error("Read returned no outcomes")
		}
	})

	t.Run("HDD with serialized I/O", func(t *testing.T) {
		disk := NewDiskWithContention()

		// Set arrival rates
		disk.SetArrivalRate("Read", 50.0)  // 50 RPS for reads
		disk.SetArrivalRate("Write", 25.0) // 25 RPS for writes

		// Verify rates are set
		if total := disk.GetTotalArrivalRate(); total != 75.0 {
			t.Errorf("Expected total arrival rate 75, got %f", total)
		}

		// Verify queue is configured with total rate
		if disk.queue == nil {
			t.Fatal("Expected HDD to use MM1Queue")
		}
		if disk.queue.ArrivalRate != 75.0 {
			t.Errorf("Expected queue arrival rate 75, got %f", disk.queue.ArrivalRate)
		}

		// Test write operation
		writeOutcomes := disk.Write()
		if writeOutcomes == nil || len(writeOutcomes.Buckets) == 0 {
			t.Error("Write returned no outcomes")
		}
	})

	t.Run("No arrival rate set", func(t *testing.T) {
		disk := NewDiskWithContention()

		// Default should be 0
		if rate := disk.GetArrivalRate("Read"); rate != 0 {
			t.Errorf("Expected default Read arrival rate 0, got %f", rate)
		}
		if total := disk.GetTotalArrivalRate(); total != 0 {
			t.Errorf("Expected default total arrival rate 0, got %f", total)
		}
	})
}
