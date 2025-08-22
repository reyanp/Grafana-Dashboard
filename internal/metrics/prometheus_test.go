package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	
	if registry == nil {
		t.Fatal("Expected registry to be created, got nil")
	}
	
	if registry.registry == nil {
		t.Fatal("Expected prometheus registry to be initialized")
	}
}

func TestRecordHTTPRequest(t *testing.T) {
	registry := NewRegistry()
	
	// Record a test HTTP request
	registry.RecordHTTPRequest("GET", "/api/v1/ping", 200, 100*time.Millisecond)
	
	// Get the metrics handler and make a request to it
	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	body := w.Body.String()
	
	// Check that the metrics are present
	if !strings.Contains(body, "http_requests_total") {
		t.Error("Expected http_requests_total metric to be present")
	}
	
	if !strings.Contains(body, "http_request_duration_seconds") {
		t.Error("Expected http_request_duration_seconds metric to be present")
	}
	
	// Check that our specific metric was recorded
	if !strings.Contains(body, `http_requests_total{method="GET",route="/api/v1/ping",status="200"} 1`) {
		t.Error("Expected specific http_requests_total metric to be recorded")
	}
}

func TestWorkMetrics(t *testing.T) {
	registry := NewRegistry()
	
	// Test work jobs inflight
	registry.IncWorkJobsInflight()
	registry.IncWorkJobsInflight()
	registry.DecWorkJobsInflight()
	
	// Test work failures
	registry.IncWorkFailures("simulate_work")
	
	// Get the metrics
	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	body := w.Body.String()
	
	// Check work metrics are present
	if !strings.Contains(body, "work_jobs_inflight") {
		t.Error("Expected work_jobs_inflight metric to be present")
	}
	
	if !strings.Contains(body, "work_failures_total") {
		t.Error("Expected work_failures_total metric to be present")
	}
	
	// Check that work_jobs_inflight shows 1 (2 inc - 1 dec)
	if !strings.Contains(body, "work_jobs_inflight 1") {
		t.Error("Expected work_jobs_inflight to be 1")
	}
	
	// Check that work_failures_total shows 1
	if !strings.Contains(body, `work_failures_total{operation="simulate_work"} 1`) {
		t.Error("Expected work_failures_total to be 1 for simulate_work operation")
	}
}

func TestGoMetrics(t *testing.T) {
	registry := NewRegistry()
	
	// Get the metrics
	handler := registry.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	body := w.Body.String()
	
	// Check that Go runtime metrics are present
	if !strings.Contains(body, "go_goroutines") {
		t.Error("Expected go_goroutines metric to be present")
	}
	
	if !strings.Contains(body, "process_cpu_seconds_total") {
		t.Error("Expected process_cpu_seconds_total metric to be present")
	}
	
	if !strings.Contains(body, "process_resident_memory_bytes") {
		t.Error("Expected process_resident_memory_bytes metric to be present")
	}
}

func TestGetInflightJobs(t *testing.T) {
	registry := NewRegistry()
	
	// Initially should be 0
	if jobs := registry.GetInflightJobs(); jobs != 0 {
		t.Errorf("Expected 0 inflight jobs, got %f", jobs)
	}
	
	// Increment and check
	registry.IncWorkJobsInflight()
	if jobs := registry.GetInflightJobs(); jobs != 1 {
		t.Errorf("Expected 1 inflight job, got %f", jobs)
	}
	
	// Increment again and check
	registry.IncWorkJobsInflight()
	if jobs := registry.GetInflightJobs(); jobs != 2 {
		t.Errorf("Expected 2 inflight jobs, got %f", jobs)
	}
	
	// Decrement and check
	registry.DecWorkJobsInflight()
	if jobs := registry.GetInflightJobs(); jobs != 1 {
		t.Errorf("Expected 1 inflight job, got %f", jobs)
	}
	
	// Decrement to 0 and check
	registry.DecWorkJobsInflight()
	if jobs := registry.GetInflightJobs(); jobs != 0 {
		t.Errorf("Expected 0 inflight jobs, got %f", jobs)
	}
}

func TestFlush(t *testing.T) {
	registry := NewRegistry()
	
	// Record some metrics
	registry.RecordHTTPRequest("GET", "/test", 200, 100*time.Millisecond)
	registry.IncWorkJobsInflight()
	registry.IncWorkFailures("test_operation")
	
	// Test flush - should not return error
	err := registry.Flush()
	if err != nil {
		t.Errorf("Flush() returned error: %v", err)
	}
	
	// Verify metrics are still accessible after flush
	families, err := registry.GetRegistry().Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics after flush: %v", err)
	}
	
	if len(families) == 0 {
		t.Error("Expected metrics to still be available after flush")
	}
}