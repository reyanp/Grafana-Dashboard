#!/bin/bash

# Baseline Load Test Script
# Generates normal traffic to establish baseline metrics
# Requirements: 7.1, 7.5

set -e

# Configuration
TARGET_URL="http://localhost:8080"
DURATION="5m"
RATE="50"
OUTPUT_DIR="./load-test-results"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Baseline Load Test${NC}"
echo "Target: $TARGET_URL"
echo "Duration: $DURATION"
echo "Rate: $RATE RPS"
echo "Output Directory: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Check if vegeta is installed
if ! command -v vegeta &> /dev/null; then
    echo -e "${RED}Error: vegeta is not installed${NC}"
    echo "Please install vegeta: https://github.com/tsenart/vegeta"
    echo "On macOS: brew install vegeta"
    echo "On Linux: Download from GitHub releases"
    exit 1
fi

# Check if target is reachable
echo -e "${YELLOW}Checking if target is reachable...${NC}"
if ! curl -s "$TARGET_URL/healthz" > /dev/null; then
    echo -e "${RED}Error: Target $TARGET_URL is not reachable${NC}"
    echo "Make sure the application is running with: docker-compose up -d"
    exit 1
fi

echo -e "${GREEN}Target is reachable. Starting load test...${NC}"

# Create vegeta targets file for baseline load
cat > "$OUTPUT_DIR/baseline-targets.txt" << EOF
GET $TARGET_URL/api/v1/ping
GET $TARGET_URL/api/v1/work?ms=50&jitter=25
GET $TARGET_URL/api/v1/work?ms=100&jitter=50
GET $TARGET_URL/api/v1/work?ms=150&jitter=75
GET $TARGET_URL/healthz
GET $TARGET_URL/readyz
EOF

# Run the load test
echo -e "${YELLOW}Running baseline load test for $DURATION at $RATE RPS...${NC}"
vegeta attack \
    -targets="$OUTPUT_DIR/baseline-targets.txt" \
    -duration="$DURATION" \
    -rate="$RATE" \
    -output="$OUTPUT_DIR/baseline-results.bin"

# Generate reports
echo -e "${YELLOW}Generating reports...${NC}"

# Text report
vegeta report < "$OUTPUT_DIR/baseline-results.bin" > "$OUTPUT_DIR/baseline-report.txt"

# JSON report for programmatic analysis
vegeta report -type=json < "$OUTPUT_DIR/baseline-results.bin" > "$OUTPUT_DIR/baseline-report.json"

# Histogram report
vegeta report -type=hist[0,50ms,100ms,200ms,500ms,1s,2s] < "$OUTPUT_DIR/baseline-results.bin" > "$OUTPUT_DIR/baseline-histogram.txt"

# Plot report (if gnuplot is available)
if command -v gnuplot &> /dev/null; then
    vegeta plot < "$OUTPUT_DIR/baseline-results.bin" > "$OUTPUT_DIR/baseline-plot.html"
    echo -e "${GREEN}Plot generated: $OUTPUT_DIR/baseline-plot.html${NC}"
fi

echo -e "${GREEN}Baseline load test completed!${NC}"
echo ""
echo "Results:"
echo "- Text report: $OUTPUT_DIR/baseline-report.txt"
echo "- JSON report: $OUTPUT_DIR/baseline-report.json"
echo "- Histogram: $OUTPUT_DIR/baseline-histogram.txt"
if command -v gnuplot &> /dev/null; then
    echo "- Plot: $OUTPUT_DIR/baseline-plot.html"
fi
echo ""

# Display summary
echo -e "${YELLOW}Summary:${NC}"
cat "$OUTPUT_DIR/baseline-report.txt"

echo ""
echo -e "${GREEN}Check Grafana dashboards at http://localhost:3000 to see the metrics${NC}"
echo -e "${GREEN}Check Prometheus at http://localhost:9090 to see raw metrics${NC}"