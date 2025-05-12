package notifier

import (
	sdl "github.com/panyam/sdl/core"
)

// DefineCDCDelay creates an Outcomes distribution representing the
// latency between a message being saved to MessageStore and the AsyncProcessor
// being triggered for it.
func DefineCDCDelay() *sdl.Outcomes[sdl.Duration] {
	// Example: Mostly fast, but with occasional delays
	delay := (&sdl.Outcomes[sdl.Duration]{And: func(a, b sdl.Duration) sdl.Duration { return a + b }}).
		Add(90, sdl.Millis(50)). // 90% < 50ms
		Add(8, sdl.Millis(500)). // 8% < 500ms
		Add(2, sdl.Millis(2000)) // 2% up to 2s delay
	return delay
}

// AddCDCDelay is a helper to combine an AccessResult outcome with the CDC delay.
func AddCDCDelay(prevOutcome *sdl.Outcomes[sdl.AccessResult], cdcDelay *sdl.Outcomes[sdl.Duration]) *sdl.Outcomes[sdl.AccessResult] {
	if prevOutcome == nil {
		return nil
	}
	if cdcDelay == nil || cdcDelay.Len() == 0 {
		return prevOutcome
	} // No delay if undefined

	// Combine using And - map Duration to AccessResult first
	delayAsAccessResult := sdl.Map(cdcDelay, func(d sdl.Duration) sdl.AccessResult {
		// Assume adding delay itself doesn't cause failure, just adds time
		return sdl.AccessResult{Success: true, Latency: d}
	})

	combined := sdl.And(prevOutcome, delayAsAccessResult, sdl.AndAccessResults)

	// Apply reduction if needed (And increases bucket count)
	maxLen := 15 // Or get from config
	trimmer := sdl.TrimToSize(100, maxLen)
	successes, failures := combined.Split(sdl.AccessResult.IsSuccess)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalOutcome := (&sdl.Outcomes[sdl.AccessResult]{And: combined.And}).Append(trimmedSuccesses, trimmedFailures)

	return finalOutcome
}
