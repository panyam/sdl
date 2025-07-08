package recipe

import (
	"testing"
)

func TestParseRecipe(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		expectedCommands int
		expectedErrors   int
		expectedTypes    []RecipeCommandType
	}{
		{
			name: "simple valid recipe",
			content: `# This is a comment
echo "Starting the process"
sdl load system.sdl
read
sdl use MySystem
echo "Process complete"`,
			expectedCommands: 6,
			expectedErrors:   0,
			expectedTypes: []RecipeCommandType{
				CommandTypeComment,
				CommandTypeEcho,
				CommandTypeCommand,
				CommandTypePause,
				CommandTypeCommand,
				CommandTypeEcho,
			},
		},
		{
			name: "empty lines and comments",
			content: `
# First comment

echo "Hello"

# Second comment
`,
			expectedCommands: 7, // Including the trailing empty line
			expectedErrors:   0,
			expectedTypes: []RecipeCommandType{
				CommandTypeEmpty,
				CommandTypeComment,
				CommandTypeEmpty,
				CommandTypeEcho,
				CommandTypeEmpty,
				CommandTypeComment,
				CommandTypeEmpty, // Trailing empty line
			},
		},
		{
			name: "SDL commands with various arguments",
			content: `sdl load "my file.sdl"
sdl use system1
sdl gen add myGen component.method 10.5
sdl canvas create`,
			expectedCommands: 4,
			expectedErrors:   0,
			expectedTypes: []RecipeCommandType{
				CommandTypeCommand,
				CommandTypeCommand,
				CommandTypeCommand,
				CommandTypeCommand,
			},
		},
		{
			name: "validation errors",
			content: `echo
echo $VAR
read -p "prompt"
sdl
sdl invalid_command
for i in range; do echo $i; done
ls -la`,
			expectedCommands: 7,
			expectedErrors:   7,
			expectedTypes: []RecipeCommandType{
				CommandTypeEcho,    // echo (empty, has error)
				CommandTypeEcho,    // echo $VAR (has error)
				CommandTypePause,   // read -p "prompt" (has error)
				CommandTypeCommand, // sdl (no args, has error)
				CommandTypeCommand, // sdl invalid_command (has error)
				CommandTypeComment, // for loop becomes comment
				CommandTypeComment, // ls -la becomes comment
			},
		},
		{
			name: "quoted strings with spaces",
			content: `echo "This is a quoted string"
sdl load "file with spaces.sdl"
echo 'Single quoted string'`,
			expectedCommands: 3,
			expectedErrors:   0,
			expectedTypes: []RecipeCommandType{
				CommandTypeEcho,
				CommandTypeCommand,
				CommandTypeEcho,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRecipe(tt.content)

			if len(result.Commands) != tt.expectedCommands {
				t.Errorf("ParseRecipe() commands = %d, expected %d", len(result.Commands), tt.expectedCommands)
			}

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ParseRecipe() errors = %d, expected %d", len(result.Errors), tt.expectedErrors)
				for _, err := range result.Errors {
					t.Logf("Error: Line %d - %s", err.LineNumber, err.Message)
				}
			}

			if len(tt.expectedTypes) > 0 {
				for i, expectedType := range tt.expectedTypes {
					if i >= len(result.Commands) {
						t.Errorf("Not enough commands, expected type %s at index %d", expectedType, i)
						break
					}
					if result.Commands[i].Type != expectedType {
						t.Errorf("Command %d type = %s, expected %s", i, result.Commands[i].Type, expectedType)
					}
				}
			}
		})
	}
}

func TestParseCommandLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple command",
			input:    "sdl load file.sdl",
			expected: []string{"sdl", "load", "file.sdl"},
		},
		{
			name:     "quoted string with spaces",
			input:    `sdl load "file with spaces.sdl"`,
			expected: []string{"sdl", "load", "file with spaces.sdl"},
		},
		{
			name:     "single quoted string",
			input:    `echo 'Hello World'`,
			expected: []string{"echo", "Hello World"},
		},
		{
			name:     "mixed quotes",
			input:    `sdl gen add "my generator" component.method 'rate value'`,
			expected: []string{"sdl", "gen", "add", "my generator", "component.method", "rate value"},
		},
		{
			name:     "no quotes",
			input:    "sdl use MySystem",
			expected: []string{"sdl", "use", "MySystem"},
		},
		{
			name:     "empty quoted string",
			input:    `echo ""`,
			expected: []string{"echo", ""},
		},
		{
			name:     "tabs and spaces",
			input:    "sdl\tload\t\tfile.sdl   ",
			expected: []string{"sdl", "load", "file.sdl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCommandLine(tt.input)
			
			if len(result) != len(tt.expected) {
				t.Errorf("ParseCommandLine() length = %d, expected %d", len(result), len(tt.expected))
				t.Logf("Got: %v", result)
				t.Logf("Expected: %v", tt.expected)
				return
			}

			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("ParseCommandLine() part %d = %q, expected %q", i, part, tt.expected[i])
				}
			}
		})
	}
}

func TestValidateRecipeContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		errorCount  int
	}{
		{
			name: "valid recipe",
			content: `# Comment
echo "Hello"
sdl load file.sdl
read`,
			expectError: false,
			errorCount:  0,
		},
		{
			name: "empty echo",
			content: `echo`,
			expectError: true,
			errorCount:  1,
		},
		{
			name: "variable expansion",
			content: `echo $HOME`,
			expectError: true,
			errorCount:  1,
		},
		{
			name: "read with variables",
			content: `read -p "Enter value:"`,
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid SDL command",
			content: `sdl invalid_command`,
			expectError: true,
			errorCount:  1,
		},
		{
			name: "unsupported shell syntax",
			content: `if [ -f file ]; then echo "exists"; fi`,
			expectError: true,
			errorCount:  1,
		},
		{
			name: "unsupported command",
			content: `ls -la`,
			expectError: true,
			errorCount:  1,
		},
		{
			name: "multiple errors",
			content: `echo
echo $VAR
ls -la
for i in range; do echo $i; done`,
			expectError: true,
			errorCount:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRecipe(tt.content)

			hasErrors := result.HasErrors()
			if hasErrors != tt.expectError {
				t.Errorf("ParseRecipe() hasErrors = %v, expected %v", hasErrors, tt.expectError)
			}

			if len(result.Errors) != tt.errorCount {
				t.Errorf("ParseRecipe() error count = %d, expected %d", len(result.Errors), tt.errorCount)
				for _, err := range result.Errors {
					t.Logf("Error: Line %d - %s", err.LineNumber, err.Message)
				}
			}
		})
	}
}

func TestRecipeParseResultMethods(t *testing.T) {
	// Test HasErrors method
	result := &RecipeParseResult{
		Commands: []RecipeCommand{},
		Errors:   []RecipeValidationError{},
	}

	if result.HasErrors() {
		t.Error("HasErrors() should return false for empty errors")
	}

	result.Errors = append(result.Errors, RecipeValidationError{
		LineNumber: 1,
		Message:    "test error",
		Severity:   "error",
	})

	if !result.HasErrors() {
		t.Error("HasErrors() should return true when errors exist")
	}
}

func TestRecipeValidationErrorImplementsError(t *testing.T) {
	err := RecipeValidationError{
		LineNumber: 5,
		Message:    "test error message",
		Severity:   "error",
	}

	expected := "Line 5: test error message"
	if err.Error() != expected {
		t.Errorf("Error() = %q, expected %q", err.Error(), expected)
	}
}