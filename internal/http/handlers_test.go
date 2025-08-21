package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"monitoring-dashboard-automation/internal/health"
	"monitoring-dashboard-automation/internal/metrics"

	"go.uber.org/zap"
)

func TestNewHealthHandlers(t *testing.T) {
	checker := health.NewChecker()
	handlers := NewHealthHandlers(checker)
	
	if handlers == nil {
		t.Fatal("NewHealthHandlers() returned nil")
	}
	
	if handlers.checker != checker {
		t.Error("NewHealthHandlers() did not set checker correctly")
	}
}

func TestHealthHandlers_Liveness(t *testing.T) {
	checker := health.NewChecker()
	handlers := NewHealthHandlers(checker)
	
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	
	handlers.Liveness(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
	
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}
}

func TestHealthHandlers_Readiness_Success(t *testing.T) {
	checker := health.NewChecker()
	handlers := NewHealthHandlers(checker)
	
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	
	handlers.Readiness(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	if w.Body.String() != "Ready" {
		t.Errorf("Expected body 'Ready', got '%s'", w.Body.String())
	}
	
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}
}

func TestHealthHandlers_Readiness_Failure(t *testing.T) {
	checker := health.NewChecker()
	checker.AddCheck("test", func(ctx context.Context) error {
		return errors.New("test failure")
	})
	handlers := NewHealthHandlers(checker)
	
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	
	handlers.Readiness(w, req)
	
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
	
	body := w.Body.String()
	if !contains(body, "Not Ready") {
		t.Errorf("Expected body to contain 'Not Ready', got '%s'", body)
	}
	
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}
}

func TestHealthHandlers_ToggleReadiness_EnableFailure(t *testing.T) {
	checker := health.NewChecker()
	handlers := NewHealthHandlers(checker)
	
	reqBody := map[string]bool{
		"force_failure": true,
	}
	jsonBody, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/api/v1/toggles/readiness", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handlers.ToggleReadiness(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["force_failure"] != true {
		t.Errorf("Expected force_failure to be true, got %v", response["force_failure"])
	}
	
	if !checker.IsForceFailure() {
		t.Error("Expected checker to have force failure enabled")
	}
}

func TestHealthHandlers_ToggleReadiness_DisableFailure(t *testing.T) {
	checker := health.NewChecker()
	checker.SetForceFailure(true) // Start with failure enabled
	handlers := NewHealthHandlers(checker)
	
	reqBody := map[string]bool{
		"force_failure": false,
	}
	jsonBody, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/api/v1/toggles/readiness", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handlers.ToggleReadiness(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["force_failure"] != false {
		t.Errorf("Expected force_failure to be false, got %v", response["force_failure"])
	}
	
	if checker.IsForceFailure() {
		t.Error("Expected checker to have force failure disabled")
	}
}

func TestHealthHandlers_ToggleReadiness_InvalidJSON(t *testing.T) {
	checker := health.NewChecker()
	handlers := NewHealthHandlers(checker)
	
	req := httptest.NewRequest("POST", "/api/v1/toggles/readiness", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handlers.ToggleReadiness(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	
	body := w.Body.String()
	if !contains(body, "Invalid JSON") {
		t.Errorf("Expected body to contain 'Invalid JSON', got '%s'", body)
	}
}

func TestHealthHandlers_Integration_ToggleAndCheck(t *testing.T) {
	checker := health.NewChecker()
	handlers := NewHealthHandlers(checker)
	
	// First, verify readiness is OK
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	handlers.Readiness(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected initial readiness to be OK, got %d", w.Code)
	}
	
	// Toggle to force failure
	reqBody := map[string]bool{"force_failure": true}
	jsonBody, _ := json.Marshal(reqBody)
	req = httptest.NewRequest("POST", "/api/v1/toggles/readiness", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handlers.ToggleReadiness(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected toggle to succeed, got %d", w.Code)
	}
	
	// Now check readiness should fail
	req = httptest.NewRequest("GET", "/readyz", nil)
	w = httptest.NewRecorder()
	handlers.Readiness(w, req)
	
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected readiness to fail after toggle, got %d", w.Code)
	}
	
	// Toggle back to normal
	reqBody = map[string]bool{"force_failure": false}
	jsonBody, _ = json.Marshal(reqBody)
	req = httptest.NewRequest("POST", "/api/v1/toggles/readiness", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handlers.ToggleReadiness(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected toggle back to succeed, got %d", w.Code)
	}
	
	// Readiness should be OK again
	req = httptest.NewRequest("GET", "/readyz", nil)
	w = httptest.NewRecorder()
	handlers.Readiness(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected readiness to be OK after toggle back, got %d", w.Code)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// API Handler Tests

func TestNewAPIHandlers(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	if handlers == nil {
		t.Fatal("NewAPIHandlers() returned nil")
	}
	
	if handlers.logger != logger {
		t.Error("NewAPIHandlers() did not set logger correctly")
	}
	
	if handlers.metrics != metricsRegistry {
		t.Error("NewAPIHandlers() did not set metrics correctly")
	}
}

func TestAPIHandlers_Ping(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	req := httptest.NewRequest("GET", "/api/v1/ping", nil)
	w := httptest.NewRecorder()
	
	handlers.Ping(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["message"] != "pong" {
		t.Errorf("Expected message 'pong', got %v", response["message"])
	}
	
	if _, ok := response["timestamp"]; !ok {
		t.Error("Expected timestamp field in response")
	}
}

func TestAPIHandlers_Work_DefaultParameters(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	req := httptest.NewRequest("GET", "/api/v1/work", nil)
	w := httptest.NewRecorder()
	
	start := time.Now()
	handlers.Work(w, req)
	duration := time.Since(start)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["message"] != "work completed" {
		t.Errorf("Expected message 'work completed', got %v", response["message"])
	}
	
	if response["requested_ms"] != float64(100) {
		t.Errorf("Expected requested_ms 100, got %v", response["requested_ms"])
	}
	
	if response["jitter_ms"] != float64(0) {
		t.Errorf("Expected jitter_ms 0, got %v", response["jitter_ms"])
	}
	
	// Check that it actually took approximately the right amount of time
	if duration < 90*time.Millisecond || duration > 150*time.Millisecond {
		t.Errorf("Expected duration around 100ms, got %v", duration)
	}
	
	if _, ok := response["timestamp"]; !ok {
		t.Error("Expected timestamp field in response")
	}
	
	if _, ok := response["actual_duration_ms"]; !ok {
		t.Error("Expected actual_duration_ms field in response")
	}
}

func TestAPIHandlers_Work_CustomParameters(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	// Test with custom ms and jitter parameters
	params := url.Values{}
	params.Add("ms", "200")
	params.Add("jitter", "50")
	
	req := httptest.NewRequest("GET", "/api/v1/work?"+params.Encode(), nil)
	w := httptest.NewRecorder()
	
	start := time.Now()
	handlers.Work(w, req)
	duration := time.Since(start)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["requested_ms"] != float64(200) {
		t.Errorf("Expected requested_ms 200, got %v", response["requested_ms"])
	}
	
	if response["jitter_ms"] != float64(50) {
		t.Errorf("Expected jitter_ms 50, got %v", response["jitter_ms"])
	}
	
	// Duration should be between 200ms and 250ms (200ms + up to 50ms jitter)
	if duration < 190*time.Millisecond || duration > 280*time.Millisecond {
		t.Errorf("Expected duration between 200-250ms, got %v", duration)
	}
}

func TestAPIHandlers_Work_InvalidParameters(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	// Test with invalid parameters - should use defaults
	params := url.Values{}
	params.Add("ms", "invalid")
	params.Add("jitter", "also_invalid")
	
	req := httptest.NewRequest("GET", "/api/v1/work?"+params.Encode(), nil)
	w := httptest.NewRecorder()
	
	handlers.Work(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	// Should fall back to defaults
	if response["requested_ms"] != float64(100) {
		t.Errorf("Expected requested_ms 100 (default), got %v", response["requested_ms"])
	}
	
	if response["jitter_ms"] != float64(0) {
		t.Errorf("Expected jitter_ms 0 (default), got %v", response["jitter_ms"])
	}
}

func TestAPIHandlers_Work_NegativeParameters(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	// Test with negative parameters - should use defaults
	params := url.Values{}
	params.Add("ms", "-100")
	params.Add("jitter", "-50")
	
	req := httptest.NewRequest("GET", "/api/v1/work?"+params.Encode(), nil)
	w := httptest.NewRecorder()
	
	handlers.Work(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	// Should fall back to defaults for negative values
	if response["requested_ms"] != float64(100) {
		t.Errorf("Expected requested_ms 100 (default), got %v", response["requested_ms"])
	}
	
	if response["jitter_ms"] != float64(0) {
		t.Errorf("Expected jitter_ms 0 (default), got %v", response["jitter_ms"])
	}
}

func TestAPIHandlers_Work_ContextCancellation(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	// Create a request with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	params := url.Values{}
	params.Add("ms", "200") // Request 200ms of work, but timeout after 50ms
	
	req := httptest.NewRequest("GET", "/api/v1/work?"+params.Encode(), nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	
	handlers.Work(w, req)
	
	if w.Code != http.StatusRequestTimeout {
		t.Errorf("Expected status %d, got %d", http.StatusRequestTimeout, w.Code)
	}
	
	body := w.Body.String()
	if !contains(body, "Work simulation cancelled") {
		t.Errorf("Expected body to contain 'Work simulation cancelled', got '%s'", body)
	}
}

func TestAPIHandlers_Work_ZeroParameters(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	// Test with zero parameters
	params := url.Values{}
	params.Add("ms", "0")
	params.Add("jitter", "0")
	
	req := httptest.NewRequest("GET", "/api/v1/work?"+params.Encode(), nil)
	w := httptest.NewRecorder()
	
	start := time.Now()
	handlers.Work(w, req)
	duration := time.Since(start)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if response["requested_ms"] != float64(0) {
		t.Errorf("Expected requested_ms 0, got %v", response["requested_ms"])
	}
	
	if response["jitter_ms"] != float64(0) {
		t.Errorf("Expected jitter_ms 0, got %v", response["jitter_ms"])
	}
	
	// Should complete very quickly
	if duration > 10*time.Millisecond {
		t.Errorf("Expected duration to be very short, got %v", duration)
	}
}

func TestAPIHandlers_SimulateWork(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	// Test normal completion
	ctx := context.Background()
	duration := 50 * time.Millisecond
	
	start := time.Now()
	err := handlers.simulateWork(ctx, duration)
	elapsed := time.Since(start)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if elapsed < 40*time.Millisecond || elapsed > 70*time.Millisecond {
		t.Errorf("Expected duration around 50ms, got %v", elapsed)
	}
}

func TestAPIHandlers_SimulateWork_Cancellation(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	duration := 200 * time.Millisecond
	
	// Cancel after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	
	start := time.Now()
	err := handlers.simulateWork(ctx, duration)
	elapsed := time.Since(start)
	
	if err == nil {
		t.Error("Expected context cancellation error")
	}
	
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
	
	// Should complete in around 50ms, not 200ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected early completion due to cancellation, got %v", elapsed)
	}
}

func TestAPIHandlers_SimulateWork_Timeout(t *testing.T) {
	logger := zap.NewNop()
	metricsRegistry := metrics.NewRegistry()
	handlers := NewAPIHandlers(logger, metricsRegistry)
	
	// Test context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	duration := 200 * time.Millisecond
	
	start := time.Now()
	err := handlers.simulateWork(ctx, duration)
	elapsed := time.Since(start)
	
	if err == nil {
		t.Error("Expected context timeout error")
	}
	
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
	
	// Should complete in around 50ms, not 200ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected early completion due to timeout, got %v", elapsed)
	}
}

func TestToggleHandlers_ErrorRate_ValidRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Create mock error toggle
	mockToggle := &mockToggleInterface{
		enabled:    false,
		rate:       0.0,
		statusCode: 500,
	}
	
	handlers := NewToggleHandlers(logger, mockToggle)
	
	// Create valid request
	reqBody := `{"enabled": true, "rate": 0.5, "status_code": 503}`
	req := httptest.NewRequest("POST", "/api/v1/toggles/error-rate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute request
	handlers.ErrorRate(w, req)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Check that toggle was updated
	if !mockToggle.enabled {
		t.Error("Expected toggle to be enabled")
	}
	if mockToggle.rate != 0.5 {
		t.Errorf("Expected rate to be 0.5, got %f", mockToggle.rate)
	}
	if mockToggle.statusCode != 503 {
		t.Errorf("Expected status code to be 503, got %d", mockToggle.statusCode)
	}
	
	// Check response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["enabled"] != true {
		t.Errorf("Expected enabled to be true in response, got %v", response["enabled"])
	}
	if response["rate"] != 0.5 {
		t.Errorf("Expected rate to be 0.5 in response, got %v", response["rate"])
	}
	if response["status_code"] != float64(503) {
		t.Errorf("Expected status_code to be 503 in response, got %v", response["status_code"])
	}
}

func TestToggleHandlers_ErrorRate_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Create mock error toggle
	mockToggle := &mockToggleInterface{
		enabled:    false,
		rate:       0.0,
		statusCode: 500,
	}
	
	handlers := NewToggleHandlers(logger, mockToggle)
	
	// Create invalid JSON request
	reqBody := `{"enabled": true, "rate": invalid}`
	req := httptest.NewRequest("POST", "/api/v1/toggles/error-rate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute request
	handlers.ErrorRate(w, req)
	
	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestToggleHandlers_ErrorRate_InvalidRate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Create mock error toggle
	mockToggle := &mockToggleInterface{
		enabled:    false,
		rate:       0.0,
		statusCode: 500,
	}
	
	handlers := NewToggleHandlers(logger, mockToggle)
	
	// Create request with invalid rate (> 1.0)
	reqBody := `{"enabled": true, "rate": 1.5, "status_code": 503}`
	req := httptest.NewRequest("POST", "/api/v1/toggles/error-rate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute request
	handlers.ErrorRate(w, req)
	
	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	
	// Test with negative rate
	reqBody = `{"enabled": true, "rate": -0.1, "status_code": 503}`
	req = httptest.NewRequest("POST", "/api/v1/toggles/error-rate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	handlers.ErrorRate(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for negative rate, got %d", w.Code)
	}
}

func TestToggleHandlers_ErrorRate_InvalidStatusCode(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Create mock error toggle
	mockToggle := &mockToggleInterface{
		enabled:    false,
		rate:       0.0,
		statusCode: 500,
	}
	
	handlers := NewToggleHandlers(logger, mockToggle)
	
	// Create request with invalid status code (< 500)
	reqBody := `{"enabled": true, "rate": 0.5, "status_code": 400}`
	req := httptest.NewRequest("POST", "/api/v1/toggles/error-rate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute request
	handlers.ErrorRate(w, req)
	
	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	
	// Test with status code > 599
	reqBody = `{"enabled": true, "rate": 0.5, "status_code": 600}`
	req = httptest.NewRequest("POST", "/api/v1/toggles/error-rate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	handlers.ErrorRate(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for status code > 599, got %d", w.Code)
	}
}

// Mock toggle interface for testing
type mockToggleInterface struct {
	enabled    bool
	rate       float64
	statusCode int
}

func (m *mockToggleInterface) SetConfig(enabled bool, rate float64, statusCode int) {
	m.enabled = enabled
	m.rate = rate
	m.statusCode = statusCode
}

func (m *mockToggleInterface) GetConfig() (bool, float64, int) {
	return m.enabled, m.rate, m.statusCode
}