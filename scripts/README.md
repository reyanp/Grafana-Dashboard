# Load Testing and Demo Scripts

This directory contains scripts for load testing and demonstrating the monitoring system capabilities.

## Scripts Overview

| Script | Purpose | Duration | Requirements |
|--------|---------|----------|--------------|
| `demo-scenario.sh` | Complete interactive demo | 30-45 min | 7.4, 7.5 |
| `load-test-baseline.sh` | Baseline load testing | 5 min | 7.1, 7.5 |
| `load-test-latency-spike.sh` | Trigger latency alerts | 12 min | 7.1, 7.2 |
| `trigger-error-alerts.sh` | Trigger error rate alerts | 12 min | 7.2, 7.3 |
| `trigger-instance-down-alerts.sh` | Trigger instance down alerts | 3 min | 7.3 |
| `run-demo.bat` | Windows batch file for demo | Variable | 7.4, 7.5 |

## Quick Start

### Linux/macOS
```bash
# Make scripts executable
chmod +x scripts/*.sh

# Run complete demo
./scripts/demo-scenario.sh

# Or run individual tests
./scripts/load-test-baseline.sh
./scripts/load-test-latency-spike.sh
./scripts/trigger-error-alerts.sh
./scripts/trigger-instance-down-alerts.sh
```

### Windows
```cmd
# Run complete demo
scripts\run-demo.bat

# Or use bash (Git Bash, WSL, etc.)
bash scripts/demo-scenario.sh
```

### Using Makefile
```bash
# Run complete demo
make demo

# Run individual tests
make load-test-baseline
make load-test-latency
make load-test-errors
make load-test-instance-down
```

## Prerequisites

1. **Docker and Docker Compose** - for running the monitoring stack
2. **Vegeta** - for HTTP load testing
3. **jq** - for JSON processing (recommended)
4. **bc** - for mathematical calculations
5. **curl** - for HTTP requests (usually pre-installed)

See [DEMO_GUIDE.md](../DEMO_GUIDE.md) for detailed installation instructions.

## Environment Variables

- `ADMIN_TOKEN` - Admin token for error injection (default: "changeme")
- `SLACK_WEBHOOK_URL` - Slack webhook for notifications (optional)
- `DISCORD_WEBHOOK_URL` - Discord webhook for notifications (optional)

## Output

All scripts generate results in the `./load-test-results/` directory:

- `*-report.txt` - Human-readable summary
- `*-report.json` - Machine-readable results
- `*-histogram.txt` - Latency distribution
- `*-plot.html` - Visual plot (if gnuplot available)

## Alert Expectations

| Test | Expected Alert | Threshold | Time to Fire |
|------|----------------|-----------|--------------|
| Latency Spike | HighLatencyP95 | p95 > 500ms | 10 minutes |
| Error Injection | HighErrorRate | Error rate > 2% | 10 minutes |
| Instance Down | InstanceDown | Service down | 2 minutes |
| Instance Down | UptimeProbeFail | Probe failure | 3 minutes |

## Monitoring URLs

- **Application**: http://localhost:8080
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **AlertManager**: http://localhost:9093

## Troubleshooting

1. **Scripts not executable**: Run `chmod +x scripts/*.sh`
2. **Vegeta not found**: Install vegeta (see DEMO_GUIDE.md)
3. **Target not reachable**: Ensure `docker-compose up -d` is running
4. **Alerts not firing**: Check Prometheus targets and alert rules
5. **Container not found**: Verify container names with `docker ps`

For detailed troubleshooting, see [DEMO_GUIDE.md](../DEMO_GUIDE.md).