# Screenshots and Documentation Images

This directory contains screenshots and images used in the project documentation.

## Required Screenshots

To complete the documentation, please add the following screenshots:

### 1. Grafana Dashboard Screenshots

**File**: `grafana-dashboard-overview.png`
- **Description**: Main monitoring dashboard showing all panels
- **How to capture**: 
  1. Start the monitoring stack: `make run`
  2. Generate some traffic: `make load-test-baseline`
  3. Open http://localhost:3000 (admin/admin)
  4. Navigate to the "Monitoring Dashboard"
  5. Take a full-page screenshot

**File**: `grafana-service-overview.png`
- **Description**: Service Overview panel showing service status
- **Size**: Focus on the top row of panels

**File**: `grafana-metrics-panels.png`
- **Description**: Request rate, error rate, and latency panels
- **Size**: Middle section of the dashboard

**File**: `grafana-system-resources.png`
- **Description**: System resources and probe status panels
- **Size**: Bottom section of the dashboard

### 2. Prometheus Screenshots

**File**: `prometheus-targets.png`
- **Description**: Prometheus targets page showing all scrape targets
- **How to capture**:
  1. Open http://localhost:9090/targets
  2. Take screenshot showing all targets in "UP" state

**File**: `prometheus-alerts.png`
- **Description**: Prometheus alerts page showing alert rules
- **How to capture**:
  1. Open http://localhost:9090/alerts
  2. Take screenshot showing alert rules and their states

**File**: `prometheus-graph.png`
- **Description**: Prometheus graph showing a sample query
- **How to capture**:
  1. Open http://localhost:9090/graph
  2. Query: `rate(http_requests_total[5m])`
  3. Take screenshot of the graph

### 3. AlertManager Screenshots

**File**: `alertmanager-overview.png`
- **Description**: AlertManager main page
- **How to capture**:
  1. Open http://localhost:9093
  2. Take screenshot of the main interface

**File**: `alertmanager-firing-alert.png`
- **Description**: AlertManager showing a firing alert
- **How to capture**:
  1. Trigger an alert: `make load-test-errors`
  2. Wait for alert to fire (10+ minutes)
  3. Take screenshot of the firing alert

### 4. Alert Notification Screenshots

**File**: `slack-alert-notification.png`
- **Description**: Example Slack alert notification
- **Requirements**: Configure Slack webhook and capture actual notification

**File**: `discord-alert-notification.png`
- **Description**: Example Discord alert notification
- **Requirements**: Configure Discord webhook and capture actual notification

### 5. Load Testing Screenshots

**File**: `vegeta-load-test-results.png`
- **Description**: Terminal output showing Vegeta load test results
- **How to capture**:
  1. Run: `make load-test-baseline`
  2. Take screenshot of the terminal output

**File**: `load-test-dashboard-impact.png`
- **Description**: Grafana dashboard during load testing
- **How to capture**:
  1. Open Grafana dashboard
  2. Run load test in another terminal
  3. Take screenshot showing metrics during load

### 6. Application Screenshots

**File**: `application-healthz.png`
- **Description**: Browser showing /healthz endpoint response
- **How to capture**:
  1. Open http://localhost:8080/healthz in browser
  2. Take screenshot

**File**: `application-metrics.png`
- **Description**: Browser showing /metrics endpoint (partial)
- **How to capture**:
  1. Open http://localhost:8080/metrics in browser
  2. Take screenshot of first 20-30 lines

## Screenshot Guidelines

### Technical Requirements
- **Format**: PNG preferred, JPG acceptable
- **Resolution**: Minimum 1920x1080 for full-page screenshots
- **Quality**: High quality, readable text
- **File Size**: Optimize for web (< 500KB per image)

### Content Guidelines
- **Clean Interface**: Close unnecessary browser tabs/windows
- **Readable Text**: Ensure all text is clearly visible
- **Relevant Data**: Show meaningful data, not empty states
- **Consistent Timing**: Take screenshots when system is stable

### Naming Convention
- Use lowercase with hyphens: `grafana-dashboard-overview.png`
- Include service name: `prometheus-targets.png`
- Be descriptive: `alertmanager-firing-alert.png`

## Usage in Documentation

Screenshots are referenced in documentation using:

```markdown
![Grafana Dashboard Overview](docs/screenshots/grafana-dashboard-overview.png)
```

## Updating Screenshots

When updating the system:
1. **Version Changes**: Update screenshots when UI changes significantly
2. **Feature Additions**: Add new screenshots for new features
3. **Configuration Changes**: Update if default configurations change
4. **Consistency**: Ensure all screenshots use the same theme/settings

## Alternative: Placeholder Images

If screenshots cannot be provided immediately, use placeholder text:

```markdown
*Screenshot: Grafana Dashboard Overview*
*This would show the main monitoring dashboard with all panels populated with metrics data.*
```

## Automation

Consider creating a script to automatically capture screenshots:

```bash
#!/bin/bash
# screenshot-automation.sh
# Requires: puppeteer or similar headless browser tool

# Start services
make run
sleep 30

# Generate traffic
make load-test-baseline &

# Capture screenshots
# (Implementation would depend on chosen tool)

# Cleanup
make clean
```

This directory structure helps maintain organized documentation assets and provides clear guidelines for contributors.