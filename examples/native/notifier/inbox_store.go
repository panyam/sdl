package notifier

import (
	"github.com/panyam/sdl/components"
	sdl "github.com/panyam/sdl/core"
)

// InboxStore handles storage and retrieval of messages within recipient inboxes.
type InboxStore struct {
	Name          string
	InboxIndex    components.LSMTree // Example: Using LSMTree for write optimization
	MaxOutcomeLen int
}

// Init initializes the InboxStore.
func (is *InboxStore) Init(name string, index components.LSMTree) *InboxStore {
	is.Name = name
	is.MaxOutcomeLen = 15

	// Assume index is Init()ed externally
	is.InboxIndex = index
	is.InboxIndex.MaxOutcomeLen = is.MaxOutcomeLen

	return is
}

// NewInboxStore convenience constructor
func NewInboxStore(name string) *InboxStore {
	lsm := components.NewLSMTree()
	lsm.Init()
	// Configure LSM if needed
	// lsm.NumRecords = 5_000_000_000 // Potentially huge number of inbox items total

	return (&InboxStore{}).Init(name, *lsm)
}

// SaveToInbox simulates writing a message copy to a recipient's inbox. Uses LSM Write.
func (is *InboxStore) SaveToInbox(recipientID string, messageID string) *sdl.Outcomes[sdl.AccessResult] {
	// LSM Write is typically fast for ingest.
	return is.InboxIndex.Write()
}

// GetMessages simulates fetching messages for a recipient's inbox. Uses LSM Read.
func (is *InboxStore) GetMessages(recipientID string) *sdl.Outcomes[sdl.AccessResult] {
	// LSM Read performance depends on where data is found (memtable, levels).
	return is.InboxIndex.Read()
}
