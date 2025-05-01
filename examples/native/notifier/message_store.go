package notifier

import (
	"github.com/panyam/leetcoach/sdl/components"
	sdl "github.com/panyam/leetcoach/sdl/core"
)

// MessageStore handles storage and retrieval of original messages.
type MessageStore struct {
	Name          string
	PKIndex       components.HashIndex // For GetMessageDetails by MessageID
	SenderIndex   components.HashIndex // For MyMessages by SenderID (Example uses HashIndex)
	MaxOutcomeLen int
}

// Init initializes the MessageStore with concrete indexes.
func (ms *MessageStore) Init(name string, pkProfile, senderProfile components.HashIndex) *MessageStore {
	ms.Name = name
	ms.MaxOutcomeLen = 15 // Default, can be configured

	// Assume profiles are already Init()ed externally if needed
	ms.PKIndex = pkProfile
	ms.SenderIndex = senderProfile

	ms.PKIndex.MaxOutcomeLen = ms.MaxOutcomeLen     // Propagate setting
	ms.SenderIndex.MaxOutcomeLen = ms.MaxOutcomeLen // Propagate setting

	return ms
}

// NewMessageStore convenience constructor
func NewMessageStore(name string) *MessageStore {
	// Create default indexes (e.g., SSD based)
	pkIndex := components.NewHashIndex() // Uses defaults
	pkIndex.Init()                       // Ensure disk is initialized within the index
	senderIndex := components.NewHashIndex()
	senderIndex.Init()

	// Configure indexes if needed (e.g., NumRecords)
	// pkIndex.NumRecords = 1_000_000_000 // Example: Large number of messages
	// senderIndex.NumRecords = 1_000_000_000

	return (&MessageStore{}).Init(name, *pkIndex, *senderIndex)
}

// SaveMessage simulates writing the message initially. Uses the PK index insert.
func (ms *MessageStore) SaveMessage(senderID string, messageID string, recipients []string, content string) *sdl.Outcomes[sdl.AccessResult] {
	// In reality, might write to multiple indexes.
	// Simplification: Cost dominated by writing the main record via PK index.
	// We pass data conceptually; the implementation uses the index's Insert profile.
	return ms.PKIndex.Insert()
}

// GetMessageDetails simulates fetching message content and recipients by ID. Uses PK index.
func (ms *MessageStore) GetMessageDetails(messageID string) *sdl.Outcomes[sdl.AccessResult] {
	// This simulates finding the message data needed by the async processor.
	return ms.PKIndex.Find()
}

// GetMessagesBySender simulates fetching messages sent by a specific user. Uses Sender index.
func (ms *MessageStore) GetMessagesBySender(senderID string) *sdl.Outcomes[sdl.AccessResult] {
	// This simulates the lookup for the "MyMessages" feature.
	return ms.SenderIndex.Find() // Assumes Find simulates finding multiple relevant records
}
