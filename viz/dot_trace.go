package viz

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/panyam/sdl/runtime"
)

// DotTraceGenerator generates a DOT graph from a trace.
type DotTraceGenerator struct{}

// Generate creates a DOT graph string from trace data.
func (g *DotTraceGenerator) Generate(trace *runtime.TraceData) (string, error) {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("digraph \"%s_Trace\" {\n", trace.System))
	b.WriteString("  rankdir=TB;\n")
	b.WriteString(fmt.Sprintf("  label=\"Dynamic Trace for: %s\";\n", trace.EntryPoint))
	b.WriteString("  node [shape=box, style=rounded];\n")

	eventMap := make(map[int]*runtime.TraceEvent)
	for _, event := range trace.Events {
		eventMap[event.ID] = event
	}

	participants := g.discoverParticipants(trace, eventMap)
	for _, p := range participants {
		b.WriteString(fmt.Sprintf("  \"%s\";\n", p))
	}
	b.WriteString("\n")

	traceCounter := 1
	for _, event := range trace.Events {
		if event.Kind != runtime.EventEnter {
			continue
		}

		caller := g.findCaller(event, eventMap)
		callee, method := getParticipantAndMethod(event.Target)
		if callee == "self" {
			callee = caller
		}

		label := fmt.Sprintf("%d: %s(%s)", traceCounter, method, strings.Join(event.Arguments, ", "))
		b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", caller, callee, label))
		traceCounter++
	}
	
	// Add wait notes separately
	for _, event := range trace.Events {
		if event.Kind == runtime.EventWait {
			caller := g.findCaller(event, eventMap)
			aggregator := "wait"
			if len(event.Arguments) > 0 && event.Arguments[0] != "" {
				aggregator = fmt.Sprintf("wait using %s", event.Arguments[0])
			}
			b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\", style=dotted, arrowhead=none];\n", caller, caller, aggregator))
		}
	}


	b.WriteString("}\n")
	return b.String(), nil
}

// findCaller traverses up the event chain to find the true originating participant,
// skipping over async event types like 'go'.
func (g *DotTraceGenerator) findCaller(event *runtime.TraceEvent, eventMap map[int]*runtime.TraceEvent) string {
	curr := event
	for curr != nil {
		parent, ok := eventMap[curr.ParentID]
		if !ok {
			return "User" // Reached the root
		}
		
		// If the parent is a 'go' event, we need to find *its* caller.
		// If it's a normal 'enter' event, its target is the caller.
		if parent.Kind == runtime.EventEnter {
			caller, _ := getParticipantAndMethod(parent.Target)
			return caller
		}
		curr = parent
	}
	return "User"
}

// discoverParticipants finds all unique participants in topological order.
func (g *DotTraceGenerator) discoverParticipants(trace *runtime.TraceData, eventMap map[int]*runtime.TraceEvent) []string {
	participantList := []string{"User"}
	participantSet := map[string]bool{"User": true}

	addParticipant := func(name string) {
		if name != "self" && !participantSet[name] {
			participantList = append(participantList, name)
			participantSet[name] = true
		}
	}

	for _, event := range trace.Events {
		if event.Kind == runtime.EventEnter {
			caller := g.findCaller(event, eventMap)
			callee, _ := getParticipantAndMethod(event.Target)
			if callee == "self" {
				callee = caller
			}
			addParticipant(callee)
		}
	}
	return participantList
}
