package runtime

import (
	"sync"

	"github.com/panyam/sdl/core"
)

// ExecutionTracer records the execution flow of a single simulation run.
type ExecutionTracer struct {
	mu      sync.Mutex
	Events  []*TraceEvent
	nextID  int64
	stack   []int64
	runtime *Runtime // Reference to runtime for metrics processing
}

// NewExecutionTracer creates a new tracer.
func NewExecutionTracer() *ExecutionTracer {
	return &ExecutionTracer{
		Events: make([]*TraceEvent, 0),
		nextID: 1,
		stack:  []int64{0},
	}
}

// SetRuntime sets the runtime reference for metrics processing
func (t *ExecutionTracer) SetRuntime(runtime *Runtime) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.runtime = runtime
}

// PushParentID manually pushes a parent ID onto the stack.
// Used by the aggregator to set the context for evaluating futures.
func (t *ExecutionTracer) PushParentID(id int64) {
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

func (t *ExecutionTracer) currentParentID() int64 {
	if len(t.stack) == 0 {
		return 0
	}
	return t.stack[len(t.stack)-1]
}

// Enter logs the entry into a function or block.
// It returns the ID of the newly created event.
func (t *ExecutionTracer) Enter(ts core.Duration, kind TraceEventKind, comp *ComponentInstance, method *MethodDecl, args ...string) int64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	eventID := t.nextID
	t.nextID++

	event := &TraceEvent{
		Kind:      kind,
		ID:        eventID,
		ParentID:  t.currentParentID(),
		Timestamp: ts,
		Component: comp,
		Method:    method,
		Arguments: args,
	}

	// Set computed fields for JSON serialization
	event.ComponentName = event.GetComponentName()
	event.MethodName = event.GetMethodName()

	t.Events = append(t.Events, event)

	// If it's a standard call, it becomes the new parent for subsequent nested calls.
	// For 'go' and 'wait', they are instantaneous events, not parent scopes.
	if kind == EventEnter {
		t.stack = append(t.stack, eventID)
	}

	return eventID
}

// Exit logs the exit from a function or block.
func (t *ExecutionTracer) Exit(ts core.Duration, duration core.Duration, comp *ComponentInstance, method *MethodDecl, retVal Value, err error) {
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
		Component: comp,
		Method:    method,
	}
	t.nextID++

	// Set computed fields for JSON serialization
	event.ComponentName = event.GetComponentName()
	event.MethodName = event.GetMethodName()

	if !retVal.IsNil() {
		event.ReturnValue = retVal.String()
	}
	if err != nil {
		event.ErrorMessage = err.Error()
	}

	t.Events = append(t.Events, event)

	// Process metrics if enabled
	// if t.runtime != nil && t.runtime.metricStore != nil { t.runtime.metricStore.ProcessTraceEvent(event) }
}
