package http

import (
	"monitoring-dashboard-automation/internal/config"
	"monitoring-dashboard-automation/internal/health"
	"monitoring-dashboard-automation/internal/metrics"
	"monitoring-dashboard-automation/internal/toggles"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// NewRouter creates and configures the HTTP router
func NewRouter(cfg *config.Config, logger *zap.Logger, metricsRegistry *metrics.Registry) *chi.Mux {
	r := chi.NewRouter()

	// Create error toggle for error injection
	errorToggle := toggles.NewErrorToggle()

	// Apply middleware stack in order
	r.Use(middleware.RequestID)           // Chi's built-in request ID middleware
	r.Use(RequestIDMiddleware)            // Our custom request ID middleware
	r.Use(PanicRecoveryMiddleware(logger)) // Panic recovery with logging
	r.Use(LoggingMiddleware(logger))      // Structured logging
	r.Use(PrometheusMiddleware(metricsRegistry)) // Prometheus instrumentation
	r.Use(middleware.Timeout(60))         // Request timeout

	// Create health checker and handlers
	healthChecker := health.NewChecker()
	healthHandlers := NewHealthHandlers(healthChecker)
	
	// Create API handlers
	apiHandlers := NewAPIHandlers(logger, metricsRegistry)
	
	// Create toggle handlers
	toggleHandlers := NewToggleHandlers(logger, errorToggle)

	// Health check routes (no error injection)
	r.Get("/healthz", healthHandlers.Liveness)
	r.Get("/readyz", healthHandlers.Readiness)

	// Metrics endpoint (no error injection)
	r.Handle("/metrics", metricsRegistry.GetHandler())

	// API routes with error injection middleware
	r.Route("/api/v1", func(r chi.Router) {
		// Apply error injection middleware to API routes
		r.Use(ErrorInjectionMiddleware(errorToggle))
		
		r.Get("/ping", apiHandlers.Ping)
		r.Get("/work", apiHandlers.Work)

		// Admin routes with bearer token authentication
		r.Route("/toggles", func(r chi.Router) {
			// Apply bearer token authentication to admin routes
			r.Use(BearerTokenAuthMiddleware(cfg.AdminToken))
			
			r.Post("/error-rate", toggleHandlers.ErrorRate)
			r.Post("/readiness", healthHandlers.ToggleReadiness)
		})
	})

	return r
}