package services

import (
	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
)

// WorkspacePage is the Go interface that mirrors the WorkspacePage proto service.
// The WorkspacePresenter pushes typed updates through this interface when state changes.
// Implementations: BrowserWorkspacePage (WASM), ConsoleWorkspacePage (CLI), mock (tests).
type WorkspacePage interface {
	// System panel: active system has changed
	OnSystemChanged(systemName string, availableSystems []string)

	// System panel: available systems list updated (e.g. after loading a new file)
	OnAvailableSystemsChanged(systemNames []string)

	// Diagram panel: system topology updated
	UpdateDiagram(diagram *SystemDiagram)

	// Generator panel: upsert a generator by name
	UpdateGenerator(name string, generator *protos.Generator)

	// Generator panel: remove a generator by name
	RemoveGenerator(name string)

	// Metric panel: upsert a metric by name
	UpdateMetric(name string, metric *protos.Metric)

	// Metric panel: remove a metric by name
	RemoveMetric(name string)

	// Flow panel: flow rates updated
	UpdateFlowRates(rates map[string]float64, strategy string)

	// Console panel: log a message
	LogMessage(level string, message string, source string)
}
