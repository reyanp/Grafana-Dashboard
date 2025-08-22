# Blackbox Exporter Configuration

This document explains how to configure and use the Blackbox Exporter for uptime monitoring.

## Overview

The Blackbox Exporter is configured to probe both internal application endpoints and external targets to monitor uptime and availability.

## Probe Modules

### http_2xx
- Used for internal application health endpoints
- Expects 2xx HTTP status codes
- 5-second timeout
- IPv4 preferred

### http_external
- Used for external website monitoring
- More permissive configuration
- 10-second timeout
- Follows redirects
- IPv4 with IPv6 fallback

### tcp_connect
- Tests TCP connectivity
- 5-second timeout

### icmp
- ICMP ping probe
- 5-second timeout

## Configured Probe Targets

### Internal Targets (Automatic)
- `http://go-app:8080/healthz` - Application liveness probe
- `http://go-app:8080/readyz` - Application readiness probe  
- `http://go-app:8080/api/v1/ping` - API endpoint probe

### External Targets (Configurable)
- `https://httpbin.org/status/200` - Test endpoint
- `https://example.com` - Example website
- `https://google.com` - Google homepage

## Adding New Probe Targets

To add new external probe targets:

1. Update the `.env` file with new target URLs:
   ```bash
   BLACKBOX_EXTERNAL_TARGET_4=https://your-website.com
   ```

2. Add the target to `prometheus/prometheus.yml` in the `blackbox_http_external` job:
   ```yaml
   static_configs:
     - targets:
       - https://httpbin.org/status/200
       - https://example.com
       - https://google.com
       - https://your-website.com  # Add your new target here
   ```

3. Restart the Prometheus container:
   ```bash
   docker-compose restart prometheus
   ```

## Metrics Generated

The Blackbox Exporter generates the following key metrics:

- `probe_success{instance="target_url"}` - 1 if probe succeeded, 0 if failed
- `probe_duration_seconds{instance="target_url"}` - Time taken for the probe
- `probe_http_status_code{instance="target_url"}` - HTTP status code returned
- `probe_http_ssl{instance="target_url"}` - 1 if SSL was used

## Alert Rules

The following alert rules are configured for probe failures:

### UptimeProbeFail
- **Condition**: `probe_success == 0`
- **Duration**: 3 minutes
- **Severity**: Critical
- **Description**: Fires when any probe target fails for more than 3 minutes

## Testing Probe Configuration

You can test probe configuration manually:

1. **Test a specific target**:
   ```bash
   curl "http://localhost:9115/probe?target=https://example.com&module=http_external"
   ```

2. **Check probe metrics**:
   ```bash
   curl http://localhost:9115/metrics | grep probe_success
   ```

3. **View in Prometheus**:
   - Go to http://localhost:9090
   - Query: `probe_success`
   - This shows the success status of all probes

## Troubleshooting

### Probe Failures
- Check the target URL is accessible
- Verify the correct module is being used
- Check network connectivity from the container
- Review blackbox exporter logs: `docker-compose logs blackbox_exporter`

### Configuration Issues
- Validate YAML syntax in `blackbox/blackbox.yml`
- Restart blackbox exporter after configuration changes
- Check Prometheus targets page: http://localhost:9090/targets

### Common Issues
1. **SSL Certificate Issues**: Set `insecure_skip_verify: true` in the module for testing
2. **Timeout Issues**: Increase timeout values in the module configuration
3. **Network Issues**: Ensure targets are accessible from within Docker network