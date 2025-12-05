package systemdetail

import (
	"fmt"
	"time"

	"github.com/panyam/sdl/lib/runtime"
	"github.com/panyam/sdl/services"
	"github.com/panyam/sdl/tools/shared/recipe"
)

// RecipeStep represents a step in recipe execution
type RecipeStep struct {
	Index      int      `json:"index"`
	LineNumber int      `json:"lineNumber"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	Status     string   `json:"status"` // pending, running, completed, failed, skipped
	Output     string   `json:"output,omitempty"`
	Error      string   `json:"error,omitempty"`
	StartTime  int64    `json:"startTime,omitempty"`
	EndTime    int64    `json:"endTime,omitempty"`
}

// RecipeExecState represents the current state of recipe execution
type RecipeExecState struct {
	IsRunning   bool         `json:"isRunning"`
	CurrentStep int          `json:"currentStep"`
	TotalSteps  int          `json:"totalSteps"`
	Steps       []RecipeStep `json:"steps"`
	Mode        string       `json:"mode"` // step, auto
	FileName    string       `json:"fileName"`
}

// SystemDiagram represents a system architecture diagram
type SystemDiagram struct {
	Systems     []string               `json:"systems"`
	Components  []DiagramComponent     `json:"components"`
	Connections []DiagramConnection    `json:"connections"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DiagramComponent represents a component in the system diagram
type DiagramComponent struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	System     string                 `json:"system"`
	Properties map[string]interface{} `json:"properties"`
}

// DiagramConnection represents a connection between components
type DiagramConnection struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Method string  `json:"method"`
	Rate   float64 `json:"rate,omitempty"`
}

// CompileResult represents the result of SDL compilation
type CompileResult struct {
	Success  bool     `json:"success"`
	Systems  []string `json:"systems"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// Callbacks for external environment integration
type Callbacks struct {
	OnRecipeStep      func(step RecipeStep)
	OnMeasurementData func(data map[string]interface{})
	OnError           func(error string)
	OnInfo            func(message string)
	OnSuccess         func(message string)
}

// SystemDetailTool provides SDL compilation, recipe execution, and diagram generation
// This is a reusable tool that can be used in WASM, CLI, tests, or server environments
type SystemDetailTool struct {
	// Core state
	systemID      string
	sdlContent    string
	recipeContent string

	// Execution state
	execState     *RecipeExecState
	diagram       *SystemDiagram
	compileResult *CompileResult

	// SDL runtime components
	canvas   *services.Canvas
	runtime  *runtime.Runtime
	canvasID string

	// External callbacks
	callbacks *Callbacks
}

// NewSystemDetailTool creates a new SystemDetailTool instance
func NewSystemDetailTool() *SystemDetailTool {
	// For now, create with minimal setup
	// The actual loader will be created when SDL content is set
	tool := &SystemDetailTool{
		canvasID: fmt.Sprintf("systemdetail-%d", time.Now().UnixNano()),
		execState: &RecipeExecState{
			Steps: make([]RecipeStep, 0),
		},
		callbacks: &Callbacks{},
	}

	return tool
}

// SetCallbacks sets the callback functions for external environment integration
func (t *SystemDetailTool) SetCallbacks(callbacks *Callbacks) {
	if callbacks != nil {
		t.callbacks = callbacks
	}
}

// Initialize initializes the tool with system data
func (t *SystemDetailTool) Initialize(systemID, sdlContent, recipeContent string) error {
	t.systemID = systemID
	t.sdlContent = sdlContent
	t.recipeContent = recipeContent

	// Reset state
	t.execState = &RecipeExecState{
		Steps: make([]RecipeStep, 0),
	}
	t.diagram = nil
	t.compileResult = nil

	return nil
}

// GetSystemID returns the current system ID
func (t *SystemDetailTool) GetSystemID() string {
	return t.systemID
}

// GetSDLContent returns the current SDL content
func (t *SystemDetailTool) GetSDLContent() string {
	return t.sdlContent
}

// Note: SetSDLContent is implemented in sdl.go

// GetRecipeContent returns the current recipe content
func (t *SystemDetailTool) GetRecipeContent() string {
	return t.recipeContent
}

// SetRecipeContent sets the recipe content and validates it
func (t *SystemDetailTool) SetRecipeContent(content string) error {
	t.recipeContent = content

	// Reset execution state
	t.execState = &RecipeExecState{
		Steps: make([]RecipeStep, 0),
	}

	// Parse and validate recipe
	return t.parseRecipe()
}

// parseRecipe parses the recipe content into steps using the shared recipe parser
func (t *SystemDetailTool) parseRecipe() error {
	if t.recipeContent == "" {
		return nil
	}

	// Parse the recipe using the shared parser
	parseResult := recipe.ParseRecipe(t.recipeContent)

	// Check for validation errors
	if parseResult.HasErrors() {
		errorMsg := "Recipe validation errors:\n"
		for _, err := range parseResult.Errors {
			errorMsg += fmt.Sprintf("Line %d: %s\n", err.LineNumber, err.Message)
		}
		return fmt.Errorf("%s", errorMsg)
	}

	// Convert recipe commands to execution steps
	steps := make([]RecipeStep, 0)

	for _, cmd := range parseResult.Commands {
		// Only create steps for executable commands
		if cmd.Type == recipe.CommandTypeCommand || cmd.Type == recipe.CommandTypeEcho || cmd.Type == recipe.CommandTypePause {
			step := RecipeStep{
				Index:      len(steps),
				LineNumber: cmd.LineNumber,
				Status:     "pending",
			}

			// Set command and args based on type
			switch cmd.Type {
			case recipe.CommandTypeCommand:
				step.Command = cmd.Command
				step.Args = cmd.Args
			case recipe.CommandTypeEcho:
				step.Command = "echo"
				if cmd.Description != "" {
					step.Args = []string{cmd.Description}
				}
			case recipe.CommandTypePause:
				step.Command = "read"
				step.Args = []string{}
			}

			steps = append(steps, step)
		}
	}

	t.execState.Steps = steps
	t.execState.TotalSteps = len(steps)

	return nil
}

// GetExecState returns the current recipe execution state
func (t *SystemDetailTool) GetExecState() *RecipeExecState {
	return t.execState
}

// GetDiagram returns the current system diagram
func (t *SystemDetailTool) GetDiagram() *SystemDiagram {
	return t.diagram
}

// GetCompileResult returns the last compilation result
func (t *SystemDetailTool) GetCompileResult() *CompileResult {
	return t.compileResult
}

// Helper methods for callbacks
func (t *SystemDetailTool) emitError(message string) {
	if t.callbacks.OnError != nil {
		t.callbacks.OnError(message)
	}
}

func (t *SystemDetailTool) emitInfo(message string) {
	if t.callbacks.OnInfo != nil {
		t.callbacks.OnInfo(message)
	}
}

func (t *SystemDetailTool) emitSuccess(message string) {
	if t.callbacks.OnSuccess != nil {
		t.callbacks.OnSuccess(message)
	}
}

func (t *SystemDetailTool) emitRecipeStep(step RecipeStep) {
	if t.callbacks.OnRecipeStep != nil {
		t.callbacks.OnRecipeStep(step)
	}
}
