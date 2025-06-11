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
	ParentID     int            `json:"parent_id,omitempty"`
	ID           int            `json:"id"`
	Timestamp    float64        `json:"ts"`
	Duration     float64        `json:"dur,omitempty"`
	Target       string         `json:"target"`
	Arguments    []string       `json:"args,omitempty"`
	ReturnValue  string         `json:"ret,omitempty"`
	ErrorMessage string         `json:"err,omitempty"`
}

// TraceData is the top-level structure for a trace file.
type TraceData struct {
	System     string        `json:"system"`
	EntryPoint string        `json:"entry_point"`
	Events     []*TraceEvent `json:"events"`
}

// ExecutionTracer records the execution flow of a single simulation run.
type ExecutionTracer struct {
	mu     sync.Mutex
	Events []*TraceEvent
	nextID int
	stack  []int
}

// NewExecutionTracer creates a new tracer.
func NewExecutionTracer() *ExecutionTracer {
	return &ExecutionTracer{
		Events: make([]*TraceEvent, 0),
		nextID: 1,
		stack:  []int{0},
	}
}

// PushParentID manually pushes a parent ID onto the stack.
// Used by the aggregator to set the context for evaluating futures.
func (t *ExecutionTracer) PushParentID(id int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stack = append(t.stack, id)
}

// PopParent removes the most recent event ID from the stack.
func (t *ExecutionTracer) PopParent() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.stack) > 1 {
		t.stack = t.stack[:len(t.stack)-1]
	}
}

func (t *ExecutionTracer) currentParentID() int {
	if len(t.stack) == 0 {
		return 0
	}
	return t.stack[len(t.stack)-1]
}

// Enter logs the entry into a function or block.
// It returns the ID of the newly created event.
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

	// If it's a standard call, it becomes the new parent for subsequent nested calls.
	// For 'go' and 'wait', they are instantaneous events, not parent scopes.
	if kind == EventEnter {
		t.stack = append(t.stack, eventID)
	}

	return eventID
}

// Exit logs the exit from a function or block.
func (t *ExecutionTracer) Exit(ts float64, duration core.Duration, retVal Value, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Pop the corresponding "enter" event from the parent stack.
	if len(t.stack) > 1 {
		t.stack = t.stack[:len(t.stack)-1]
	}

	event := &TraceEvent{
		Kind:      EventExit,
		ID:        t.nextID,
		ParentID:  t.currentParentID(),
		Timestamp: ts,
		Duration:  duration,
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
