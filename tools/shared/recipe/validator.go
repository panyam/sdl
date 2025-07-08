package recipe

import (
	"regexp"
	"strings"
)

// UnsupportedPattern represents a pattern that is not supported in recipes
type UnsupportedPattern struct {
	Pattern *regexp.Regexp
	Message string
}

// UnsupportedPatterns contains all the shell syntax patterns that are not allowed
var UnsupportedPatterns = []UnsupportedPattern{
	{regexp.MustCompile(`^\s*if\s+`), "if statements not supported"},
	{regexp.MustCompile(`^\s*for\s+`), "for loops not supported"},
	{regexp.MustCompile(`^\s*while\s+`), "while loops not supported"},
	{regexp.MustCompile(`^\s*case\s+`), "case statements not supported"},
	{regexp.MustCompile(`^\s*function\s+`), "function definitions not supported"},
	{regexp.MustCompile(`.*\|.*`), "pipes not supported"},
	{regexp.MustCompile(`.*>>?.*`), "redirections not supported"},
	{regexp.MustCompile(`.*<.*`), "input redirection not supported"},
	{regexp.MustCompile(`.*\$\(.*\)`), "command substitution not supported"},
	{regexp.MustCompile(".*`.*`"), "backtick command substitution not supported"},
	{regexp.MustCompile(`^\s*export\s+`), "export not supported"},
	{regexp.MustCompile(`^\s*source\s+`), "source not supported"},
	{regexp.MustCompile(`^\s*\.\s+`), "source (.) not supported"},
	{regexp.MustCompile(`.*&\s*$`), "background jobs not supported"},
	{regexp.MustCompile(`^\s*\[.*\]`), "test expressions not supported"},
	{regexp.MustCompile(`^\s*\[\[.*\]\]`), "test expressions not supported"},
	{regexp.MustCompile(`.*\$\(\(.*\)\)`), "arithmetic expansion not supported"},
}

// ContainsUnquotedVariable checks if text contains variables outside of quoted strings
// Variables are indicated by $ characters not inside single or double quotes
func ContainsUnquotedVariable(text string) bool {
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for _, char := range text {
		if escaped {
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		} else if char == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		} else if char == '$' && !inSingleQuote && !inDoubleQuote {
			return true
		}
	}

	return false
}

// CheckUnsupportedPatterns checks if a line contains any unsupported shell syntax
func CheckUnsupportedPatterns(line string) *UnsupportedPattern {
	trimmed := strings.TrimSpace(line)
	
	for _, pattern := range UnsupportedPatterns {
		if pattern.Pattern.MatchString(trimmed) {
			return &pattern
		}
	}
	
	return nil
}

// ValidateEchoContent validates the content of an echo command
func ValidateEchoContent(content string, lineNumber int) []RecipeValidationError {
	var errors []RecipeValidationError
	
	if content == "" {
		errors = append(errors, RecipeValidationError{
			LineNumber: lineNumber,
			Message:    "Empty echo statement",
			Severity:   "error",
		})
	}
	
	if ContainsUnquotedVariable(content) {
		errors = append(errors, RecipeValidationError{
			LineNumber: lineNumber,
			Message:    "Variable expansion not supported in echo statements",
			Severity:   "error",
		})
	}
	
	return errors
}

// ValidateReadCommand validates a read command
func ValidateReadCommand(line string, lineNumber int) []RecipeValidationError {
	var errors []RecipeValidationError
	
	trimmed := strings.TrimSpace(line)
	if strings.Contains(trimmed, " ") && trimmed != "read" {
		errors = append(errors, RecipeValidationError{
			LineNumber: lineNumber,
			Message:    "'read' with variables not supported. Use plain 'read' for pause points",
			Severity:   "error",
		})
	}
	
	return errors
}

// ValidateSDLCommand validates an SDL command and its arguments
func ValidateSDLCommand(command string, args []string, lineNumber int, fullLine string) []RecipeValidationError {
	var errors []RecipeValidationError
	
	if len(args) == 0 {
		errors = append(errors, RecipeValidationError{
			LineNumber: lineNumber,
			Message:    "SDL command missing arguments",
			Severity:   "error",
		})
		return errors
	}
	
	// First arg should be the SDL subcommand
	sdlCommand := args[0]
	if !IsValidSDLCommand(sdlCommand) {
		errors = append(errors, RecipeValidationError{
			LineNumber: lineNumber,
			Message:    "Unknown SDL command '" + sdlCommand + "'. Valid commands: " + strings.Join(ValidSDLCommands, ", "),
			Severity:   "error",
		})
	}
	
	// Check for variable expansion in SDL commands
	if ContainsUnquotedVariable(fullLine) {
		errors = append(errors, RecipeValidationError{
			LineNumber: lineNumber,
			Message:    "Variable expansion not supported in SDL commands",
			Severity:   "error",
		})
	}
	
	return errors
}

// ValidateGenericCommand validates commands that are not sdl, echo, or read
func ValidateGenericCommand(command string, lineNumber int, fullLine string) []RecipeValidationError {
	var errors []RecipeValidationError
	
	// Check for unsupported shell patterns first
	if pattern := CheckUnsupportedPatterns(fullLine); pattern != nil {
		errors = append(errors, RecipeValidationError{
			LineNumber: lineNumber,
			Message:    pattern.Message + " - " + strings.TrimSpace(fullLine),
			Severity:   "error",
		})
		return errors
	}
	
	// Check for variables outside quotes
	if ContainsUnquotedVariable(fullLine) {
		errors = append(errors, RecipeValidationError{
			LineNumber: lineNumber,
			Message:    "variables not supported outside of quoted strings",
			Severity:   "error",
		})
		return errors
	}
	
	// Check if it's an unsupported command
	if IsUnsupportedCommand(command) {
		errors = append(errors, RecipeValidationError{
			LineNumber: lineNumber,
			Message:    "Command '" + command + "' not supported. Only 'sdl', 'echo', and 'read' commands are allowed",
			Severity:   "error",
		})
		return errors
	}
	
	// Check if it's not one of the allowed commands
	if !IsAllowedCommand(command) {
		errors = append(errors, RecipeValidationError{
			LineNumber: lineNumber,
			Message:    "Unknown command '" + command + "'. Only 'sdl', 'echo', and 'read' commands are supported",
			Severity:   "error",
		})
		return errors
	}
	
	return errors
}

// RemoveQuotes removes surrounding quotes from a string
func RemoveQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}