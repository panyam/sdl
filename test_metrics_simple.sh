#!/bin/bash

# Simple test script for SDL metrics system using CLI commands
# This assumes the server is already running on port 8080

echo "=== SDL Metrics Test (Simple) ==="
echo "Prerequisites:"
echo "  1. Build SDL: make build"
echo "  2. Start server: ./sdl serve --port 8080"
echo "  3. Run this script in another terminal"
echo

# Step 1: Load and use system
echo "1. Setting up system..."
./sdl load examples/contacts/contacts.sdl
./sdl use ContactsSystem
echo

# Step 2: List measurements (should be empty)
echo "2. Current measurements (should be empty):"
./sdl measure list
echo

# Step 3: Add measurements
echo "3. Adding measurements..."

# Add throughput measurement
echo "   Adding throughput measurement..."
./sdl measure add server_qps server.Lookup count --aggregation rate --window 10s

# Add latency measurement
echo "   Adding latency measurement..."
./sdl measure add server_p95 server.Lookup latency --aggregation p95 --window 10s
echo

# List measurements again
echo "   Current measurements:"
./sdl measure list
echo

# Step 4: Generate some traffic
echo "4. Generating traffic..."
./sdl run results server.Lookup --runs 100 --workers 10
echo

# Step 5: Check data
echo "5. Checking metrics data..."

echo "   Raw data points:"
./sdl measure data server_qps

echo "   Statistics for QPS:"
./sdl measure stats server_qps

echo "   Statistics for P95 latency:"
./sdl measure stats server_p95
echo

# Step 6: Using traffic generators for continuous metrics
echo "6. Using traffic generators for continuous metrics..."
./sdl gen add lookup server.Lookup 20
./sdl gen start
echo "   Traffic running for 5 seconds..."
sleep 5
./sdl gen stop
echo

echo "   Updated statistics:"
./sdl measure stats server_qps
./sdl measure stats server_p95
echo

# Step 7: Cleanup
echo "7. Cleaning up..."
./sdl measure remove server_qps server_p95
./sdl gen remove lookup
echo "   Done!"
echo

echo "=== Test Complete ==="
echo
echo "Try these CLI commands yourself:"
echo "  # Add a measurement"
echo "  sdl measure add <id> <component.method> <metric> [--aggregation <type>] [--window <duration>]"
echo "  # List measurements"
echo "  sdl measure list"
echo "  # View raw data"
echo "  sdl measure data <id>"
echo "  # View statistics"
echo "  sdl measure stats <id>"
echo "  # Remove measurement"
echo "  sdl measure remove <id>"