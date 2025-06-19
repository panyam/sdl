#!/bin/bash

# Quick test of metrics - copy and paste these commands

echo "=== SDL Metrics Quick Test ==="
echo "Run these commands in order:"
echo

cat << 'EOF'
# Terminal 1: Start server
./sdl serve --port 8080

# Terminal 2: Run these commands

# 1. Load system
./sdl load examples/contacts/contacts.sdl
./sdl use ContactsSystem

# 2. Add a simple measurement
./sdl measure add test1 server.Lookup count

# 3. Generate some traffic
./sdl run test server.Lookup --runs 50

# 4. Check the data
./sdl measure data test1

# 5. Get statistics
./sdl measure stats test1

# 6. Clean up
./sdl measure remove test1

# More examples:

# Add latency measurement with P95
./sdl measure add latency_test server.Lookup latency --aggregation p95

# Add error rate measurement
./sdl measure add error_test server.Lookup count --aggregation rate --result-value "Val(Bool: false)"

# List all measurements
./sdl measure list

# Clear all measurements
./sdl measure clear
EOF