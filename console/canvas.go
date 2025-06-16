package console

import (
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/panyam/sdl/components"
	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader"
	"github.com/panyam/sdl/runtime"
	"github.com/panyam/sdl/types"
	"github.com/panyam/sdl/viz"
)

// Canvas provides a stateful, interactive environment for loading,
// modifying, and analyzing SDL models. It acts as the core engine
// for both scripted tests and future interactive tools like a REPL.
type Canvas struct {
	loader              *loader.Loader
	runtime             *runtime.Runtime
	activeFile          *loader.FileStatus
	activeSystem        *runtime.SystemInstance
	sessionVars         map[string]any
	loadedFiles         map[string]*loader.FileStatus
	genManager          *generatorManager
	measManager         *measurementManager
	systemParameters    map[string]interface{} // Track parameter changes
	tsdb                *DuckDBTimeSeriesStore // Time-series database for measurements
	measurementTracer   *MeasurementTracer     // Current measurement tracer
	currentFlowContext  *runtime.FlowContext   // Current flow state (applied/active)
	proposedFlowContext *runtime.FlowContext   // Proposed flow state (being evaluated)
}

// NewCanvas creates a new interactive canvas session.
func NewCanvas() *Canvas {
	l := loader.NewLoader(nil, nil, 10)
	r := runtime.NewRuntime(l)
	return &Canvas{
		loader:           l,
		runtime:          r,
		sessionVars:      make(map[string]any),
		loadedFiles:      make(map[string]*loader.FileStatus),
		systemParameters: make(map[string]interface{}),
	}
}

// Load parses, validates, and loads an SDL file and all its imports
// into the session. If the file is already loaded, it will be re-loaded
// and re-validated to ensure freshness.
func (c *Canvas) Load(filePath string) error {
	// TODO: Handle hot-reloading. For now, we load once.
	fileStatus, err := c.loader.LoadFile(filePath, "", 0)
	if err != nil {
		return fmt.Errorf("error loading file '%s': %w", filePath, err)
	}

	if !c.loader.Validate(fileStatus) {
		// Collect errors into a single error object
		// For now, just return a generic error.
		fileStatus.PrintErrors()
		return fmt.Errorf("validation failed for file '%s'", filePath)
	}

	c.loadedFiles[fileStatus.FullPath] = fileStatus
	c.activeFile = fileStatus

	// Invalidate flow contexts since file changed
	c.invalidateFlowContexts()

	return nil
}

// Use sets the active system for subsequent commands.
func (c *Canvas) Use(systemName string) error {
	if c.activeFile == nil {
		return fmt.Errorf("no file loaded. Call Load() before Use()")
	}

	fileInstance := c.runtime.LoadFile(c.activeFile.FullPath)
	if fileInstance == nil {
		return fmt.Errorf("could not get file instance for '%s'", c.activeFile.FullPath)
	}

	system := fileInstance.NewSystem(systemName)
	if system == nil {
		return fmt.Errorf("system '%s' not found in file '%s'", systemName, c.activeFile.FullPath)
	}
	c.activeSystem = system

	// Initialize the system
	var totalSimTime runtime.Duration
	env := fileInstance.Env()
	eval := runtime.NewSimpleEval(fileInstance, nil)
	eval.EvalInitSystem(system, env, &totalSimTime)
	if eval.HasErrors() {
		// eval.PrintErrors()
		return fmt.Errorf("errors found while initializing system '%s'", systemName)
	}

	// Crucially, assign the populated environment back to the active system
	c.activeSystem.Env = env

	// Initialize flow contexts for the new system
	c.initializeFlowContexts()

	return nil
}

// Set modifies a component parameter at runtime.
// The path is a dot-separated string, e.g., "app.cache.HitRate".
func (c *Canvas) Set(path string, value any) error {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid path for Set: '%s'. Must be at least <instance>.<field>", path)
	}

	if c.activeSystem == nil || c.activeSystem.Env == nil {
		return fmt.Errorf("no active system. Call Use() before Set()")
	}

	instanceName := parts[0]
	instanceVal, ok := c.activeSystem.Env.Get(instanceName)
	if !ok {
		return fmt.Errorf("instance '%s' not found in active system", instanceName)
	}

	// Start with the top-level component instance
	var currentComp *runtime.ComponentInstance
	if comp, ok := instanceVal.Value.(*runtime.ComponentInstance); ok {
		currentComp = comp
	} else {
		return fmt.Errorf("'%s' is not a component instance", instanceName)
	}

	// Traverse the path parts[1:len(parts)-1] to find the target component
	for i := 1; i < len(parts)-1; i++ {
		depName := parts[i]
		// Debug removed for cleaner output
		depVal, ok := currentComp.Get(depName) // Get the dependency by name
		if !ok || depVal.IsNil() {
			// Debug removed
			return fmt.Errorf("dependency '%s' not found in component '%s'", depName, currentComp.ComponentDecl.Name.Value)
		}
		// Debug removed

		if nextComp, ok := depVal.Value.(*runtime.ComponentInstance); ok {
			currentComp = nextComp
		} else {
			return fmt.Errorf("dependency '%s' in '%s' is not a component instance", depName, currentComp.ComponentDecl.Name.Value)
		}
	}

	// Now currentComp is the component on which we need to set the final parameter
	finalParamName := parts[len(parts)-1]

	// Debug removed

	if currentComp.IsNative {
		// For native components, use reflection to set the field on the underlying Go struct.
		err := c.setField(currentComp.NativeInstance, []string{finalParamName}, value)
		if err == nil {
			// Track parameter change for state persistence
			c.systemParameters[path] = value
		}
		return err
	} else {
		// For user-defined components, set the parameter in their runtime environment.
		var newValue decl.Value
		var err error
		switch v := value.(type) {
		case int:
			newValue, err = decl.NewValue(decl.IntType, int64(v))
		case int64:
			newValue, err = decl.NewValue(decl.IntType, v)
		case float64:
			newValue, err = decl.NewValue(decl.FloatType, v)
		case bool:
			newValue, err = decl.NewValue(decl.BoolType, v)
		case string:
			newValue, err = decl.NewValue(decl.StrType, v)
		default:
			err = fmt.Errorf("unsupported value type for Set on user-defined component: %T", value)
		}

		if err != nil {
			return err
		}
		err = currentComp.Set(finalParamName, newValue)
		if err == nil {
			// Track parameter change for state persistence
			c.systemParameters[path] = value
		}
		return err
	}
}

func (c *Canvas) setField(obj any, path []string, value any) error {
	objVal := reflect.ValueOf(obj)

	// Dereference pointers until we get to a struct or the end of the chain
	for objVal.Kind() == reflect.Ptr {
		if objVal.IsNil() {
			return fmt.Errorf("cannot set field on nil pointer in path %v", path)
		}
		objVal = objVal.Elem()
	}

	// If the current object is a native component wrapper, get the actual wrapped component.
	if objVal.Kind() == reflect.Struct {
		wrappedField := objVal.FieldByName("Wrapped")
		if wrappedField.IsValid() {
			objVal = wrappedField
			// Again, dereference pointers of the wrapped value
			for objVal.Kind() == reflect.Ptr {
				if objVal.IsNil() {
					return fmt.Errorf("the 'Wrapped' field is a nil pointer")
				}
				objVal = objVal.Elem()
			}
		}
	}

	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty when setting field")
	}

	fieldName := path[0]
	field := objVal.FieldByName(fieldName)

	if !field.IsValid() {
		return fmt.Errorf("field '%s' not found in struct type %s", fieldName, objVal.Type())
	}

	if len(path) == 1 {
		// This is the final field to set
		if !field.CanSet() {
			return fmt.Errorf("field '%s' cannot be set", fieldName)
		}
		valToSet := reflect.ValueOf(value)
		if valToSet.Type().ConvertibleTo(field.Type()) {
			field.Set(valToSet.Convert(field.Type()))
			return nil
		}
		return fmt.Errorf("cannot set field '%s': value of type %T is not assignable to field of type %s", fieldName, value, field.Type())
	}

	// We need to go deeper
	return c.setField(field.Interface(), path[1:], value)
}

// Run executes a simulation and stores the results in a session variable.
// target is the method to call, e.g., "app.Redirect".
// If measurements are registered, automatically injects MeasurementTracer for data collection.
func (c *Canvas) Run(varName string, target string, opts ...RunOption) error {
	cfg := &RunConfig{
		Runs:    1000, // Default runs
		Workers: 50,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	parts := strings.Split(target, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid target for Run: '%s'. Must be <instance>.<method>", target)
	}
	instanceName, methodName := parts[0], parts[1]

	batchSize := cfg.Runs / 100
	if batchSize == 0 {
		batchSize = 1
	}
	numBatches := (cfg.Runs + batchSize - 1) / batchSize

	allResults := make([]types.RunResult, 0, cfg.Runs)

	onBatch := func(batch int, batchVals []runtime.Value) {
		ts := time.Now().UnixMilli() // Use wall time for timestamp for now
		for _, val := range batchVals {
			allResults = append(allResults, types.RunResult{
				Timestamp:   ts,
				Latency:     val.Time * 1000, // to ms
				ResultValue: val.String(),
				IsError:     false, // TODO
			})
		}
	}

	// Check if measurements are registered - if so, use tracer-enabled execution
	if c.HasMeasurements() {
		err := c.runWithMeasurementTracer(instanceName, methodName, numBatches, batchSize, cfg.Workers, onBatch)
		if err != nil {
			return fmt.Errorf("measurement-enabled run failed: %w", err)
		}
	} else {
		// Standard execution without measurements
		runtime.RunCallInBatches(c.activeSystem, instanceName, methodName, numBatches, batchSize, cfg.Workers, onBatch)
	}

	c.sessionVars[varName] = allResults
	return nil
}

// runWithMeasurementTracer executes simulations with measurement tracing enabled
func (c *Canvas) runWithMeasurementTracer(instanceName, methodName string, numBatches, batchSize, numWorkers int, onBatch func(int, []runtime.Value)) error {
	// Generate unique run ID for this simulation
	runID := fmt.Sprintf("run_%d", time.Now().UnixMilli())

	// Ensure measurement tracer is initialized
	tracer, err := c.GetMeasurementTracer("./data")
	if err != nil {
		return fmt.Errorf("failed to initialize measurement tracer: %w", err)
	}

	// Set the run ID for this session
	tracer.SetRunID(runID)

	// Create a custom simulation runner with tracer support
	fi := c.activeSystem.File
	se := runtime.NewSimpleEval(fi, tracer.ExecutionTracer)

	// Use the existing system environment
	var env *runtime.Env[runtime.Value]
	if c.activeSystem.Env != nil {
		env = c.activeSystem.Env
	} else {
		env = fi.Env()
	}

	// Run the simulation in batches with the tracer
	for batch := 0; batch < numBatches; batch++ {
		batchVals := make([]runtime.Value, 0, batchSize)

		for i := 0; i < batchSize; i++ {
			// Create call expression for instance.method
			var runLatency runtime.Duration
			ce := &runtime.CallExpr{
				Function: &runtime.MemberAccessExpr{
					Receiver: &runtime.IdentifierExpr{Value: instanceName},
					Member:   &runtime.IdentifierExpr{Value: methodName},
				},
			}

			// Execute single call with tracer
			result, _ := se.Eval(ce, env, &runLatency)
			result.Time = runLatency // Capture the latency of this single run

			if se.HasErrors() {
				return fmt.Errorf("simulation error in batch %d: %v", batch, se.Errors)
			}
			batchVals = append(batchVals, result)
		}

		// Process batch results
		onBatch(batch, batchVals)
	}

	// Post-process the tracer events to extract measurements
	err = c.processTracerEvents(tracer)
	if err != nil {
		return fmt.Errorf("failed to process tracer events: %w", err)
	}

	return nil
}

// processTracerEvents extracts measurements from tracer events and stores them in the database
func (c *Canvas) processTracerEvents(tracer *MeasurementTracer) error {
	events := tracer.ExecutionTracer.Events
	measurements := tracer.GetMeasurements()

	// Match Enter/Exit events to extract measurements
	enterStack := make(map[int]*runtime.TraceEvent) // eventID -> Enter event

	for _, event := range events {
		if event.Kind == runtime.EventEnter {
			enterStack[event.ID] = event
		} else if event.Kind == runtime.EventExit && len(enterStack) > 0 {
			// Find the corresponding Enter event by walking back through the stack
			var enterEvent *runtime.TraceEvent
			for id, enter := range enterStack {
				// Simple approach: match the most recent Enter event that hasn't been matched
				// This assumes proper nesting of Enter/Exit events
				enterEvent = enter
				delete(enterStack, id)
				break
			}

			if enterEvent != nil {
				// Check if this target is being measured
				target := enterEvent.Target
				if measurement, exists := measurements[target]; exists && measurement.Enabled {
					// Extract metric and store in database
					err := c.storeMeasurementFromEvent(tracer, measurement, enterEvent, event)
					if err != nil {
						fmt.Printf("Warning: Failed to store measurement for %s: %v\n", target, err)
					}
				}
			}
		}
	}

	return nil
}

// storeMeasurementFromEvent extracts measurement data from tracer events and stores in database
func (c *Canvas) storeMeasurementFromEvent(tracer *MeasurementTracer, measurement *MeasurementConfig, enterEvent, exitEvent *runtime.TraceEvent) error {
	// Use current wall clock time instead of simulation timestamp
	timestampNs := time.Now().UnixNano()

	var metricValue float64
	switch measurement.MetricType {
	case "latency":
		metricValue = float64(exitEvent.Duration)
	case "throughput":
		metricValue = 1.0 // Each call represents one unit of throughput
	case "errors":
		// TODO: Check if there was an error in the execution
		metricValue = 0.0 // No error detection yet
	default:
		metricValue = float64(exitEvent.Duration)
	}

	point := TracePoint{
		Timestamp:   timestampNs,
		Target:      measurement.Target,
		Duration:    metricValue,
		ReturnValue: "", // TODO: Extract return value if needed
		Error:       "", // TODO: Extract error if needed
		Args:        enterEvent.Arguments,
		RunID:       tracer.runID,
	}

	return tracer.tsdb.Insert(point)
}

// Plot generates a visualization from data stored in session variables.
func (c *Canvas) Plot(opts ...PlotOption) error {
	cfg := &PlotConfig{
		YAxis: "latency",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.OutputFile == "" {
		return fmt.Errorf("output file must be specified for plot")
	}

	var vizSeries []viz.DataSeries
	for _, seriesInfo := range cfg.Series {
		data, ok := c.sessionVars[seriesInfo.From]
		if !ok {
			return fmt.Errorf("session variable '%s' not found for plotting", seriesInfo.From)
		}

		results, ok := data.([]types.RunResult)
		if !ok {
			return fmt.Errorf("session variable '%s' is not of type []RunResult", seriesInfo.From)
		}
		// TODO: This logic is duplicated from plot.go, can be refactored
		buckets := make(map[int64][]float64)
		for _, r := range results {
			if !r.IsError {
				bucketTime := (r.Timestamp / 1000) * 1000
				buckets[bucketTime] = append(buckets[bucketTime], r.Latency)
			}
		}

		var points []viz.DataPoint
		for ts, latencies := range buckets {
			sort.Float64s(latencies)
			if len(latencies) > 0 {
				points = append(points, viz.DataPoint{
					X: ts,
					Y: latencies[int(float64(len(latencies))*0.9)], // P90
				})
			}
		}
		sort.Slice(points, func(i, j int) bool { return points[i].X < points[j].X })
		vizSeries = append(vizSeries, viz.DataSeries{Name: seriesInfo.Name, Points: points})
	}

	plotter := viz.NewSVGPlotter(viz.DefaultPlotConfig())
	svg, err := plotter.Generate(vizSeries, cfg.Title, "Time", "P90 Latency (ms)")
	if err != nil {
		return err
	}
	return os.WriteFile(cfg.OutputFile, []byte(svg), 0644)
}

// --- Option types for Run/Plot ---

type RunConfig struct {
	Runs    int
	Workers int
}
type RunOption func(*RunConfig)

func WithRuns(n int) RunOption {
	return func(cfg *RunConfig) {
		cfg.Runs = n
	}
}

type SeriesInfo struct {
	Name string
	From string // Variable name from sessionVars
}

type PlotConfig struct {
	Title      string
	YAxis      string
	OutputFile string
	Series     []SeriesInfo
}

type PlotOption func(*PlotConfig)

func WithSeries(name, fromVar string) PlotOption {
	return func(cfg *PlotConfig) {
		cfg.Series = append(cfg.Series, SeriesInfo{Name: name, From: fromVar})
	}
}

func WithOutput(path string) PlotOption {
	return func(cfg *PlotConfig) {
		cfg.OutputFile = path
	}
}

// SystemDiagram represents the topology of a system using viz package types
type SystemDiagram struct {
	SystemName string     `json:"systemName"`
	Nodes      []viz.Node `json:"nodes"`
	Edges      []viz.Edge `json:"edges"`
}

// GetAvailableSystemNames returns all system names from loaded SDL files
func (c *Canvas) GetAvailableSystemNames() []string {
	var systemNames []string

	// Iterate through all loaded files
	for _, fileStatus := range c.loadedFiles {
		if fileStatus == nil || fileStatus.FileDecl == nil {
			continue
		}

		// Get systems from this file
		systems, err := fileStatus.FileDecl.GetSystems()
		if err != nil {
			continue // Skip files with errors
		}

		// Add system names to our list
		for _, system := range systems {
			systemNames = append(systemNames, system.Name.Value)
		}
	}

	return systemNames
}

// GetSystemDiagram returns the topology of the currently active system
func (c *Canvas) GetSystemDiagram() (*SystemDiagram, error) {
	if c.activeSystem == nil {
		return nil, fmt.Errorf("no active system set. Use Load() and Use() commands first")
	}

	systemName := c.activeSystem.System.Name.Value

	// Get current flow rates (no need to recalculate - they are kept up to date)
	currentFlowRates := c.GetCurrentFlowRates()

	// Extract nodes and edges from the system declaration
	var nodes []viz.Node
	var edges []viz.Edge
	instanceNameToID := make(map[string]string)

	// Build nodes from instance declarations
	for _, item := range c.activeSystem.System.Body {
		if instDecl, ok := item.(*decl.InstanceDecl); ok {
			nodeID := instDecl.Name.Value
			instanceNameToID[nodeID] = nodeID

			// Get component methods using runtime system
			var methods []viz.MethodInfo
			if c.runtime != nil {
				// Use runtime system to resolve component - it has already resolved imports
				fileInstance := c.runtime.LoadFile(c.activeFile.FullPath)
				if fileInstance != nil {
					component, err := fileInstance.GetComponentDecl(instDecl.ComponentName.Value)
					if err == nil && component != nil {
						componentMethods, _ := component.Methods()
						for methodName, methodDecl := range componentMethods {
							returnType := "void"
							if methodDecl.ReturnType != nil {
								returnType = methodDecl.ReturnType.Name
							}

							// Get traffic rate for this method from current flow rates
							methodTarget := fmt.Sprintf("%s.%s", nodeID, methodName)
							methodTraffic := currentFlowRates[methodTarget]

							// Only include methods with non-zero traffic to reduce clutter
							if methodTraffic > 0 {
								methods = append(methods, viz.MethodInfo{
									Name:       methodName,
									ReturnType: returnType,
									Traffic:    methodTraffic,
								})
							}
						}
					}
				}
			}

			// Calculate total traffic for this component
			componentTotalRPS := c.GetComponentTotalRPS(nodeID)
			var trafficDisplay string
			if componentTotalRPS > 0 {
				trafficDisplay = fmt.Sprintf("%.1f rps", componentTotalRPS)
			} else {
				trafficDisplay = "0 rps"
			}

			nodes = append(nodes, viz.Node{
				ID:      nodeID,
				Name:    instDecl.Name.Value,
				Type:    instDecl.ComponentName.Value,
				Methods: methods,
				Traffic: trafficDisplay,
			})
		}
	}

	// Build edges from instance overrides/dependencies
	for _, item := range c.activeSystem.System.Body {
		if instDecl, ok := item.(*decl.InstanceDecl); ok {
			fromNodeID := instanceNameToID[instDecl.Name.Value]

			// Add edges from instance overrides (system-level dependencies)
			for _, assignment := range instDecl.Overrides {
				if targetIdent, okIdent := assignment.Value.(*decl.IdentifierExpr); okIdent {
					if toNodeID, isInstance := instanceNameToID[targetIdent.Value]; isInstance {
						edges = append(edges, viz.Edge{
							FromID: fromNodeID,
							ToID:   toNodeID,
							Label:  assignment.Var.Value,
						})
					}
				}
			}
		}
	}

	// Add flow-based edges from FlowContext if available
	if c.currentFlowContext != nil && len(c.currentFlowContext.FlowPaths) > 0 {
		log.Printf("Canvas: Processing %d flow paths for edges", len(c.currentFlowContext.FlowPaths))

		// Clear existing edges to replace with flow-based edges
		edges = []viz.Edge{}

		// Process each flow path
		for pathKey, pathInfo := range c.currentFlowContext.FlowPaths {
			// Parse the flow path key (e.g., "server.HandleLookup->contactCache.Read")
			parts := strings.Split(pathKey, "->")
			if len(parts) != 2 {
				continue
			}

			fromParts := strings.Split(parts[0], ".")
			toParts := strings.Split(parts[1], ".")

			if len(fromParts) >= 2 && len(toParts) >= 2 {
				fromComponent := fromParts[0]
				fromMethod := fromParts[1]
				toComponent := toParts[0]
				toMethod := toParts[1]

				// Create label with order and condition
				label := fmt.Sprintf("%.0f", pathInfo.Order)
				if pathInfo.Condition != "" {
					if pathInfo.Order == math.Trunc(pathInfo.Order) {
						label = fmt.Sprintf("%.0f: %s", pathInfo.Order, pathInfo.Condition)
					} else {
						label = fmt.Sprintf("%.1f: %s", pathInfo.Order, pathInfo.Condition)
					}
				}

				log.Printf("Canvas: Creating edge %s.%s -> %s.%s with label %s",
					fromComponent, fromMethod, toComponent, toMethod, label)

				edges = append(edges, viz.Edge{
					FromID:      fromComponent,
					ToID:        toComponent,
					FromMethod:  fromMethod,
					ToMethod:    toMethod,
					Label:       label,
					Order:       pathInfo.Order,
					Condition:   pathInfo.Condition,
					Probability: pathInfo.Probability,
				})
			}
		}

		// If no flow paths, fall back to dependency edges
		if len(edges) == 0 {
			log.Printf("Canvas: No flow edges created, using dependency edges")
			// Re-create dependency edges
			for _, bodyItem := range c.activeSystem.System.Body {
				if instance, ok := bodyItem.(*decl.InstanceDecl); ok {
					fromNodeID := instance.Name.Value

					// Process overrides to find dependencies
					for _, assignment := range instance.Overrides {
						if targetIdent, okIdent := assignment.Value.(*decl.IdentifierExpr); okIdent {
							if toNodeID, isInstance := instanceNameToID[targetIdent.Value]; isInstance {
								edges = append(edges, viz.Edge{
									FromID: fromNodeID,
									ToID:   toNodeID,
									Label:  assignment.Var.Value,
								})
							}
						}
					}
				}
			}
		}
	}

	return &SystemDiagram{
		SystemName: systemName,
		Nodes:      nodes,
		Edges:      edges,
	}, nil
}

// initMeasurementTracing initializes the time-series database and measurement tracer
func (c *Canvas) initMeasurementTracing(dataDir string) error {
	if c.tsdb != nil {
		return nil // Already initialized
	}

	// Create time-series database
	tsdb, err := NewDuckDBTimeSeriesStore(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize time-series database: %w", err)
	}
	c.tsdb = tsdb

	// Create measurement tracer
	c.measurementTracer = NewMeasurementTracer(tsdb, "default_run")

	return nil
}

// CreateMeasurementTracer creates a new measurement tracer for the Canvas
func (c *Canvas) CreateMeasurementTracer(dataDir string) (*MeasurementTracer, error) {
	err := c.initMeasurementTracing(dataDir)
	if err != nil {
		return nil, err
	}
	return c.measurementTracer, nil
}

// GetMeasurementTracer returns the current measurement tracer, initializing if needed
func (c *Canvas) GetMeasurementTracer(dataDir string) (*MeasurementTracer, error) {
	if c.measurementTracer == nil {
		_, err := c.CreateMeasurementTracer(dataDir)
		if err != nil {
			return nil, err
		}
	}
	return c.measurementTracer, nil
}

// AddMeasurement adds a new measurement target to the tracer
func (c *Canvas) AddCanvasMeasurement(id, name, target, metricType string, enabled bool) error {
	tracer, err := c.GetMeasurementTracer("") // Use default location
	if err != nil {
		return err
	}

	measurement := &MeasurementConfig{
		ID:         id,
		Name:       name,
		Target:     target,
		MetricType: metricType,
		Enabled:    enabled,
	}

	tracer.AddMeasurement(measurement)
	return nil
}

// RemoveMeasurement removes a measurement target from the tracer
func (c *Canvas) RemoveCanvasMeasurement(target string) error {
	if c.measurementTracer == nil {
		return nil // No tracer, nothing to remove
	}
	c.measurementTracer.RemoveMeasurement(target)
	return nil
}

// GetCanvasMeasurements returns all registered measurements
func (c *Canvas) GetCanvasMeasurements() map[string]*MeasurementConfig {
	if c.measurementTracer == nil {
		return make(map[string]*MeasurementConfig)
	}
	return c.measurementTracer.GetMeasurements()
}

// ClearMeasurements removes all measurement targets
func (c *Canvas) ClearMeasurements() {
	if c.measurementTracer != nil {
		c.measurementTracer.ClearMeasurements()
	}
}

// HasMeasurements returns true if any measurements are registered
func (c *Canvas) HasMeasurements() bool {
	if c.measurementTracer == nil {
		return false
	}
	return c.measurementTracer.HasMeasurements()
}

// SetMeasurementRunID updates the run ID for the current measurement session
func (c *Canvas) SetMeasurementRunID(runID string) {
	if c.measurementTracer != nil {
		c.measurementTracer.SetRunID(runID)
	}
}

// GetMeasurementMetrics retrieves recent metrics for a target
func (c *Canvas) GetMeasurementMetrics(target string, since time.Time) ([]TracePoint, error) {
	if c.measurementTracer == nil {
		return nil, fmt.Errorf("measurement tracer not initialized")
	}
	return c.measurementTracer.GetMetrics(target, since)
}

// GetMeasurementPercentiles calculates percentiles for a target
func (c *Canvas) GetMeasurementPercentiles(target string, since time.Time) (p50, p90, p95, p99 float64, err error) {
	if c.measurementTracer == nil {
		return 0, 0, 0, 0, fmt.Errorf("measurement tracer not initialized")
	}
	return c.measurementTracer.GetPercentiles(target, since)
}

// ExecuteMeasurementSQL runs a custom SQL query on measurement data
func (c *Canvas) ExecuteMeasurementSQL(query string, args ...interface{}) ([]map[string]interface{}, error) {
	if c.measurementTracer == nil {
		return nil, fmt.Errorf("measurement tracer not initialized")
	}
	return c.measurementTracer.ExecuteSQL(query, args...)
}

// GetMeasurementStats returns statistics about stored measurements
func (c *Canvas) GetMeasurementStats() (map[string]interface{}, error) {
	if c.measurementTracer == nil {
		return nil, fmt.Errorf("measurement tracer not initialized")
	}
	return c.measurementTracer.GetStats()
}

// CanvasStats represents Canvas statistics for server monitoring
type CanvasStats struct {
	LoadedFiles        int `json:"loaded_files"`
	ActiveSystems      int `json:"active_systems"`
	ActiveGenerators   int `json:"active_generators"`
	ActiveMeasurements int `json:"active_measurements"`
	TotalRuns          int `json:"total_runs"`
}

// GetStats returns current Canvas statistics
func (c *Canvas) GetStats() CanvasStats {
	stats := CanvasStats{
		LoadedFiles:   len(c.loadedFiles),
		ActiveSystems: 0,
		TotalRuns:     len(c.sessionVars),
	}

	if c.activeSystem != nil {
		stats.ActiveSystems = 1
	}

	if c.genManager != nil {
		stats.ActiveGenerators = len(c.genManager.generators)
	}

	if c.measManager != nil {
		stats.ActiveMeasurements = len(c.measManager.measurements)
	}

	return stats
}

// evaluateProposedFlows calculates what the system flows would be with current generator settings
func (c *Canvas) evaluateProposedFlows() error {
	if c.activeSystem == nil {
		return fmt.Errorf("no active system available for flow calculation")
	}

	// Build entry points from enabled generators
	entryPoints := make(map[string]float64)
	if c.genManager != nil {
		for _, gen := range c.genManager.generators {
			if gen.Enabled {
				entryPoints[gen.Target] = float64(gen.Rate)
			}
		}
	}

	// Create proposed flow context
	c.proposedFlowContext = runtime.NewFlowContext(c.activeSystem.System, c.systemParameters)
	c.registerNativeComponents(c.proposedFlowContext)

	// Run FlowEval to calculate proposed system-wide flows
	if len(entryPoints) > 0 {
		runtime.SolveSystemFlows(entryPoints, c.proposedFlowContext)
	}

	return nil
}

// applyProposedFlows moves the proposed flow context to current (accepting the new flow state)
func (c *Canvas) applyProposedFlows() {
	if c.proposedFlowContext != nil {
		c.currentFlowContext = c.proposedFlowContext
		c.proposedFlowContext = nil
	}
}

// invalidateFlowContexts clears flow contexts when system state changes
func (c *Canvas) invalidateFlowContexts() {
	c.currentFlowContext = nil
	c.proposedFlowContext = nil
}

// initializeFlowContexts sets up initial flow contexts for a new system
func (c *Canvas) initializeFlowContexts() {
	if c.activeSystem == nil {
		return
	}

	// Initialize current context with zero flows
	c.currentFlowContext = runtime.NewFlowContext(c.activeSystem.System, c.systemParameters)
	c.registerNativeComponents(c.currentFlowContext)

	// Clear proposed context
	c.proposedFlowContext = nil
}

// GetCurrentFlowRates returns the current (applied) flow rates
func (c *Canvas) GetCurrentFlowRates() map[string]float64 {
	if c.currentFlowContext == nil {
		return make(map[string]float64)
	}
	return c.currentFlowContext.ArrivalRates
}

// GetProposedFlowRates returns the proposed flow rates (being evaluated)
func (c *Canvas) GetProposedFlowRates() map[string]float64 {
	if c.proposedFlowContext == nil {
		return make(map[string]float64)
	}
	return c.proposedFlowContext.ArrivalRates
}

// GetComponentTotalRPS calculates total RPS for a component by summing all its methods
func (c *Canvas) GetComponentTotalRPS(componentID string) float64 {
	rates := c.GetCurrentFlowRates()
	total := 0.0
	prefix := componentID + "."

	for target, rps := range rates {
		if strings.HasPrefix(target, prefix) {
			total += rps
		}
	}

	return total
}

// Close closes the Canvas and cleans up resources
func (c *Canvas) Close() error {
	if c.measurementTracer != nil {
		err := c.measurementTracer.Close()
		c.measurementTracer = nil
		c.tsdb = nil
		return err
	}
	return nil
}

// registerNativeComponents registers native component instances in the FlowContext
// for flow analysis delegation
func (c *Canvas) registerNativeComponents(flowContext *runtime.FlowContext) {
	if c.activeSystem == nil || flowContext == nil {
		return
	}

	// Iterate through system instances and register native components
	for _, item := range c.activeSystem.System.Body {
		if instDecl, ok := item.(*decl.InstanceDecl); ok {
			componentName := instDecl.ComponentName.Value
			instanceName := instDecl.Name.Value

			// Check if this is a known native component type and create an instance
			var nativeComponent components.FlowAnalyzable
			switch componentName {
			case "ResourcePool":
				nativeComponent = &components.ResourcePool{
					Name:        instanceName,
					Size:        5, // Default size - could be overridden by parameters
					ArrivalRate: 1.0,
					AvgHoldTime: 0.01, // 10ms default hold time
				}
				nativeComponent.(*components.ResourcePool).Init()

			case "HashIndex":
				nativeComponent = &components.HashIndex{}
				nativeComponent.(*components.HashIndex).Init()

			case "Cache":
				nativeComponent = &components.Cache{
					HitRate: 0.6, // Default 60% hit rate matching the original simulation
				}
				nativeComponent.(*components.Cache).Init()
			}

			// Register the native component in the FlowContext
			if nativeComponent != nil {
				flowContext.SetNativeComponent(instanceName, nativeComponent)
				log.Printf("Canvas: Registered native component '%s' (type %s) for flow analysis", instanceName, componentName)
			}
		}
	}
}
