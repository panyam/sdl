// sdl/examples/notifier/async_processor.go
package notifier

import (
	"log" // Using log for warnings/errors in manual approach

	sdl "github.com/panyam/sdl/core"
)

// AsyncProcessor simulates the logic that processes a single message
// after it's picked up by CDC.
type AsyncProcessor struct {
	Name          string
	MessageStore  *MessageStore
	InboxService  *InboxService
	MaxOutcomeLen int
	// Could add ResourcePool dependency if processing is resource-constrained
}

// Init initializes the AsyncProcessor.
func (ap *AsyncProcessor) Init(name string, ms *MessageStore, is *InboxService) *AsyncProcessor {
	ap.Name = name
	ap.MessageStore = ms
	ap.InboxService = is
	ap.MaxOutcomeLen = 15
	return ap
}

// ProcessMessage simulates processing one message for a distribution of recipients.
// This is where we handle the variable fan-out challenge.
func (ap *AsyncProcessor) ProcessMessage(messageID string, recipientCountDist *sdl.Outcomes[int]) *sdl.Outcomes[sdl.AccessResult] {

	// Step 1: Get Message Details (Optional, depends on what CDC provides)
	// Let's assume we need it for recipient list / content.
	getDetailsOutcome := ap.MessageStore.GetMessageDetails(messageID)

	// --- Step 2: Fan-out and Save to Inboxes (Manual Expansion Approach) ---
	// We'll manually expand for a few representative recipient counts.
	// More sophisticated: Use FanoutAnd or Repeat operator if available.

	finalCombinedOutcome := &sdl.Outcomes[sdl.AccessResult]{And: sdl.AndAccessResults}
	trimmer := sdl.TrimToSize(100, ap.MaxOutcomeLen)

	// Get the outcome for saving to ONE inbox
	saveOneOutcome := ap.InboxService.SaveToInbox( /*recipientID*/ "", messageID)

	for _, recipientBucket := range recipientCountDist.Buckets {
		numRecipients := recipientBucket.Value
		probOfThisCount := recipientBucket.Weight / recipientCountDist.TotalWeight() // Renormalize just in case

		if numRecipients <= 0 {
			continue
		} // Skip zero recipients case

		// Estimate cost of saving to N inboxes.
		// Simplification: Assume parallel saves, cost dominated by the MAX latency of N saves.
		// A very rough approximation: P99 latency of saving to N might be slightly higher
		// than P99 of saving to 1. Let's just reuse saveOneOutcome for simplicity for now,
		// but scale success probability. A better model would use Repeat(N, Parallel).
		// Success requires ALL N saves to succeed (approx = singleSaveAvail ^ N).
		// Latency is roughly the latency of a single save (max of parallel).
		// THIS IS A MAJOR SIMPLIFICATION for the manual approach.
		saveAllEstimateOutcome := sdl.Map(saveOneOutcome, func(ar sdl.AccessResult) sdl.AccessResult {
			if ar.Success {
				// Reduce success prob exponentially? Very approximate.
				singleSuccessProb := sdl.Availability(saveOneOutcome)
				if singleSuccessProb < 0 {
					singleSuccessProb = 0
				} // Clamp
				if singleSuccessProb > 1 {
					singleSuccessProb = 1
				}

				// Avoid pow(0,0) -> NaN
				/*
					allSuccessProb := 0.0
					if singleSuccessProb > 1e-12 || numRecipients == 0 {
						allSuccessProb = math.Pow(singleSuccessProb, float64(numRecipients))
					}
				*/

				// We can't easily change the success *value* within the outcome distribution here.
				// This approach is flawed for calculating combined success properly.
				// For now, let's just return the single save outcome, acknowledging inaccuracy.
				// TODO: Replace this with a proper Repeat(N, Parallel) or FanoutAnd model later.
				return ar // Placeholder: returns single save profile
			}
			// If single save failed, assume all fail (simplification)
			return ar
		})
		log.Printf("WARN: Using simplified fan-out cost (single save profile) for N=%d recipients in AsyncProcessor.", numRecipients)

		// Combine: GetDetails -> SaveAllEstimate
		pathOutcomeN := sdl.And(getDetailsOutcome, saveAllEstimateOutcome, sdl.AndAccessResults)

		// Scale the weights of this path's outcomes by the probability of having N recipients
		scaledPathOutcomeN := pathOutcomeN.Copy() // Create copy to scale
		scaledPathOutcomeN.ScaleWeights(probOfThisCount)

		// Append to the final result
		finalCombinedOutcome.Append(scaledPathOutcomeN)

		// Reduce intermediate results to prevent explosion if many recipient counts
		successes, failures := finalCombinedOutcome.Split(sdl.AccessResult.IsSuccess)
		trimmedSuccesses := trimmer(successes)
		trimmedFailures := trimmer(failures)
		finalCombinedOutcome = (&sdl.Outcomes[sdl.AccessResult]{And: sdl.AndAccessResults}).Append(trimmedSuccesses, trimmedFailures)

	}

	return finalCombinedOutcome
}
