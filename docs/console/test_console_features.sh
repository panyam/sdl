#!/bin/bash

echo "Testing SDL Console with go-prompt features"
echo "==========================================="
echo ""
echo "Features to test:"
echo "1. Tab completion for commands"
echo "2. Tab completion for file paths"
echo "3. Tab completion for parameter paths"
echo "4. Arrow key navigation (up/down for history)"
echo "5. Context-aware prompt showing active system"
echo "6. Rich completions with descriptions"
echo ""
echo "Starting SDL console..."
echo ""

# Start the console
./sdl console --port 8080