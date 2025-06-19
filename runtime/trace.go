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
	Kind         TraceEventKind     `json:"kind"`
	ParentID     int                `json:"parent_id,omitempty"`
	ID           int                `json:"id"`
	Timestamp    core.Duration      `json:"ts"`       // Virtual time in simulation
	Duration     core.Duration      `json:"dur,omitempty"` // Duration in virtual time
	Component    *ComponentInstance `json:"-"`        // Component instance (nil for native/global methods)
	Method       *MethodDecl        `json:"-"`        // Method declaration
	Arguments    []string           `json:"args,omitempty"`
	ReturnValue  string             `json:"ret,omitempty"`
	ErrorMessage string             `json:"err,omitempty"`
	// Computed fields for JSON serialization
	ComponentName string             `json:"component,omitempty"`
	MethodName    string             `json:"method,omitempty"`
}

// GetComponentName returns the component name for metrics/display
func (e *TraceEvent) GetComponentName() string {
	if e.Component != nil {
		return e.Component.ComponentDecl.Name.Value
	}
	return ""
}

// GetMethodName returns the method name for metrics/display
func (e *TraceEvent) GetMethodName() string {
	if e.Method != nil {
		return e.Method.Name.Value
	}
	return ""
}

// Target returns the target string for backward compatibility with viz package
func (e *TraceEvent) Target() string {
	compName := e.GetComponentName()
	methodName := e.GetMethodName()
	
	if compName != "" && methodName != "" {
		return compName + "." + methodName
	} else if methodName != "" {
		return methodName
	}
	return ""
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
	runtime *Runtime // Reference to runtime for metrics processing
}

// NewExecutionTracer creates a new tracer.
func NewExecutionTracer() *ExecutionTracer {
	return &ExecutionTracer{
		Events: make([]*TraceEvent, 0),
		nextID: 1,
		stack:  []int{0},
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
func (t *ExecutionTracer) Enter(ts core.Duration, kind TraceEventKind, comp *ComponentInstance, method *MethodDecl, args ...string) int {
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

// EnterString logs entry for non-method calls (like "wait", "goroutine_for_").
// This is for backward compatibility with existing code.
func (t *ExecutionTracer) EnterString(ts core.Duration, kind TraceEventKind, target string, args ...string) int {
	return t.Enter(ts, kind, nil, nil, args...)
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
	if t.runtime != nil && t.runtime.metricStore != nil {
		t.runtime.metricStore.ProcessTraceEvent(event)
	}
}
