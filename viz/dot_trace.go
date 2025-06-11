package viz

import (
	"bytes"
	"fmt"
	"sort"
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

	// 1. Discover all participants
	participants := g.discoverParticipants(trace)
	for _, p := range participants {
		b.WriteString(fmt.Sprintf("  \"%s\";\n", p))
	}
	b.WriteString("\n")

	// 2. Generate edges from events
	scopeParticipant := map[int]string{0: "User"} // Maps event ID to the participant owning that scope
	traceCounter := 1

	for _, event := range trace.Events {
		caller, parentFound := scopeParticipant[event.ParentID]
		if !parentFound && event.ParentID != 0 {
			caller = "Unknown" // Should not happen in a valid trace
		}

		switch event.Kind {
		case runtime.EventEnter:
			callee, method := getParticipantAndMethod(event.Target)
			if callee == "self" {
				callee = caller
			}
			scopeParticipant[event.ID] = callee

			// Create a detailed label for the edge
			label := fmt.Sprintf("%d: %s(%s)", traceCounter, method, strings.Join(event.Arguments, ", "))
			b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", caller, callee, label))
			traceCounter++

		case runtime.EventWait:
			aggregator := "wait"
			if len(event.Arguments) > 0 {
				aggregator = fmt.Sprintf("wait using %s", event.Arguments[0])
			}
			// Represent wait as a self-loop with a dotted style
			b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\", style=dotted, arrowhead=none];\n", caller, caller, aggregator))

		case runtime.EventGo:
			// For now, we don't add a special node, but this is where logic could go
			// The subsequent "enter" events will be children of this "go" event's parent.
			scopeParticipant[event.ID] = caller
		}
	}

	b.WriteString("}\n")
	return b.String(), nil
}

// discoverParticipants finds all unique participants in topological order.
func (g *DotTraceGenerator) discoverParticipants(trace *runtime.TraceData) []string {
	participantList := []string{"User"}
	participantSet := map[string]bool{"User": true}
	scopeParticipant := map[int]string{0: "User"}

	sort.SliceStable(trace.Events, func(i, j int) bool {
		return trace.Events[i].Timestamp < trace.Events[j].Timestamp
	})

	addParticipant := func(name string) {
		if name != "self" && !participantSet[name] {
			participantList = append(participantList, name)
			participantSet[name] = true
		}
	}

	for _, event := range trace.Events {
		if event.Kind == runtime.EventEnter {
			caller := scopeParticipant[event.ParentID]
			callee, _ := getParticipantAndMethod(event.Target)
			if callee == "self" {
				callee = caller
			}
			addParticipant(callee)
			scopeParticipant[event.ID] = callee
		} else if event.Kind == runtime.EventGo {
			caller := scopeParticipant[event.ParentID]
			scopeParticipant[event.ID] = caller
		}
	}
	return participantList
}
