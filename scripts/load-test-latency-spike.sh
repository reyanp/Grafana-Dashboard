#!/bin/bash

# Latency Spike Load Test Script
# Generates high latency traffic to trigger p95 latency alerts
# Requirements: 7.1, 7.2

set -e

# Configuration
TARGET_URL="http://localhost:8080"
DURATION="12m"  # Run longer than alert threshold (10m)
RATE="30"       # Moderate rate to focus on latency
OUTPUT_DIR="./load-test-results"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Latency Spike Load Test${NC}"
echo "Target: $TARGET_URL"
echo "Duration: $DURATION"
echo "Rate: $RATE RPS"
echo "Output Directory: $OUTPUT_DIR"
echo ""
echo -e "${YELLOW}This test will generate high latency requests to trigger p95 latency alerts${NC}"
echo -e "${YELLOW}Alert threshold: p95 > 500ms for 10 minutes${NC}"
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

echo -e "${GREEN}Target is reachable. Starting latency spike test...${NC}"

# Create vegeta targets file for high latency requests
cat > "$OUTPUT_DIR/latency-spike-targets.txt" << EOF
GET $TARGET_URL/api/v1/work?ms=600&jitter=200
GET $TARGET_URL/api/v1/work?ms=700&jitter=300
GET $TARGET_URL/api/v1/work?ms=800&jitter=400
GET $TARGET_URL/api/v1/work?ms=900&jitter=500
GET $TARGET_URL/api/v1/work?ms=1000&jitter=500
GET $TARGET_URL/api/v1/work?ms=1200&jitter=600
EOF

# Run the load test
echo -e "${YELLOW}Running latency spike test for $DURATION at $RATE RPS...${NC}"
echo -e "${YELLOW}Expected to trigger HighLatencyP95 alert after 10 minutes${NC}"

vegeta attack \
    -targets="$OUTPUT_DIR/latency-spike-targets.txt" \
    -duration="$DURATION" \
    -rate="$RATE" \
    -output="$OUTPUT_DIR/latency-spike-results.bin"

# Generate reports
echo -e "${YELLOW}Generating reports...${NC}"

# Text report
vegeta report < "$OUTPUT_DIR/latency-spike-results.bin" > "$OUTPUT_DIR/latency-spike-report.txt"

# JSON report for programmatic analysis
vegeta report -type=json < "$OUTPUT_DIR/latency-spike-results.bin" > "$OUTPUT_DIR/latency-spike-report.json"

# Histogram report with focus on high latencies
vegeta report -type=hist[0,100ms,200ms,500ms,1s,2s,5s] < "$OUTPUT_DIR/latency-spike-results.bin" > "$OUTPUT_DIR/latency-spike-histogram.txt"

# Plot report (if gnuplot is available)
if command -v gnuplot &> /dev/null; then
    vegeta plot < "$OUTPUT_DIR/latency-spike-results.bin" > "$OUTPUT_DIR/latency-spike-plot.html"
    echo -e "${GREEN}Plot generated: $OUTPUT_DIR/latency-spike-plot.html${NC}"
fi

echo -e "${GREEN}Latency spike load test completed!${NC}"
echo ""
echo "Results:"
echo "- Text report: $OUTPUT_DIR/latency-spike-report.txt"
echo "- JSON report: $OUTPUT_DIR/latency-spike-report.json"
echo "- Histogram: $OUTPUT_DIR/latency-spike-histogram.txt"
if command -v gnuplot &> /dev/null; then
    echo "- Plot: $OUTPUT_DIR/latency-spike-plot.html"
fi
echo ""

# Display summary
echo -e "${YELLOW}Summary:${NC}"
cat "$OUTPUT_DIR/latency-spike-report.txt"

echo ""
echo -e "${GREEN}Check Grafana dashboards at http://localhost:3000 to see the latency metrics${NC}"
echo -e "${GREEN}Check AlertManager at http://localhost:9093 for HighLatencyP95 alerts${NC}"
echo -e "${GREEN}Check Prometheus at http://localhost:9090 to see alert status${NC}"

# Check if p95 latency exceeded threshold
P95_LATENCY=$(vegeta report -type=json < "$OUTPUT_DIR/latency-spike-results.bin" | jq -r '.latencies.p95 / 1000000')
THRESHOLD=500

echo ""
if (( $(echo "$P95_LATENCY > $THRESHOLD" | bc -l) )); then
    echo -e "${GREEN}✓ P95 latency (${P95_LATENCY}ms) exceeded threshold (${THRESHOLD}ms)${NC}"
    echo -e "${GREEN}✓ HighLatencyP95 alert should be triggered${NC}"
else
    echo -e "${YELLOW}⚠ P95 latency (${P95_LATENCY}ms) did not exceed threshold (${THRESHOLD}ms)${NC}"
    echo -e "${YELLOW}⚠ You may need to increase latency parameters or run longer${NC}"
fi