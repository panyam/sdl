// sdl/dsl/driver.go
package dsl

import (
	"fmt"
	"strconv" // Need for parseLiteralValue

	"github.com/panyam/leetcoach/sdl/components" // Needed for component constructors & internal funcs
	"github.com/panyam/leetcoach/sdl/core"
)

// --- Analysis Result Wrapper (Reinstated) ---
// AnalysisResultWrapper holds the raw outcome and calculated metrics for an analyze block.
type AnalysisResultWrapper struct {
	Name              string                      // Name from the analyze block
	Outcome           interface{}                 // The raw *core.Outcomes[V] result
	Metrics           map[core.MetricType]float64 // Use core.MetricType as key
	Error             error                       // Any error during analysis evaluation
	Skipped           bool                        // True if analysis couldn't run (e.g., target error)
	Messages          []string                    // Log messages during analysis
	AnalysisPerformed bool                        // Added field similar to core.AnalysisResult
}

// Helper to add a log message
func (arw *AnalysisResultWrapper) addMsg(format string, args ...interface{}) {
	arw.Messages = append(arw.Messages, fmt.Sprintf(format, args...))
}

// componentConstructor defines the type for functions that create component instances.
type componentConstructor func(name string, params map[string]interface{}) (interface{}, error)

// registerBuiltinComponents registers Go component constructors in the interpreter's global environment.
func registerBuiltinComponents(interpreter *Interpreter) {
	// The value stored is the constructor function itself.
	// The key is the name used in the DSL (e.g., "Disk").
	interpreter.Env().Set("Disk", componentConstructor(func(name string, params map[string]interface{}) (interface{}, error) {
		profile := "SSD" // Default
		if p, ok := params["ProfileName"].(string); ok {
			profile = p
		}
		return components.NewDisk(profile), nil
	}))
	interpreter.Env().Set("Cache", componentConstructor(func(name string, params map[string]interface{}) (interface{}, error) {
		cache := components.NewCache()
		// TODO: Apply params map to cache fields if necessary
		return cache, nil
	}))
	interpreter.Env().Set("ResourcePool", componentConstructor(func(name string, params map[string]interface{}) (interface{}, error) {
		// TODO: Extract Size, ArrivalRate, AvgHoldTime from params map with type checks/defaults
		sizeVal, _ := params["Size"].(int64)
		lambdaVal, _ := params["ArrivalRate"].(float64)
		holdTimeVal, _ := params["AvgHoldTime"].(core.Duration)
		// Add basic validation/defaults
		if sizeVal <= 0 {
			sizeVal = 1
		}
		if lambdaVal <= 0 {
			lambdaVal = 1e-9
		}
		if holdTimeVal <= 0 {
			holdTimeVal = 1e-9
		}
		return components.NewResourcePool(name, uint(sizeVal), lambdaVal, holdTimeVal), nil
	}))
	// ... register other components ...
}

// registerCoreInternalFunctions populates the registry with functions
// needed by the declarative components' ASTs.
func registerCoreInternalFunctions(interpreter *Interpreter) {
	// Disk Functions
	interpreter.RegisterInternalFunc("GetDiskReadProfile", func(i *Interpreter, args []interface{}) (interface{}, error) {
		profileName := "SSD"
		if len(args) > 0 {
			nameOutcome, ok := args[0].(*core.Outcomes[string])
			if ok {
				detName, okName := nameOutcome.GetValue()
				if okName {
					profileName = detName
				}
			}
		}
		tempDisk := components.NewDisk(profileName)
		return tempDisk.Read(), nil
	})
	interpreter.RegisterInternalFunc("GetDiskWriteProfile", func(i *Interpreter, args []interface{}) (interface{}, error) {
		profileName := "SSD"
		if len(args) > 0 {
			nameOutcome, ok := args[0].(*core.Outcomes[string])
			if ok {
				detName, okName := nameOutcome.GetValue()
				if okName {
					profileName = detName
				}
			}
		}
		tempDisk := components.NewDisk(profileName)
		return tempDisk.Write(), nil
	})
	interpreter.RegisterInternalFunc("GetRecordProcessingTime", func(i *Interpreter, args []interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 arg for GetRecordProcessingTime")
		}
		return args[0], nil
	})
	interpreter.RegisterInternalFunc("ScaleLatency", func(i *Interpreter, args []interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 args for ScaleLatency (outcome, factor)")
		}
		inputOutcomeRaw := args[0]
		factorOutcomeRaw := args[1]
		factorOutcome, ok := factorOutcomeRaw.(*core.Outcomes[float64])
		if !ok {
			return nil, fmt.Errorf("ScaleLatency factor must be float64 outcome, got %T", factorOutcomeRaw)
		}
		factor, ok := factorOutcome.GetValue()
		if !ok {
			return nil, fmt.Errorf("ScaleLatency factor must be deterministic")
		}
		switch inputOutcome := inputOutcomeRaw.(type) {
		case *core.Outcomes[core.AccessResult]:
			mapper := func(ar core.AccessResult) core.AccessResult {
				if ar.Success {
					ar.Latency *= factor
				}
				return ar
			}
			return core.Map(inputOutcome, mapper), nil
		case *core.Outcomes[core.Duration]:
			mapper := func(d core.Duration) core.Duration { return d * factor }
			return core.Map(inputOutcome, mapper), nil
		default:
			return nil, fmt.Errorf("unsupported outcome type %T for ScaleLatency", inputOutcomeRaw)
		}
	})
	// Add other internal funcs needed...
}

// calculateAndStoreMetrics populates the Metrics map within AnalysisResultWrapper.
func calculateAndStoreMetrics(resultWrapper *AnalysisResultWrapper) {
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
		resultWrapper.addMsg("Cannot calculate metrics for outcome type %T", o)
		resultWrapper.AnalysisPerformed = false // Cannot calculate metrics
	}
}

// parseLiteralValue converts a LiteralExpr value string to a basic Go type.
func parseLiteralValue(lit *LiteralExpr) (interface{}, error) {
	switch lit.Kind {
	case "STRING":
		return lit.Value, nil
	case "INT":
		return strconv.ParseInt(lit.Value, 10, 64)
	case "FLOAT":
		return strconv.ParseFloat(lit.Value, 64)
	case "BOOL":
		return strconv.ParseBool(lit.Value)
	// TODO: case "DURATION":
	default:
		return nil, fmt.Errorf("cannot parse literal kind %s yet", lit.Kind)
	}
}

// RunDSL executes the logic defined in an AST, focusing on System and Analyze blocks.
// Returns a map of analysis names to their corresponding AnalysisResultWrapper.
func RunDSL(astRoot Node /* options? */) (map[string]*AnalysisResultWrapper, error) {
	analysisResults := make(map[string]*AnalysisResultWrapper)

	// --- Basic Setup ---
	systemDecl, ok := astRoot.(*SystemDecl)
	if !ok {
		return analysisResults, fmt.Errorf("RunDSL currently expects *SystemDecl as root, got %T", astRoot)
	}

	interpreter := NewInterpreter(15)          // Default max buckets
	registerBuiltinComponents(interpreter)     // Register component constructors
	registerCoreInternalFunctions(interpreter) // Register helpers needed by decl components

	// --- Process System ---
	systemEnv := NewEnclosedEnvironment(interpreter.env)
	interpreter.env = systemEnv // Switch to system env for instantiation

	// Instantiate components
	for _, node := range systemDecl.Body {
		instanceDecl, isInstance := node.(*InstanceDecl)
		if !isInstance {
			continue
		}
		// ... (Instantiation logic remains the same) ...
		constructorRaw, typeFound := systemEnv.Get(instanceDecl.ComponentType)
		if !typeFound {
			// Store error and continue? Or return immediately? Return for now.
			err := fmt.Errorf("component type '%s' not found for instance '%s'", instanceDecl.ComponentType, instanceDecl.Name)
			// We don't have a result wrapper yet to store this error in.
			return analysisResults, err
		}
		constructor, isConstructor := constructorRaw.(componentConstructor)
		if !isConstructor {
			err := fmt.Errorf("type '%s' is not a constructible component", instanceDecl.ComponentType)
			return analysisResults, err
		}
		params := make(map[string]interface{})
		for _, p := range instanceDecl.Params {
			if lit, okLit := p.Value.(*LiteralExpr); okLit {
				parsedVal, err := parseLiteralValue(lit)
				if err != nil {
					return analysisResults, fmt.Errorf("error parsing param '%s' for instance '%s': %w", p.Name, instanceDecl.Name, err)
				}
				params[p.Name] = parsedVal
			} else {
				return analysisResults, fmt.Errorf("parameter '%s' for instance '%s' must be a literal (for now)", p.Name, instanceDecl.Name)
			}
		}
		instanceObj, err := constructor(instanceDecl.Name, params)
		if err != nil {
			return analysisResults, fmt.Errorf("error constructing instance '%s': %w", instanceDecl.Name, err)
		}
		systemEnv.Set(instanceDecl.Name, instanceObj)
	} // End instantiation loop

	// Run Analyze blocks
	for _, node := range systemDecl.Body {
		analyzeDecl, isAnalyze := node.(*AnalyzeDecl)
		if !isAnalyze {
			continue
		}

		analysisName := analyzeDecl.Name
		// Create the result wrapper
		resultWrapper := &AnalysisResultWrapper{
			Name:    analysisName,
			Metrics: make(map[core.MetricType]float64), // Use correct key type
		}
		analysisResults[analysisName] = resultWrapper

		resultWrapper.addMsg("Starting analysis '%s'...", analysisName)
		interpreter.ClearStack() // Ensure clean stack

		// Evaluate the target expression
		_, evalErr := interpreter.Eval(analyzeDecl.Target)

		if evalErr != nil {
			resultWrapper.addMsg("Evaluation error: %v", evalErr)
			resultWrapper.Error = evalErr
			resultWrapper.Skipped = true
			resultWrapper.AnalysisPerformed = false // Mark as not performed due to eval error
			continue
		}

		// Get final outcome from stack
		finalOutcome, stackErr := interpreter.GetFinalResult()
		if stackErr != nil {
			resultWrapper.addMsg("Stack error after evaluation: %v", stackErr)
			resultWrapper.Error = stackErr
			resultWrapper.Skipped = true
			resultWrapper.AnalysisPerformed = false // Mark as not performed due to stack error
			continue
		}
		resultWrapper.Outcome = finalOutcome // Store the raw *core.Outcomes[V] object

		// Calculate Metrics
		calculateAndStoreMetrics(resultWrapper) // Pass the wrapper
		// Log some results for feedback
		if resultWrapper.AnalysisPerformed {
			resultWrapper.addMsg("Analysis complete. Availability: %.6f, P99: %.6fs",
				resultWrapper.Metrics[core.AvailabilityMetric], resultWrapper.Metrics[core.P99LatencyMetric])
		} else {
			resultWrapper.addMsg("Analysis finished, but metrics could not be calculated (Outcome type: %T).", resultWrapper.Outcome)
		}

	} // End analyze loop

	interpreter.env = systemEnv.outer // Restore global env
	return analysisResults, nil
}
