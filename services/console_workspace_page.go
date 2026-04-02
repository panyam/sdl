package services

import (
	"fmt"
	"sync"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// ConsoleWorkspacePage implements WorkspacePage for CLI and test usage.
// It records all received updates (for test assertions) and optionally
// prints to stdout (for CLI output). Follows the lilbattle Base*Panel
// pattern where data-only implementations are used for both CLI and tests.
type ConsoleWorkspacePage struct {
	mu sync.Mutex

	// Verbose enables printing to stdout (CLI mode). Disabled for tests.
	Verbose bool

	// Recorded state — latest snapshot of what the presenter pushed
	ActiveSystem     string
	AvailableSystems []string
	Diagram          *SystemDiagram
	Generators       map[string]*protos.Generator
	Metrics          map[string]*protos.Metric
	FlowRates        map[string]float64
	FlowStrategy     string
	LogEntries       []LogEntry
}

// LogEntry records a single console log message.
type LogEntry struct {
	Level, Message, Source string
}

// NewConsoleWorkspacePage creates a console page, optionally verbose.
func NewConsoleWorkspacePage(verbose bool) *ConsoleWorkspacePage {
	return &ConsoleWorkspacePage{
		Verbose:    verbose,
		Generators: make(map[string]*protos.Generator),
		Metrics:    make(map[string]*protos.Metric),
	}
}

func (c *ConsoleWorkspacePage) OnSystemChanged(systemName string, availableSystems []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ActiveSystem = systemName
	c.AvailableSystems = availableSystems
	// Clear previous system's state — new system will push fresh generators/metrics
	c.Generators = make(map[string]*protos.Generator)
	c.Metrics = make(map[string]*protos.Metric)
	c.Diagram = nil
	c.FlowRates = nil
	if c.Verbose {
		fmt.Printf("System changed: %s\n", systemName)
	}
}

func (c *ConsoleWorkspacePage) OnAvailableSystemsChanged(systemNames []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AvailableSystems = systemNames
	if c.Verbose {
		fmt.Printf("Available systems: %v\n", systemNames)
	}
}

func (c *ConsoleWorkspacePage) UpdateDiagram(diagram *SystemDiagram) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Diagram = diagram
	if c.Verbose && diagram != nil {
		fmt.Printf("Diagram: %s (%d nodes, %d edges)\n", diagram.SystemName, len(diagram.Nodes), len(diagram.Edges))
	}
}

func (c *ConsoleWorkspacePage) UpdateGenerator(name string, generator *protos.Generator) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Generators[name] = generator
	if c.Verbose {
		status := "stopped"
		if generator.Enabled {
			status = "running"
		}
		fmt.Printf("Generator %s: rate=%.1f %s\n", name, generator.Rate, status)
	}
}

func (c *ConsoleWorkspacePage) RemoveGenerator(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Generators, name)
	if c.Verbose {
		fmt.Printf("Generator removed: %s\n", name)
	}
}

func (c *ConsoleWorkspacePage) UpdateMetric(name string, metric *protos.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Metrics[name] = metric
	if c.Verbose {
		fmt.Printf("Metric %s: type=%s aggregation=%s\n", name, metric.MetricType, metric.Aggregation)
	}
}

func (c *ConsoleWorkspacePage) RemoveMetric(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Metrics, name)
	if c.Verbose {
		fmt.Printf("Metric removed: %s\n", name)
	}
}

func (c *ConsoleWorkspacePage) UpdateFlowRates(rates map[string]float64, strategy string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FlowRates = rates
	c.FlowStrategy = strategy
	if c.Verbose {
		fmt.Printf("Flow rates (%s): %d entries\n", strategy, len(rates))
	}
}

func (c *ConsoleWorkspacePage) LogMessage(level string, message string, source string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LogEntries = append(c.LogEntries, LogEntry{level, message, source})
	if c.Verbose {
		fmt.Printf("[%s] %s\n", level, message)
	}
}
