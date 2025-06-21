#!/bin/bash
# End-to-end test for metrics functionality

set -e

echo "=== SDL Metrics End-to-End Test ==="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to check if server is running
check_server() {
    if ! curl -s http://localhost:9090/healthz > /dev/null 2>&1; then
        echo -e "${RED}✗ Server not running on port 9090${NC}"
        echo "Please start the server with: sdl serve"
        exit 1
    fi
    echo -e "${GREEN}✓ Server is running${NC}"
}

# Function to run command and check result
run_test() {
    local cmd="$1"
    local desc="$2"
    echo -n "Testing: $desc... "
    if eval "$cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        echo "Failed command: $cmd"
        exit 1
    fi
}

# Check server
check_server

# Test 1: Load SDL file
run_test "./sdl load examples/contacts/contacts.sdl" "Load contacts SDL file"

# Test 2: Use system
run_test "./sdl use ContactsSystem" "Use ContactsSystem"

# Test 3: Add metrics
run_test "./sdl metrics add latency_p95 server Lookup --type latency --aggregation p95 --window 5" "Add P95 latency metric"
run_test "./sdl metrics add request_rate server Lookup --type count --aggregation sum --window 10" "Add request count metric"
run_test "./sdl metrics add db_latency database Query --type latency --aggregation avg --window 5" "Add database latency metric"

# Test 4: List metrics
echo "Listing metrics:"
./sdl metrics list

# Test 5: Start generator
run_test "./sdl gen add test_gen server.Lookup 10" "Add generator at 10 RPS"
run_test "./sdl gen start test_gen" "Start generator"

# Let it run for a bit
echo "Collecting metrics for 5 seconds..."
sleep 5

# Test 6: Query metrics
echo ""
echo "Querying metrics:"
./sdl metrics query latency_p95 --duration 30s --limit 10

# Test 7: Stop generator
run_test "./sdl gen stop test_gen" "Stop generator"

# Test 8: Remove metric (test the fix)
run_test "./sdl metrics remove request_rate" "Remove request_rate metric"

# Test 9: List metrics again (should not crash)
echo ""
echo "Listing metrics after removal:"
./sdl metrics list

# Test 10: Clean up
run_test "./sdl gen remove test_gen" "Remove generator"
run_test "./sdl metrics remove latency_p95" "Remove latency_p95 metric"
run_test "./sdl metrics remove db_latency" "Remove db_latency metric"

echo ""
echo -e "${GREEN}=== All tests passed! ===${NC}"