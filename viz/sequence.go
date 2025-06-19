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

	eventMap := make(map[int]*runtime.TraceEvent)
	childrenMap := make(map[int][]*runtime.TraceEvent)
	for _, event := range trace.Events {
		eventMap[event.ID] = event
		childrenMap[event.ParentID] = append(childrenMap[event.ParentID], event)
	}

	scopeOwner := g.calculateScopeOwners(trace, eventMap)
	participants := g.discoverParticipants(trace, scopeOwner)

	for _, p := range participants {
		b.WriteString(fmt.Sprintf("  participant %s\n", p))
	}
	b.WriteString("\n")

	processedEvents := make(map[int]bool)
	var renderBranch func(event *runtime.TraceEvent)

	renderBranch = func(event *runtime.TraceEvent) {
		if processedEvents[event.ID] {
			return
		}
		processedEvents[event.ID] = true

		caller := scopeOwner[event.ParentID]
		callee := scopeOwner[event.ID]

		switch event.Kind {
		case runtime.EventEnter:
			_, method := getParticipantAndMethod(event.Target())
			b.WriteString(fmt.Sprintf("  %s->>%s: %s(%s)\n", caller, callee, method, strings.Join(event.Arguments, ", ")))
			b.WriteString(fmt.Sprintf("  activate %s\n", callee))

			for _, child := range childrenMap[event.ID] {
				renderBranch(child)
			}

			b.WriteString(fmt.Sprintf("  deactivate %s\n", callee))

		case runtime.EventGo:
			loopCount := "N"
			if len(event.Arguments) > 0 {
				loopCount = event.Arguments[0]
			}
			b.WriteString(fmt.Sprintf("  loop %s times\n", loopCount))
			for _, child := range childrenMap[event.ID] {
				// The child's caller is the owner of the 'go' event's scope
				renderBranch(child)
			}
			b.WriteString("  end\n")

		case runtime.EventWait:
			aggregator := "wait"
			if len(event.Arguments) > 0 && event.Arguments[0] != "" {
				aggregator = fmt.Sprintf("wait using %s", event.Arguments[0])
			}
			b.WriteString(fmt.Sprintf("  Note over %s: %s\n", caller, aggregator))
		}
	}

	for _, event := range trace.Events {
		if !processedEvents[event.ID] {
			renderBranch(event)
		}
	}

	return b.String(), nil
}

func (g *MermaidSequenceGenerator) calculateScopeOwners(trace *runtime.TraceData, eventMap map[int]*runtime.TraceEvent) map[int]string {
	scopeOwner := make(map[int]string)
	scopeOwner[0] = "User"

	for _, event := range trace.Events {
		owner := scopeOwner[event.ParentID]
		if owner == "" {
			owner = "User"
		}

		currentOwner := owner
		if event.Kind == runtime.EventEnter {
			callee, _ := getParticipantAndMethod(event.Target())
			if callee != "self" {
				currentOwner = callee
			}
		}
		scopeOwner[event.ID] = currentOwner
	}
	return scopeOwner
}

func (g *MermaidSequenceGenerator) discoverParticipants(trace *runtime.TraceData, scopeOwner map[int]string) []string {
	participantList := []string{"User"}
	participantSet := map[string]bool{"User": true}

	for _, event := range trace.Events {
		owner := scopeOwner[event.ID]
		if owner != "" && !participantSet[owner] {
			participantList = append(participantList, owner)
			participantSet[owner] = true
		}
	}
	return participantList
}

func getParticipantAndMethod(target string) (participant, method string) {
	parts := strings.SplitN(target, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return target, ""
}
