# Monitoring Dashboard Automation

A production-style monitoring system that exposes metrics from a Go web service, visualizes them with Grafana, and triggers automated alerts to Slack/Discord when performance thresholds are breached.

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── http/
│   │   ├── router.go           # Chi router setup
│   │   ├── handlers.go         # HTTP handlers
│   │   └── middleware.go       # Request middleware
│   ├── metrics/
│   │   └── prometheus.go       # Prometheus instrumentation
│   ├── toggles/
│   │   └── errors.go           # Error injection logic
│   └── health/
│       └── readiness.go        # Health check logic
├── go.mod                      # Go module definition
├── .env.example               # Environment configuration template
└── README.md                  # Project documentation
```

## Configuration

Copy `.env.example` to `.env` and update the values as needed:

```bash
cp .env.example .env
```

## Getting Started

1. Install dependencies:
```bash
go mod tidy
```

2. Run the application:
```bash
go run cmd/api/main.go
```

The application will start on port 8080 by default.

## Environment Variables

See `.env.example` for all available configuration options.