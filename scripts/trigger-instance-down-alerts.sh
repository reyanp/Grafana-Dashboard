#!/bin/bash

# Instance Down Alert Trigger Script
# Stops service containers to trigger instance down and probe failure alerts
# Requirements: 7.3

set -e

# Configuration
CONTAINER_NAME="monitoring-dashboard-automation-go-app-1"
DOWN_DURATION="3m"  # Keep service down longer than alert threshold (2m)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Instance Down Alert Trigger Test${NC}"
echo "Container: $CONTAINER_NAME"
echo "Down Duration: $DOWN_DURATION"
echo ""
echo -e "${YELLOW}This test will stop the Go application container to trigger instance down alerts${NC}"
echo -e "${YELLOW}Alert thresholds:${NC}"
echo -e "${YELLOW}  - InstanceDown: service down for 2 minutes${NC}"
echo -e "${YELLOW}  - UptimeProbeFail: probe failure for 3 minutes${NC}"
echo ""

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null && ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: docker-compose or docker is not installed${NC}"
    exit 1
fi

# Check if container exists and is running
echo -e "${YELLOW}Checking container status...${NC}"
if ! docker ps | grep -q "$CONTAINER_NAME"; then
    echo -e "${RED}Error: Container $CONTAINER_NAME is not running${NC}"
    echo "Make sure the application is running with: docker-compose up -d"
    exit 1
fi

echo -e "${GREEN}Container is running. Preparing to stop it...${NC}"

# Function to restart container on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Restarting container...${NC}"
    docker-compose up -d go-app
    
    # Wait for container to be healthy
    echo -e "${YELLOW}Waiting for container to be healthy...${NC}"
    sleep 10
    
    # Check if container is responding
    for i in {1..30}; do
        if curl -s http://localhost:8080/healthz > /dev/null 2>&1; then
            echo -e "${GREEN}Container is healthy and responding${NC}"
            break
        fi
        echo -n "."
        sleep 2
    done
    echo ""
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Stop the container
echo -e "${YELLOW}Stopping container $CONTAINER_NAME...${NC}"
docker stop "$CONTAINER_NAME"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}Container stopped successfully${NC}"
else
    echo -e "${RED}Failed to stop container${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Container will remain down for $DOWN_DURATION${NC}"
echo -e "${YELLOW}Expected alerts:${NC}"
echo -e "${YELLOW}  - InstanceDown alert should fire after 2 minutes${NC}"
echo -e "${YELLOW}  - UptimeProbeFail alert should fire after 3 minutes${NC}"
echo ""
echo -e "${GREEN}Monitor the following while container is down:${NC}"
echo "- Grafana dashboards: http://localhost:3000"
echo "- AlertManager: http://localhost:9093"
echo "- Prometheus: http://localhost:9090"
echo ""

# Show countdown
echo -e "${YELLOW}Countdown:${NC}"
SECONDS_TOTAL=$(echo "$DOWN_DURATION" | sed 's/m/ * 60/' | bc)
for ((i=SECONDS_TOTAL; i>0; i--)); do
    MINUTES=$((i / 60))
    SECONDS=$((i % 60))
    printf "\r${YELLOW}Time remaining: %02d:%02d${NC}" $MINUTES $SECONDS
    sleep 1
done
echo ""

echo -e "${GREEN}Down period completed. Container will be restarted by cleanup function.${NC}"

# Check alert status in Prometheus
echo -e "${YELLOW}Checking alert status in Prometheus...${NC}"
if command -v curl &> /dev/null; then
    echo ""
    echo -e "${YELLOW}Active alerts:${NC}"
    curl -s "http://localhost:9090/api/v1/alerts" | jq -r '.data.alerts[] | select(.state == "firing") | "- \(.labels.alertname): \(.annotations.summary // .annotations.description)"' 2>/dev/null || echo "Could not fetch alerts (jq may not be installed)"
    
    echo ""
    echo -e "${YELLOW}Alert rules status:${NC}"
    curl -s "http://localhost:9090/api/v1/rules" | jq -r '.data.groups[].rules[] | select(.type == "alerting") | "- \(.name): \(.state)"' 2>/dev/null || echo "Could not fetch rules (jq may not be installed)"
fi

echo ""
echo -e "${GREEN}Instance down test completed!${NC}"
echo ""
echo -e "${GREEN}What to verify:${NC}"
echo "1. InstanceDown alert should have fired in AlertManager"
echo "2. UptimeProbeFail alert should have fired in AlertManager"
echo "3. Grafana dashboards should show the service as down during the test period"
echo "4. Webhook notifications should have been sent (if configured)"
echo "5. Alerts should resolve after the container is restarted"
echo ""
echo -e "${GREEN}Check AlertManager at http://localhost:9093 to see alert history${NC}"