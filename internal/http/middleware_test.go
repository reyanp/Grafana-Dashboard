package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"monitoring-dashboard-automation/internal/metrics"

	"github.com/go-chi/chi/v5"
)

func TestPrometheusMiddleware(t *testing.T) {
	// Create a metrics registry
	metricsRegistry := metrics.NewRegistry()
	
	// Create a test router with the middleware
	r := chi.NewRouter()
	r.Use(PrometheusMiddleware(metricsRegistry))
	
	// Add a test route
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})
	
	// Make a request to the test route
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	// Check that the request was successful
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Now check that metrics were recorded
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	metricsW := httptest.NewRecorder()
	
	metricsHandler := metricsRegistry.GetHandler()
	metricsHandler.ServeHTTP(metricsW, metricsReq)
	
	metricsBody := metricsW.Body.String()
	
	// Check that the HTTP request was recorded in metrics
	if !strings.Contains(metricsBody, "http_requests_total") {
		t.Error("Expected http_requests_total metric to be present")
	}
	
	if !strings.Contains(metricsBody, "http_request_duration_seconds") {
		t.Error("Expected http_request_duration_seconds metric to be present")
	}
	
	// Check that our specific request was recorded
	if !strings.Contains(metricsBody, `method="GET"`) {
		t.Error("Expected GET method to be recorded in metrics")
	}
	
	if !strings.Contains(metricsBody, `status="200"`) {
		t.Error("Expected 200 status to be recorded in metrics")
	}
}

func TestGetRoutePattern(t *testing.T) {
	// Test with chi router context
	r := chi.NewRouter()
	r.Get("/api/v1/test/{id}", func(w http.ResponseWriter, r *http.Request) {
		pattern := getRoutePattern(r)
		if pattern != "/api/v1/test/{id}" {
			t.Errorf("Expected route pattern '/api/v1/test/{id}', got '%s'", pattern)
		}
		w.WriteHeader(http.StatusOK)
	})
	
	req := httptest.NewRequest("GET", "/api/v1/test/123", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	// Test without chi router context (fallback to path)
	plainReq := httptest.NewRequest("GET", "/plain/path", nil)
	pattern := getRoutePattern(plainReq)
	if pattern != "/plain/path" {
		t.Errorf("Expected route pattern '/plain/path', got '%s'", pattern)
	}
}

func TestBearerTokenAuthMiddleware_ValidToken(t *testing.T) {
	adminToken := "test-admin-token"
	
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authorized"))
	})

	// Wrap with bearer token auth middleware
	middleware := BearerTokenAuthMiddleware(adminToken)
	wrappedHandler := middleware(handler)

	// Create test request with valid token
	req := httptest.NewRequest("POST", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "authorized" {
		t.Errorf("Expected 'authorized', got %s", w.Body.String())
	}
}

func TestBearerTokenAuthMiddleware_InvalidToken(t *testing.T) {
	adminToken := "test-admin-token"
	
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authorized"))
	})

	// Wrap with bearer token auth middleware
	middleware := BearerTokenAuthMiddleware(adminToken)
	wrappedHandler := middleware(handler)

	// Create test request with invalid token
	req := httptest.NewRequest("POST", "/admin", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestBearerTokenAuthMiddleware_MissingHeader(t *testing.T) {
	adminToken := "test-admin-token"
	
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authorized"))
	})

	// Wrap with bearer token auth middleware
	middleware := BearerTokenAuthMiddleware(adminToken)
	wrappedHandler := middleware(handler)

	// Create test request without Authorization header
	req := httptest.NewRequest("POST", "/admin", nil)
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestBearerTokenAuthMiddleware_InvalidFormat(t *testing.T) {
	adminToken := "test-admin-token"
	
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authorized"))
	})

	// Wrap with bearer token auth middleware
	middleware := BearerTokenAuthMiddleware(adminToken)
	wrappedHandler := middleware(handler)

	// Create test request with invalid format (missing "Bearer ")
	req := httptest.NewRequest("POST", "/admin", nil)
	req.Header.Set("Authorization", adminToken)
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// Mock error toggle for testing
type mockErrorToggle struct {
	shouldInject bool
	statusCode   int
}

func (m *mockErrorToggle) ShouldInjectError() (bool, int) {
	return m.shouldInject, m.statusCode
}

func TestErrorInjectionMiddleware_NoInjection(t *testing.T) {
	// Create mock error toggle that doesn't inject errors
	toggle := &mockErrorToggle{shouldInject: false, statusCode: 0}
	
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with error injection middleware
	middleware := ErrorInjectionMiddleware(toggle)
	wrappedHandler := middleware(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check response - should pass through normally
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "success" {
		t.Errorf("Expected 'success', got %s", w.Body.String())
	}
}

func TestErrorInjectionMiddleware_WithInjection(t *testing.T) {
	// Create mock error toggle that injects 503 errors
	toggle := &mockErrorToggle{shouldInject: true, statusCode: 503}
	
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with error injection middleware
	middleware := ErrorInjectionMiddleware(toggle)
	wrappedHandler := middleware(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check response - should return injected error
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Injected error for testing") {
		t.Errorf("Expected error message, got %s", w.Body.String())
	}
}

func TestErrorInjectionMiddleware_InvalidToggle(t *testing.T) {
	// Create invalid toggle (doesn't implement the interface)
	toggle := "invalid"
	
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with error injection middleware
	middleware := ErrorInjectionMiddleware(toggle)
	wrappedHandler := middleware(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(w, req)

	// Check response - should pass through normally (no-op middleware)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "success" {
		t.Errorf("Expected 'success', got %s", w.Body.String())
	}
}