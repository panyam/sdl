package recipe

import (
	"strings"
)

// ParseRecipe parses a recipe file content and returns commands and validation errors
func ParseRecipe(content string) *RecipeParseResult {
	lines := strings.Split(content, "\n")
	commands := []RecipeCommand{}
	errors := []RecipeValidationError{}

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		lineNumber := index + 1

		// Empty line
		if trimmed == "" {
			commands = append(commands, RecipeCommand{
				LineNumber: lineNumber,
				RawLine:    line,
				Type:       CommandTypeEmpty,
			})
			continue
		}

		// Comment line
		if strings.HasPrefix(trimmed, "#") {
			commands = append(commands, RecipeCommand{
				LineNumber:  lineNumber,
				RawLine:     line,
				Type:        CommandTypeComment,
				Description: strings.TrimSpace(trimmed[1:]),
			})
			continue
		}

		// Echo statement (description)
		if strings.HasPrefix(trimmed, "echo ") {
			echoContent := strings.TrimSpace(trimmed[5:])
			
			// Validate echo content
			echoErrors := ValidateEchoContent(echoContent, lineNumber)
			errors = append(errors, echoErrors...)
			
			commands = append(commands, RecipeCommand{
				LineNumber:  lineNumber,
				RawLine:     line,
				Type:        CommandTypeEcho,
				Description: RemoveQuotes(echoContent),
			})
			continue
		}

		// Echo with just "echo" (no content)
		if trimmed == "echo" {
			// Validate echo content (empty)
			echoErrors := ValidateEchoContent("", lineNumber)
			errors = append(errors, echoErrors...)
			
			commands = append(commands, RecipeCommand{
				LineNumber: lineNumber,
				RawLine:    line,
				Type:       CommandTypeEcho,
			})
			continue
		}

		// Read command (pause point)
		if trimmed == "read" || strings.HasPrefix(trimmed, "read ") {
			// Validate read command
			readErrors := ValidateReadCommand(trimmed, lineNumber)
			errors = append(errors, readErrors...)
			
			commands = append(commands, RecipeCommand{
				LineNumber:  lineNumber,
				RawLine:     line,
				Type:        CommandTypePause,
				Description: "Press continue to proceed",
			})
			continue
		}

		// SDL command - handles both "sdl ..." and standalone "sdl"
		if strings.HasPrefix(trimmed, "sdl ") || trimmed == "sdl" {
			parts := ParseCommandLine(trimmed)
			if len(parts) > 1 {
				// Validate SDL command
				sdlErrors := ValidateSDLCommand(parts[0], parts[1:], lineNumber, trimmed)
				errors = append(errors, sdlErrors...)
				
				commands = append(commands, RecipeCommand{
					LineNumber: lineNumber,
					RawLine:    line,
					Type:       CommandTypeCommand,
					Command:    parts[0],
					Args:       parts[1:],
				})
			} else {
				// "sdl" with no arguments
				errors = append(errors, RecipeValidationError{
					LineNumber: lineNumber,
					Message:    "SDL command missing arguments",
					Severity:   "error",
				})
				
				// Still create a command entry for it
				commands = append(commands, RecipeCommand{
					LineNumber: lineNumber,
					RawLine:    line,
					Type:       CommandTypeCommand,
					Command:    "sdl",
					Args:       []string{},
				})
			}
			continue
		}

		// Check for unsupported shell syntax
		if pattern := CheckUnsupportedPatterns(trimmed); pattern != nil {
			errors = append(errors, RecipeValidationError{
				LineNumber: lineNumber,
				Message:    pattern.Message + " - " + trimmed,
				Severity:   "error",
			})
			
			// Add as comment for unsupported lines
			commands = append(commands, RecipeCommand{
				LineNumber:  lineNumber,
				RawLine:     line,
				Type:        CommandTypeComment,
				Description: "[Unsupported: " + truncateString(trimmed, 50) + "]",
			})
			continue
		}

		// Check for other executable commands (not comments)
		if !strings.HasPrefix(trimmed, "#") {
			firstWord := strings.Fields(trimmed)[0]
			
			// Validate generic command
			genericErrors := ValidateGenericCommand(firstWord, lineNumber, trimmed)
			errors = append(errors, genericErrors...)
			
			// Add as comment for unsupported lines
			commands = append(commands, RecipeCommand{
				LineNumber:  lineNumber,
				RawLine:     line,
				Type:        CommandTypeComment,
				Description: "[Unsupported: " + truncateString(trimmed, 50) + "]",
			})
		}
	}

	return &RecipeParseResult{
		Commands: commands,
		Errors:   errors,
	}
}

// ParseCommandLine parses a command line with quoted string support
func ParseCommandLine(line string) []string {
	parts := []string{}
	current := ""
	inQuote := false
	quoteChar := byte(0)
	wasQuoted := false

	for i := 0; i < len(line); i++ {
		char := line[i]

		if inQuote {
			if char == quoteChar {
				inQuote = false
				quoteChar = 0
				wasQuoted = true
			} else {
				current += string(char)
			}
		} else {
			if char == '"' || char == '\'' {
				inQuote = true
				quoteChar = char
				wasQuoted = true
			} else if char == ' ' || char == '\t' {
				if current != "" || wasQuoted {
					parts = append(parts, current)
					current = ""
					wasQuoted = false
				}
			} else {
				current += string(char)
			}
		}
	}

	if current != "" || wasQuoted {
		parts = append(parts, current)
	}

	return parts
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}