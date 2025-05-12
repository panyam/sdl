package notifier

import (
	"fmt"
	"testing"

	sdl "github.com/panyam/sdl/core"
)

// Setup function
func setupNotifierSystem(t *testing.T) (*NotifierService, *AsyncProcessor, *InboxService) {
	msgStore := NewMessageStore("PrimaryMsgStore")
	inboxSvc := NewInboxService("InboxSvc") // Create the service
	notifierSvc := (&NotifierService{}).Init("NotifierAPI", msgStore)
	asyncProcessor := (&AsyncProcessor{}).Init("MsgProcessor", msgStore, inboxSvc) // Async needs stores

	// Configure components if needed (e.g., record counts)
	// msgStore.PKIndex.NumRecords = 1e9
	// inboxSvc.InboxIndex.NumRecords = 5e9

	return notifierSvc, asyncProcessor, inboxSvc
}

func TestNotifier_EndToEndDelivery(t *testing.T) {
	notifierSvc, asyncProcessor, _ := setupNotifierSystem(t) // InboxService not directly called in this test path

	// --- Define Inputs ---
	senderID := "user_A"
	messageID := "msg_123"
	content := "Hello"
	// Distribution for number of recipients
	recipientCountDist := (&sdl.Outcomes[int]{And: func(a, b int) int { return a + b }}). // And func needed for type
												Add(90, 10).   // 90% -> 10 recipients
												Add(9, 1000).  // 9%  -> 1000 recipients
												Add(1, 100000) // 1%  -> 100k recipients (expensive!)

	cdcDelay := DefineCDCDelay()

	// --- Simulate End-to-End Flow Manually ---

	// 1. Initial Send (synchronous part)
	sendOutcome := notifierSvc.SendMessage(senderID, messageID, recipientCountDist, content)

	// 2. Add CDC Delay
	afterCdcOutcome := AddCDCDelay(sendOutcome, cdcDelay)

	// 3. Simulate Async Processing (includes GetDetails + Fanout Save)
	//    This uses the manual expansion/approximation within ProcessMessage
	asyncProcessingOutcome := asyncProcessor.ProcessMessage(messageID, recipientCountDist)

	// 4. Combine CDC result with Async processing result
	//    Success determined by BOTH succeeding (implicitly handled in asyncProcessingOutcome already?)
	//    Let's assume ProcessMessage failure includes GetDetails failure.
	//    The success probability of the fan-out save needs careful consideration.
	//    Using the simplified ProcessMessage, final success relies on its outcome.
	endToEndOutcome := sdl.And(afterCdcOutcome, asyncProcessingOutcome, sdl.AndAccessResults)

	// Reduce final result
	maxLen := 20 // Allow more buckets for this complex flow
	trimmer := sdl.TrimToSize(150, maxLen)
	successes, failures := endToEndOutcome.Split(sdl.AccessResult.IsSuccess)
	trimmedSuccesses := trimmer(successes)
	trimmedFailures := trimmer(failures)
	finalTrimmedOutcome := (&sdl.Outcomes[sdl.AccessResult]{And: sdl.AndAccessResults}).Append(trimmedSuccesses, trimmedFailures)

	// --- Analyze Results ---
	endToEndSLOMillis := 5000.0 // Example: P99 within 5 seconds

	// Availability will be impacted by initial save, get details, *and* the fanout save estimate
	// It will likely be lower than simple cases due to the fanout success approximation.
	expectations := []sdl.Expectation{
		sdl.ExpectAvailability(sdl.GTE, 0.80), // Expect lower avail due to complex path & fanout approx
		sdl.ExpectP99(sdl.LT, sdl.Millis(endToEndSLOMillis)),
	}

	analysisName := fmt.Sprintf("Notifier EndToEnd Delivery")
	analysisResult := sdl.Analyze(analysisName, func() *sdl.Outcomes[sdl.AccessResult] {
		// We pass the *already calculated* final outcome to Analyze
		return finalTrimmedOutcome
	}, expectations...)

	analysisResult.Assert(t)

	finalP99 := analysisResult.Metrics[sdl.P99LatencyMetric]
	if finalP99 < sdl.Millis(endToEndSLOMillis) {
		t.Logf("SLO MET: End-to-End P99 Latency (%.6fs) < target (%.3fs)", finalP99, sdl.Millis(endToEndSLOMillis))
	} else {
		t.Logf("SLO MISSED: End-to-End P99 Latency (%.6fs) >= target (%.3fs)", finalP99, sdl.Millis(endToEndSLOMillis))
	}
	t.Logf("WARN: End-to-End results use simplified fan-out cost model in AsyncProcessor.")

}

// TODO: Add tests for MyMessages and InboxService.GetMessages

// TODO: Add tests for MyMessages and InboxService.GetMessages
func TestNotifier_GetMessages(t *testing.T) {
	_, _, inboxSvc := setupNotifierSystem(t)
	recipientID := "user_B"

	// Define expectations for GetMessages (likely fast if using LSM)
	getMessagesExpectations := []sdl.Expectation{
		sdl.ExpectAvailability(sdl.GTE, 0.99), // Depends on LSM Read profile
		sdl.ExpectP99(sdl.LT, sdl.Millis(50)), // Should be relatively fast
	}

	// Analyze
	analysisResult := sdl.Analyze(fmt.Sprintf("Inbox GetMessages (%s)", recipientID), func() *sdl.Outcomes[sdl.AccessResult] {
		return inboxSvc.GetMessages(recipientID)
	}, getMessagesExpectations...)

	analysisResult.Assert(t)
}

func TestNotifier_MyMessages(t *testing.T) {
	notifierSvc, _, _ := setupNotifierSystem(t)
	senderID := "user_A"

	// Define expectations for MyMessages (depends on MessageStore SenderIndex Find)
	myMessagesExpectations := []sdl.Expectation{
		sdl.ExpectAvailability(sdl.GTE, 0.99), // Depends on HashIndex Find profile
		sdl.ExpectP99(sdl.LT, sdl.Millis(30)), // HashIndex Find should be fast
	}

	// Analyze
	analysisResult := sdl.Analyze(fmt.Sprintf("Notifier MyMessages (%s)", senderID), func() *sdl.Outcomes[sdl.AccessResult] {
		return notifierSvc.MyMessages(senderID)
	}, myMessagesExpectations...)

	analysisResult.Assert(t)
}
