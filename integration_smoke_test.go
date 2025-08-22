package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SmokeTestSuite contains basic smoke tests for integration test components
type SmokeTestSuite struct {
	suite.Suite
}

// TestIntegrationTestStructure tests that the integration test structure is valid
func (suite *SmokeTestSuite) TestIntegrationTestStructure() {
	// Test that we can create the integration test suite
	integrationSuite := &IntegrationTestSuite{}
	assert.NotNil(suite.T(), integrationSuite)
	
	// Test that service URLs are properly formatted
	integrationSuite.goAppURL = "http://localhost:8080"
	integrationSuite.prometheusURL = "http://localhost:9090"
	integrationSuite.grafanaURL = "http://localhost:3000"
	integrationSuite.alertmanagerURL = "http://localhost:9093"
	integrationSuite.nodeExporterURL = "http://localhost:9100"
	integrationSuite.blackboxURL = "http://localhost:9115"
	
	assert.Contains(suite.T(), integrationSuite.goAppURL, "localhost:8080")
	assert.Contains(suite.T(), integrationSuite.prometheusURL, "localhost:9090")
	assert.Contains(suite.T(), integrationSuite.grafanaURL, "localhost:3000")
}

// TestWebhookPayloadStructure tests the webhook payload structure
func (suite *SmokeTestSuite) TestWebhookPayloadStructure() {
	payload := WebhookPayload{
		Timestamp: time.Now(),
		Headers:   map[string]string{"Content-Type": "application/json"},
		Body:      `{"test": "data"}`,
		URL:       "/test",
	}
	
	assert.NotZero(suite.T(), payload.Timestamp)
	assert.Equal(suite.T(), "application/json", payload.Headers["Content-Type"])
	assert.Contains(suite.T(), payload.Body, "test")
	assert.Equal(suite.T(), "/test", payload.URL)
}

// TestIntegrationSmoke runs the smoke test suite
func TestIntegrationSmoke(t *testing.T) {
	suite.Run(t, new(SmokeTestSuite))
}