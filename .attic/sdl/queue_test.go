package sdl

import (
	"math"
	"testing"
	// Ensure metrics helpers are accessible
)

func TestQueue_Init_Defaults(t *testing.T) {
	q := NewQueue()
	if q.AvgArrivalRate <= 0 || q.AvgServiceRate <= 0 {
		t.Error("Default rates should be positive")
	}
	if q.AvgArrivalRate >= q.AvgServiceRate {
		t.Error("Default service rate should be > arrival rate for stability")
	}
	if q.NumberOfServers != 1 {
		t.Error("Default number of servers should be 1")
	}
	if q.enqueueOutcomes == nil || q.enqueueOutcomes.Len() == 0 {
		t.Fatal("Enqueue outcomes not calculated")
	}
}

func TestQueue_ConfigureRates(t *testing.T) {
	q := NewQueue().ConfigureRates(50.0, 100.0, 2)
	if q.AvgArrivalRate != 50.0 {
		t.Error("Arrival rate mismatch")
	}
	if q.AvgServiceRate != 100.0 {
		t.Error("Service rate mismatch")
	}
	if q.NumberOfServers != 2 {
		t.Error("Num servers mismatch")
	}
}

func TestQueue_Enqueue(t *testing.T) {
	q := NewQueue()
	q.QueueFailProb = 0.01       // 1% fail prob
	q.EnqueueFailProb = 0.05     // Not really used yet for infinite queue
	q.MaxQueueSize = 10          // Make it bounded for testing EnqueueFailProb concept
	q.calculateEnqueueOutcomes() // Recalculate with new probs

	outcomes := q.Enqueue()
	if outcomes == nil || outcomes.Len() == 0 {
		t.Fatal("Enqueue returned nil/empty")
	}

	avail := Availability(outcomes) // Success = successfully enqueued
	t.Logf("Enqueue Avail: %.4f", avail)

	// Expected success = (1 - QueueFailProb) * (1 - EnqueueFailProb) -- if fail probs independent
	// Current simple model only uses QueueFailProb for infinite queue.
	// If MaxQueueSize > 0, it uses EnqueueFailProb but needs proper calculation based on state.
	// Let's just check against QueueFailProb for now.
	expectedAvail := 1.0 - q.QueueFailProb
	if !approxEqualTest(avail, expectedAvail, 1e-9) {
		t.Errorf("Enqueue availability mismatch: expected %.4f, got %.4f", expectedAvail, avail)
	}
}

func TestQueue_Dequeue_Stable_MM1(t *testing.T) {
	lambda := 8.0                                 // items/sec
	mu := 10.0                                    // items/sec
	q := NewQueue().ConfigureRates(lambda, mu, 1) // M/M/1

	waitOutcomes := q.Dequeue()
	if waitOutcomes == nil || waitOutcomes.Len() == 0 {
		t.Fatal("Dequeue (stable) returned nil/empty")
	}

	// Calculate expected average wait Wq for M/M/1: lambda / (mu * (mu - lambda))
	rho := lambda / mu
	expectedWq := lambda / (mu * (mu - lambda))
	if rho >= 1.0 {
		t.Fatal("Test setup error: queue should be stable")
	}

	// Calculate mean latency from the returned outcomes distribution
	// Need Map to AccessResult temporarily for MeanLatency helper
	waitAccessOutcomes := Map(waitOutcomes, func(d Duration) AccessResult { return AccessResult{true, d} })
	calculatedMeanWait := MeanLatency(waitAccessOutcomes)

	t.Logf("Dequeue M/M/1 (rho=%.2f): Expected Wq=%.6fs, Calculated Mean Wait=%.6fs (Buckets: %d)", rho, expectedWq, calculatedMeanWait, waitOutcomes.Len())

	// Check if calculated mean is reasonably close to theoretical Wq
	if !approxEqualTest(calculatedMeanWait, expectedWq, expectedWq*0.2) { // Allow 20% tolerance due to exponential approximation
		t.Errorf("Mean waiting time mismatch: expected %.6f, got %.6f", expectedWq, calculatedMeanWait)
	}

	// Check if P99 is significantly higher than mean for exponential
	p99Wait := PercentileLatency(waitAccessOutcomes, 0.99)
	t.Logf("Dequeue M/M/1 P99 Wait: %.6fs", p99Wait)
	if p99Wait < calculatedMeanWait*2.0 { // P99 of exponential should be >> mean
		t.Errorf("P99 Wait (%.6f) seems too low compared to mean (%.6f)", p99Wait, calculatedMeanWait)
	}

}

func TestQueue_Dequeue_Stable_MMc(t *testing.T) {
	lambda := 16.0 // items/sec
	mu := 10.0     // items/sec per server
	c := 2         // servers
	// rho = 16 / (2*10) = 0.8 < 1.0 -> Stable
	q := NewQueue().ConfigureRates(lambda, mu, c) // M/M/2

	waitOutcomes := q.Dequeue()
	if waitOutcomes == nil || waitOutcomes.Len() == 0 {
		t.Fatal("Dequeue (M/M/c stable) returned nil/empty")
	}

	// Calculate expected average wait Wq for M/M/c (using formulas from Dequeue)
	rho := lambda / (float64(c) * mu)
	sumTerm := 0.0
	for k := 0; k < c; k++ {
		sumTerm += math.Pow(lambda/mu, float64(k)) / factorial(uint64(k))
	}
	lastTermNum := math.Pow(lambda/mu, float64(c))
	lastTermDenom := factorial(uint64(c)) * (1.0 - rho)
	p0 := 1.0 / (sumTerm + (lastTermNum / lastTermDenom))
	lq := (p0 * math.Pow(lambda/mu, float64(c)) * rho) / (factorial(uint64(c)) * math.Pow(1.0-rho, 2))
	expectedWq := 0.0
	if lambda > 1e-9 {
		expectedWq = lq / lambda
	}

	waitAccessOutcomes := Map(waitOutcomes, func(d Duration) AccessResult { return AccessResult{true, d} })
	calculatedMeanWait := MeanLatency(waitAccessOutcomes)

	t.Logf("Dequeue M/M/c (rho=%.2f, c=%d): Expected Wq=%.6fs, Calculated Mean Wait=%.6fs (Buckets: %d)", rho, c, expectedWq, calculatedMeanWait, waitOutcomes.Len())

	if !approxEqualTest(calculatedMeanWait, expectedWq, expectedWq*0.2) { // Allow 20% tolerance
		t.Errorf("M/M/c Mean waiting time mismatch: expected %.6f, got %.6f", expectedWq, calculatedMeanWait)
	}
}

func TestQueue_Dequeue_Unstable(t *testing.T) {
	lambda := 12.0                                // items/sec
	mu := 10.0                                    // items/sec
	q := NewQueue().ConfigureRates(lambda, mu, 1) // M/M/1, rho = 1.2 -> Unstable

	waitOutcomes := q.Dequeue()

	// Expect empty outcomes for unstable queue
	if waitOutcomes == nil {
		t.Fatal("Dequeue(unstable) returned nil, expected empty Outcomes")
	}
	if waitOutcomes.Len() != 0 {
		t.Errorf("Dequeue(unstable) should return empty outcomes, got %d buckets", waitOutcomes.Len())
	}
	t.Logf("Dequeue M/M/1 (rho=%.2f): Returned %d buckets as expected for unstable queue.", lambda/mu, waitOutcomes.Len())
}

func TestQueue_Dequeue_NoLoad(t *testing.T) {
	lambda := 0.0                                 // items/sec
	mu := 10.0                                    // items/sec
	q := NewQueue().ConfigureRates(lambda, mu, 1) // M/M/1, rho = 0

	waitOutcomes := q.Dequeue()
	if waitOutcomes == nil || waitOutcomes.Len() == 0 {
		t.Fatal("Dequeue (no load) returned nil/empty")
	}

	waitAccessOutcomes := Map(waitOutcomes, func(d Duration) AccessResult { return AccessResult{true, d} })
	calculatedMeanWait := MeanLatency(waitAccessOutcomes)
	t.Logf("Dequeue M/M/1 (rho=0): Mean Wait=%.6fs (Buckets: %d)", calculatedMeanWait, waitOutcomes.Len())

	// Expect zero wait time
	if !approxEqualTest(calculatedMeanWait, 0.0, 1e-9) {
		t.Errorf("Mean waiting time mismatch: expected 0.0, got %.6f", calculatedMeanWait)
	}
	if waitOutcomes.Len() != 1 {
		t.Errorf("Expected 1 bucket for zero wait time, got %d", waitOutcomes.Len())
	}
}
