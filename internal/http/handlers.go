package http

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"monitoring-dashboard-automation/internal/health"
	"monitoring-dashboard-automation/internal/metrics"

	"go.uber.org/zap"
)

// HealthHandlers contains all health-related HTTP handlers
type HealthHandlers struct {
	checker *health.Checker
}

// NewHealthHandlers creates new health handlers
func NewHealthHandlers(checker *health.Checker) *HealthHandlers {
	return &HealthHandlers{
		checker: checker,
	}
}

// Liveness handles GET /healthz - always returns 200 OK
func (h *HealthHandlers) Liveness(w http.ResponseWriter, r *http.Request) {
	health.LivenessHandler(w, r)
}

// Readiness handles GET /readyz - checks dependencies
func (h *HealthHandlers) Readiness(w http.ResponseWriter, r *http.Request) {
	handler := health.ReadinessHandler(h.checker)
	handler(w, r)
}

// ToggleReadiness handles POST /api/v1/toggles/readiness - for testing
func (h *HealthHandlers) ToggleReadiness(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ForceFailure bool `json:"force_failure"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	h.checker.SetForceFailure(req.ForceFailure)

	response := map[string]interface{}{
		"force_failure": req.ForceFailure,
		"message":      "Readiness check toggle updated",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// APIHandlers contains all API-related HTTP handlers
type APIHandlers struct {
	logger  *zap.Logger
	metrics *metrics.Registry
}

// NewAPIHandlers creates new API handlers
func NewAPIHandlers(logger *zap.Logger, metrics *metrics.Registry) *APIHandlers {
	return &APIHandlers{
		logger:  logger,
		metrics: metrics,
	}
}

// Ping handles GET /api/v1/ping - simple ping endpoint
func (h *APIHandlers) Ping(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "pong",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Work handles GET /api/v1/work - simulates work with configurable duration and jitter
func (h *APIHandlers) Work(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	msParam := r.URL.Query().Get("ms")
	jitterParam := r.URL.Query().Get("jitter")

	// Default values
	baseDuration := 100 * time.Millisecond
	jitterDuration := time.Duration(0)

	// Parse ms parameter
	if msParam != "" {
		if ms, err := strconv.Atoi(msParam); err == nil && ms >= 0 {
			baseDuration = time.Duration(ms) * time.Millisecond
		}
	}

	// Parse jitter parameter
	if jitterParam != "" {
		if jitter, err := strconv.Atoi(jitterParam); err == nil && jitter >= 0 {
			jitterDuration = time.Duration(jitter) * time.Millisecond
		}
	}

	// Calculate total duration with jitter
	totalDuration := baseDuration
	if jitterDuration > 0 {
		// Add random jitter between 0 and jitterDuration
		jitter := time.Duration(rand.Int63n(int64(jitterDuration)))
		totalDuration += jitter
	}

	// Increment inflight jobs metric
	h.metrics.IncWorkJobsInflight()
	defer h.metrics.DecWorkJobsInflight()

	// Simulate work with context cancellation support
	startTime := time.Now()
	if err := h.simulateWork(r.Context(), totalDuration); err != nil {
		// Work was cancelled or failed
		h.metrics.IncWorkFailures("simulate_work")
		h.logger.Warn("Work simulation failed", 
			zap.Error(err),
			zap.Duration("requested_duration", totalDuration),
			zap.Duration("actual_duration", time.Since(startTime)))
		
		http.Error(w, "Work simulation cancelled", http.StatusRequestTimeout)
		return
	}

	actualDuration := time.Since(startTime)

	response := map[string]interface{}{
		"message":           "work completed",
		"requested_ms":      int(baseDuration.Milliseconds()),
		"jitter_ms":         int(jitterDuration.Milliseconds()),
		"actual_duration_ms": int(actualDuration.Milliseconds()),
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// simulateWork simulates work for the given duration, respecting context cancellation
func (h *APIHandlers) simulateWork(ctx context.Context, duration time.Duration) error {
	select {
	case <-time.After(duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ToggleHandlers contains all toggle-related HTTP handlers
type ToggleHandlers struct {
	logger      *zap.Logger
	errorToggle interface {
		SetConfig(enabled bool, rate float64, statusCode int)
		GetConfig() (bool, float64, int)
	}
}

// NewToggleHandlers creates new toggle handlers
func NewToggleHandlers(logger *zap.Logger, errorToggle interface {
	SetConfig(enabled bool, rate float64, statusCode int)
	GetConfig() (bool, float64, int)
}) *ToggleHandlers {
	return &ToggleHandlers{
		logger:      logger,
		errorToggle: errorToggle,
	}
}

// ErrorRate handles POST /api/v1/toggles/error-rate - configures error injection
func (h *ToggleHandlers) ErrorRate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled    bool    `json:"enabled"`
		Rate       float64 `json:"rate"`
		StatusCode int     `json:"status_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode error rate toggle request", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate rate is between 0.0 and 1.0
	if req.Rate < 0.0 || req.Rate > 1.0 {
		http.Error(w, "Rate must be between 0.0 and 1.0", http.StatusBadRequest)
		return
	}

	// Validate status code is a valid 5xx error code
	if req.StatusCode < 500 || req.StatusCode > 599 {
		http.Error(w, "Status code must be between 500 and 599", http.StatusBadRequest)
		return
	}

	// Update the error toggle configuration
	h.errorToggle.SetConfig(req.Enabled, req.Rate, req.StatusCode)

	h.logger.Info("Error injection toggle updated",
		zap.Bool("enabled", req.Enabled),
		zap.Float64("rate", req.Rate),
		zap.Int("status_code", req.StatusCode),
	)

	response := map[string]interface{}{
		"enabled":     req.Enabled,
		"rate":        req.Rate,
		"status_code": req.StatusCode,
		"message":     "Error injection toggle updated",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}