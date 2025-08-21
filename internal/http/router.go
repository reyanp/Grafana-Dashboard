package http

import (
	"net/http"

	"monitoring-dashboard-automation/internal/config"
	"monitoring-dashboard-automation/internal/metrics"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// NewRouter creates and configures the HTTP router
func NewRouter(cfg *config.Config, logger *zap.Logger, metricsRegistry *metrics.Registry) *chi.Mux {
	r := chi.NewRouter()

	// Apply middleware stack in order
	r.Use(middleware.RequestID)           // Chi's built-in request ID middleware
	r.Use(RequestIDMiddleware)            // Our custom request ID middleware
	r.Use(PanicRecoveryMiddleware(logger)) // Panic recovery with logging
	r.Use(LoggingMiddleware(logger))      // Structured logging
	r.Use(middleware.Timeout(60))         // Request timeout

	// Health check routes
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for readiness check - will be implemented in later task
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	// Metrics endpoint
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for metrics endpoint - will be implemented in later task
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics endpoint placeholder"))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for ping endpoint - will be implemented in later task
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("pong"))
		})

		r.Get("/work", func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for work endpoint - will be implemented in later task
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("work simulation placeholder"))
		})

		// Admin routes (will need authentication middleware in later task)
		r.Route("/toggles", func(r chi.Router) {
			r.Post("/error-rate", func(w http.ResponseWriter, r *http.Request) {
				// Placeholder for error rate toggle - will be implemented in later task
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("error rate toggle placeholder"))
			})
		})
	})

	return r
}