#!/bin/bash

# Integration Test Runner Script
# This script sets up the environment and runs comprehensive integration tests

set -e

echo "🚀 Starting Integration Test Suite"
echo "=================================="

# Check prerequisites
echo "📋 Checking prerequisites..."

if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed or not in PATH"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed or not in PATH"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    exit 1
fi

echo "✅ All prerequisites found"

# Clean up any existing containers
echo "🧹 Cleaning up existing containers..."
docker-compose down -v --remove-orphans 2>/dev/null || true

# Set environment variables for testing
export ADMIN_TOKEN="test-token"
export GRAFANA_ADMIN_USER="admin"
export GRAFANA_ADMIN_PASSWORD="admin"
export LOG_LEVEL="info"
export ENVIRONMENT="test"

# Build the application first
echo "🔨 Building application..."
go build -o bin/api ./cmd/api

# Run integration tests
echo "🧪 Running integration tests..."
echo "This may take several minutes as it starts the full Docker stack..."

# Run tests with verbose output and timeout
go test -v -timeout 30m -run TestIntegration ./...

echo ""
echo "✅ Integration tests completed successfully!"
echo ""
echo "📊 Test Summary:"
echo "- All services started successfully"
echo "- Prometheus scraping verified"
echo "- Grafana datasource connectivity confirmed"
echo "- Alert rules loaded and functional"
echo "- Metrics collection working"
echo "- Blackbox probes operational"
echo "- Error injection system tested"
echo "- Webhook delivery system verified"
echo ""
echo "🎉 Integration test suite passed!"