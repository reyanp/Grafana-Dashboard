# Monitoring System Demo Guide

This guide provides step-by-step instructions for demonstrating the complete monitoring system capabilities, including load testing, alert triggering, and system recovery verification.

## Prerequisites

### Required Tools

1. **Vegeta** - HTTP load testing tool
   ```bash
   # macOS
   brew install vegeta
   
   # Linux - Download from GitHub releases
   wget https://github.com/tsenart/vegeta/releases/download/v12.8.4/vegeta_12.8.4_linux_amd64.tar.gz
   tar -xzf vegeta_12.8.4_linux_amd64.tar.gz
   sudo mv vegeta /usr/local/bin/
   ```

2. **jq** - JSON processor (for parsing results)
   ```bash
   # macOS
   brew install jq
   
   # Linux
   sudo apt-get install jq  # Ubuntu/Debian
   sudo yum install jq      # CentOS/RHEL
   ```

3. **bc** - Basic calculator (for mathematical operations)
   ```bash
   # Usually pre-installed, but if needed:
   # macOS
   brew install bc
   
   # Linux
   sudo apt-get install bc  # Ubuntu/Debian
   sudo yum install bc      # CentOS/RHEL
   ```

### System Setup

1. **Start the monitoring stack**:
   ```bash
   docker-compose up -d
   ```

2. **Verify all services are running**:
   ```bash
   docker-compose ps
   ```

3. **Wait for services to be ready** (about 30 seconds):
   - Go application: http://localhost:8080/healthz
   - Grafana: http://localhost:3000 (admin/admin)
   - Prometheus: http://localhost:9090
   - AlertManager: http://localhost:9093

## Demo Scripts Overview

### 1. Complete Demo Scenario (`scripts/demo-scenario.sh`)

**Purpose**: Runs through all monitoring scenarios in a guided, interactive manner.

**Usage**:
```bash
chmod +x scripts/demo-scenario.sh
./scripts/demo-scenario.sh
```

**What it does**:
- Performs system health checks
- Runs baseline load testing
- Triggers latency alerts
- Triggers error rate alerts
- Triggers instance down alerts
- Verifies system recovery

**Duration**: 30-45 minutes (depending on which tests you choose to run)

### 2. Baseline Load Test (`scripts/load-test-baseline.sh`)

**Purpose**: Generates normal traffic to establish baseline metrics.

**Usage**:
```bash
chmod +x scripts/load-test-baseline.sh
./scripts/load-test-baseline.sh
```

**Configuration**:
- Duration: 5 minutes
- Rate: 50 RPS
- Endpoints: Mixed traffic to `/ping`, `/work`, `/healthz`, `/readyz`

**Expected Results**:
- P95 latency < 300ms
- Error rate < 1%
- Successful metrics collection in Prometheus
- Dashboard population in Grafana

### 3. Latency Spike Test (`scripts/load-test-latency-spike.sh`)

**Purpose**: Generates high-latency requests to trigger HighLatencyP95 alerts.

**Usage**:
```bash
chmod +x scripts/load-test-latency-spike.sh
./scripts/load-test-latency-spike.sh
```

**Configuration**:
- Duration: 12 minutes
- Rate: 30 RPS
- Latency: 600-1200ms with jitter

**Expected Results**:
- P95 latency > 500ms
- HighLatencyP95 alert fires after 10 minutes
- Alert visible in AlertManager and Grafana

### 4. Error Injection Test (`scripts/trigger-error-alerts.sh`)

**Purpose**: Uses error injection to trigger HighErrorRate alerts.

**Usage**:
```bash
# Set admin token if different from default
export ADMIN_TOKEN="your-admin-token"
chmod +x scripts/trigger-error-alerts.sh
./scripts/trigger-error-alerts.sh
```

**Configuration**:
- Duration: 12 minutes
- Rate: 40 RPS
- Error rate: 5% (above 2% threshold)
- Error status: 500

**Expected Results**:
- Error rate > 2%
- HighErrorRate alert fires after 10 minutes
- Error injection automatically disabled after test

### 5. Instance Down Test (`scripts/trigger-instance-down-alerts.sh`)

**Purpose**: Stops the application container to trigger instance down alerts.

**Usage**:
```bash
chmod +x scripts/trigger-instance-down-alerts.sh
./scripts/trigger-instance-down-alerts.sh
```

**Configuration**:
- Down duration: 3 minutes
- Container: monitoring-dashboard-automation-go-app-1

**Expected Results**:
- InstanceDown alert fires after 2 minutes
- UptimeProbeFail alert fires after 3 minutes
- Container automatically restarted after test

## Alert Thresholds

The system is configured with the following alert thresholds:

| Alert | Condition | Duration | Description |
|-------|-----------|----------|-------------|
| InstanceDown | `up{job=~"go-app\|node"} == 0` | 2 minutes | Service instance is down |
| HighErrorRate | `5xx error rate > 2%` | 10 minutes | High rate of server errors |
| HighLatencyP95 | `p95 latency > 500ms` | 10 minutes | High response latency |
| UptimeProbeFail | `probe_success == 0` | 3 minutes | Blackbox probe failure |

## Monitoring URLs

During the demo, monitor these URLs to see the system in action:

- **Grafana Dashboards**: http://localhost:3000
  - Username: admin
  - Password: admin
  - Look for: Service Overview, Request Rate, Error Rate, Latency panels

- **Prometheus**: http://localhost:9090
  - Check: Targets status, Alert rules, Query metrics

- **AlertManager**: http://localhost:9093
  - Check: Active alerts, Alert history, Silences

- **Application**: http://localhost:8080
  - Endpoints: `/healthz`, `/readyz`, `/metrics`, `/api/v1/ping`

## Step-by-Step Demo Instructions

### Quick Demo (15 minutes)

1. **Start the system**:
   ```bash
   docker-compose up -d
   ```

2. **Run baseline test**:
   ```bash
   ./scripts/load-test-baseline.sh
   ```

3. **Check Grafana dashboards** - verify metrics are flowing

4. **Trigger one alert** (choose one):
   ```bash
   # Option A: Quick error test (5 minutes)
   timeout 5m ./scripts/trigger-error-alerts.sh
   
   # Option B: Quick instance down test (3 minutes)
   ./scripts/trigger-instance-down-alerts.sh
   ```

5. **Verify alert in AlertManager** and **check recovery**

### Full Demo (45 minutes)

1. **Run the complete demo scenario**:
   ```bash
   ./scripts/demo-scenario.sh
   ```

2. **Follow the interactive prompts** - the script will guide you through each step

3. **Monitor all URLs** as indicated by the script

## Troubleshooting

### Common Issues

1. **"vegeta: command not found"**
   - Install vegeta using the instructions in Prerequisites

2. **"Target is not reachable"**
   - Ensure Docker Compose stack is running: `docker-compose up -d`
   - Wait 30 seconds for services to start
   - Check service health: `curl http://localhost:8080/healthz`

3. **"Failed to enable error injection"**
   - Check admin token: `export ADMIN_TOKEN="changeme"`
   - Verify application is responding: `curl http://localhost:8080/healthz`

4. **"Container not found"**
   - Check container name: `docker ps | grep go-app`
   - Adjust container name in script if needed

5. **Alerts not firing**
   - Check Prometheus targets: http://localhost:9090/targets
   - Verify alert rules: http://localhost:9090/rules
   - Ensure test duration exceeds alert thresholds

### Verification Commands

```bash
# Check all services are healthy
curl -s http://localhost:8080/healthz
curl -s http://localhost:9090/-/healthy
curl -s http://localhost:3000/api/health
curl -s http://localhost:9093/-/healthy

# Check Prometheus targets
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Check active alerts
curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.state == "firing") | .labels.alertname'

# Check metrics are being collected
curl -s http://localhost:8080/metrics | grep http_requests_total
```

## Results Analysis

### Load Test Results

Results are saved in `./load-test-results/` directory:

- `*-report.txt` - Human-readable summary
- `*-report.json` - Machine-readable results
- `*-histogram.txt` - Latency distribution
- `*-plot.html` - Visual plot (if gnuplot available)

### Key Metrics to Monitor

1. **Request Rate**: Should match test configuration (RPS)
2. **Error Rate**: Should be <1% for baseline, ~5% for error injection
3. **Latency**: P95 should be <300ms for baseline, >500ms for spike test
4. **Success Rate**: Should be >99% for baseline, ~95% for error injection

### Alert Verification

Check that alerts fire and resolve correctly:

1. **Alert Firing**: Visible in AlertManager within expected timeframe
2. **Alert Resolution**: Alerts resolve when conditions return to normal
3. **Webhook Delivery**: Notifications sent to configured channels (if set up)
4. **Dashboard Updates**: Grafana panels reflect alert conditions

## Webhook Configuration (Optional)

To receive actual notifications during the demo:

1. **Set up Slack webhook**:
   ```bash
   export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
   ```

2. **Set up Discord webhook**:
   ```bash
   export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK"
   ```

3. **Restart AlertManager**:
   ```bash
   docker-compose restart alertmanager
   ```

## Performance Expectations

### System Requirements

- **CPU**: 2+ cores recommended
- **Memory**: 4GB+ RAM recommended
- **Disk**: 1GB+ free space for logs and metrics

### Expected Performance

- **Application**: Handles 200+ RPS on laptop hardware
- **Prometheus**: Scrapes metrics every 5 seconds
- **Grafana**: Dashboard updates every 5-10 seconds
- **AlertManager**: Evaluates rules every 15 seconds

## Cleanup

After the demo:

```bash
# Stop all services
docker-compose down

# Remove volumes (optional - removes all metrics data)
docker-compose down -v

# Clean up test results
rm -rf ./load-test-results/
```

## Next Steps

After completing the demo:

1. **Explore Grafana dashboards** - customize panels and queries
2. **Modify alert thresholds** - adjust rules in `prometheus/alerts.yml`
3. **Add custom metrics** - instrument your own applications
4. **Set up real webhooks** - configure Slack/Discord notifications
5. **Scale the system** - add more application instances
6. **Monitor real services** - adapt configuration for production use

This monitoring system demonstrates production-ready observability patterns that can be adapted for real-world applications and infrastructure.