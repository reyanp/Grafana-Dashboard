# Configuration Guide

This document provides comprehensive information about configuring the monitoring dashboard automation system.

## 📋 Table of Contents

- [Environment Variables](#environment-variables)
- [Service Configuration](#service-configuration)
- [Alert Rules Configuration](#alert-rules-configuration)
- [Dashboard Configuration](#dashboard-configuration)
- [Webhook Configuration](#webhook-configuration)
- [Security Configuration](#security-configuration)
- [Performance Tuning](#performance-tuning)
- [Production Considerations](#production-considerations)

## Environment Variables

### Application Configuration

```bash
# .env file
APP_PORT=8080                    # HTTP server port
ADMIN_TOKEN=changeme             # Bearer token for admin endpoints
LOG_LEVEL=info                   # Logging level: debug, info, warn, error
ENVIRONMENT=development          # Environment identifier
```

**APP_PORT**: The port on which the Go application listens for HTTP requests.
- Default: `8080`
- Valid range: `1024-65535`
- Note: Must match the port in `docker-compose.yml`

**ADMIN_TOKEN**: Bearer token required for accessing admin endpoints like error injection.
- Default: `changeme`
- Security: Use a strong, random token in production
- Usage: `curl -H "Authorization: Bearer $ADMIN_TOKEN" ...`

**LOG_LEVEL**: Controls the verbosity of application logging.
- `debug`: Detailed debugging information
- `info`: General information messages (default)
- `warn`: Warning messages only
- `error`: Error messages only

**ENVIRONMENT**: Environment identifier used in logs and metrics labels.
- Common values: `development`, `staging`, `production`
- Used for filtering and routing in monitoring systems

### Webhook Configuration

```bash
# Alert notification webhooks
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK
```

**SLACK_WEBHOOK_URL**: Slack incoming webhook URL for alert notifications.
- Format: `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX`
- Setup: Create in Slack App settings > Incoming Webhooks
- Optional: Leave empty to disable Slack notifications

**DISCORD_WEBHOOK_URL**: Discord webhook URL for alert notifications.
- Format: `https://discord.com/api/webhooks/123456789/abcdefghijklmnopqrstuvwxyz`
- Setup: Server Settings > Integrations > Webhooks > New Webhook
- Optional: Leave empty to disable Discord notifications

### Monitoring Targets

```bash
# Internal application endpoints (automatically monitored)
BLACKBOX_INTERNAL_TARGET_1=http://go-app:8080/healthz
BLACKBOX_INTERNAL_TARGET_2=http://go-app:8080/readyz
BLACKBOX_INTERNAL_TARGET_3=http://go-app:8080/api/v1/ping

# External endpoints for uptime monitoring
BLACKBOX_EXTERNAL_TARGET_1=https://httpbin.org/status/200
BLACKBOX_EXTERNAL_TARGET_2=https://example.com
BLACKBOX_EXTERNAL_TARGET_3=https://google.com
```

**Internal Targets**: Endpoints within the Docker network that are monitored for availability.
- Use container names (e.g., `go-app`) for internal communication
- Common endpoints: health checks, API endpoints, metrics endpoints

**External Targets**: Public endpoints monitored for external connectivity.
- Use full URLs with protocol (https:// or http://)
- Examples: Customer-facing websites, third-party APIs, CDNs

### Service Configuration

```bash
# Grafana settings
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=admin

# Prometheus settings
PROMETHEUS_RETENTION_TIME=15d
PROMETHEUS_SCRAPE_INTERVAL=5s

# AlertManager settings
ALERT_GROUP_WAIT=30s
ALERT_GROUP_INTERVAL=5m
ALERT_REPEAT_INTERVAL=4h
```

## Service Configuration

### Prometheus Configuration

**File**: `prometheus/prometheus.yml`

```yaml
global:
  scrape_interval: 5s           # How often to scrape targets
  evaluation_interval: 15s      # How often to evaluate rules
  external_labels:
    cluster: 'monitoring-demo'
    replica: 'A'

rule_files:
  - "alerts.yml"

scrape_configs:
  - job_name: 'go-app'
    static_configs:
      - targets: ['go-app:8080']
    scrape_interval: 5s
    metrics_path: /metrics
    
  - job_name: 'node'
    static_configs:
      - targets: ['node_exporter:9100']
    scrape_interval: 5s
    
  - job_name: 'blackbox'
    static_configs:
      - targets: ['blackbox_exporter:9115']
    scrape_interval: 5s

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

**Key Settings**:
- `scrape_interval`: Frequency of metrics collection (5s for demo, 15-30s for production)
- `evaluation_interval`: How often alert rules are evaluated
- `external_labels`: Labels added to all metrics (useful for federation)

### AlertManager Configuration

**File**: `alertmanager/alertmanager.yml`

```yaml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alertmanager@example.org'

route:
  group_by: ['alertname', 'instance']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  receiver: 'default'

receivers:
  - name: 'default'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#alerts'
        username: 'AlertManager'
        title: 'Alert: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
        
    discord_configs:
      - webhook_url: '${DISCORD_WEBHOOK_URL}'
        title: 'Alert: {{ .GroupLabels.alertname }}'
        message: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'instance']
```

**Key Settings**:
- `group_wait`: Wait time before sending first notification for a group
- `group_interval`: Wait time between notifications for the same group
- `repeat_interval`: How often to resend notifications for firing alerts
- `inhibit_rules`: Suppress lower-severity alerts when higher-severity ones are firing

### Grafana Configuration

**Datasource**: `grafana/provisioning/datasources/datasource.yml`

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
```

**Dashboard Provisioning**: `grafana/provisioning/dashboards/dashboard.yml`

```yaml
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
```

## Alert Rules Configuration

**File**: `prometheus/alerts.yml`

### Instance Down Alert

```yaml
groups:
  - name: instance
    rules:
      - alert: InstanceDown
        expr: up{job=~"go-app|node"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Instance {{ $labels.instance }} down"
          description: "{{ $labels.instance }} of job {{ $labels.job }} has been down for more than 2 minutes."
```

**Configuration Options**:
- `expr`: PromQL expression that triggers the alert
- `for`: Duration the condition must be true before firing
- `severity`: Alert severity level (info, warning, critical)
- `summary`: Short description of the alert
- `description`: Detailed description with template variables

### High Error Rate Alert

```yaml
      - alert: HighErrorRate
        expr: |
          (
            rate(http_requests_total{status=~"5.."}[5m]) /
            rate(http_requests_total[5m])
          ) > 0.02
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High error rate on {{ $labels.instance }}"
          description: "Error rate is {{ $value | humanizePercentage }} for the last 10 minutes."
```

### High Latency Alert

```yaml
      - alert: HighLatencyP95
        expr: |
          histogram_quantile(0.95,
            rate(http_request_duration_seconds_bucket[5m])
          ) > 0.5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High latency on {{ $labels.instance }}"
          description: "95th percentile latency is {{ $value }}s for the last 10 minutes."
```

### Probe Failure Alert

```yaml
      - alert: UptimeProbeFail
        expr: probe_success == 0
        for: 3m
        labels:
          severity: warning
        annotations:
          summary: "Probe failed for {{ $labels.instance }}"
          description: "Probe {{ $labels.instance }} has been failing for more than 3 minutes."
```

## Dashboard Configuration

### Panel Configuration Examples

**Request Rate Panel**:
```json
{
  "title": "Request Rate",
  "type": "graph",
  "targets": [
    {
      "expr": "rate(http_requests_total[5m])",
      "legendFormat": "{{ method }} {{ route }}"
    }
  ],
  "yAxes": [
    {
      "label": "Requests/sec",
      "min": 0
    }
  ]
}
```

**Error Rate Panel**:
```json
{
  "title": "Error Rate",
  "type": "stat",
  "targets": [
    {
      "expr": "rate(http_requests_total{status=~\"4..|5..\"}[5m]) / rate(http_requests_total[5m]) * 100",
      "legendFormat": "Error Rate %"
    }
  ],
  "fieldConfig": {
    "defaults": {
      "unit": "percent",
      "thresholds": {
        "steps": [
          {"color": "green", "value": 0},
          {"color": "yellow", "value": 1},
          {"color": "red", "value": 5}
        ]
      }
    }
  }
}
```

## Webhook Configuration

### Slack Webhook Setup

1. **Create Slack App**:
   - Go to https://api.slack.com/apps
   - Click "Create New App"
   - Choose "From scratch"
   - Name your app and select workspace

2. **Enable Incoming Webhooks**:
   - Go to "Incoming Webhooks"
   - Toggle "Activate Incoming Webhooks" to On
   - Click "Add New Webhook to Workspace"
   - Select channel and authorize

3. **Configure Webhook URL**:
   ```bash
   export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
   ```

### Discord Webhook Setup

1. **Create Webhook**:
   - Go to Server Settings > Integrations
   - Click "Create Webhook"
   - Choose channel and customize name/avatar
   - Copy webhook URL

2. **Configure Webhook URL**:
   ```bash
   export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/123456789/abcdefghijklmnopqrstuvwxyz"
   ```

### Custom Webhook Payload

```json
{
  "receiver": "default",
  "status": "firing",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "HighErrorRate",
        "instance": "go-app:8080",
        "job": "go-app",
        "severity": "warning"
      },
      "annotations": {
        "description": "Error rate is 5.2% for the last 10 minutes",
        "summary": "High error rate detected"
      },
      "startsAt": "2023-01-01T12:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus:9090/graph?g0.expr=..."
    }
  ],
  "groupLabels": {
    "alertname": "HighErrorRate"
  },
  "commonLabels": {
    "alertname": "HighErrorRate",
    "instance": "go-app:8080",
    "job": "go-app",
    "severity": "warning"
  },
  "commonAnnotations": {
    "description": "Error rate is 5.2% for the last 10 minutes",
    "summary": "High error rate detected"
  },
  "externalURL": "http://alertmanager:9093",
  "version": "4",
  "groupKey": "{}:{alertname=\"HighErrorRate\"}"
}
```

## Security Configuration

### Authentication

**Grafana Authentication**:
```bash
# Environment variables
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=secure_password_here

# Or in docker-compose.yml
environment:
  - GF_SECURITY_ADMIN_USER=admin
  - GF_SECURITY_ADMIN_PASSWORD=secure_password_here
```

**Application Admin Token**:
```bash
# Generate secure token
ADMIN_TOKEN=$(openssl rand -hex 32)
echo "ADMIN_TOKEN=$ADMIN_TOKEN" >> .env
```

### Network Security

**Docker Network Isolation**:
```yaml
# docker-compose.yml
networks:
  monitoring:
    driver: bridge
    internal: false  # Set to true to isolate from external network
```

**Port Exposure**:
```yaml
# Only expose necessary ports
services:
  go-app:
    ports:
      - "8080:8080"  # Application
  grafana:
    ports:
      - "3000:3000"  # Dashboard access
  # Don't expose internal ports like Prometheus, AlertManager
```

### TLS Configuration

**Grafana HTTPS**:
```yaml
# docker-compose.yml
grafana:
  environment:
    - GF_SERVER_PROTOCOL=https
    - GF_SERVER_CERT_FILE=/etc/ssl/certs/grafana.crt
    - GF_SERVER_CERT_KEY=/etc/ssl/private/grafana.key
  volumes:
    - ./certs:/etc/ssl/certs:ro
    - ./private:/etc/ssl/private:ro
```

## Performance Tuning

### Prometheus Optimization

**Storage Configuration**:
```yaml
# docker-compose.yml
prometheus:
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.path=/prometheus'
    - '--storage.tsdb.retention.time=15d'
    - '--storage.tsdb.retention.size=10GB'
    - '--web.console.libraries=/etc/prometheus/console_libraries'
    - '--web.console.templates=/etc/prometheus/consoles'
    - '--web.enable-lifecycle'
    - '--storage.tsdb.wal-compression'
```

**Scrape Interval Tuning**:
```yaml
# prometheus.yml
global:
  scrape_interval: 15s      # Production: 15-30s, Demo: 5s
  evaluation_interval: 15s  # Should match or be multiple of scrape_interval

scrape_configs:
  - job_name: 'high-frequency'
    scrape_interval: 5s     # Critical services
  - job_name: 'normal'
    scrape_interval: 15s    # Standard services
  - job_name: 'low-frequency'
    scrape_interval: 60s    # Batch jobs, less critical
```

### Resource Limits

**Container Resource Limits**:
```yaml
# docker-compose.yml
services:
  go-app:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
          
  prometheus:
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '1.0'
        reservations:
          memory: 1G
          cpus: '0.5'
```

### Grafana Optimization

**Dashboard Performance**:
```json
{
  "refresh": "30s",           // Reduce refresh frequency
  "time": {
    "from": "now-1h",         // Shorter time ranges
    "to": "now"
  },
  "templating": {
    "list": [
      {
        "query": "label_values(job)",
        "refresh": 2,         // Refresh on dashboard load and time range change
        "sort": 1             // Sort alphabetically
      }
    ]
  }
}
```

## Production Considerations

### High Availability

**Prometheus HA Setup**:
```yaml
# Multiple Prometheus instances
prometheus-1:
  image: prom/prometheus
  external_labels:
    replica: 'A'
    
prometheus-2:
  image: prom/prometheus
  external_labels:
    replica: 'B'
```

**Grafana HA Setup**:
```yaml
grafana:
  environment:
    - GF_DATABASE_TYPE=postgres
    - GF_DATABASE_HOST=postgres:5432
    - GF_DATABASE_NAME=grafana
    - GF_DATABASE_USER=grafana
    - GF_DATABASE_PASSWORD=password
```

### Backup and Recovery

**Prometheus Data Backup**:
```bash
# Backup script
#!/bin/bash
docker-compose exec prometheus \
  tar -czf /prometheus/backup-$(date +%Y%m%d).tar.gz /prometheus/data

# Copy backup out of container
docker cp prometheus_container:/prometheus/backup-$(date +%Y%m%d).tar.gz ./backups/
```

**Grafana Configuration Backup**:
```bash
# Export dashboards
curl -u admin:admin http://localhost:3000/api/search | \
  jq -r '.[] | select(.type == "dash-db") | .uid' | \
  xargs -I {} curl -u admin:admin http://localhost:3000/api/dashboards/uid/{} > dashboard-{}.json
```

### Monitoring the Monitoring

**Meta-monitoring Metrics**:
```yaml
# Additional scrape configs for monitoring the monitoring stack
scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
      
  - job_name: 'grafana'
    static_configs:
      - targets: ['grafana:3000']
    metrics_path: /metrics
    
  - job_name: 'alertmanager'
    static_configs:
      - targets: ['alertmanager:9093']
```

**Health Check Alerts**:
```yaml
# Alert if monitoring components are down
- alert: PrometheusDown
  expr: up{job="prometheus"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Prometheus is down"
    
- alert: GrafanaDown
  expr: up{job="grafana"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Grafana is down"
```

This configuration guide provides comprehensive information for customizing and optimizing the monitoring system for different environments and use cases.