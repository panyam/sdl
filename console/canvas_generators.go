package console

import (
	"fmt"
	"sync"
	"time"

	"github.com/panyam/sdl/loader"
)

// GeneratorConfig represents a traffic generator configuration
type GeneratorConfig struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Target   string                 `json:"target"` // e.g., "app.Redirect"
	Rate     int                    `json:"rate"`   // requests per second
	Duration time.Duration          `json:"duration,omitempty"`
	Enabled  bool                   `json:"enabled"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// MeasurementConfig represents a measurement configuration
type MeasurementConfig struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	MetricType string                 `json:"metricType"` // "latency", "throughput", "errors", etc.
	Target     string                 `json:"target"`     // component to measure
	Interval   time.Duration          `json:"interval"`
	Enabled    bool                   `json:"enabled"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// MetricSnapshot represents a point-in-time metric value
type MetricSnapshot struct {
	Timestamp  int64   `json:"timestamp"`  // Unix timestamp in milliseconds
	MetricType string  `json:"metricType"` // e.g., "latency", "qps", "errorRate"
	Value      float64 `json:"value"`      // The metric value
	Source     string  `json:"source"`     // Source component/measurement
}

// CanvasState represents the complete state of a Canvas session
type CanvasState struct {
	LoadedFiles      []string                      `json:"loadedFiles"`
	ActiveFile       string                        `json:"activeFile"`
	ActiveSystem     string                        `json:"activeSystem"`
	Generators       map[string]*GeneratorConfig   `json:"generators"`
	Measurements     map[string]*MeasurementConfig `json:"measurements"`
	SessionVars      map[string]interface{}        `json:"sessionVars"`
	LastRunResult    interface{}                   `json:"lastRunResult,omitempty"`
	SystemParameters map[string]interface{}        `json:"systemParameters,omitempty"` // Current parameter values
	MetricsHistory   []MetricSnapshot              `json:"metricsHistory,omitempty"`   // Recent metrics for charts
}

// generatorManager manages traffic generators
type generatorManager struct {
	mu         sync.RWMutex
	generators map[string]*GeneratorConfig
	running    map[string]bool
	stopChans  map[string]chan struct{}
}

// measurementManager manages measurements
type measurementManager struct {
	mu           sync.RWMutex
	measurements map[string]*MeasurementConfig
	active       map[string]bool
	stopChans    map[string]chan struct{}
}

// initManagers initializes measurement manager only (generators now managed by SystemInstance)
func (c *Canvas) initManagers() {
	if c.measManager == nil {
		c.measManager = &measurementManager{
			measurements: make(map[string]*MeasurementConfig),
			active:       make(map[string]bool),
			stopChans:    make(map[string]chan struct{}),
		}
	}
}

// AddGenerator delegates to SystemInstance (legacy Canvas compatibility)
func (c *Canvas) AddGenerator(config *GeneratorConfig) error {
	if c.activeSystem == nil {
		return fmt.Errorf("no active system. Call Use() before adding generators")
	}

	if config.ID == "" {
		return fmt.Errorf("generator ID cannot be empty")
	}

	// Convert Canvas GeneratorConfig to runtime GeneratorConfig
	return c.activeSystem.AddGenerator(config.ID, config.Name, config.Target, float64(config.Rate))
}

// RemoveGenerator delegates to SystemInstance
func (c *Canvas) RemoveGenerator(id string) error {
	if c.activeSystem == nil {
		return fmt.Errorf("no active system available")
	}

	return c.activeSystem.RemoveGenerator(id)
}

// UpdateGenerator delegates to SystemInstance
func (c *Canvas) UpdateGenerator(config *GeneratorConfig) error {
	if c.activeSystem == nil {
		return fmt.Errorf("no active system available")
	}

	return c.activeSystem.UpdateGenerator(config.ID, float64(config.Rate), config.Enabled)
}

// PauseGenerator delegates to SystemInstance
func (c *Canvas) PauseGenerator(id string) error {
	if c.activeSystem == nil {
		return fmt.Errorf("no active system available")
	}

	gen, exists := c.activeSystem.GetGenerator(id)
	if !exists {
		return fmt.Errorf("generator with ID '%s' not found", id)
	}

	return c.activeSystem.UpdateGenerator(id, gen.Rate, false) // Disable the generator
}

// ResumeGenerator delegates to SystemInstance  
func (c *Canvas) ResumeGenerator(id string) error {
	if c.activeSystem == nil {
		return fmt.Errorf("no active system available")
	}

	gen, exists := c.activeSystem.GetGenerator(id)
	if !exists {
		return fmt.Errorf("generator with ID '%s' not found", id)
	}

	return c.activeSystem.UpdateGenerator(id, gen.Rate, true) // Enable the generator
}

// StartGenerators - simplified for flow analysis (no actual simulation running)
func (c *Canvas) StartGenerators() error {
	if c.activeSystem == nil {
		return fmt.Errorf("no active system. Call Use() before starting generators")
	}

	// For flow analysis, we just trigger flow recalculation
	return c.activeSystem.AnalyzeFlows()
}

// StopGenerators - simplified for flow analysis
func (c *Canvas) StopGenerators() error {
	// For flow analysis, generators don't actually "run" - this is a no-op
	return nil
}

// GetGenerators returns all generator configurations from SystemInstance
func (c *Canvas) GetGenerators() map[string]*GeneratorConfig {
	if c.activeSystem == nil {
		return make(map[string]*GeneratorConfig)
	}

	activeGenerators := c.activeSystem.GetActiveGenerators()
	result := make(map[string]*GeneratorConfig)
	
	// Convert runtime.GeneratorConfig to console.GeneratorConfig
	for _, gen := range activeGenerators {
		result[gen.ID] = &GeneratorConfig{
			ID:      gen.ID,
			Name:    gen.Name,
			Target:  gen.Target,
			Rate:    int(gen.Rate),
			Enabled: gen.Enabled,
			Options: gen.Options,
		}
	}
	
	return result
}

// AddMeasurement adds a new measurement configuration
func (c *Canvas) AddMeasurement(config *MeasurementConfig) error {
	c.initManagers()

	if config.ID == "" {
		return fmt.Errorf("measurement ID cannot be empty")
	}

	c.measManager.mu.Lock()
	defer c.measManager.mu.Unlock()

	if _, exists := c.measManager.measurements[config.ID]; exists {
		return fmt.Errorf("measurement with ID '%s' already exists", config.ID)
	}

	// Set default interval if not specified to prevent panic
	if config.Interval == 0 {
		config.Interval = 5 * time.Second // Default 5 second interval
	}

	c.measManager.measurements[config.ID] = config
	return nil
}

// RemoveMeasurement removes a measurement
func (c *Canvas) RemoveMeasurement(id string) error {
	c.initManagers()

	c.measManager.mu.Lock()
	defer c.measManager.mu.Unlock()

	if _, exists := c.measManager.measurements[id]; !exists {
		return fmt.Errorf("measurement with ID '%s' not found", id)
	}

	// Stop if active
	if c.measManager.active[id] {
		if stopChan, ok := c.measManager.stopChans[id]; ok {
			close(stopChan)
			delete(c.measManager.stopChans, id)
		}
		delete(c.measManager.active, id)
	}

	delete(c.measManager.measurements, id)
	return nil
}

// GetMeasurements returns all measurement configurations
func (c *Canvas) GetMeasurements() map[string]*MeasurementConfig {
	c.initManagers()

	c.measManager.mu.RLock()
	defer c.measManager.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*MeasurementConfig)
	for k, v := range c.measManager.measurements {
		result[k] = v
	}
	return result
}

// Save serializes the current Canvas state
func (c *Canvas) Save() (*CanvasState, error) {
	c.initManagers()

	state := &CanvasState{
		LoadedFiles:      make([]string, 0, len(c.loadedFiles)),
		Generators:       c.GetGenerators(),
		Measurements:     c.GetMeasurements(),
		SessionVars:      make(map[string]interface{}),
		SystemParameters: make(map[string]interface{}),
		MetricsHistory:   make([]MetricSnapshot, 0),
	}

	// Copy loaded files
	for path := range c.loadedFiles {
		state.LoadedFiles = append(state.LoadedFiles, path)
	}

	// Set active file and system
	if c.activeFile != nil {
		state.ActiveFile = c.activeFile.FullPath
	}
	if c.activeSystem != nil {
		state.ActiveSystem = c.activeSystem.System.Name.Value
	}

	// Copy session vars (shallow copy for now)
	for k, v := range c.sessionVars {
		state.SessionVars[k] = v
	}

	// Copy system parameters from SystemInstance
	if c.activeSystem != nil && c.activeSystem.FlowContext != nil {
		for k, v := range c.activeSystem.FlowContext.Parameters {
			state.SystemParameters[k] = v
		}
	}

	return state, nil
}

// recomputeSystemFlows evaluates proposed flows and automatically applies them
func (c *Canvas) recomputeSystemFlows() error {
	// Evaluate what the new flows would be
	err := c.evaluateProposedFlows()
	if err != nil {
		return err
	}
	
	// Automatically apply the proposed flows
	// In the future, we could add confirmation logic here
	c.applyProposedFlows()
	
	return nil
}


// Restore loads a previously saved Canvas state
func (c *Canvas) Restore(state *CanvasState) error {
	// Stop all generators and measurements first
	c.StopGenerators()

	// Clear current state
	c.loadedFiles = make(map[string]*loader.FileStatus)
	c.sessionVars = make(map[string]any)
	c.initManagers()

	// Reload files
	for _, filePath := range state.LoadedFiles {
		if err := c.Load(filePath); err != nil {
			return fmt.Errorf("failed to reload file '%s': %w", filePath, err)
		}
	}

	// Set active system
	if state.ActiveSystem != "" && state.ActiveFile != "" {
		// First ensure the correct file is active
		if fileStatus, ok := c.loadedFiles[state.ActiveFile]; ok {
			c.activeFile = fileStatus
			if err := c.Use(state.ActiveSystem); err != nil {
				return fmt.Errorf("failed to restore active system '%s': %w", state.ActiveSystem, err)
			}
		}
	}

	// Restore generators
	for _, gen := range state.Generators {
		if err := c.AddGenerator(gen); err != nil {
			return fmt.Errorf("failed to restore generator '%s': %w", gen.ID, err)
		}
	}

	// Restore measurements
	for _, meas := range state.Measurements {
		if err := c.AddMeasurement(meas); err != nil {
			return fmt.Errorf("failed to restore measurement '%s': %w", meas.ID, err)
		}
	}

	// Restore session vars
	for k, v := range state.SessionVars {
		c.sessionVars[k] = v
	}

	// Restore system parameters (reapply all parameter changes)
	for path, value := range state.SystemParameters {
		if err := c.Set(path, value); err != nil {
			// Log error but continue with other parameters
			fmt.Printf("Warning: failed to restore parameter %s: %v\n", path, err)
		}
	}

	return nil
}

// runGenerator is the internal method that executes a generator
func (c *Canvas) runGenerator(id string, config *GeneratorConfig, stopChan <-chan struct{}) {
	// Run simulations in batches at regular intervals
	// We generate config.Rate requests per second worth of load
	ticker := time.NewTicker(time.Second) // Check every second
	defer ticker.Stop()

	fmt.Printf("🚀 Generator %s started: %s @ %d rps\n", id, config.Target, config.Rate)

	for {
		select {
		case <-stopChan:
			fmt.Printf("🛑 Generator %s stopped\n", id)
			return
		case <-ticker.C:
			if config.Enabled {
				// Generate a batch of requests based on the rate
				// This simulates config.Rate requests happening over 1 second
				varName := fmt.Sprintf("gen_%s_%d", id, time.Now().Unix())

				// Run simulation batch - this will auto-inject measurements if configured
				err := c.Run(varName, config.Target, WithRuns(config.Rate))
				if err != nil {
					// Log error but continue generating
					fmt.Printf("❌ Generator %s error: %v\n", id, err)
				} else {
					fmt.Printf("✅ Generator %s: executed %d calls to %s\n", id, config.Rate, config.Target)
				}
			}
		}
	}
}

// startMeasurement starts collecting measurements
func (c *Canvas) startMeasurement(config *MeasurementConfig) {
	// This is a placeholder implementation
	// In a real implementation, this would collect metrics according to the config
	stopChan := make(chan struct{})
	c.measManager.stopChans[config.ID] = stopChan
	c.measManager.active[config.ID] = true

	go func() {
		ticker := time.NewTicker(config.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				// Collect measurement here
				// This would query the target component and record metrics
			}
		}
	}()
}
