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

	participantList := []string{"User"}
	participantSet := map[string]bool{"User": true}
	scopeParticipant := map[int]string{0: "User"}

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
		}
	}

	for _, p := range participantList {
		b.WriteString(fmt.Sprintf("  participant %s\n", p))
	}
	b.WriteString("\n")

	activeParticipantStack := []string{"User"}
	inParallelBlock := false

	for _, event := range trace.Events {
		if len(activeParticipantStack) == 0 {
			activeParticipantStack = append(activeParticipantStack, "User")
		}
		caller := activeParticipantStack[len(activeParticipantStack)-1]

		switch event.Kind {
		case runtime.EventEnter:
			callee, method := getParticipantAndMethod(event.Target)
			if callee == "self" {
				callee = caller
			}

			if inParallelBlock {
				b.WriteString(fmt.Sprintf("  %s->>%s: %s(%s)\n", caller, callee, method, strings.Join(event.Arguments, ", ")))
			} else {
				b.WriteString(fmt.Sprintf("  %s->>%s: %s(%s)\n", caller, callee, method, strings.Join(event.Arguments, ", ")))
				b.WriteString(fmt.Sprintf("  activate %s\n", callee))
				activeParticipantStack = append(activeParticipantStack, callee)
			}

		case runtime.EventExit:
			if !inParallelBlock {
				if len(activeParticipantStack) > 1 {
					participantToDeactivate := activeParticipantStack[len(activeParticipantStack)-1]
					activeParticipantStack = activeParticipantStack[:len(activeParticipantStack)-1]
					b.WriteString(fmt.Sprintf("  deactivate %s\n", participantToDeactivate))
				}
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
				b.WriteString("  end\n")
				inParallelBlock = false
			}
			aggregator := ""
			if len(event.Arguments) > 0 {
				aggregator = event.Arguments[0]
			}
			b.WriteString(fmt.Sprintf("  Note over %s: wait using %s\n", caller, aggregator))
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
