# Troubleshooting Guide

This guide provides solutions for common issues encountered when running the monitoring dashboard automation system.

## 🚨 Quick Diagnostics

### System Health Check Script

```bash
#!/bin/bash
# Save as check-system-health.sh and run: chmod +x check-system-health.sh && ./check-system-health.sh

echo "=== Monitoring System Health Check ==="
echo

# Check Docker
echo "1. Checking Docker..."
if command -v docker >/dev/null 2>&1; then
    echo "✅ Docker is installed: $(docker --version)"
    if docker info >/dev/null 2>&1; then
        echo "✅ Docker daemon is running"
    else
        echo "❌ Docker daemon is not running"
        exit 1
    fi
else
    echo "❌ Docker is not installed"
    exit 1
fi

# Check Docker Compose
echo
echo "2. Checking Docker Compose..."
if command -v docker-compose >/dev/null 2>&1; then
    echo "✅ Docker Compose is installed: $(docker-compose --version)"
else
    echo "❌ Docker Compose is not installed"
    exit 1
fi

# Check ports
echo
echo "3. Checking port availability..."
ports=(3000 8080 9090 9093 9100 9115)
for port in "${ports[@]}"; do
    if lsof -i :$port >/dev/null 2>&1; then
        echo "⚠️  Port $port is in use"
    else
        echo "✅ Port $port is available"
    fi
done

# Check running containers
echo
echo "4. Checking running containers..."
if docker-compose ps | grep -q "Up"; then
    echo "✅ Some containers are running:"
    docker-compose ps
else
    echo "ℹ️  No containers are currently running"
fi

# Check service health
echo
echo "5. Checking service health..."
services=(
    "http://localhost:8080/healthz|Application"
    "http://localhost:9090/-/healthy|Prometheus"
    "http://localhost:3000/api/health|Grafana"
    "http://localhost:9093/-/healthy|AlertManager"
)

for service in "${services[@]}"; do
    url=$(echo $service | cut -d'|' -f1)
    name=$(echo $service | cut -d'|' -f2)
    
    if curl -f -s "$url" >/dev/null 2>&1; then
        echo "✅ $name is healthy"
    else
        echo "❌ $name is not responding"
    fi
done

echo
echo "=== Health check complete ==="
```

## 🔧 Common Issues and Solutions

### 1. Services Not Starting

#### Issue: Docker Compose fails to start services

**Symptoms**:
- `docker-compose up -d` fails
- Containers exit immediately
- Port binding errors

**Solutions**:

```bash
# Check for port conflicts
netstat -tulpn | grep -E ':(3000|8080|9090|9093|9100|9115)'

# Kill processes using required ports
sudo lsof -ti:8080 | xargs kill -9

# Check Docker daemon status
sudo systemctl status docker

# Restart Docker daemon
sudo systemctl restart docker

# Clean up old containers and networks
docker-compose down -v
docker system prune -f

# Start with verbose logging
docker-compose up --verbose
```

#### Issue: Insufficient resources

**Symptoms**:
- Containers keep restarting
- Out of memory errors
- Slow performance

**Solutions**:

```bash
# Check available resources
free -h
df -h

# Increase Docker memory limit (Docker Desktop)
# Settings > Resources > Memory > 4GB+

# Check container resource usage
docker stats

# Reduce resource usage by disabling non-essential services temporarily
docker-compose up -d go-app prometheus grafana
```

### 2. Metrics Collection Issues

#### Issue: Prometheus not scraping targets

**Symptoms**:
- Empty dashboards in Grafana
- Targets showing as "DOWN" in Prometheus
- No metrics data

**Diagnostic Commands**:
```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health, lastError: .lastError}'

# Check application metrics endpoint
curl http://localhost:8080/metrics | head -20

# Verify Prometheus configuration
docker-compose exec prometheus cat /etc/prometheus/prometheus.yml
```

**Solutions**:

```bash
# Restart Prometheus
docker-compose restart prometheus

# Check network connectivity between containers
docker-compose exec prometheus ping go-app

# Verify service discovery
docker-compose exec prometheus nslookup go-app

# Check Prometheus logs
docker-compose logs prometheus | grep -i error
```

#### Issue: Application not exposing metrics

**Symptoms**:
- `/metrics` endpoint returns 404
- Prometheus shows scrape errors
- No application-specific metrics

**Solutions**:

```bash
# Check if application is running
curl http://localhost:8080/healthz

# Verify metrics endpoint
curl -v http://localhost:8080/metrics

# Check application logs
docker-compose logs go-app | grep -i metric

# Restart application
docker-compose restart go-app
```

### 3. Grafana Dashboard Issues

#### Issue: Grafana shows "No data" or empty panels

**Symptoms**:
- Dashboard panels are empty
- "No data" messages
- Query errors

**Diagnostic Commands**:
```bash
# Check Grafana datasource
curl -u admin:admin http://localhost:3000/api/datasources

# Test Prometheus connection from Grafana
curl -u admin:admin "http://localhost:3000/api/datasources/proxy/1/api/v1/query?query=up"

# Check dashboard provisioning
docker-compose exec grafana ls -la /etc/grafana/provisioning/dashboards/
```

**Solutions**:

```bash
# Restart Grafana
docker-compose restart grafana

# Re-provision datasources
docker-compose exec grafana rm -rf /var/lib/grafana/grafana.db
docker-compose restart grafana

# Check Grafana logs
docker-compose logs grafana | grep -i error

# Manually configure datasource
# Go to http://localhost:3000/datasources
# Add Prometheus: http://prometheus:9090
```

#### Issue: Dashboard not loading or corrupted

**Solutions**:

```bash
# Reset Grafana data
docker-compose down
docker volume rm monitoring-dashboard-automation_grafana_data
docker-compose up -d

# Re-import dashboard manually
# 1. Go to http://localhost:3000/dashboard/import
# 2. Upload grafana/provisioning/dashboards/monitoring-dashboard.json

# Check dashboard JSON syntax
jq . grafana/provisioning/dashboards/monitoring-dashboard.json
```

### 4. Alert System Problems

#### Issue: Alerts not firing

**Symptoms**:
- No alerts in AlertManager
- Expected alerts don't trigger
- Alert rules not loading

**Diagnostic Commands**:
```bash
# Check alert rules in Prometheus
curl http://localhost:9090/api/v1/rules | jq '.data.groups[].rules[] | select(.type=="alerting") | {name: .name, state: .state}'

# Check active alerts
curl http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | {name: .labels.alertname, state: .state}'

# Check AlertManager status
curl http://localhost:9093/api/v1/status
```

**Solutions**:

```bash
# Verify alert rules syntax
docker-compose exec prometheus promtool check rules /etc/prometheus/alerts.yml

# Restart Prometheus to reload rules
docker-compose restart prometheus

# Check Prometheus logs for rule errors
docker-compose logs prometheus | grep -i alert

# Manually trigger alert condition
# For error rate alert:
curl -X POST -H "Authorization: Bearer changeme" \
     -H "Content-Type: application/json" \
     -d '{"enabled": true, "rate": 0.5, "status_code": 500}' \
     http://localhost:8080/api/v1/toggles/error-rate
```

#### Issue: Webhook notifications not working

**Symptoms**:
- Alerts fire but no notifications received
- Webhook delivery failures
- AlertManager shows routing errors

**Solutions**:

```bash
# Check AlertManager configuration
docker-compose exec alertmanager cat /etc/alertmanager/alertmanager.yml

# Test webhook URLs manually
curl -X POST -H "Content-Type: application/json" \
     -d '{"text": "Test notification"}' \
     "$SLACK_WEBHOOK_URL"

# Check AlertManager logs
docker-compose logs alertmanager | grep -i webhook

# Verify environment variables
docker-compose exec alertmanager env | grep WEBHOOK
```

### 5. Load Testing Issues

#### Issue: Vegeta not installed or not working

**Symptoms**:
- `vegeta: command not found`
- Load test scripts fail
- Permission denied errors

**Solutions**:

```bash
# Install Vegeta on macOS
brew install vegeta

# Install Vegeta on Linux
wget https://github.com/tsenart/vegeta/releases/download/v12.8.4/vegeta_12.8.4_linux_amd64.tar.gz
tar -xzf vegeta_12.8.4_linux_amd64.tar.gz
sudo mv vegeta /usr/local/bin/

# Make scripts executable
chmod +x scripts/*.sh

# Check Vegeta installation
vegeta -version
```

#### Issue: Load tests not generating expected results

**Symptoms**:
- No traffic visible in dashboards
- Tests complete too quickly
- Error rates don't match expectations

**Solutions**:

```bash
# Verify application is responding
curl http://localhost:8080/api/v1/ping

# Check if error injection is working
curl -X POST -H "Authorization: Bearer changeme" \
     -H "Content-Type: application/json" \
     -d '{"enabled": true, "rate": 0.1, "status_code": 500}' \
     http://localhost:8080/api/v1/toggles/error-rate

# Run a simple manual test
echo "GET http://localhost:8080/api/v1/ping" | vegeta attack -duration=30s -rate=10 | vegeta report

# Check load test results directory
ls -la ./load-test-results/
```

### 6. Integration Test Failures

#### Issue: Integration tests timing out or failing

**Symptoms**:
- Tests fail with timeout errors
- Services not ready in time
- Container startup failures

**Solutions**:

```bash
# Run tests with increased timeout
go test -v -timeout 45m -run TestIntegration ./...

# Check test logs for specific failures
go test -v -run TestIntegration ./... 2>&1 | tee test-output.log

# Run individual test functions
go test -v -run TestServicesStartSuccessfully ./...

# Clean up before running tests
docker-compose down -v
docker system prune -f
```

### 7. Performance Issues

#### Issue: High resource usage or slow response times

**Symptoms**:
- High CPU or memory usage
- Slow dashboard loading
- Application timeouts

**Diagnostic Commands**:
```bash
# Check container resource usage
docker stats

# Check system resources
top
htop
free -h
df -h

# Check application performance
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/api/v1/ping

# Create curl-format.txt:
cat > curl-format.txt << 'EOF'
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
EOF
```

**Solutions**:

```bash
# Reduce Prometheus scrape frequency
# Edit prometheus/prometheus.yml:
# scrape_interval: 15s  # Instead of 5s

# Limit container resources
# Add to docker-compose.yml:
# deploy:
#   resources:
#     limits:
#       memory: 512M
#       cpus: '0.5'

# Reduce Grafana refresh rate
# In dashboard settings: Refresh every 30s instead of 5s

# Clean up old data
docker-compose exec prometheus rm -rf /prometheus/data/*
docker-compose restart prometheus
```

## 🔍 Advanced Debugging

### Enable Debug Logging

```bash
# Application debug logging
export LOG_LEVEL=debug
docker-compose restart go-app

# Prometheus debug logging
# Add to prometheus/prometheus.yml:
# global:
#   log.level: debug

# Grafana debug logging
# Add to docker-compose.yml grafana environment:
# - GF_LOG_LEVEL=debug
```

### Network Debugging

```bash
# Check container networking
docker network ls
docker network inspect monitoring-dashboard-automation_monitoring

# Test inter-container connectivity
docker-compose exec go-app ping prometheus
docker-compose exec prometheus ping go-app

# Check DNS resolution
docker-compose exec go-app nslookup prometheus
```

### Database and Storage Issues

```bash
# Check Prometheus data directory
docker-compose exec prometheus ls -la /prometheus/

# Check Grafana database
docker-compose exec grafana ls -la /var/lib/grafana/

# Reset all data (WARNING: This will delete all metrics and dashboards)
docker-compose down -v
docker volume prune -f
docker-compose up -d
```

## 📞 Getting Help

### Collecting Debug Information

Before seeking help, collect this information:

```bash
# System information
uname -a
docker --version
docker-compose --version

# Service status
docker-compose ps
docker-compose logs --tail=50

# Configuration
cat .env
cat docker-compose.yml

# Resource usage
docker stats --no-stream
free -h
df -h

# Network status
netstat -tulpn | grep -E ':(3000|8080|9090|9093|9100|9115)'
```

### Log Collection Script

```bash
#!/bin/bash
# Save as collect-logs.sh

echo "Collecting monitoring system debug information..."
mkdir -p debug-info

# System info
uname -a > debug-info/system-info.txt
docker --version >> debug-info/system-info.txt
docker-compose --version >> debug-info/system-info.txt

# Service status
docker-compose ps > debug-info/service-status.txt

# Logs
docker-compose logs > debug-info/all-logs.txt
docker-compose logs go-app > debug-info/go-app-logs.txt
docker-compose logs prometheus > debug-info/prometheus-logs.txt
docker-compose logs grafana > debug-info/grafana-logs.txt

# Configuration
cp .env debug-info/ 2>/dev/null || echo "No .env file" > debug-info/env-missing.txt
cp docker-compose.yml debug-info/

# Resource usage
docker stats --no-stream > debug-info/resource-usage.txt
free -h > debug-info/memory-usage.txt
df -h > debug-info/disk-usage.txt

echo "Debug information collected in debug-info/ directory"
tar -czf debug-info.tar.gz debug-info/
echo "Created debug-info.tar.gz for sharing"
```

### Common Error Messages and Solutions

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `bind: address already in use` | Port conflict | Kill process using port or change port |
| `no such host` | DNS/networking issue | Check container networking |
| `connection refused` | Service not running | Start service or check health |
| `permission denied` | File permissions | Fix file permissions with chmod |
| `out of memory` | Insufficient RAM | Increase Docker memory limit |
| `no space left on device` | Disk full | Clean up disk space |
| `context deadline exceeded` | Timeout | Increase timeout or check performance |

This troubleshooting guide should help resolve most common issues. For additional help, check the project documentation or create an issue with the debug information collected above.