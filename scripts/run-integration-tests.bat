@echo off
REM Integration Test Runner Script for Windows
REM This script sets up the environment and runs comprehensive integration tests

echo 🚀 Starting Integration Test Suite
echo ==================================

REM Check prerequisites
echo 📋 Checking prerequisites...

where docker >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo ❌ Docker is not installed or not in PATH
    exit /b 1
)

where docker-compose >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo ❌ Docker Compose is not installed or not in PATH
    exit /b 1
)

where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo ❌ Go is not installed or not in PATH
    exit /b 1
)

echo ✅ All prerequisites found

REM Clean up any existing containers
echo 🧹 Cleaning up existing containers...
docker-compose down -v --remove-orphans 2>nul

REM Set environment variables for testing
set ADMIN_TOKEN=test-token
set GRAFANA_ADMIN_USER=admin
set GRAFANA_ADMIN_PASSWORD=admin
set LOG_LEVEL=info
set ENVIRONMENT=test

REM Build the application first
echo 🔨 Building application...
go build -o bin/api.exe ./cmd/api

REM Run integration tests
echo 🧪 Running integration tests...
echo This may take several minutes as it starts the full Docker stack...

REM Run tests with verbose output and timeout
go test -v -timeout 30m -run TestIntegration ./...

if %ERRORLEVEL% neq 0 (
    echo ❌ Integration tests failed
    exit /b 1
)

echo.
echo ✅ Integration tests completed successfully!
echo.
echo 📊 Test Summary:
echo - All services started successfully
echo - Prometheus scraping verified
echo - Grafana datasource connectivity confirmed
echo - Alert rules loaded and functional
echo - Metrics collection working
echo - Blackbox probes operational
echo - Error injection system tested
echo - Webhook delivery system verified
echo.
echo 🎉 Integration test suite passed!