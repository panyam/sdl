package runtime

import (
	"sync"
	"github.com/panyam/sdl/core"
)

// TraceEventKind defines the type of a trace event.
type TraceEventKind string

const (
	EventEnter TraceEventKind = "enter"
	EventExit  TraceEventKind = "exit"
	EventGo    TraceEventKind = "go"
	EventWait  TraceEventKind = "wait"
)

// TraceEvent represents a single event in an execution trace.
type TraceEvent struct {
	Kind         TraceEventKind `json:"kind"`
	ParentID     int            `json:"parent_id,omitempty"` // ID of the parent event
	ID           int            `json:"id"`                  // Unique ID for this event
	Timestamp    float64        `json:"ts"`                  // Simulation time at which event occurred
	Duration     float64        `json:"dur,omitempty"`       // Duration of the event (especially for exits)
	Target       string         `json:"target"`              // e.g., "MyService.doWork" or "future1"
	Arguments    []string       `json:"args,omitempty"`      // String representation of arguments
	ReturnValue  string         `json:"ret,omitempty"`       // String representation of return value
	ErrorMessage string         `json:"err,omitempty"`       // Error message if any
}

// TraceData is the top-level structure for a trace file.
type TraceData struct {
	System     string        `json:"system"`
	EntryPoint string        `json:"entry_point"`
	Events     []*TraceEvent `json:"events"`
}


// ExecutionTracer records the execution flow of a single simulation run.
type ExecutionTracer struct {
	mu         sync.Mutex
	Events     []*TraceEvent
	nextID     int
	stack      []int // Stores the ID of the current parent event
}

// NewExecutionTracer creates a new tracer.
func NewExecutionTracer() *ExecutionTracer {
	return &ExecutionTracer{
		Events: make([]*TraceEvent, 0),
		nextID: 1,
		stack:  []int{0}, // Start with a root event "parent" of 0
	}
}

// currentParentID returns the ID of the event at the top of the stack.
func (t *ExecutionTracer) currentParentID() int {
	if len(t.stack) == 0 {
		return 0 // Root
	}
	return t.stack[len(t.stack)-1]
}

// pushParent puts a new event ID onto the stack.
func (t *ExecutionTracer) pushParent(id int) {
	t.stack = append(t.stack, id)
}

// popParent removes the most recent event ID from the stack.
func (t *ExecutionTracer) popParent() {
	if len(t.stack) > 1 { // Don't pop the root
		t.stack = t.stack[:len(t.stack)-1]
	}
}

// Enter logs the entry into a function or block.
// It returns the ID of the newly created "enter" event, which should be used
// for the corresponding Exit call.
func (t *ExecutionTracer) Enter(ts float64, kind TraceEventKind, target string, args ...string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	eventID := t.nextID
	t.nextID++

	event := &TraceEvent{
		Kind:      kind,
		ID:        eventID,
		ParentID:  t.currentParentID(),
		Timestamp: ts,
		Target:    target,
		Arguments: args,
	}
	t.Events = append(t.Events, event)

	// The new event becomes the parent for subsequent events.
	t.pushParent(eventID)

	return eventID
}

// Exit logs the exit from a function or block.
func (t *ExecutionTracer) Exit(eventID int, ts float64, duration core.Duration, retVal Value, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// Pop this event from the parent stack first.
	t.popParent()

	// Find the corresponding "enter" event to update its duration,
	// or create a new "exit" event. Let's create a new event for simplicity.
	event := &TraceEvent{
		Kind:      EventExit,
		ID:        t.nextID,
		ParentID:  t.currentParentID(), 
		Timestamp: ts,
		Duration:  duration,
		Target:    "", // Target is known from the enter event
	}
	t.nextID++

	if !retVal.IsNil() {
		event.ReturnValue = retVal.String()
	}
	if err != nil {
		event.ErrorMessage = err.Error()
	}

	t.Events = append(t.Events, event)
}
