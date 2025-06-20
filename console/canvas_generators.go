package console

// GeneratorConfig represents a traffic generator configuration
/*
 */

// MetricSnapshot represents a point-in-time metric value
type MetricSnapshot struct {
	Timestamp  int64   `json:"timestamp"`  // Unix timestamp in milliseconds
	MetricType string  `json:"metricType"` // e.g., "latency", "qps", "errorRate"
	Value      float64 `json:"value"`      // The metric value
	Source     string  `json:"source"`     // Source component/measurement
}

// CanvasState represents the complete state of a Canvas session
/*
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

// Save serializes the current Canvas state
func (c *Canvas) Save() (*CanvasState, error) {
	c.initManagers()

	state := &CanvasState{
		LoadedFiles:      make([]string, 0, len(c.loadedFiles)),
		Generators:       c.ListGenerators(),
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

	// Copy system parameters
	for k, v := range c.systemParameters {
		state.SystemParameters[k] = v
	}

	return state, nil
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

	Start("Generator %s started: %s @ %d rps", id, config.Target, config.Rate)

	for {
		select {
		case <-stopChan:
			Stop("Generator %s stopped", id)
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
					Failure("Generator %s error: %v", id, err)
				} else {
					Debug("Generator %s: executed %d calls to %s", id, config.Rate, config.Target)
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
*/
