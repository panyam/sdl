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
	b.WriteString("  rankdir=LR;\n")
	b.WriteString(fmt.Sprintf("  label=\"Dynamic Trace for: %s\";\n", trace.EntryPoint))
	b.WriteString("  node [shape=box, style=rounded];\n")

	sort.SliceStable(trace.Events, func(i, j int) bool {
		return trace.Events[i].Timestamp < trace.Events[j].Timestamp
	})

	// 1. First pass: Determine the owner of each event's scope
	scopeOwner := make(map[int]string)
	scopeOwner[0] = "User" // Root scope is owned by the User
	for _, event := range trace.Events {
		owner := scopeOwner[event.ParentID]
		if owner == "" {
			owner = "User" // Should only happen for the first event
		}

		if event.Kind == runtime.EventEnter {
			callee, _ := getParticipantAndMethod(event.Target())
			if callee == "self" {
				scopeOwner[event.ID] = owner // 'self' call remains within the owner's scope
			} else {
				scopeOwner[event.ID] = callee
			}
		} else {
			// For other event types, they are just part of their parent's scope
			scopeOwner[event.ID] = owner
		}
	}

	// 2. Discover all participants in topological order
	participantList := []string{"User"}
	participantSet := map[string]bool{"User": true}
	for _, event := range trace.Events {
		if event.Kind == runtime.EventEnter {
			callee, _ := getParticipantAndMethod(event.Target())
			if callee != "self" && !participantSet[callee] {
				participantList = append(participantList, callee)
				participantSet[callee] = true
			}
		}
	}

	// Declare nodes
	for _, p := range participantList {
		b.WriteString(fmt.Sprintf("  \"%s\";\n", p))
	}
	b.WriteString("\n")

	// 3. Generate edges
	traceCounter := 1
	for _, event := range trace.Events {
		caller := scopeOwner[event.ParentID]

		if event.Kind == runtime.EventEnter {
			callee, method := getParticipantAndMethod(event.Target())
			if callee == "self" {
				callee = caller
			}

			label := fmt.Sprintf("%d: %s(%s)", traceCounter, method, strings.Join(event.Arguments, ", "))
			b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"];\n", caller, callee, label))
			traceCounter++
		} else if event.Kind == runtime.EventWait {
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
