package bitly

import (
	sdl "github.com/panyam/sdl/core"
)

// IDGenerator component simulates generating unique short codes.
type IDGenerator struct {
	Name string
	// Configuration for latency/availability (could be more complex later)
	Latency        sdl.Outcomes[sdl.Duration]
	FailureProb    float64
	FailureLatency sdl.Outcomes[sdl.Duration]

	// Pre-calculated outcomes
	genOutcomes *sdl.Outcomes[sdl.AccessResult]
}

// Init initializes the IDGenerator.
func (ig *IDGenerator) Init(name string) *IDGenerator {
	ig.Name = name
	// Defaults: Very fast and reliable
	ig.Latency.Add(100, sdl.Micros(50))         // 50us
	ig.FailureProb = 0.00001                    // 0.001% failure (1 in 100k)
	ig.FailureLatency.Add(100, sdl.Micros(100)) // Fast failure detection

	// Set And functions if needed for composition later (though unlikely for Duration)
	ig.Latency.And = func(a, b sdl.Duration) sdl.Duration { return a + b }
	ig.FailureLatency.And = func(a, b sdl.Duration) sdl.Duration { return a + b }

	ig.calculateOutcomes()
	return ig
}

func (ig *IDGenerator) calculateOutcomes() {
	outcomes := &sdl.Outcomes[sdl.AccessResult]{And: sdl.AndAccessResults}
	totalProb := 1.0

	// Failures
	failProb := ig.FailureProb
	if failProb > 1e-9 {
		baseFailProb := failProb
		failWeightTotal := ig.FailureLatency.TotalWeight()
		if failWeightTotal > 1e-9 {
			for _, b := range ig.FailureLatency.Buckets {
				prob := baseFailProb * (b.Weight / failWeightTotal)
				if prob > 1e-9 {
					outcomes.Add(prob, sdl.AccessResult{Success: false, Latency: b.Value})
				}
			}
		} else { // Handle empty FailureLatency
			outcomes.Add(baseFailProb, sdl.AccessResult{Success: false, Latency: 0})
		}
		totalProb -= failProb
	}

	// Successes
	succProb := totalProb
	if succProb > 1e-9 {
		baseSuccProb := succProb
		succWeightTotal := ig.Latency.TotalWeight()
		if succWeightTotal > 1e-9 {
			for _, b := range ig.Latency.Buckets {
				prob := baseSuccProb * (b.Weight / succWeightTotal)
				if prob > 1e-9 {
					outcomes.Add(prob, sdl.AccessResult{Success: true, Latency: b.Value})
				}
			}
		} else { // Handle empty Latency
			outcomes.Add(baseSuccProb, sdl.AccessResult{Success: true, Latency: 0})
		}
	}

	ig.genOutcomes = outcomes
}

// GenerateID simulates creating a new unique ID.
func (ig *IDGenerator) GenerateID() *sdl.Outcomes[sdl.AccessResult] {
	if ig.genOutcomes == nil {
		ig.calculateOutcomes()
	} // Defensive
	return ig.genOutcomes
}
