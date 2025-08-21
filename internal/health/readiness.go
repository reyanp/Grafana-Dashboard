package health

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// CheckFunc represents a health check function
type CheckFunc func(ctx context.Context) error

// Checker manages health checks for the application
type Checker struct {
	checks map[string]CheckFunc
	mu     sync.RWMutex
	
	// Toggle for testing - allows forcing readiness to fail
	forceFailure bool
	failureMu    sync.RWMutex
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
	return &Checker{
		checks: make(map[string]CheckFunc),
	}
}

// AddCheck adds a named health check
func (c *Checker) AddCheck(name string, check CheckFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// RemoveCheck removes a named health check
func (c *Checker) RemoveCheck(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.checks, name)
}

// SetForceFailure allows toggling readiness check failure for testing
func (c *Checker) SetForceFailure(fail bool) {
	c.failureMu.Lock()
	defer c.failureMu.Unlock()
	c.forceFailure = fail
}

// IsForceFailure returns whether force failure is enabled
func (c *Checker) IsForceFailure() bool {
	c.failureMu.RLock()
	defer c.failureMu.RUnlock()
	return c.forceFailure
}

// CheckReadiness runs all registered health checks
func (c *Checker) CheckReadiness(ctx context.Context) error {
	// Check if force failure is enabled for testing
	if c.IsForceFailure() {
		return &HealthCheckError{
			Component: "forced",
			Message:   "readiness check forced to fail for testing",
		}
	}

	c.mu.RLock()
	checks := make(map[string]CheckFunc, len(c.checks))
	for name, check := range c.checks {
		checks[name] = check
	}
	c.mu.RUnlock()

	// Run all checks with a timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for name, check := range checks {
		if err := check(ctx); err != nil {
			return &HealthCheckError{
				Component: name,
				Message:   err.Error(),
			}
		}
	}

	return nil
}

// HealthCheckError represents a health check failure
type HealthCheckError struct {
	Component string
	Message   string
}

func (e *HealthCheckError) Error() string {
	return "health check failed for " + e.Component + ": " + e.Message
}

// LivenessHandler always returns 200 OK - used for liveness probes
func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// ReadinessHandler checks readiness and returns appropriate status
func ReadinessHandler(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		if err := checker.CheckReadiness(ctx); err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready: " + err.Error()))
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	}
}