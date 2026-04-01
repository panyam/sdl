//go:build js && wasm
// +build js,wasm

package main

import (
	"context"
	"log"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	wasmservices "github.com/panyam/sdl/gen/wasm/go/sdl/v1/services"
	"github.com/panyam/sdl/services"
)

// DevEnvPageForwarder implements DevEnvPageHandler by forwarding calls to the
// generated DevEnvPageClient (WASM browser channel). This is the browser-specific
// implementation of the page handler interface — following the constraint that
// service-layer code uses Go interfaces, and WASM-specific types live here.
type DevEnvPageForwarder struct {
	Client *wasmservices.DevEnvPageClient
}

func NewDevEnvPageForwarder(client *wasmservices.DevEnvPageClient) *DevEnvPageForwarder {
	return &DevEnvPageForwarder{Client: client}
}

func (f *DevEnvPageForwarder) OnSystemChanged(systemName string, availableSystems []string) {
	if f.Client == nil {
		return
	}
	_, err := f.Client.OnSystemChanged(context.Background(), &protos.DevEnvSystemChangedRequest{
		SystemName:       systemName,
		AvailableSystems: availableSystems,
	})
	if err != nil {
		log.Printf("DevEnvPageForwarder: OnSystemChanged error: %v", err)
	}
}

func (f *DevEnvPageForwarder) OnAvailableSystemsChanged(systemNames []string) {
	if f.Client == nil {
		return
	}
	_, err := f.Client.OnAvailableSystemsChanged(context.Background(), &protos.DevEnvAvailableSystemsRequest{
		SystemNames: systemNames,
	})
	if err != nil {
		log.Printf("DevEnvPageForwarder: OnAvailableSystemsChanged error: %v", err)
	}
}

func (f *DevEnvPageForwarder) UpdateDiagram(diagram *services.SystemDiagram) {
	if f.Client == nil {
		return
	}
	_, err := f.Client.UpdateDiagram(context.Background(), &protos.UpdateDiagramRequest{
		Diagram: services.ToProtoSystemDiagram(diagram),
	})
	if err != nil {
		log.Printf("DevEnvPageForwarder: UpdateDiagram error: %v", err)
	}
}

func (f *DevEnvPageForwarder) UpdateGenerator(name string, generator *protos.Generator) {
	if f.Client == nil {
		return
	}
	_, err := f.Client.UpdateGenerator(context.Background(), &protos.DevEnvUpdateGeneratorRequest{
		Name:      name,
		Generator: generator,
	})
	if err != nil {
		log.Printf("DevEnvPageForwarder: UpdateGenerator error: %v", err)
	}
}

func (f *DevEnvPageForwarder) RemoveGenerator(name string) {
	if f.Client == nil {
		return
	}
	_, err := f.Client.RemoveGenerator(context.Background(), &protos.DevEnvRemoveGeneratorRequest{
		Name: name,
	})
	if err != nil {
		log.Printf("DevEnvPageForwarder: RemoveGenerator error: %v", err)
	}
}

func (f *DevEnvPageForwarder) UpdateMetric(name string, metric *protos.Metric) {
	if f.Client == nil {
		return
	}
	_, err := f.Client.UpdateMetric(context.Background(), &protos.DevEnvUpdateMetricRequest{
		Name:   name,
		Metric: metric,
	})
	if err != nil {
		log.Printf("DevEnvPageForwarder: UpdateMetric error: %v", err)
	}
}

func (f *DevEnvPageForwarder) RemoveMetric(name string) {
	if f.Client == nil {
		return
	}
	_, err := f.Client.RemoveMetric(context.Background(), &protos.DevEnvRemoveMetricRequest{
		Name: name,
	})
	if err != nil {
		log.Printf("DevEnvPageForwarder: RemoveMetric error: %v", err)
	}
}

func (f *DevEnvPageForwarder) UpdateFlowRates(rates map[string]float64, strategy string) {
	if f.Client == nil {
		return
	}
	_, err := f.Client.UpdateFlowRates(context.Background(), &protos.UpdateFlowRatesRequest{
		Rates:    rates,
		Strategy: strategy,
	})
	if err != nil {
		log.Printf("DevEnvPageForwarder: UpdateFlowRates error: %v", err)
	}
}

func (f *DevEnvPageForwarder) LogMessage(level string, message string, source string) {
	if f.Client == nil {
		return
	}
	_, err := f.Client.LogMessage(context.Background(), &protos.LogMessageRequest{
		Level:   level,
		Message: message,
		Source:  source,
	})
	if err != nil {
		log.Printf("DevEnvPageForwarder: LogMessage error: %v", err)
	}
}
