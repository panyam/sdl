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

	// 1. Discover all participants in topological order
	participants := g.discoverParticipants(trace)
	for _, p := range participants {
		b.WriteString(fmt.Sprintf("  participant %s\n", p))
	}
	b.WriteString("\n")

	// 2. Process events using a proper stack to manage activation state
	activeStack := []string{"User"}
	inParallelBlock := false
	goEventParent := ""

	for _, event := range trace.Events {
		if len(activeStack) == 0 {
			// This is a failsafe, should not happen in a valid trace starting from User
			activeStack = append(activeStack, "User")
		}
		caller := activeStack[len(activeStack)-1]

		switch event.Kind {
		case runtime.EventEnter:
			callee, method := getParticipantAndMethod(event.Target)
			if callee == "self" {
				callee = caller
			}
			
			// If we are in a parallel block, the caller is the entity that started the gobatch
			if inParallelBlock {
				caller = goEventParent
			}

			b.WriteString(fmt.Sprintf("  %s->>%s: %s(%s)\n", caller, callee, method, strings.Join(event.Arguments, ", ")))
			
			if !inParallelBlock {
				b.WriteString(fmt.Sprintf("  activate %s\n", callee))
				activeStack = append(activeStack, callee)
			}

		case runtime.EventExit:
			if !inParallelBlock && len(activeStack) > 1 {
				participantToDeactivate := activeStack[len(activeStack)-1]
				activeStack = activeStack[:len(activeStack)-1]
				b.WriteString(fmt.Sprintf("  deactivate %s\n", participantToDeactivate))
			}

		case runtime.EventGo:
			loopCount := "N"
			if len(event.Arguments) > 0 {
				loopCount = event.Arguments[0]
			}
			b.WriteString(fmt.Sprintf("  loop %s times\n", loopCount))
			inParallelBlock = true
			goEventParent = caller // Remember who started the parallel block

		case runtime.EventWait:
			if inParallelBlock {
				b.WriteString("  end\n")
				inParallelBlock = false
				goEventParent = ""
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

// discoverParticipants finds all unique participants in topological order from the trace.
func (g *MermaidSequenceGenerator) discoverParticipants(trace *runtime.TraceData) []string {
	participantList := []string{"User"}
	participantSet := map[string]bool{"User": true}

	addParticipant := func(name string) {
		if name != "self" && !participantSet[name] {
			participantList = append(participantList, name)
			participantSet[name] = true
		}
	}

	// This map tracks the "owner" of a scope, which is needed to resolve 'self' calls
	scopeOwner := make(map[int]string)
	scopeOwner[0] = "User" 

	for _, event := range trace.Events {
		owner := scopeOwner[event.ParentID]
		if owner == "" {
			owner = "User" // Default for safety
		}
		
		currentOwner := owner
		if event.Kind == runtime.EventEnter {
			callee, _ := getParticipantAndMethod(event.Target)
			if callee != "self" {
				currentOwner = callee
			}
		}
		scopeOwner[event.ID] = currentOwner
		addParticipant(currentOwner)
	}

	return participantList
}


// getParticipantAndMethod extracts the participant and method from a target string like "instance.method".
func getParticipantAndMethod(target string) (participant, method string) {
	parts := strings.SplitN(target, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return target, ""
}
