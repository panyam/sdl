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

// BrowserDevEnvPage implements DevEnvPageHandler by forwarding calls to the
// generated DevEnvPageClient (WASM browser channel). Follows the lilbattle
// Browser*Panel convention where browser-specific WASM types live in
// cmd/wasm/browser.go and service-layer code uses Go interfaces.
type BrowserDevEnvPage struct {
	DevEnvPage *wasmservices.DevEnvPageClient
}

func NewBrowserDevEnvPage(devEnvPage *wasmservices.DevEnvPageClient) *BrowserDevEnvPage {
	return &BrowserDevEnvPage{DevEnvPage: devEnvPage}
}

func (f *BrowserDevEnvPage) OnSystemChanged(systemName string, availableSystems []string) {
	if f.DevEnvPage == nil {
		return
	}
	_, err := f.DevEnvPage.OnSystemChanged(context.Background(), &protos.DevEnvSystemChangedRequest{
		SystemName:       systemName,
		AvailableSystems: availableSystems,
	})
	if err != nil {
		log.Printf("BrowserDevEnvPage: OnSystemChanged error: %v", err)
	}
}

func (f *BrowserDevEnvPage) OnAvailableSystemsChanged(systemNames []string) {
	if f.DevEnvPage == nil {
		return
	}
	_, err := f.DevEnvPage.OnAvailableSystemsChanged(context.Background(), &protos.DevEnvAvailableSystemsRequest{
		SystemNames: systemNames,
	})
	if err != nil {
		log.Printf("BrowserDevEnvPage: OnAvailableSystemsChanged error: %v", err)
	}
}

func (f *BrowserDevEnvPage) UpdateDiagram(diagram *services.SystemDiagram) {
	if f.DevEnvPage == nil {
		return
	}
	_, err := f.DevEnvPage.UpdateDiagram(context.Background(), &protos.UpdateDiagramRequest{
		Diagram: services.ToProtoSystemDiagram(diagram),
	})
	if err != nil {
		log.Printf("BrowserDevEnvPage: UpdateDiagram error: %v", err)
	}
}

func (f *BrowserDevEnvPage) UpdateGenerator(name string, generator *protos.Generator) {
	if f.DevEnvPage == nil {
		return
	}
	_, err := f.DevEnvPage.UpdateGenerator(context.Background(), &protos.DevEnvUpdateGeneratorRequest{
		Name:      name,
		Generator: generator,
	})
	if err != nil {
		log.Printf("BrowserDevEnvPage: UpdateGenerator error: %v", err)
	}
}

func (f *BrowserDevEnvPage) RemoveGenerator(name string) {
	if f.DevEnvPage == nil {
		return
	}
	_, err := f.DevEnvPage.RemoveGenerator(context.Background(), &protos.DevEnvRemoveGeneratorRequest{
		Name: name,
	})
	if err != nil {
		log.Printf("BrowserDevEnvPage: RemoveGenerator error: %v", err)
	}
}

func (f *BrowserDevEnvPage) UpdateMetric(name string, metric *protos.Metric) {
	if f.DevEnvPage == nil {
		return
	}
	_, err := f.DevEnvPage.UpdateMetric(context.Background(), &protos.DevEnvUpdateMetricRequest{
		Name:   name,
		Metric: metric,
	})
	if err != nil {
		log.Printf("BrowserDevEnvPage: UpdateMetric error: %v", err)
	}
}

func (f *BrowserDevEnvPage) RemoveMetric(name string) {
	if f.DevEnvPage == nil {
		return
	}
	_, err := f.DevEnvPage.RemoveMetric(context.Background(), &protos.DevEnvRemoveMetricRequest{
		Name: name,
	})
	if err != nil {
		log.Printf("BrowserDevEnvPage: RemoveMetric error: %v", err)
	}
}

func (f *BrowserDevEnvPage) UpdateFlowRates(rates map[string]float64, strategy string) {
	if f.DevEnvPage == nil {
		return
	}
	_, err := f.DevEnvPage.UpdateFlowRates(context.Background(), &protos.UpdateFlowRatesRequest{
		Rates:    rates,
		Strategy: strategy,
	})
	if err != nil {
		log.Printf("BrowserDevEnvPage: UpdateFlowRates error: %v", err)
	}
}

func (f *BrowserDevEnvPage) LogMessage(level string, message string, source string) {
	if f.DevEnvPage == nil {
		return
	}
	_, err := f.DevEnvPage.LogMessage(context.Background(), &protos.LogMessageRequest{
		Level:   level,
		Message: message,
		Source:  source,
	})
	if err != nil {
		log.Printf("BrowserDevEnvPage: LogMessage error: %v", err)
	}
}
