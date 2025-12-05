package runtime

import (
	"github.com/panyam/sdl/lib/core"
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
	ParentID     int64              `json:"parent_id,omitempty"`
	ID           int64              `json:"id"`
	Timestamp    core.Duration      `json:"ts"`            // Virtual time in simulation
	Duration     core.Duration      `json:"dur,omitempty"` // Duration in virtual time
	Component    *ComponentInstance `json:"-"`             // Component instance (nil for native/global methods)
	Method       *MethodDecl        `json:"-"`             // Method declaration
	Arguments    []string           `json:"args,omitempty"`
	ReturnValue  string             `json:"ret,omitempty"`
	ErrorMessage string             `json:"err,omitempty"`
	// Computed fields for JSON serialization
	ComponentName string `json:"component,omitempty"`
	MethodName    string `json:"method,omitempty"`
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
