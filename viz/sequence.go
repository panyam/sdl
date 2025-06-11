package viz

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/panyam/sdl/runtime"
)

// MermaidSequenceGenerator generates Mermaid sequence diagrams from a trace.
type MermaidSequenceGenerator struct{}

// Generate creates a Mermaid sequence diagram as a string.
func (g *MermaidSequenceGenerator) Generate(trace *runtime.TraceData) (string, error) {
	var b bytes.Buffer
	b.WriteString("sequenceDiagram\n")

	sort.SliceStable(trace.Events, func(i, j int) bool {
		return trace.Events[i].Timestamp < trace.Events[j].Timestamp
	})

	// 1. First pass: Determine owner of each scope and discover all participants
	scopeOwner := make(map[int]string)
	scopeOwner[0] = "User"
	participantList := []string{"User"}
	participantSet := map[string]bool{"User": true}

	addParticipant := func(name string) {
		if !participantSet[name] {
			participantList = append(participantList, name)
			participantSet[name] = true
		}
	}

	for _, event := range trace.Events {
		owner := scopeOwner[event.ParentID]
		if owner == "" {
			owner = "User" // Default for safety
		}

		if event.Kind == runtime.EventEnter {
			callee, _ := getParticipantAndMethod(event.Target)
			if callee == "self" {
				scopeOwner[event.ID] = owner
				addParticipant(owner)
			} else {
				scopeOwner[event.ID] = callee
				addParticipant(callee)
			}
		} else {
			scopeOwner[event.ID] = owner
		}
	}

	// 2. Declare all participants in the discovered order
	for _, p := range participantList {
		b.WriteString(fmt.Sprintf("  participant %s\n", p))
	}
	b.WriteString("\n")

	// 3. Process events using the scope map to generate the sequence
	inParallelBlock := false
	for _, event := range trace.Events {
		caller := scopeOwner[event.ParentID]

		switch event.Kind {
		case runtime.EventEnter:
			callee, method := getParticipantAndMethod(event.Target)
			if callee == "self" {
				callee = caller
			}

			b.WriteString(fmt.Sprintf("  %s->>%s: %s(%s)\n", caller, callee, method, strings.Join(event.Arguments, ", ")))
			if !inParallelBlock {
				b.WriteString(fmt.Sprintf("  activate %s\n", callee))
			}

		case runtime.EventExit:
			// The owner of the exit event is the participant that needs to be deactivated.
			participantToDeactivate := scopeOwner[event.ID]
			if !inParallelBlock {
				b.WriteString(fmt.Sprintf("  deactivate %s\n", participantToDeactivate))
			}

		case runtime.EventGo:
			loopCount := "N"
			if len(event.Arguments) > 0 {
				loopCount = event.Arguments[0]
			}
			b.WriteString(fmt.Sprintf("  loop %s times\n", loopCount))
			inParallelBlock = true

		case runtime.EventWait:
			if inParallelBlock {
				b.WriteString("  end\n") // End the 'loop' block
				inParallelBlock = false
			}
			aggregator := "wait"
			if len(event.Arguments) > 0 && event.Arguments[0] != "" {
				aggregator = fmt.Sprintf("wait using %s", event.Arguments[0])
			}
			b.WriteString(fmt.Sprintf("  Note over %s: %s\n", caller, aggregator))
		}
	}

	return b.String(), nil
}

// getParticipantAndMethod extracts the participant and method from a target string like "instance.method".
func getParticipantAndMethod(target string) (participant, method string) {
	parts := strings.SplitN(target, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return target, ""
}
