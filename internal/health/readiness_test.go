package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewChecker(t *testing.T) {
	checker := NewChecker()
	if checker == nil {
		t.Fatal("NewChecker() returned nil")
	}
	if checker.checks == nil {
		t.Fatal("NewChecker() did not initialize checks map")
	}
}

func TestChecker_AddCheck(t *testing.T) {
	checker := NewChecker()
	
	checkFunc := func(ctx context.Context) error {
		return nil
	}
	
	checker.AddCheck("test", checkFunc)
	
	if len(checker.checks) != 1 {
		t.Errorf("Expected 1 check, got %d", len(checker.checks))
	}
	
	if _, exists := checker.checks["test"]; !exists {
		t.Error("Check 'test' was not added")
	}
}

func TestChecker_RemoveCheck(t *testing.T) {
	checker := NewChecker()
	
	checkFunc := func(ctx context.Context) error {
		return nil
	}
	
	checker.AddCheck("test", checkFunc)
	checker.RemoveCheck("test")
	
	if len(checker.checks) != 0 {
		t.Errorf("Expected 0 checks, got %d", len(checker.checks))
	}
}

func TestChecker_SetForceFailure(t *testing.T) {
	checker := NewChecker()
	
	// Initially should be false
	if checker.IsForceFailure() {
		t.Error("Expected force failure to be false initially")
	}
	
	// Set to true
	checker.SetForceFailure(true)
	if !checker.IsForceFailure() {
		t.Error("Expected force failure to be true after setting")
	}
	
	// Set back to false
	checker.SetForceFailure(false)
	if checker.IsForceFailure() {
		t.Error("Expected force failure to be false after resetting")
	}
}

func TestChecker_CheckReadiness_Success(t *testing.T) {
	checker := NewChecker()
	
	// Add a successful check
	checker.AddCheck("test", func(ctx context.Context) error {
		return nil
	})
	
	ctx := context.Background()
	err := checker.CheckReadiness(ctx)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestChecker_CheckReadiness_Failure(t *testing.T) {
	checker := NewChecker()
	
	// Add a failing check
	checker.AddCheck("test", func(ctx context.Context) error {
		return errors.New("test failure")
	})
	
	ctx := context.Background()
	err := checker.CheckReadiness(ctx)
	
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	healthErr, ok := err.(*HealthCheckError)
	if !ok {
		t.Errorf("Expected HealthCheckError, got %T", err)
	}
	
	if healthErr.Component != "test" {
		t.Errorf("Expected component 'test', got '%s'", healthErr.Component)
	}
}

func TestChecker_CheckReadiness_ForceFailure(t *testing.T) {
	checker := NewChecker()
	
	// Add a successful check
	checker.AddCheck("test", func(ctx context.Context) error {
		return nil
	})
	
	// Force failure
	checker.SetForceFailure(true)
	
	ctx := context.Background()
	err := checker.CheckReadiness(ctx)
	
	if err == nil {
		t.Error("Expected error due to force failure, got nil")
	}
	
	healthErr, ok := err.(*HealthCheckError)
	if !ok {
		t.Errorf("Expected HealthCheckError, got %T", err)
	}
	
	if healthErr.Component != "forced" {
		t.Errorf("Expected component 'forced', got '%s'", healthErr.Component)
	}
}

func TestChecker_CheckReadiness_Timeout(t *testing.T) {
	checker := NewChecker()
	
	// Add a slow check that will timeout
	checker.AddCheck("slow", func(ctx context.Context) error {
		select {
		case <-time.After(10 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	
	ctx := context.Background()
	err := checker.CheckReadiness(ctx)
	
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestHealthCheckError_Error(t *testing.T) {
	err := &HealthCheckError{
		Component: "database",
		Message:   "connection failed",
	}
	
	expected := "health check failed for database: connection failed"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestLivenessHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	
	LivenessHandler(w, req)
	
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

func TestReadinessHandler_Success(t *testing.T) {
	checker := NewChecker()
	handler := ReadinessHandler(checker)
	
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
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

func TestReadinessHandler_Failure(t *testing.T) {
	checker := NewChecker()
	checker.AddCheck("test", func(ctx context.Context) error {
		return errors.New("test failure")
	})
	
	handler := ReadinessHandler(checker)
	
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
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

func TestReadinessHandler_ForceFailure(t *testing.T) {
	checker := NewChecker()
	checker.SetForceFailure(true)
	
	handler := ReadinessHandler(checker)
	
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
	
	body := w.Body.String()
	if !contains(body, "Not Ready") {
		t.Errorf("Expected body to contain 'Not Ready', got '%s'", body)
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