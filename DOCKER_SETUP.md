# Docker Compose Monitoring Stack

This directory contains a complete monitoring stack using Docker Compose with the following services:

## Services

### Core Application
- **go-app** (port 8080): The main Go web service with metrics endpoints
- **prometheus** (port 9090): Metrics collection and alerting
- **grafana** (port 3000): Visualization dashboards
- **alertmanager** (port 9093): Alert routing and notifications

### Monitoring Components
- **node_exporter** (port 9100): System metrics collection
- **blackbox_exporter** (port 9115): External probe monitoring

## Quick Start

1. Copy environment variables:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` file with your webhook URLs:
   ```bash
   # Update these with your actual webhook URLs
   SLACK_WEBHOOK_URL=https://hooks.slack.com/services/TXXXXXXXX/BXXXXXXXX/XXXXXXXXXXXXXXXXXXXXXXXX
   DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/123456789012345678/XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
   ```

3. Start the stack:
   ```bash
   docker-compose up -d
   ```

4. Access the services:
   - Application: http://localhost:8080
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)
   - AlertManager: http://localhost:9093

## Configuration Files

### Prometheus (`prometheus/`)
- `prometheus.yml`: Scrape configuration for all services
- `alerts.yml`: Alert rules for monitoring conditions

### Grafana (`grafana/provisioning/`)
- `datasources/datasource.yml`: Prometheus datasource configuration
- `dashboards/dashboard.yml`: Dashboard provisioning configuration
- `dashboards/monitoring-dashboard.json`: Main monitoring dashboard

### AlertManager (`alertmanager/`)
- `alertmanager.yml`: Alert routing and notification configuration

### Blackbox Exporter (`blackbox/`)
- `blackbox.yml`: HTTP probe configuration

## Volumes

The following Docker volumes are created for data persistence:
- `prometheus_data`: Prometheus time-series data
- `grafana_data`: Grafana dashboards and settings
- `alertmanager_data`: AlertManager state and silences

## Network

All services run on the `monitoring` bridge network for internal communication.

## Health Checks

- Go application includes health check on `/healthz` endpoint
- All services have restart policies configured
- Prometheus monitors all service health via scraping

## Stopping the Stack

```bash
# Stop services but keep data
docker-compose down

# Stop services and remove volumes (data loss)
docker-compose down -v
```

## Troubleshooting

1. **Services not starting**: Check logs with `docker-compose logs <service-name>`
2. **Metrics not appearing**: Verify Prometheus targets at http://localhost:9090/targets
3. **Alerts not firing**: Check AlertManager status at http://localhost:9093
4. **Dashboard empty**: Ensure Grafana can connect to Prometheus datasource

## Customization

- Modify scrape intervals in `prometheus/prometheus.yml`
- Add custom alert rules in `prometheus/alerts.yml`
- Update dashboard panels in `grafana/provisioning/dashboards/monitoring-dashboard.json`
- Configure additional notification channels in `alertmanager/alertmanager.yml`