package console

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/panyam/sdl/cmd/sdl/commands"
	"github.com/panyam/sdl/decl"
	"github.com/panyam/sdl/loader"
	"github.com/panyam/sdl/runtime"
	"github.com/panyam/sdl/viz"
)

// Canvas provides a stateful, interactive environment for loading,
// modifying, and analyzing SDL models. It acts as the core engine
// for both scripted tests and future interactive tools like a REPL.
type Canvas struct {
	loader       *loader.Loader
	runtime      *runtime.Runtime
	activeFile   *loader.FileStatus
	activeSystem *runtime.SystemInstance
	sessionVars  map[string]any
	loadedFiles  map[string]*loader.FileStatus
}

// NewCanvas creates a new interactive canvas session.
func NewCanvas() *Canvas {
	l := loader.NewLoader(nil, nil, 10)
	r := runtime.NewRuntime(l)
	return &Canvas{
		loader:      l,
		runtime:     r,
		sessionVars: make(map[string]any),
		loadedFiles: make(map[string]*loader.FileStatus),
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
		// Debug removed
		return c.setField(currentComp.NativeInstance, []string{finalParamName}, value)
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
		return currentComp.Set(finalParamName, newValue)
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

	allResults := make([]commands.RunResult, 0, cfg.Runs)

	onBatch := func(batch int, batchVals []runtime.Value) {
		ts := time.Now().UnixMilli() // Use wall time for timestamp for now
		for _, val := range batchVals {
			allResults = append(allResults, commands.RunResult{
				Timestamp:   ts,
				Latency:     val.Time * 1000, // to ms
				ResultValue: val.String(),
				IsError:     false, // TODO
			})
		}
	}

	runtime.RunCallInBatches(c.activeSystem, instanceName, methodName, numBatches, batchSize, cfg.Workers, onBatch)
	c.sessionVars[varName] = allResults
	return nil
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

		results, ok := data.([]commands.RunResult)
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
