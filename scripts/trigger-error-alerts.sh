#!/bin/bash

# Error Alert Trigger Script
# Uses error injection to trigger error rate alerts
# Requirements: 7.2, 7.3

set -e

# Configuration
TARGET_URL="http://localhost:8080"
ADMIN_TOKEN="${ADMIN_TOKEN:-changeme}"
DURATION="12m"  # Run longer than alert threshold (10m)
RATE="40"       # Moderate rate to generate sufficient errors
ERROR_RATE="0.05"  # 5% error rate (above 2% threshold)
STATUS_CODE="500"
OUTPUT_DIR="./load-test-results"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Error Alert Trigger Test${NC}"
echo "Target: $TARGET_URL"
echo "Duration: $DURATION"
echo "Rate: $RATE RPS"
echo "Error Rate: $ERROR_RATE (${ERROR_RATE%.*}%)"
echo "Error Status Code: $STATUS_CODE"
echo "Output Directory: $OUTPUT_DIR"
echo ""
echo -e "${YELLOW}This test will enable error injection to trigger error rate alerts${NC}"
echo -e "${YELLOW}Alert threshold: 5xx error rate > 2% for 10 minutes${NC}"
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

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is not installed${NC}"
    echo "Please install jq for JSON processing"
    echo "On macOS: brew install jq"
    echo "On Linux: apt-get install jq or yum install jq"
    exit 1
fi

# Check if target is reachable
echo -e "${YELLOW}Checking if target is reachable...${NC}"
if ! curl -s "$TARGET_URL/healthz" > /dev/null; then
    echo -e "${RED}Error: Target $TARGET_URL is not reachable${NC}"
    echo "Make sure the application is running with: docker-compose up -d"
    exit 1
fi

echo -e "${GREEN}Target is reachable. Configuring error injection...${NC}"

# Enable error injection
echo -e "${YELLOW}Enabling error injection (${ERROR_RATE%.*}% error rate)...${NC}"
curl -s -X POST "$TARGET_URL/api/v1/toggles/error-rate" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"enabled\": true, \"rate\": $ERROR_RATE, \"status_code\": $STATUS_CODE}" \
    | jq '.'

if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to enable error injection${NC}"
    echo "Check that ADMIN_TOKEN is correct and the service is running"
    exit 1
fi

echo -e "${GREEN}Error injection enabled successfully${NC}"

# Create vegeta targets file for mixed traffic
cat > "$OUTPUT_DIR/error-test-targets.txt" << EOF
GET $TARGET_URL/api/v1/ping
GET $TARGET_URL/api/v1/work?ms=100&jitter=50
GET $TARGET_URL/api/v1/work?ms=200&jitter=100
GET $TARGET_URL/healthz
GET $TARGET_URL/readyz
EOF

# Function to cleanup error injection on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Disabling error injection...${NC}"
    curl -s -X POST "$TARGET_URL/api/v1/toggles/error-rate" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"enabled\": false, \"rate\": 0.0, \"status_code\": 500}" \
        | jq '.'
    echo -e "${GREEN}Error injection disabled${NC}"
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Run the load test
echo -e "${YELLOW}Running error injection test for $DURATION at $RATE RPS...${NC}"
echo -e "${YELLOW}Expected to trigger HighErrorRate alert after 10 minutes${NC}"

vegeta attack \
    -targets="$OUTPUT_DIR/error-test-targets.txt" \
    -duration="$DURATION" \
    -rate="$RATE" \
    -output="$OUTPUT_DIR/error-test-results.bin"

# Generate reports
echo -e "${YELLOW}Generating reports...${NC}"

# Text report
vegeta report < "$OUTPUT_DIR/error-test-results.bin" > "$OUTPUT_DIR/error-test-report.txt"

# JSON report for programmatic analysis
vegeta report -type=json < "$OUTPUT_DIR/error-test-results.bin" > "$OUTPUT_DIR/error-test-report.json"

# Histogram report
vegeta report -type=hist[0,50ms,100ms,200ms,500ms,1s,2s] < "$OUTPUT_DIR/error-test-results.bin" > "$OUTPUT_DIR/error-test-histogram.txt"

# Plot report (if gnuplot is available)
if command -v gnuplot &> /dev/null; then
    vegeta plot < "$OUTPUT_DIR/error-test-results.bin" > "$OUTPUT_DIR/error-test-plot.html"
    echo -e "${GREEN}Plot generated: $OUTPUT_DIR/error-test-plot.html${NC}"
fi

echo -e "${GREEN}Error injection load test completed!${NC}"
echo ""
echo "Results:"
echo "- Text report: $OUTPUT_DIR/error-test-report.txt"
echo "- JSON report: $OUTPUT_DIR/error-test-report.json"
echo "- Histogram: $OUTPUT_DIR/error-test-histogram.txt"
if command -v gnuplot &> /dev/null; then
    echo "- Plot: $OUTPUT_DIR/error-test-plot.html"
fi
echo ""

# Display summary
echo -e "${YELLOW}Summary:${NC}"
cat "$OUTPUT_DIR/error-test-report.txt"

echo ""
echo -e "${GREEN}Check Grafana dashboards at http://localhost:3000 to see the error metrics${NC}"
echo -e "${GREEN}Check AlertManager at http://localhost:9093 for HighErrorRate alerts${NC}"
echo -e "${GREEN}Check Prometheus at http://localhost:9090 to see alert status${NC}"

# Check if error rate exceeded threshold
SUCCESS_RATE=$(vegeta report -type=json < "$OUTPUT_DIR/error-test-results.bin" | jq -r '.success')
ERROR_RATE_ACTUAL=$(echo "1 - $SUCCESS_RATE" | bc -l)
ERROR_RATE_PERCENT=$(echo "$ERROR_RATE_ACTUAL * 100" | bc -l)
THRESHOLD=2

echo ""
if (( $(echo "$ERROR_RATE_PERCENT > $THRESHOLD" | bc -l) )); then
    echo -e "${GREEN}✓ Error rate (${ERROR_RATE_PERCENT%.*}%) exceeded threshold (${THRESHOLD}%)${NC}"
    echo -e "${GREEN}✓ HighErrorRate alert should be triggered${NC}"
else
    echo -e "${YELLOW}⚠ Error rate (${ERROR_RATE_PERCENT%.*}%) did not exceed threshold (${THRESHOLD}%)${NC}"
    echo -e "${YELLOW}⚠ You may need to increase error injection rate${NC}"
fi

# Error injection will be disabled by the cleanup trap