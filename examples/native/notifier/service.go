// sdl/examples/notifier/notifier_service.go
package notifier

import (
	"fmt"

	sdl "github.com/panyam/sdl/core"
)

// NotifierService handles incoming messages and user queries for sent messages.
type NotifierService struct {
	Name         string
	MessageStore *MessageStore
	// Dependencies for async path (CDC, Processor, InboxStore) are not directly called
}

// Init initializes the NotifierService.
func (ns *NotifierService) Init(name string, ms *MessageStore) *NotifierService {
	ns.Name = name
	if ms == nil {
		panic(fmt.Sprintf("NotifierService '%s' initialized with nil MessageStore", name))
	}
	ns.MessageStore = ms
	return ns
}

// SendMessage simulates the initial, synchronous part of sending a message.
// It writes to the primary message store. The recipient count distribution is passed
// conceptually but not used in *this* immediate synchronous result.
func (ns *NotifierService) SendMessage(senderID string, messageID string, recipientsDist *sdl.Outcomes[int], content string) *sdl.Outcomes[sdl.AccessResult] {
	// The immediate result is just the latency/success of the initial save.
	// The asynchronous delivery is modeled separately for end-to-end analysis.
	dummyRecipients := []string{} // The actual list isn't used by SaveMessage's performance model here
	return ns.MessageStore.SaveMessage(senderID, messageID, dummyRecipients, content)
}

// MyMessages simulates fetching messages sent by the user.
func (ns *NotifierService) MyMessages(senderID string) *sdl.Outcomes[sdl.AccessResult] {
	return ns.MessageStore.GetMessagesBySender(senderID)
}
