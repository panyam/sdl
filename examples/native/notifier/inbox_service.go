package notifier

import (
	"fmt"

	sdl "github.com/panyam/sdl/core"
)

// InboxService provides operations related to user inboxes.
// It acts as a wrapper around the underlying InboxStore.
type InboxService struct {
	Name       string
	InboxStore *InboxStore // Dependency on the concrete inbox storage
}

// Init initializes the InboxService.
func (isvc *InboxService) Init(name string, store *InboxStore) *InboxService {
	isvc.Name = name
	if store == nil {
		panic(fmt.Sprintf("InboxService '%s' initialized with nil InboxStore", name))
	}
	isvc.InboxStore = store
	return isvc
}

// NewInboxService convenience constructor
func NewInboxService(name string) *InboxService {
	// Create the underlying store with defaults
	store := NewInboxStore(name + "_Store") // Give store a related name
	return (&InboxService{}).Init(name, store)
}

// SaveToInbox simulates writing a message to a specific recipient's inbox.
// Delegates to the InboxStore.
func (isvc *InboxService) SaveToInbox(recipientID string, messageID string) *sdl.Outcomes[sdl.AccessResult] {
	// The service layer might add minimal overhead or validation, but for performance
	// modeling, we often delegate directly to the storage layer's profile.
	return isvc.InboxStore.SaveToInbox(recipientID, messageID)
}

// GetMessages simulates fetching messages for a user's inbox.
// Delegates to the InboxStore.
func (isvc *InboxService) GetMessages(recipientID string) *sdl.Outcomes[sdl.AccessResult] {
	return isvc.InboxStore.GetMessages(recipientID)
}
