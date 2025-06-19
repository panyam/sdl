#!/bin/bash

# Test script for SDL metrics system

SDL=$GOBIN/sdl

echo "=== SDL Metrics System Test ==="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Server URL
SERVER_URL="${CANVAS_SERVER_URL:-http://localhost:8080}"

echo -e "${BLUE}1. Starting SDL server...${NC}"
echo "Run this in another terminal:"
echo "  $SDL serve --port 8080"
echo
read -p "Press enter when server is running..."

echo -e "\n${BLUE}2. Loading SDL file...${NC}"
$SDL load examples/contacts/contacts.sdl
echo

echo -e "${BLUE}3. Using ContactsSystem...${NC}"
$SDL use ContactsSystem
echo

echo -e "${BLUE}4. Testing Metrics API with curl...${NC}"
echo

# List measurements (should be empty)
echo -e "${GREEN}Listing measurements (should be empty):${NC}"
$SDL measure list
echo

# Add a count measurement
echo -e "${GREEN}Adding server throughput measurement:${NC}"
$SDL measure add server_throughput server.Lookup count --aggregation rate --window 10s
echo

# Add a latency measurement
echo -e "${GREEN}Adding server latency measurement:${NC}"
$SDL measure add server_latency server.Lookup latency --aggregation p95 --window 10s
echo

# Add an error rate measurement
echo -e "${GREEN}Adding server error rate measurement:${NC}"
$SDL measure add server_errors server.Lookup count --aggregation rate --window 30s --result-value "Val(Bool: false)"
echo

# List measurements again
echo -e "${GREEN}Listing measurements (should have 3):${NC}"
$SDL measure list
echo

# Run some simulations to generate data
echo -e "${BLUE}5. Running simulations to generate metrics data...${NC}"
$SDL gen add lookup server.Lookup 10
$SDL gen start
sleep 3
$SDL gen stop
echo

# Get measurement data
echo -e "${BLUE}6. Retrieving metrics data...${NC}"
echo

echo -e "${GREEN}Server throughput data points:${NC}"
$SDL measure data server_throughput
echo

echo -e "${GREEN}Server latency data points:${NC}"
$SDL measure data server_latency
echo

# Get aggregated data
echo -e "${GREEN}Server throughput statistics:${NC}"
$SDL measure stats server_throughput
echo

echo -e "${GREEN}Server latency statistics:${NC}"
$SDL measure stats server_latency
echo

echo -e "${GREEN}Server error rate statistics:${NC}"
$SDL measure stats server_errors
echo

# Test error cases
echo -e "${BLUE}7. Testing error cases...${NC}"
echo

echo -e "${GREEN}Getting non-existent measurement:${NC}"
$SDL measure stats does_not_exist || echo "✓ Error handled correctly"
echo

echo -e "${GREEN}Adding measurement with invalid component:${NC}"
$SDL measure add bad_measurement nonexistent.Method count || echo "✓ Error handled correctly"
echo

echo -e "${GREEN}Adding measurement with invalid metric type:${NC}"
$SDL measure add bad_metric server.Lookup invalid_type || echo "✓ Error handled correctly"
echo

# Clean up
echo -e "\n${BLUE}8. Cleaning up...${NC}"
echo

echo -e "${GREEN}Removing measurements individually:${NC}"
$SDL measure remove server_throughput server_latency server_errors
echo

echo -e "${GREEN}Verifying cleanup:${NC}"
$SDL measure list
echo

# Test clear all
echo -e "${GREEN}Testing clear all command:${NC}"
$SDL measure add temp1 server.Lookup count
$SDL measure add temp2 server.Lookup latency
echo "Added 2 temporary measurements"
$SDL measure clear
echo

echo -e "${GREEN}Final verification:${NC}"
$SDL measure list
echo

echo -e "\n${BLUE}=== Test Complete ===${NC}"
