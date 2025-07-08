package recipe

import (
	"io/ioutil"
	"testing"
)

func TestParseRealBitlyRecipe(t *testing.T) {
	// Read the real Bitly recipe file
	content, err := ioutil.ReadFile("/Users/sri/personal/golang/sdl/examples/bitly/mvp.recipe")
	if err != nil {
		t.Fatalf("Failed to read Bitly recipe: %v", err)
	}

	result := ParseRecipe(string(content))

	t.Logf("Parsed %d commands from Bitly recipe", len(result.Commands))
	t.Logf("Found %d validation errors", len(result.Errors))

	// Count different command types
	counts := map[RecipeCommandType]int{}
	for _, cmd := range result.Commands {
		counts[cmd.Type]++
	}

	t.Logf("Command type breakdown:")
	for cmdType, count := range counts {
		t.Logf("  %s: %d", cmdType, count)
	}

	// Log any validation errors
	if len(result.Errors) > 0 {
		t.Logf("Validation errors found:")
		for _, err := range result.Errors {
			t.Logf("  Line %d: %s", err.LineNumber, err.Message)
		}
	}

	// Basic sanity checks
	if len(result.Commands) == 0 {
		t.Error("No commands parsed from recipe")
	}

	// Count executable commands (commands that would actually be run)
	executableCount := 0
	for _, cmd := range result.Commands {
		if cmd.Type == CommandTypeCommand || cmd.Type == CommandTypeEcho || cmd.Type == CommandTypePause {
			executableCount++
		}
	}

	t.Logf("Executable commands: %d", executableCount)

	// Verify we found SDL commands
	sdlCommandCount := 0
	for _, cmd := range result.Commands {
		if cmd.Type == CommandTypeCommand && cmd.Command == "sdl" {
			sdlCommandCount++
		}
	}

	if sdlCommandCount == 0 {
		t.Error("No SDL commands found in recipe")
	}

	t.Logf("SDL commands found: %d", sdlCommandCount)

	// The recipe should parse without critical errors for the main functionality
	criticalErrors := 0
	for _, err := range result.Errors {
		// These are acceptable since the recipe contains some commented-out and invalid commands
		if err.Severity == "error" {
			criticalErrors++
		}
	}

	// The Bitly recipe has some intentionally invalid/commented commands, so we expect some errors
	// but not too many that would prevent execution of core functionality
	if criticalErrors > len(result.Commands)/2 {
		t.Errorf("Too many critical errors (%d) relative to total commands (%d)", criticalErrors, len(result.Commands))
	}
}

func TestParseSpecificBitlyCommands(t *testing.T) {
	// Test specific command patterns from the Bitly recipe
	testCases := []struct {
		name           string
		line           string
		expectedType   RecipeCommandType
		expectError    bool
		expectedArgs   []string
	}{
		{
			name:         "SDL load command",
			line:         "sdl load examples/bitly/mvp.sdl",
			expectedType: CommandTypeCommand,
			expectError:  false,
			expectedArgs: []string{"load", "examples/bitly/mvp.sdl"},
		},
		{
			name:         "SDL use command",
			line:         "sdl use Bitly",
			expectedType: CommandTypeCommand,
			expectError:  false,
			expectedArgs: []string{"use", "Bitly"},
		},
		{
			name:         "SDL metrics add with flags",
			line:         "sdl metrics add request_latency webserver RequestRide --type latency --window=1 --aggregation=p90",
			expectedType: CommandTypeCommand,
			expectError:  false,
			expectedArgs: []string{"metrics", "add", "request_latency", "webserver", "RequestRide", "--type", "latency", "--window=1", "--aggregation=p90"},
		},
		{
			name:         "SDL set command",
			line:         "sdl set database.pool.AvgHoldTime 0.15",
			expectedType: CommandTypeCommand,
			expectError:  false,
			expectedArgs: []string{"set", "database.pool.AvgHoldTime", "0.15"},
		},
		{
			name:         "SDL gen add with flags",
			line:         "sdl gen add baseline webserver.RequestRide 5 --apply-flows",
			expectedType: CommandTypeCommand,
			expectError:  false,
			expectedArgs: []string{"gen", "add", "baseline", "webserver.RequestRide", "5", "--apply-flows"},
		},
		{
			name:         "Echo command",
			line:         `echo "Let's reset back to 10k for our demo..."`,
			expectedType: CommandTypeEcho,
			expectError:  false,
		},
		{
			name:         "Comment line",
			line:         "# This is a comment",
			expectedType: CommandTypeComment,
			expectError:  false,
		},
		{
			name:         "Shebang line",
			line:         "#!/usr/bin/env bash",
			expectedType: CommandTypeComment,
			expectError:  false, // Shebang is treated as comment, not error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseRecipe(tc.line)

			if len(result.Commands) != 1 {
				t.Fatalf("Expected 1 command, got %d", len(result.Commands))
			}

			cmd := result.Commands[0]
			if cmd.Type != tc.expectedType {
				t.Errorf("Expected type %s, got %s", tc.expectedType, cmd.Type)
			}

			hasError := len(result.Errors) > 0
			if hasError != tc.expectError {
				t.Errorf("Expected error: %v, got error: %v", tc.expectError, hasError)
				if hasError {
					for _, err := range result.Errors {
						t.Logf("Error: %s", err.Message)
					}
				}
			}

			if tc.expectedArgs != nil && cmd.Type == CommandTypeCommand {
				if len(cmd.Args) != len(tc.expectedArgs) {
					t.Errorf("Expected %d args, got %d", len(tc.expectedArgs), len(cmd.Args))
				} else {
					for i, expectedArg := range tc.expectedArgs {
						if i < len(cmd.Args) && cmd.Args[i] != expectedArg {
							t.Errorf("Arg %d: expected %q, got %q", i, expectedArg, cmd.Args[i])
						}
					}
				}
			}
		})
	}
}