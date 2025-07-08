package recipe

import "fmt"

// RecipeCommandType represents the type of a recipe command
type RecipeCommandType string

const (
	CommandTypeEmpty   RecipeCommandType = "empty"
	CommandTypeComment RecipeCommandType = "comment"
	CommandTypeEcho    RecipeCommandType = "echo"
	CommandTypePause   RecipeCommandType = "pause"
	CommandTypeCommand RecipeCommandType = "command"
)

// RecipeCommand represents a single command in a recipe
type RecipeCommand struct {
	LineNumber  int               `json:"lineNumber"`
	RawLine     string            `json:"rawLine"`
	Type        RecipeCommandType `json:"type"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Description string            `json:"description,omitempty"`
}

// RecipeValidationError represents a validation error in a recipe
type RecipeValidationError struct {
	LineNumber int    `json:"lineNumber"`
	Message    string `json:"message"`
	Severity   string `json:"severity"` // "error" or "warning"
}

// Error implements the error interface for RecipeValidationError
func (e RecipeValidationError) Error() string {
	return fmt.Sprintf("Line %d: %s", e.LineNumber, e.Message)
}

// RecipeParseResult contains the result of parsing a recipe
type RecipeParseResult struct {
	Commands []RecipeCommand          `json:"commands"`
	Errors   []RecipeValidationError  `json:"errors,omitempty"`
}

// HasErrors returns true if there are any validation errors
func (r *RecipeParseResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ValidSDLCommands lists the allowed SDL commands
var ValidSDLCommands = []string{
	"load",    // Load SDL file
	"use",     // Use system  
	"gen",     // Generator operations
	"metrics", // Metrics operations
	"set",     // Set parameters
	"canvas",  // Canvas operations
}

// UnsupportedCommands lists shell commands that are not allowed
var UnsupportedCommands = []string{
	"cd", "pwd", "ls", "mkdir", "rm", "cp", "mv", "cat", "grep", "sed", "awk",
	"find", "chmod", "chown", "curl", "wget", "git", "npm", "yarn", "python",
	"node", "bash", "sh", "zsh", "exit", "return", "break", "continue",
}

// IsValidSDLCommand checks if a command is a valid SDL command
func IsValidSDLCommand(command string) bool {
	for _, valid := range ValidSDLCommands {
		if command == valid {
			return true
		}
	}
	return false
}

// IsUnsupportedCommand checks if a command is in the unsupported list
func IsUnsupportedCommand(command string) bool {
	for _, unsupported := range UnsupportedCommands {
		if command == unsupported {
			return true
		}
	}
	return false
}

// IsAllowedCommand checks if a command is one of the three allowed types
func IsAllowedCommand(command string) bool {
	return command == "sdl" || command == "echo" || command == "read"
}