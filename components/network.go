package components

import (
	"math/rand"
	"time"

	sc "github.com/panyam/leetcoach/sdl/core"
	// "log"
)

// NetworkLink represents a network connection with latency and loss characteristics.
type NetworkLink struct {
	// Base latency for a successful transfer (e.g., round-trip time)
	BaseLatency Duration

	// Jitter represents the variability in latency.
	// We'll model this simply by adding a random duration centered around 0.
	// MaxJitter represents the maximum deviation (positive or negative) from BaseLatency.
	MaxJitter Duration

	// PacketLossProb is the probability (0.0 to 1.0) that a transfer fails entirely.
	PacketLossProb float64

	// Number of buckets to represent the latency distribution (for jitter).
	// More buckets = smoother distribution, fewer = coarser approximation.
	LatencyBuckets int

	// Pre-calculated outcomes for efficiency (can be nil initially)
	transferOutcomes *Outcomes[sc.AccessResult]

	// Random source for jitter calculation
	rng *rand.Rand
}

// Init initializes the NetworkLink with provided parameters or defaults.
func (nl *NetworkLink) Init(baseLatency, maxJitter Duration, packetLoss float64) *NetworkLink {
	nl.BaseLatency = baseLatency
	nl.MaxJitter = maxJitter
	nl.PacketLossProb = packetLoss
	nl.LatencyBuckets = 5 // Default number of buckets for jitter distribution

	// Ensure probabilities are valid
	if nl.PacketLossProb < 0.0 {
		nl.PacketLossProb = 0.0
	}
	if nl.PacketLossProb > 1.0 {
		nl.PacketLossProb = 1.0
	}
	// Ensure jitter is non-negative
	if nl.MaxJitter < 0 {
		nl.MaxJitter = 0
	}

	// Initialize random source
	nl.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Pre-calculate the outcomes distribution
	nl.calculateTransferOutcomes()

	return nl
}

// NewNetworkLink creates and initializes a new NetworkLink component.
// Uses some reasonable defaults if parameters are zero/invalid.
func NewNetworkLink(baseLatency, maxJitter Duration, packetLoss float64) *NetworkLink {
	// Basic defaults if needed
	if baseLatency <= 0 {
		baseLatency = Millis(10) // Default to 10ms latency
	}
	if maxJitter <= 0 {
		maxJitter = baseLatency * 0.2 // Default to 20% jitter
	}
	nl := &NetworkLink{}
	return nl.Init(baseLatency, maxJitter, packetLoss)
}

// calculateTransferOutcomes generates the probabilistic outcomes for a transfer.
func (nl *NetworkLink) calculateTransferOutcomes() {
	outcomes := &Outcomes[sc.AccessResult]{And: sc.AndAccessResults}

	successProb := 1.0 - nl.PacketLossProb

	if successProb > 1e-9 { // Only add success buckets if probability is non-zero
		// Distribute success probability across latency buckets based on jitter
		bucketWeight := successProb / float64(nl.LatencyBuckets)
		latencyStep := (2 * nl.MaxJitter) / float64(nl.LatencyBuckets-1) // Step between latency points

		for i := range nl.LatencyBuckets {
			// Calculate latency for this bucket: Base +/- Jitter
			// Simple linear distribution for jitter for now:
			// Bucket 0: Base - MaxJitter
			// Bucket N-1: Base + MaxJitter
			jitterAmount := -nl.MaxJitter + float64(i)*latencyStep
			latency := nl.BaseLatency + jitterAmount

			// Ensure latency is non-negative
			if latency < 0 {
				latency = 0
			}

			// Add bucket for this latency point
			outcomes.Add(bucketWeight, sc.AccessResult{
				Success: true,
				Latency: latency,
			})
		}
	}

	// Add failure bucket if loss probability is non-zero
	if nl.PacketLossProb > 1e-9 {
		// Assume failure is detected quickly (e.g., timeout slightly longer than max expected latency)
		// We could make failure latency configurable too.
		failureLatency := nl.BaseLatency + nl.MaxJitter + Millis(1) // Simple estimate
		outcomes.Add(nl.PacketLossProb, sc.AccessResult{
			Success: false,
			Latency: failureLatency,
		})
	}

	nl.transferOutcomes = outcomes
}

// Transfer simulates sending data over the network link.
// Returns the pre-calculated distribution of outcomes.
// The returned Outcomes should generally not be modified directly by callers.
func (nl *NetworkLink) Transfer() *Outcomes[sc.AccessResult] {
	if nl.transferOutcomes == nil {
		// Should have been calculated in Init, but recalculate defensively
		nl.calculateTransferOutcomes()
	}
	return nl.transferOutcomes
}
