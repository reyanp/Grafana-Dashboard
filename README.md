# Monitoring Dashboard Automation

A complete monitoring system using Go, Prometheus, Grafana, and AlertManager. Demonstrates production-ready observability patterns with automated alerting to Slack/Discord.

## 🚀 Quick Start

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd monitoring-dashboard-automation
   cp .env.example .env
   ```

2. **Start the monitoring stack**:
   ```bash
   make run
   ```

3. **Access the services**:
   - **Application**: http://localhost:8080
   - **Grafana**: http://localhost:3000 (admin/admin)
   - **Prometheus**: http://localhost:9090
   - **AlertManager**: http://localhost:9093

4. **Run a demo**:
   ```bash
   make demo
   ```

## ✨ Features

- **Go Web Service** with Prometheus metrics
- **Grafana Dashboards** for visualization
- **Prometheus** for metrics collection
- **AlertManager** with Slack/Discord notifications
- **Load Testing** scripts with Vegeta
- **Error Injection** system for testing alerts
- **Health Checks** and uptime monitoring
- **Docker Compose** setup for easy deployment

## 🛠 Available Commands

| Command | Description |
|---------|-------------|
| `make run` | Start the monitoring stack |
| `make clean` | Stop and clean up everything |
| `make test` | Run all tests |
| `make demo` | Run complete demo scenario |
| `make load-test-baseline` | Generate baseline traffic |
| `make status` | Check service health |

## 📊 What You Get

### Monitoring Stack
- **6 services** running in Docker containers
- **Automatic metrics collection** every 5 seconds
- **Pre-configured dashboards** showing key metrics
- **Alert rules** for common failure scenarios

### Key Metrics Tracked
- HTTP request rates and error rates
- Response time percentiles (P50, P95, P99)
- Service uptime and health status
- System resource usage

### Alert Conditions
- Service down for 2+ minutes
- Error rate > 2% for 10+ minutes
- High latency (P95 > 500ms) for 10+ minutes
- External probe failures

## ⚙️ Configuration

1. **Copy environment template**:
   ```bash
   cp .env.example .env
   ```

2. **Add your webhook URLs** (optional):
   ```bash
   # Edit .env file
   SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
   DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK
   ```

3. **Restart AlertManager** to pick up webhook changes:
   ```bash
   docker-compose restart alertmanager
   ```

## 🧪 Testing

### Unit Tests
```bash
make test-unit
```

### Integration Tests
```bash
make test-integration
```

### Load Testing
```bash
# Generate normal traffic
make load-test-baseline

# Test error conditions
make load-test-errors

# Test latency alerts
make load-test-latency
```

## 📁 Project Structure

```
.
├── cmd/api/              # Application entry point
├── internal/             # Go application code
├── prometheus/           # Prometheus configuration
├── grafana/             # Grafana dashboards
├── alertmanager/        # Alert routing configuration
├── scripts/             # Load testing and demo scripts
├── docs/                # Additional documentation
├── docker-compose.yml   # Complete monitoring stack
└── Makefile            # Convenient build targets
```

## 🔧 Troubleshooting

### Services not starting?
```bash
# Check service status
docker-compose ps

# View logs
docker-compose logs
```

### No data in dashboards?
```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Generate some traffic
curl http://localhost:8080/api/v1/ping
```

### Alerts not firing?
```bash
# Check alert rules
curl http://localhost:9090/api/v1/rules

# Test error injection
curl -X POST -H "Authorization: Bearer changeme" \
     -H "Content-Type: application/json" \
     -d '{"enabled": true, "rate": 0.1, "status_code": 500}' \
     http://localhost:8080/api/v1/toggles/error-rate
```

## 📚 Documentation

For detailed information, see the [docs/](docs/) directory:

- [DEMO_GUIDE.md](docs/DEMO_GUIDE.md) - Step-by-step demo walkthrough
- [CONFIGURATION.md](docs/CONFIGURATION.md) - Detailed configuration options
- [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) - Common issues and solutions

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Built With

- [Go](https://golang.org/) - Application runtime
- [Prometheus](https://prometheus.io/) - Metrics collection
- [Grafana](https://grafana.com/) - Visualization
- [Docker](https://www.docker.com/) - Containerization
- [Vegeta](https://github.com/tsenart/vegeta) - Load testing
- [Kiro](https://kiro.dev/) - Amazon's AI IDE

---

**Passion project to learn data visualize tools + test new AI IDE**