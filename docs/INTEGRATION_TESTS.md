# Integration Tests Documentation

## Overview

The integration tests verify that the complete monitoring dashboard automation system works end-to-end. These tests start the full Docker Compose stack and validate all components work together correctly.

## Test Coverage

### 1. Service Startup and Health Checks
- **TestServicesStartSuccessfully**: Verifies all services start and are reachable
  - Go application health endpoints (`/healthz`, `/readyz`, `/metrics`)
  - Prometheus ready endpoint and metrics
  - Grafana health API
  - AlertManager ready endpoint
  - Node Exporter metrics
  - Blackbox Exporter metrics

### 2. Prometheus Scraping
- **TestPrometheusScrapingTargets**: Validates Prometheus can scrape all configured targets
  - Checks all expected scrape jobs are active
  - Verifies target health status
  - Confirms scrape intervals are working

### 3. Grafana Integration
- **TestGrafanaDataSourceAndQueries**: Tests Grafana-Prometheus integration
  - Verifies Prometheus datasource is configured
  - Tests query execution through Grafana API
  - Validates data retrieval and formatting

### 4. Metrics Collection
- **TestMetricsCollection**: Confirms metrics are being collected properly
  - HTTP request metrics (`http_requests_total`, `http_request_duration_seconds`)
  - Go runtime metrics (`go_goroutines`, `process_cpu_seconds_total`)
  - Application uptime metrics (`up`)

### 5. Alert System Testing
- **TestErrorInjectionAndAlerts**: Tests error injection and alert firing
  - Enables error injection via API
  - Generates traffic to trigger error conditions
  - Verifies HighErrorRate alert detection
  - Tests alert state transitions

- **TestLatencyAlertsWithWorkSimulation**: Tests latency-based alerts
  - Uses work simulation endpoint to create high latency
  - Verifies P95 latency metric collection
  - Checks for HighLatencyP95 alert conditions

- **TestInstanceDownAlert**: Tests instance failure detection
  - Stops application container
  - Verifies InstanceDown alert detection
  - Tests service recovery

### 6. Blackbox Monitoring
- **TestBlackboxProbes**: Validates external monitoring
  - Tests probe_success metrics
  - Verifies internal and external probe configurations
  - Checks probe result reporting

### 7. Webhook Notifications
- **TestWebhookDelivery**: Tests alert notification delivery
  - Starts mock webhook server
  - Configures AlertManager with test webhooks
  - Verifies webhook payload delivery

### 8. Docker Compose Health
- **TestDockerComposeHealthChecks**: Validates container health
  - Checks container running states
  - Verifies health check status
  - Confirms service dependencies

### 9. End-to-End Flow
- **TestEndToEndMonitoringFlow**: Complete system validation
  - Traffic generation and metrics collection
  - Alert rule loading and evaluation
  - Dashboard query functionality
  - System integration verification

## Prerequisites

### Required Software
- Docker and Docker Compose
- Go 1.21 or later
- Make (optional, for convenience targets)

### System Requirements
- At least 4GB RAM available for containers
- Ports 3000, 8080, 8081, 9090, 9093, 9100, 9115 available
- Internet access for external probe targets

### Environment Variables
The tests automatically set these variables:
- `ADMIN_TOKEN=test-token`
- `GRAFANA_ADMIN_USER=admin`
- `GRAFANA_ADMIN_PASSWORD=admin`
- `SLACK_WEBHOOK_URL=http://host.docker.internal:8081/slack`
- `DISCORD_WEBHOOK_URL=http://host.docker.internal:8081/discord`

## Running Integration Tests

### Method 1: Using Make (Recommended)
```bash
make test-integration
```

### Method 2: Using Scripts
```bash
# Linux/macOS
./scripts/run-integration-tests.sh

# Windows
scripts\run-integration-tests.bat
```

### Method 3: Direct Go Test
```bash
go test -v -timeout 30m -run TestIntegration ./...
```

## Test Execution Flow

1. **Setup Phase** (2-3 minutes)
   - Start mock webhook server
   - Launch Docker Compose stack
   - Wait for all services to be ready
   - Verify service health endpoints

2. **Test Execution** (10-15 minutes)
   - Run individual test cases
   - Generate test traffic and conditions
   - Validate system responses
   - Check alert states and metrics

3. **Cleanup Phase** (1-2 minutes)
   - Stop mock webhook server
   - Tear down Docker Compose stack
   - Clean up test artifacts

## Expected Test Duration

- **Total Runtime**: 15-20 minutes
- **Service Startup**: 2-3 minutes
- **Test Execution**: 10-15 minutes
- **Cleanup**: 1-2 minutes

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   - Ensure ports 3000, 8080, 8081, 9090, 9093, 9100, 9115 are available
   - Stop any existing Docker containers using these ports

2. **Docker Resource Limits**
   - Increase Docker memory limit to at least 4GB
   - Ensure sufficient disk space for container images

3. **Network Connectivity**
   - Verify internet access for external probe targets
   - Check firewall settings for Docker networking

4. **Service Startup Timeouts**
   - Some services may take longer on slower systems
   - Tests include generous timeouts but may need adjustment

### Debug Mode

To run tests with additional debugging:
```bash
export LOG_LEVEL=debug
go test -v -timeout 30m -run TestIntegration ./...
```

### Manual Verification

After tests complete, you can manually verify the system:
```bash
# Check running containers
docker-compose ps

# Access services
open http://localhost:3000  # Grafana
open http://localhost:9090  # Prometheus
open http://localhost:9093  # AlertManager
```

## Test Data and Metrics

### Generated Traffic Patterns
- **Normal Traffic**: 20-100 requests to `/api/v1/ping` and `/api/v1/work`
- **Error Traffic**: 50-100 requests with 50-100% error injection
- **Latency Traffic**: 50 requests with 600ms+ work simulation
- **Probe Traffic**: Continuous health checks via Blackbox Exporter

### Expected Metrics
- Request rate: 1-10 RPS during tests
- Error rate: 0-100% depending on test phase
- P95 latency: 100ms-700ms depending on work simulation
- Probe success: 90-100% for internal endpoints

### Alert Conditions Tested
- **HighErrorRate**: >2% error rate for 10+ minutes
- **HighLatencyP95**: >500ms P95 latency for 10+ minutes
- **InstanceDown**: Service unavailable for 2+ minutes
- **UptimeProbeFail**: Probe failure for 3+ minutes

## Continuous Integration

For CI/CD pipelines, use:
```yaml
# Example GitHub Actions step
- name: Run Integration Tests
  run: |
    make test-integration
  timeout-minutes: 25
```

## Performance Benchmarks

### Acceptable Performance Ranges
- Service startup: < 3 minutes
- Metrics collection latency: < 30 seconds
- Alert evaluation: < 1 minute for detection
- Dashboard query response: < 5 seconds
- Webhook delivery: < 10 seconds

### Resource Usage During Tests
- CPU: 2-4 cores peak usage
- Memory: 2-4GB total container usage
- Disk: 1-2GB for images and volumes
- Network: 10-50MB for image pulls and traffic

## Test Maintenance

### Updating Tests
- Modify `integration_test.go` for new test cases
- Update service URLs if ports change
- Adjust timeouts for slower environments
- Add new metrics validation as features are added

### Version Compatibility
- Tests are designed to work with current Docker image versions
- Update image tags in `docker-compose.yml` as needed
- Verify API compatibility when updating Prometheus/Grafana

## Security Considerations

### Test Environment
- Uses test tokens and credentials
- Mock webhook server for safe notification testing
- Isolated Docker network for container communication
- No production data or credentials used

### Cleanup
- All test containers are removed after execution
- No persistent data stored outside test duration
- Environment variables reset after tests