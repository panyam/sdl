package decl

import (
	"fmt"

	"github.com/panyam/sdl/core"
)

// --- Analysis Result Wrapper (Reinstated) ---
// AnalysisResultWrapper holds the raw outcome and calculated metrics for an analyze block.
type AnalysisResultWrapper struct {
	Name              string                      // Name from the analyze block
	Outcome           any                         // The raw *core.Outcomes[V] result
	Metrics           map[core.MetricType]float64 // Use core.MetricType as key
	Error             error                       // Any error during analysis evaluation
	Skipped           bool                        // True if analysis couldn't run (e.g., target error)
	Messages          []string                    // Log messages during analysis
	AnalysisPerformed bool                        // Added field similar to core.AnalysisResult
}

// Helper to add a log message
func (arw *AnalysisResultWrapper) AddMsg(format string, args ...any) {
	arw.Messages = append(arw.Messages, fmt.Sprintf(format, args...))
}

// CalculateAndStoreMetrics populates the Metrics map within AnalysisResultWrapper.
func CalculateAndStoreMetrics(resultWrapper *AnalysisResultWrapper) {
	if resultWrapper.Outcome == nil {
		resultWrapper.AnalysisPerformed = false
		return
	}
	// Use type switch to check for known Metricable outcome types
	switch o := resultWrapper.Outcome.(type) {
	case *core.Outcomes[core.AccessResult]:
		resultWrapper.Metrics[core.AvailabilityMetric] = core.Availability(o)
		resultWrapper.Metrics[core.MeanLatencyMetric] = core.MeanLatency(o)
		resultWrapper.Metrics[core.P50LatencyMetric] = core.PercentileLatency(o, 0.50)
		resultWrapper.Metrics[core.P99LatencyMetric] = core.PercentileLatency(o, 0.99)
		resultWrapper.Metrics[core.P999LatencyMetric] = core.PercentileLatency(o, 0.999)
		resultWrapper.AnalysisPerformed = (o != nil && o.Len() > 0)
	case *core.Outcomes[core.RangedResult]:
		resultWrapper.Metrics[core.AvailabilityMetric] = core.Availability(o)
		resultWrapper.Metrics[core.MeanLatencyMetric] = core.MeanLatency(o)
		resultWrapper.Metrics[core.P50LatencyMetric] = core.PercentileLatency(o, 0.50)
		resultWrapper.Metrics[core.P99LatencyMetric] = core.PercentileLatency(o, 0.99)
		resultWrapper.Metrics[core.P999LatencyMetric] = core.PercentileLatency(o, 0.999)
		resultWrapper.AnalysisPerformed = (o != nil && o.Len() > 0)
	// Add other Metricable types here
	default:
		resultWrapper.AddMsg("Cannot calculate metrics for outcome type %T", o)
		resultWrapper.AnalysisPerformed = false // Cannot calculate metrics
	}
}
