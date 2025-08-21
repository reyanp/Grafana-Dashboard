# Implementation Plan

- [x] 1. Set up project structure and core configuration





  - Create Go module with proper directory structure (cmd/api, internal packages)
  - Implement configuration management with environment variable parsing
  - Create .env.example file with all required configuration options
  - _Requirements: 8.1, 8.2, 8.5_

- [x] 2. Implement core HTTP server and middleware





  - Set up Chi router with basic route structure
  - Implement request ID middleware for request tracing
  - Implement structured logging middleware using zap logger
  - Implement panic recovery middleware with stack trace logging
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [x] 3. Create health check endpoints









  - Implement GET /healthz endpoint that always returns 200 OK
  - Implement GET /readyz endpoint with dependency health checks
  - Create health check logic that can be toggled for testing
  - Write unit tests for health check endpoints
  - _Requirements: 1.1, 1.2_

- [x] 4. Implement Prometheus metrics instrumentation





  - Set up prometheus client_golang with custom registry
  - Create middleware for automatic HTTP request instrumentation
  - Implement http_requests_total counter with route, method, status labels
  - Implement http_request_duration_seconds histogram with route, method labels
  - Expose standard Go runtime metrics (goroutines, CPU, memory)
  - Create GET /metrics endpoint that serves Prometheus format
  - _Requirements: 1.6, 2.1, 2.2, 2.3, 2.4, 2.5_

- [x] 5. Create API endpoints with work simulation










  - Implement GET /api/v1/ping endpoint with basic response
  - Implement GET /api/v1/work endpoint with ms and jitter parameters
  - Add work simulation logic that respects context cancellation
  - Implement work_jobs_inflight gauge metric for active jobs
  - Implement work_failures_total counter for failed operations
  - Write unit tests for API endpoints and work simulation
  - _Requirements: 1.4, 1.5, 2.6, 2.7_

- [x] 6. Implement error injection system





  - Create error toggle data structure with rate and status code configuration
  - Implement POST /api/v1/toggles/error-rate endpoint with bearer token auth
  - Create middleware that injects errors based on toggle configuration
  - Implement bearer token authentication for admin routes
  - Write unit tests for error injection and authentication
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 7. Create Docker Compose infrastructure






  - Write docker-compose.yml with all required services (go-app, prometheus, grafana, alertmanager, node_exporter, blackbox_exporter)
  - Create Dockerfile for Go application with multi-stage build
  - Configure service networking and port mappings
  - Set up volume mounts for configuration files
  - _Requirements: 4.1_

- [x] 8. Configure Prometheus scraping and alerting





  - Create prometheus/prometheus.yml with scrape configurations for all services
  - Implement scrape configs for go-app, node_exporter, and blackbox_exporter with 5-second intervals
  - Create prometheus/alerts.yml with all required alert rules
  - Implement InstanceDown, HighErrorRate, HighLatencyP95, and UptimeProbeFail alert rules
  - _Requirements: 4.2, 4.3, 5.1, 5.2, 5.3, 5.4_

- [x] 9. Configure AlertManager for webhook notifications





  - Create alertmanager/alertmanager.yml with routing and receiver configuration
  - Configure Slack webhook integration with proper message formatting
  - Configure Discord webhook integration with proper message formatting
  - Set up alert grouping, timing, and repeat intervals
  - _Requirements: 4.5, 5.5, 5.6_

- [ ] 10. Set up Grafana with datasource and dashboards
  - Create grafana/provisioning/datasource.yml to configure Prometheus datasource
  - Configure Grafana to start with Prometheus connection
  - Create grafana/provisioning/dashboard.json with all required panels
  - Implement Service Overview, Request Rate, Error Rate, and Latency panels
  - Implement Node CPU/Memory, Probe Status, Top Routes, and Alert History panels
  - _Requirements: 4.4, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8_

- [ ] 11. Configure Blackbox Exporter for uptime probes
  - Create blackbox/blackbox.yml with HTTP probe configuration
  - Configure probes for application health endpoint and external targets
  - Set up Prometheus to scrape blackbox exporter metrics
  - Configure probe targets via environment variables
  - _Requirements: 4.3, 5.4_

- [ ] 12. Create load testing and demo scripts
  - Write vegeta-based load testing script that generates sufficient traffic
  - Create script to trigger latency alerts by increasing work duration
  - Create script to trigger error alerts using error injection toggle
  - Write demo scenario documentation with step-by-step instructions
  - _Requirements: 7.1, 7.2, 7.3, 7.5_

- [ ] 13. Implement graceful shutdown and signal handling
  - Add context-based shutdown handling for HTTP server
  - Implement graceful shutdown that waits for in-flight requests
  - Add signal handling for SIGTERM and SIGINT
  - Ensure metrics are properly flushed on shutdown
  - Write tests for graceful shutdown behavior
  - _Requirements: 1.5, 2.6_

- [ ] 14. Create comprehensive integration tests
  - Write integration tests that start the full Docker Compose stack
  - Test that all services start successfully and are reachable
  - Verify Prometheus can scrape metrics from all targets
  - Test that Grafana can query Prometheus and display data
  - Verify alert rules fire under expected conditions
  - Test webhook delivery to mock Slack/Discord endpoints
  - _Requirements: 7.4, 7.5_

- [ ] 15. Add Makefile and documentation
  - Create Makefile with targets for build, test, run, and cleanup
  - Write comprehensive README.md with setup and usage instructions
  - Document all environment variables and configuration options
  - Include screenshots of Grafana dashboards and alert notifications
  - Create troubleshooting guide for common issues
  - _Requirements: 8.4, 8.5_