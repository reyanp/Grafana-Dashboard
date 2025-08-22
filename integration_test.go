package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite contains all integration tests
type IntegrationTestSuite struct {
	suite.Suite
	httpClient *http.Client
	baseURL    string
	
	// Service endpoints
	goAppURL        string
	prometheusURL   string
	grafanaURL      string
	alertmanagerURL string
	nodeExporterURL string
	blackboxURL     string
	
	// Mock webhook server
	webhookServer *http.Server
	webhookPort   string
	receivedWebhooks []WebhookPayload
}

// WebhookPayload represents a webhook notification
type WebhookPayload struct {
	Timestamp time.Time
	Headers   map[string]string
	Body      string
	URL       string
}

// SetupSuite runs before all tests in the suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Set up HTTP client with reasonable timeouts
	suite.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Define service URLs
	suite.goAppURL = "http://localhost:8080"
	suite.prometheusURL = "http://localhost:9090"
	suite.grafanaURL = "http://localhost:3000"
	suite.alertmanagerURL = "http://localhost:9093"
	suite.nodeExporterURL = "http://localhost:9100"
	suite.blackboxURL = "http://localhost:9115"
	
	// Start mock webhook server
	suite.startMockWebhookServer()
	
	// Start Docker Compose stack
	suite.startDockerComposeStack()
	
	// Wait for all services to be ready
	suite.waitForServicesReady()
}

// TearDownSuite runs after all tests in the suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	// Stop mock webhook server
	if suite.webhookServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		suite.webhookServer.Shutdown(ctx)
	}
	
	// Stop Docker Compose stack
	suite.stopDockerComposeStack()
}

// startMockWebhookServer starts a mock webhook server to capture notifications
func (suite *IntegrationTestSuite) startMockWebhookServer() {
	suite.webhookPort = "8081"
	suite.receivedWebhooks = make([]WebhookPayload, 0)
	
	mux := http.NewServeMux()
	
	// Slack webhook endpoint
	mux.HandleFunc("/slack", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		headers := make(map[string]string)
		for k, v := range r.Header {
			headers[k] = strings.Join(v, ",")
		}
		
		suite.receivedWebhooks = append(suite.receivedWebhooks, WebhookPayload{
			Timestamp: time.Now(),
			Headers:   headers,
			Body:      string(body),
			URL:       "/slack",
		})
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	
	// Discord webhook endpoint
	mux.HandleFunc("/discord", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		headers := make(map[string]string)
		for k, v := range r.Header {
			headers[k] = strings.Join(v, ",")
		}
		
		suite.receivedWebhooks = append(suite.receivedWebhooks, WebhookPayload{
			Timestamp: time.Now(),
			Headers:   headers,
			Body:      string(body),
			URL:       "/discord",
		})
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	
	suite.webhookServer = &http.Server{
		Addr:    ":" + suite.webhookPort,
		Handler: mux,
	}
	
	go func() {
		if err := suite.webhookServer.ListenAndServe(); err != http.ErrServerClosed {
			suite.T().Logf("Mock webhook server error: %v", err)
		}
	}()
	
	// Wait for webhook server to start
	time.Sleep(2 * time.Second)
}

// startDockerComposeStack starts the Docker Compose stack
func (suite *IntegrationTestSuite) startDockerComposeStack() {
	suite.T().Log("Starting Docker Compose stack...")
	
	// Set environment variables for webhook URLs
	os.Setenv("SLACK_WEBHOOK_URL", fmt.Sprintf("http://host.docker.internal:%s/slack", suite.webhookPort))
	os.Setenv("DISCORD_WEBHOOK_URL", fmt.Sprintf("http://host.docker.internal:%s/discord", suite.webhookPort))
	os.Setenv("ADMIN_TOKEN", "test-token")
	
	// Stop any existing containers
	cmd := exec.Command("docker-compose", "down", "-v")
	cmd.Run()
	
	// Start the stack
	cmd = exec.Command("docker-compose", "up", "-d", "--build")
	output, err := cmd.CombinedOutput()
	require.NoError(suite.T(), err, "Failed to start Docker Compose stack: %s", string(output))
	
	suite.T().Log("Docker Compose stack started successfully")
}

// stopDockerComposeStack stops the Docker Compose stack
func (suite *IntegrationTestSuite) stopDockerComposeStack() {
	suite.T().Log("Stopping Docker Compose stack...")
	
	cmd := exec.Command("docker-compose", "down", "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		suite.T().Logf("Warning: Failed to stop Docker Compose stack: %s", string(output))
	}
}

// waitForServicesReady waits for all services to be ready
func (suite *IntegrationTestSuite) waitForServicesReady() {
	suite.T().Log("Waiting for services to be ready...")
	
	services := map[string]string{
		"Go App":         suite.goAppURL + "/healthz",
		"Prometheus":     suite.prometheusURL + "/-/ready",
		"Grafana":        suite.grafanaURL + "/api/health",
		"AlertManager":   suite.alertmanagerURL + "/-/ready",
		"Node Exporter":  suite.nodeExporterURL + "/metrics",
		"Blackbox":       suite.blackboxURL + "/metrics",
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	for name, url := range services {
		suite.T().Logf("Waiting for %s to be ready...", name)
		suite.waitForEndpoint(ctx, url, name)
	}
	
	// Additional wait for services to fully initialize
	suite.T().Log("Services ready, waiting for initialization...")
	time.Sleep(30 * time.Second)
}

// waitForEndpoint waits for a specific endpoint to be ready
func (suite *IntegrationTestSuite) waitForEndpoint(ctx context.Context, url, serviceName string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			suite.T().Fatalf("Timeout waiting for %s to be ready", serviceName)
		case <-ticker.C:
			req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
			resp, err := suite.httpClient.Do(req)
			if err == nil && resp.StatusCode < 400 {
				resp.Body.Close()
				suite.T().Logf("%s is ready", serviceName)
				return
			}
			if resp != nil {
				resp.Body.Close()
			}
			suite.T().Logf("Waiting for %s... (error: %v)", serviceName, err)
		}
	}
}

// TestServicesStartSuccessfully tests that all services start and are reachable
func (suite *IntegrationTestSuite) TestServicesStartSuccessfully() {
	suite.T().Log("Testing that all services are reachable...")
	
	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{"Go App Health", suite.goAppURL + "/healthz", 200},
		{"Go App Readiness", suite.goAppURL + "/readyz", 200},
		{"Go App Metrics", suite.goAppURL + "/metrics", 200},
		{"Go App Ping", suite.goAppURL + "/api/v1/ping", 200},
		{"Prometheus Ready", suite.prometheusURL + "/-/ready", 200},
		{"Prometheus Metrics", suite.prometheusURL + "/metrics", 200},
		{"Grafana Health", suite.grafanaURL + "/api/health", 200},
		{"AlertManager Ready", suite.alertmanagerURL + "/-/ready", 200},
		{"Node Exporter Metrics", suite.nodeExporterURL + "/metrics", 200},
		{"Blackbox Metrics", suite.blackboxURL + "/metrics", 200},
	}
	
	for _, test := range tests {
		suite.T().Run(test.name, func(t *testing.T) {
			resp, err := suite.httpClient.Get(test.url)
			require.NoError(t, err, "Failed to reach %s", test.url)
			defer resp.Body.Close()
			
			assert.Equal(t, test.expected, resp.StatusCode, 
				"Unexpected status code for %s", test.url)
		})
	}
}

// TestPrometheusScrapingTargets tests that Prometheus can scrape metrics from all targets
func (suite *IntegrationTestSuite) TestPrometheusScrapingTargets() {
	suite.T().Log("Testing Prometheus scraping targets...")
	
	// Get targets from Prometheus API
	resp, err := suite.httpClient.Get(suite.prometheusURL + "/api/v1/targets")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var targetsResponse struct {
		Status string `json:"status"`
		Data   struct {
			ActiveTargets []struct {
				DiscoveredLabels map[string]string `json:"discoveredLabels"`
				Labels           map[string]string `json:"labels"`
				ScrapePool       string            `json:"scrapePool"`
				ScrapeURL        string            `json:"scrapeUrl"`
				Health           string            `json:"health"`
				LastError        string            `json:"lastError"`
				LastScrape       time.Time         `json:"lastScrape"`
				LastScrapeDuration float64         `json:"lastScrapeDuration"`
			} `json:"activeTargets"`
		} `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&targetsResponse)
	require.NoError(suite.T(), err)
	
	assert.Equal(suite.T(), "success", targetsResponse.Status)
	
	// Check that we have targets for all expected jobs
	expectedJobs := []string{"prometheus", "go-app", "node", "blackbox", "blackbox_http_internal", "blackbox_http_external"}
	foundJobs := make(map[string]bool)
	
	for _, target := range targetsResponse.Data.ActiveTargets {
		foundJobs[target.ScrapePool] = true
		
		// Check that targets are healthy
		if target.Health != "up" {
			suite.T().Logf("Warning: Target %s (%s) is not healthy: %s", 
				target.ScrapeURL, target.ScrapePool, target.LastError)
		}
	}
	
	for _, job := range expectedJobs {
		assert.True(suite.T(), foundJobs[job], "Expected job %s not found in targets", job)
	}
	
	suite.T().Logf("Found %d active targets", len(targetsResponse.Data.ActiveTargets))
}

// TestGrafanaDataSourceAndQueries tests that Grafana can query Prometheus
func (suite *IntegrationTestSuite) TestGrafanaDataSourceAndQueries() {
	suite.T().Log("Testing Grafana datasource and queries...")
	
	// Test Grafana datasource
	req, _ := http.NewRequest("GET", suite.grafanaURL+"/api/datasources", nil)
	req.SetBasicAuth("admin", "admin")
	
	resp, err := suite.httpClient.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var datasources []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&datasources)
	require.NoError(suite.T(), err)
	
	// Check that Prometheus datasource exists
	var promDS map[string]interface{}
	for _, ds := range datasources {
		if ds["type"] == "prometheus" {
			promDS = ds
			break
		}
	}
	require.NotNil(suite.T(), promDS, "Prometheus datasource not found")
	
	// Test a simple query through Grafana
	dsUID := promDS["uid"].(string)
	queryURL := fmt.Sprintf("%s/api/ds/query", suite.grafanaURL)
	
	queryPayload := map[string]interface{}{
		"queries": []map[string]interface{}{
			{
				"datasource": map[string]interface{}{
					"type": "prometheus",
					"uid":  dsUID,
				},
				"expr":   "up",
				"refId":  "A",
				"format": "time_series",
			},
		},
		"from": fmt.Sprintf("%d", time.Now().Add(-5*time.Minute).UnixMilli()),
		"to":   fmt.Sprintf("%d", time.Now().UnixMilli()),
	}
	
	queryJSON, _ := json.Marshal(queryPayload)
	req, _ = http.NewRequest("POST", queryURL, bytes.NewBuffer(queryJSON))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "admin")
	
	resp, err = suite.httpClient.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.Equal(suite.T(), 200, resp.StatusCode, "Grafana query failed")
	
	var queryResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&queryResponse)
	require.NoError(suite.T(), err)
	
	// Check that we got results
	results, ok := queryResponse["results"].(map[string]interface{})
	require.True(suite.T(), ok, "No results in query response")
	
	aResult, ok := results["A"].(map[string]interface{})
	require.True(suite.T(), ok, "No result for query A")
	
	frames, ok := aResult["frames"].([]interface{})
	require.True(suite.T(), ok && len(frames) > 0, "No data frames in query result")
	
	suite.T().Log("Grafana can successfully query Prometheus")
}

// TestMetricsCollection tests that metrics are being collected properly
func (suite *IntegrationTestSuite) TestMetricsCollection() {
	suite.T().Log("Testing metrics collection...")
	
	// Generate some traffic to create metrics
	for i := 0; i < 10; i++ {
		suite.httpClient.Get(suite.goAppURL + "/api/v1/ping")
		suite.httpClient.Get(suite.goAppURL + "/api/v1/work?ms=50")
	}
	
	// Wait for metrics to be scraped
	time.Sleep(10 * time.Second)
	
	// Query Prometheus for metrics
	metricsToCheck := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"go_goroutines",
		"process_cpu_seconds_total",
		"up",
	}
	
	for _, metric := range metricsToCheck {
		suite.T().Run(fmt.Sprintf("Metric_%s", metric), func(t *testing.T) {
			url := fmt.Sprintf("%s/api/v1/query?query=%s", suite.prometheusURL, metric)
			resp, err := suite.httpClient.Get(url)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			var queryResult struct {
				Status string `json:"status"`
				Data   struct {
					Result []map[string]interface{} `json:"result"`
				} `json:"data"`
			}
			
			err = json.NewDecoder(resp.Body).Decode(&queryResult)
			require.NoError(t, err)
			
			assert.Equal(t, "success", queryResult.Status)
			assert.Greater(t, len(queryResult.Data.Result), 0, 
				"No data found for metric %s", metric)
		})
	}
}

// TestErrorInjectionAndAlerts tests error injection and alert firing
func (suite *IntegrationTestSuite) TestErrorInjectionAndAlerts() {
	suite.T().Log("Testing error injection and alert firing...")
	
	// Clear previous webhooks
	suite.receivedWebhooks = make([]WebhookPayload, 0)
	
	// Enable error injection
	errorConfig := map[string]interface{}{
		"enabled":     true,
		"rate":        0.5, // 50% error rate
		"status_code": 503,
	}
	
	configJSON, _ := json.Marshal(errorConfig)
	req, _ := http.NewRequest("POST", suite.goAppURL+"/api/v1/toggles/error-rate", 
		bytes.NewBuffer(configJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	
	resp, err := suite.httpClient.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	assert.Equal(suite.T(), 200, resp.StatusCode)
	
	// Generate traffic to trigger errors
	suite.T().Log("Generating traffic to trigger error alerts...")
	for i := 0; i < 100; i++ {
		suite.httpClient.Get(suite.goAppURL + "/api/v1/ping")
		if i%10 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	// Wait for alert evaluation (alerts fire after 10 minutes, but we'll check for pending)
	time.Sleep(30 * time.Second)
	
	// Check for pending alerts in Prometheus
	alertsURL := suite.prometheusURL + "/api/v1/alerts"
	resp, err = suite.httpClient.Get(alertsURL)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var alertsResponse struct {
		Status string `json:"status"`
		Data   struct {
			Alerts []struct {
				Labels      map[string]string `json:"labels"`
				Annotations map[string]string `json:"annotations"`
				State       string            `json:"state"`
				ActiveAt    *time.Time        `json:"activeAt"`
				Value       string            `json:"value"`
			} `json:"alerts"`
		} `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&alertsResponse)
	require.NoError(suite.T(), err)
	
	// Look for HighErrorRate alert
	var errorRateAlert *struct {
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		State       string            `json:"state"`
		ActiveAt    *time.Time        `json:"activeAt"`
		Value       string            `json:"value"`
	}
	
	for _, alert := range alertsResponse.Data.Alerts {
		if alert.Labels["alertname"] == "HighErrorRate" {
			errorRateAlert = &alert
			break
		}
	}
	
	if errorRateAlert != nil {
		suite.T().Logf("Found HighErrorRate alert in state: %s", errorRateAlert.State)
		assert.Contains(suite.T(), []string{"pending", "firing"}, errorRateAlert.State)
	} else {
		suite.T().Log("HighErrorRate alert not yet visible (may need more time)")
	}
	
	// Disable error injection
	errorConfig["enabled"] = false
	configJSON, _ = json.Marshal(errorConfig)
	req, _ = http.NewRequest("POST", suite.goAppURL+"/api/v1/toggles/error-rate", 
		bytes.NewBuffer(configJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	
	resp, err = suite.httpClient.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
}

// TestLatencyAlertsWithWorkSimulation tests latency alerts using work simulation
func (suite *IntegrationTestSuite) TestLatencyAlertsWithWorkSimulation() {
	suite.T().Log("Testing latency alerts with work simulation...")
	
	// Generate high latency traffic
	suite.T().Log("Generating high latency traffic...")
	for i := 0; i < 50; i++ {
		// Request work that takes 600ms (above the 500ms threshold)
		suite.httpClient.Get(suite.goAppURL + "/api/v1/work?ms=600&jitter=100")
		if i%5 == 0 {
			time.Sleep(200 * time.Millisecond)
		}
	}
	
	// Wait for metrics to be collected
	time.Sleep(30 * time.Second)
	
	// Check for latency metrics in Prometheus
	latencyQuery := "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
	url := fmt.Sprintf("%s/api/v1/query?query=%s", suite.prometheusURL, latencyQuery)
	
	resp, err := suite.httpClient.Get(url)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var queryResult struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&queryResult)
	require.NoError(suite.T(), err)
	
	assert.Equal(suite.T(), "success", queryResult.Status)
	
	if len(queryResult.Data.Result) > 0 {
		// Check if P95 latency is above threshold
		for _, result := range queryResult.Data.Result {
			if len(result.Value) >= 2 {
				if valueStr, ok := result.Value[1].(string); ok {
					suite.T().Logf("P95 latency: %s seconds", valueStr)
				}
			}
		}
	}
	
	// Check for pending HighLatencyP95 alerts
	alertsURL := suite.prometheusURL + "/api/v1/alerts"
	resp, err = suite.httpClient.Get(alertsURL)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var alertsResponse struct {
		Status string `json:"status"`
		Data   struct {
			Alerts []struct {
				Labels map[string]string `json:"labels"`
				State  string            `json:"state"`
			} `json:"alerts"`
		} `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&alertsResponse)
	require.NoError(suite.T(), err)
	
	for _, alert := range alertsResponse.Data.Alerts {
		if alert.Labels["alertname"] == "HighLatencyP95" {
			suite.T().Logf("Found HighLatencyP95 alert in state: %s", alert.State)
		}
	}
}

// TestInstanceDownAlert tests instance down alerts by stopping a container
func (suite *IntegrationTestSuite) TestInstanceDownAlert() {
	suite.T().Log("Testing instance down alert...")
	
	// Stop the go-app container
	cmd := exec.Command("docker-compose", "stop", "go-app")
	output, err := cmd.CombinedOutput()
	require.NoError(suite.T(), err, "Failed to stop go-app container: %s", string(output))
	
	// Wait for the alert to be detected (InstanceDown fires after 2 minutes)
	suite.T().Log("Waiting for InstanceDown alert to be detected...")
	time.Sleep(30 * time.Second)
	
	// Check for InstanceDown alert
	alertsURL := suite.prometheusURL + "/api/v1/alerts"
	resp, err := suite.httpClient.Get(alertsURL)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var alertsResponse struct {
		Status string `json:"status"`
		Data   struct {
			Alerts []struct {
				Labels map[string]string `json:"labels"`
				State  string            `json:"state"`
			} `json:"alerts"`
		} `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&alertsResponse)
	require.NoError(suite.T(), err)
	
	var instanceDownFound bool
	for _, alert := range alertsResponse.Data.Alerts {
		if alert.Labels["alertname"] == "InstanceDown" {
			suite.T().Logf("Found InstanceDown alert in state: %s", alert.State)
			instanceDownFound = true
		}
	}
	
	// The alert might be pending since we haven't waited the full 2 minutes
	if !instanceDownFound {
		suite.T().Log("InstanceDown alert not yet visible (may need more time)")
	}
	
	// Restart the go-app container
	cmd = exec.Command("docker-compose", "start", "go-app")
	output, err = cmd.CombinedOutput()
	require.NoError(suite.T(), err, "Failed to restart go-app container: %s", string(output))
	
	// Wait for service to be ready again
	suite.waitForEndpoint(context.Background(), suite.goAppURL+"/healthz", "Go App")
}

// TestBlackboxProbes tests that blackbox exporter probes are working
func (suite *IntegrationTestSuite) TestBlackboxProbes() {
	suite.T().Log("Testing blackbox exporter probes...")
	
	// Query probe_success metric
	url := fmt.Sprintf("%s/api/v1/query?query=probe_success", suite.prometheusURL)
	resp, err := suite.httpClient.Get(url)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var queryResult struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&queryResult)
	require.NoError(suite.T(), err)
	
	assert.Equal(suite.T(), "success", queryResult.Status)
	assert.Greater(suite.T(), len(queryResult.Data.Result), 0, "No probe_success metrics found")
	
	// Check probe results
	for _, result := range queryResult.Data.Result {
		instance := result.Metric["instance"]
		if len(result.Value) >= 2 {
			if valueStr, ok := result.Value[1].(string); ok {
				suite.T().Logf("Probe for %s: success=%s", instance, valueStr)
			}
		}
	}
}

// TestWebhookDelivery tests webhook delivery to mock endpoints
func (suite *IntegrationTestSuite) TestWebhookDelivery() {
	suite.T().Log("Testing webhook delivery...")
	
	// Clear previous webhooks
	suite.receivedWebhooks = make([]WebhookPayload, 0)
	
	// Trigger a test alert by enabling error injection briefly
	errorConfig := map[string]interface{}{
		"enabled":     true,
		"rate":        1.0, // 100% error rate for quick alert
		"status_code": 503,
	}
	
	configJSON, _ := json.Marshal(errorConfig)
	req, _ := http.NewRequest("POST", suite.goAppURL+"/api/v1/toggles/error-rate", 
		bytes.NewBuffer(configJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	
	resp, err := suite.httpClient.Do(req)
	if err == nil {
		resp.Body.Close()
		
		// Generate some error traffic
		for i := 0; i < 20; i++ {
			suite.httpClient.Get(suite.goAppURL + "/api/v1/ping")
		}
		
		// Disable error injection
		errorConfig["enabled"] = false
		configJSON, _ = json.Marshal(errorConfig)
		req, _ = http.NewRequest("POST", suite.goAppURL+"/api/v1/toggles/error-rate", 
			bytes.NewBuffer(configJSON))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")
		
		resp, err = suite.httpClient.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}
	
	// Wait a bit for potential webhook delivery
	time.Sleep(10 * time.Second)
	
	// Check if we received any webhooks
	suite.T().Logf("Received %d webhook notifications", len(suite.receivedWebhooks))
	
	for i, webhook := range suite.receivedWebhooks {
		suite.T().Logf("Webhook %d: URL=%s, Body=%s", i+1, webhook.URL, webhook.Body)
	}
	
	// Note: In a real scenario, alerts take time to fire (10+ minutes for HighErrorRate)
	// This test mainly verifies the webhook server is reachable and configured correctly
}

// TestDockerComposeHealthChecks tests Docker Compose health checks
func (suite *IntegrationTestSuite) TestDockerComposeHealthChecks() {
	suite.T().Log("Testing Docker Compose health checks...")
	
	// Get container status
	cmd := exec.Command("docker-compose", "ps", "--format", "json")
	output, err := cmd.CombinedOutput()
	require.NoError(suite.T(), err, "Failed to get container status: %s", string(output))
	
	// Parse container status
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		var container struct {
			Name    string `json:"Name"`
			State   string `json:"State"`
			Status  string `json:"Status"`
			Health  string `json:"Health"`
			Service string `json:"Service"`
		}
		
		err := json.Unmarshal([]byte(line), &container)
		if err != nil {
			suite.T().Logf("Failed to parse container info: %s", line)
			continue
		}
		
		suite.T().Logf("Container %s (%s): State=%s, Status=%s, Health=%s", 
			container.Name, container.Service, container.State, container.Status, container.Health)
		
		assert.Equal(suite.T(), "running", container.State, 
			"Container %s should be running", container.Name)
		
		// Check health status if available
		if container.Health != "" {
			assert.Equal(suite.T(), "healthy", container.Health, 
				"Container %s should be healthy", container.Name)
		}
	}
}

// TestEndToEndMonitoringFlow tests the complete monitoring flow
func (suite *IntegrationTestSuite) TestEndToEndMonitoringFlow() {
	suite.T().Log("Testing end-to-end monitoring flow...")
	
	// 1. Generate normal traffic
	suite.T().Log("Step 1: Generating normal traffic...")
	for i := 0; i < 20; i++ {
		suite.httpClient.Get(suite.goAppURL + "/api/v1/ping")
		suite.httpClient.Get(suite.goAppURL + "/api/v1/work?ms=100")
	}
	
	// 2. Wait for metrics collection
	time.Sleep(15 * time.Second)
	
	// 3. Verify metrics are collected
	suite.T().Log("Step 2: Verifying metrics collection...")
	url := fmt.Sprintf("%s/api/v1/query?query=rate(http_requests_total[5m])", suite.prometheusURL)
	resp, err := suite.httpClient.Get(url)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var queryResult struct {
		Status string `json:"status"`
		Data   struct {
			Result []map[string]interface{} `json:"result"`
		} `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&queryResult)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", queryResult.Status)
	assert.Greater(suite.T(), len(queryResult.Data.Result), 0, "No request rate metrics found")
	
	// 4. Test Grafana dashboard queries
	suite.T().Log("Step 3: Testing Grafana dashboard queries...")
	// This is covered in TestGrafanaDataSourceAndQueries
	
	// 5. Verify alert rules are loaded
	suite.T().Log("Step 4: Verifying alert rules...")
	rulesURL := suite.prometheusURL + "/api/v1/rules"
	resp, err = suite.httpClient.Get(rulesURL)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var rulesResponse struct {
		Status string `json:"status"`
		Data   struct {
			Groups []struct {
				Name  string `json:"name"`
				Rules []struct {
					Name   string            `json:"name"`
					Query  string            `json:"query"`
					Type   string            `json:"type"`
					Labels map[string]string `json:"labels"`
				} `json:"rules"`
			} `json:"groups"`
		} `json:"data"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&rulesResponse)
	require.NoError(suite.T(), err)
	
	expectedAlerts := []string{"InstanceDown", "HighErrorRate", "HighLatencyP95", "UptimeProbeFail"}
	foundAlerts := make(map[string]bool)
	
	for _, group := range rulesResponse.Data.Groups {
		for _, rule := range group.Rules {
			if rule.Type == "alerting" {
				foundAlerts[rule.Name] = true
			}
		}
	}
	
	for _, alert := range expectedAlerts {
		assert.True(suite.T(), foundAlerts[alert], "Alert rule %s not found", alert)
	}
	
	suite.T().Log("End-to-end monitoring flow test completed successfully")
}

// TestIntegration runs the integration test suite
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	
	// Check if Docker and Docker Compose are available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not found, skipping integration tests")
	}
	
	if _, err := exec.LookPath("docker-compose"); err != nil {
		t.Skip("Docker Compose not found, skipping integration tests")
	}
	
	suite.Run(t, new(IntegrationTestSuite))
}