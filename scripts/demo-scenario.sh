#!/bin/bash

# Complete Demo Scenario Script
# Runs through all monitoring scenarios to demonstrate the complete system
# Requirements: 7.4, 7.5

set -e

# Configuration
TARGET_URL="http://localhost:8080"
ADMIN_TOKEN="${ADMIN_TOKEN:-changeme}"
OUTPUT_DIR="./load-test-results"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print section headers
print_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# Function to wait for user input
wait_for_user() {
    echo -e "${YELLOW}Press Enter to continue to the next step...${NC}"
    read -r
}

# Function to check service health
check_service_health() {
    echo -e "${YELLOW}Checking service health...${NC}"
    
    # Check main application
    if curl -s "$TARGET_URL/healthz" > /dev/null; then
        echo -e "${GREEN}✓ Go application is healthy${NC}"
    else
        echo -e "${RED}✗ Go application is not responding${NC}"
        return 1
    fi
    
    # Check Prometheus
    if curl -s "http://localhost:9090/-/healthy" > /dev/null; then
        echo -e "${GREEN}✓ Prometheus is healthy${NC}"
    else
        echo -e "${RED}✗ Prometheus is not responding${NC}"
        return 1
    fi
    
    # Check Grafana
    if curl -s "http://localhost:3000/api/health" > /dev/null; then
        echo -e "${GREEN}✓ Grafana is healthy${NC}"
    else
        echo -e "${RED}✗ Grafana is not responding${NC}"
        return 1
    fi
    
    # Check AlertManager
    if curl -s "http://localhost:9093/-/healthy" > /dev/null; then
        echo -e "${GREEN}✓ AlertManager is healthy${NC}"
    else
        echo -e "${RED}✗ AlertManager is not responding${NC}"
        return 1
    fi
    
    echo -e "${GREEN}All services are healthy!${NC}"
    return 0
}

# Function to show URLs
show_urls() {
    echo -e "${GREEN}Access the monitoring stack:${NC}"
    echo "- Application: http://localhost:8080"
    echo "- Grafana: http://localhost:3000 (admin/admin)"
    echo "- Prometheus: http://localhost:9090"
    echo "- AlertManager: http://localhost:9093"
    echo ""
}

# Main demo script
main() {
    print_section "MONITORING SYSTEM DEMO"
    
    echo -e "${GREEN}This demo will walk you through all monitoring capabilities:${NC}"
    echo "1. System health check"
    echo "2. Baseline load testing"
    echo "3. Latency spike testing (triggers alerts)"
    echo "4. Error injection testing (triggers alerts)"
    echo "5. Instance down testing (triggers alerts)"
    echo "6. System recovery verification"
    echo ""
    
    show_urls
    wait_for_user
    
    # Step 1: Health Check
    print_section "STEP 1: SYSTEM HEALTH CHECK"
    
    if ! check_service_health; then
        echo -e "${RED}System health check failed. Please ensure all services are running:${NC}"
        echo "docker-compose up -d"
        exit 1
    fi
    
    echo ""
    echo -e "${GREEN}Open Grafana (http://localhost:3000) to see the dashboards${NC}"
    echo -e "${GREEN}You should see all services showing as 'up' in the Service Overview panel${NC}"
    wait_for_user
    
    # Step 2: Baseline Load Test
    print_section "STEP 2: BASELINE LOAD TESTING"
    
    echo -e "${YELLOW}Running baseline load test to establish normal metrics...${NC}"
    echo "This will generate normal traffic for 5 minutes at 50 RPS"
    echo ""
    
    if [ -f "./scripts/load-test-baseline.sh" ]; then
        chmod +x "./scripts/load-test-baseline.sh"
        "./scripts/load-test-baseline.sh"
    else
        echo -e "${RED}Baseline load test script not found${NC}"
        exit 1
    fi
    
    echo ""
    echo -e "${GREEN}Baseline test completed. Check Grafana to see:${NC}"
    echo "- Request rate metrics"
    echo "- Response time percentiles"
    echo "- Error rates (should be near 0%)"
    wait_for_user
    
    # Step 3: Latency Spike Test
    print_section "STEP 3: LATENCY SPIKE TESTING"
    
    echo -e "${YELLOW}Running latency spike test to trigger HighLatencyP95 alerts...${NC}"
    echo "This will generate high-latency requests for 12 minutes"
    echo "Expected: HighLatencyP95 alert should fire after 10 minutes"
    echo ""
    echo -e "${RED}This test takes 12 minutes. You can monitor progress in Grafana.${NC}"
    echo -e "${YELLOW}Do you want to run the latency spike test? (y/n)${NC}"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        if [ -f "./scripts/load-test-latency-spike.sh" ]; then
            chmod +x "./scripts/load-test-latency-spike.sh"
            "./scripts/load-test-latency-spike.sh"
        else
            echo -e "${RED}Latency spike test script not found${NC}"
            exit 1
        fi
        
        echo ""
        echo -e "${GREEN}Latency spike test completed. Check:${NC}"
        echo "- Grafana latency panels (should show high p95 latency)"
        echo "- AlertManager (http://localhost:9093) for HighLatencyP95 alerts"
        echo "- Webhook notifications (if configured)"
    else
        echo -e "${YELLOW}Skipping latency spike test${NC}"
    fi
    
    wait_for_user
    
    # Step 4: Error Injection Test
    print_section "STEP 4: ERROR INJECTION TESTING"
    
    echo -e "${YELLOW}Running error injection test to trigger HighErrorRate alerts...${NC}"
    echo "This will enable 5% error rate for 12 minutes"
    echo "Expected: HighErrorRate alert should fire after 10 minutes"
    echo ""
    echo -e "${RED}This test takes 12 minutes. You can monitor progress in Grafana.${NC}"
    echo -e "${YELLOW}Do you want to run the error injection test? (y/n)${NC}"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        if [ -f "./scripts/trigger-error-alerts.sh" ]; then
            chmod +x "./scripts/trigger-error-alerts.sh"
            "./scripts/trigger-error-alerts.sh"
        else
            echo -e "${RED}Error injection test script not found${NC}"
            exit 1
        fi
        
        echo ""
        echo -e "${GREEN}Error injection test completed. Check:${NC}"
        echo "- Grafana error rate panels (should show ~5% error rate)"
        echo "- AlertManager (http://localhost:9093) for HighErrorRate alerts"
        echo "- Webhook notifications (if configured)"
    else
        echo -e "${YELLOW}Skipping error injection test${NC}"
    fi
    
    wait_for_user
    
    # Step 5: Instance Down Test
    print_section "STEP 5: INSTANCE DOWN TESTING"
    
    echo -e "${YELLOW}Running instance down test to trigger InstanceDown and UptimeProbeFail alerts...${NC}"
    echo "This will stop the Go application container for 3 minutes"
    echo "Expected alerts:"
    echo "- InstanceDown alert should fire after 2 minutes"
    echo "- UptimeProbeFail alert should fire after 3 minutes"
    echo ""
    echo -e "${RED}This will temporarily stop the application!${NC}"
    echo -e "${YELLOW}Do you want to run the instance down test? (y/n)${NC}"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        if [ -f "./scripts/trigger-instance-down-alerts.sh" ]; then
            chmod +x "./scripts/trigger-instance-down-alerts.sh"
            "./scripts/trigger-instance-down-alerts.sh"
        else
            echo -e "${RED}Instance down test script not found${NC}"
            exit 1
        fi
        
        echo ""
        echo -e "${GREEN}Instance down test completed. Check:${NC}"
        echo "- Grafana service overview (should show service as down during test)"
        echo "- AlertManager (http://localhost:9093) for InstanceDown and UptimeProbeFail alerts"
        echo "- Webhook notifications (if configured)"
    else
        echo -e "${YELLOW}Skipping instance down test${NC}"
    fi
    
    wait_for_user
    
    # Step 6: System Recovery Verification
    print_section "STEP 6: SYSTEM RECOVERY VERIFICATION"
    
    echo -e "${YELLOW}Verifying system recovery...${NC}"
    
    if check_service_health; then
        echo ""
        echo -e "${GREEN}✓ All services have recovered successfully${NC}"
        
        # Wait a bit for metrics to stabilize
        echo -e "${YELLOW}Waiting 30 seconds for metrics to stabilize...${NC}"
        sleep 30
        
        echo ""
        echo -e "${GREEN}Final verification - check that:${NC}"
        echo "1. All alerts have resolved in AlertManager"
        echo "2. Grafana dashboards show normal metrics"
        echo "3. Service overview shows all services as 'up'"
        echo "4. Error rates have returned to normal (near 0%)"
        echo "5. Latency has returned to baseline levels"
        
    else
        echo -e "${RED}System recovery failed. Some services are not healthy.${NC}"
        echo "You may need to restart services manually:"
        echo "docker-compose down && docker-compose up -d"
    fi
    
    print_section "DEMO COMPLETED"
    
    echo -e "${GREEN}Monitoring system demo completed successfully!${NC}"
    echo ""
    echo -e "${GREEN}Summary of what was demonstrated:${NC}"
    echo "✓ Baseline monitoring and metrics collection"
    echo "✓ Latency alerting (HighLatencyP95)"
    echo "✓ Error rate alerting (HighErrorRate)"
    echo "✓ Instance down alerting (InstanceDown, UptimeProbeFail)"
    echo "✓ System recovery and alert resolution"
    echo ""
    echo -e "${GREEN}Key URLs for continued exploration:${NC}"
    show_urls
    echo ""
    echo -e "${GREEN}Load test results are available in: $OUTPUT_DIR${NC}"
    echo ""
    echo -e "${YELLOW}Thank you for trying the monitoring system demo!${NC}"
}

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Run main demo
main