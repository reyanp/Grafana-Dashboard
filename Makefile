# Makefile for Monitoring Dashboard Automation
# Provides convenient targets for building, testing, and running load tests

.PHONY: help build test test-unit test-integration run clean demo load-test-baseline load-test-latency load-test-errors load-test-instance-down logs status fmt lint

# Default target
help:
	@echo "Available targets:"
	@echo "  build                 - Build the Go application"
	@echo "  test                  - Run all tests"
	@echo "  test-unit             - Run unit tests only"
	@echo "  test-integration      - Run integration tests (requires Docker)"
	@echo "  run                   - Start the monitoring stack"
	@echo "  clean                 - Stop and clean up the monitoring stack"
	@echo "  demo                  - Run the complete demo scenario"
	@echo "  load-test-baseline    - Run baseline load test"
	@echo "  load-test-latency     - Run latency spike test"
	@echo "  load-test-errors      - Run error injection test"
	@echo "  load-test-instance-down - Run instance down test"
	@echo "  check-deps            - Check required dependencies"
	@echo "  logs                  - Show logs from all services"
	@echo "  status                - Show status of all services"
	@echo "  fmt                   - Format Go code"
	@echo "  lint                  - Run Go linter"

# Build the Go application
build:
	go build -o bin/api ./cmd/api

# Run all tests
test: test-unit test-integration

# Run unit tests only
test-unit:
	go test -v -short ./...

# Run integration tests (requires Docker)
test-integration: check-deps build
	@echo "Running integration tests (this may take several minutes)..."
	@chmod +x scripts/run-integration-tests.sh 2>/dev/null || true
	@if [ -f scripts/run-integration-tests.sh ]; then \
		./scripts/run-integration-tests.sh; \
	else \
		echo "Running integration tests directly..."; \
		go test -v -timeout 30m -run TestIntegration ./...; \
	fi

# Start the monitoring stack
run:
	docker-compose up -d
	@echo "Waiting for services to start..."
	@sleep 10
	@echo "Services started. Access URLs:"
	@echo "  Application: http://localhost:8080"
	@echo "  Grafana: http://localhost:3000 (admin/admin)"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  AlertManager: http://localhost:9093"

# Stop and clean up
clean:
	docker-compose down
	docker-compose down -v
	rm -rf ./load-test-results/

# Check required dependencies
check-deps:
	@echo "Checking dependencies..."
	@command -v docker >/dev/null 2>&1 || { echo "Error: docker is required but not installed"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "Error: docker-compose is required but not installed"; exit 1; }
	@command -v vegeta >/dev/null 2>&1 || { echo "Warning: vegeta is required for load testing"; }
	@command -v jq >/dev/null 2>&1 || { echo "Warning: jq is recommended for JSON processing"; }
	@command -v bc >/dev/null 2>&1 || { echo "Warning: bc is required for calculations"; }
	@echo "Dependency check completed"

# Run the complete demo scenario
demo: check-deps
	@chmod +x scripts/demo-scenario.sh 2>/dev/null || true
	./scripts/demo-scenario.sh

# Run baseline load test
load-test-baseline: check-deps
	@chmod +x scripts/load-test-baseline.sh 2>/dev/null || true
	./scripts/load-test-baseline.sh

# Run latency spike test
load-test-latency: check-deps
	@chmod +x scripts/load-test-latency-spike.sh 2>/dev/null || true
	./scripts/load-test-latency-spike.sh

# Run error injection test
load-test-errors: check-deps
	@chmod +x scripts/trigger-error-alerts.sh 2>/dev/null || true
	./scripts/trigger-error-alerts.sh

# Run instance down test
load-test-instance-down: check-deps
	@chmod +x scripts/trigger-instance-down-alerts.sh 2>/dev/null || true
	./scripts/trigger-instance-down-alerts.sh

# Show logs from all services
logs:
	docker-compose logs -f

# Show status of all services
status:
	@echo "=== Docker Compose Services ==="
	docker-compose ps
	@echo ""
	@echo "=== Service Health Checks ==="
	@curl -f -s http://localhost:8080/healthz >/dev/null && echo "✅ Application: healthy" || echo "❌ Application: unhealthy"
	@curl -f -s http://localhost:9090/-/healthy >/dev/null && echo "✅ Prometheus: healthy" || echo "❌ Prometheus: unhealthy"
	@curl -f -s http://localhost:3000/api/health >/dev/null && echo "✅ Grafana: healthy" || echo "❌ Grafana: unhealthy"
	@curl -f -s http://localhost:9093/-/healthy >/dev/null && echo "✅ AlertManager: healthy" || echo "❌ AlertManager: unhealthy"

# Format Go code
fmt:
	go fmt ./...
	goimports -w .

# Run Go linter
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run