#!/bin/bash
# Test script for automatic flow calculation

echo "=== Testing Automatic Flow Calculation with MVP Demo ==="
echo

# First, let's set up the canvas with the MVP system
echo "1. Loading MVP system..."
sdl load examples/uber/mvp.sdl
sdl use UberMVP

echo
echo "2. Adding generators with --apply-flows flag..."
echo "   This should automatically calculate and apply arrival rates"

# Add normal traffic generator
sdl gen add normal webserver.RequestRide 100 --apply-flows
echo

# Add surge traffic generator (initially stopped)
sdl gen add surge webserver.RequestRide 300 --apply-flows
echo

echo "3. Listing generators..."
sdl gen list
echo

echo "4. Checking current flow rates..."
sdl flows show
echo

echo "5. Starting surge generator with --apply-flows..."
echo "   This should recalculate flows with both generators active"
sdl gen start surge --apply-flows
echo

echo "6. Checking updated flow rates..."
sdl flows show
echo

echo "7. Stopping surge generator with --apply-flows..."
echo "   Flows should return to normal levels"
sdl gen stop surge --apply-flows
echo

echo "8. Final flow rates check..."
sdl flows show
echo

echo "=== Test Complete ==="
echo "The --apply-flows flag automatically:"
echo "- Evaluates system flows when generators are modified"
echo "- Applies calculated arrival rates to components"
echo "- Updates flows when generators are started/stopped"