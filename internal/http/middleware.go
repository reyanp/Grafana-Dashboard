package http

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"monitoring-dashboard-automation/internal/metrics"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// RequestIDKey is the context key for request ID
type contextKey string

const RequestIDKey contextKey = "requestID"

// RequestIDMiddleware generates and adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs HTTP requests with structured logging
func LoggingMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create a response writer wrapper to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			
			// Get request ID from context
			requestID, _ := r.Context().Value(RequestIDKey).(string)
			
			// Log request start
			logger.Info("Request started",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.String("request_id", requestID),
			)
			
			defer func() {
				// Log request completion
				logger.Info("Request completed",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", ww.Status()),
					zap.Int("bytes", ww.BytesWritten()),
					zap.Duration("duration", time.Since(start)),
					zap.String("request_id", requestID),
				)
			}()
			
			next.ServeHTTP(ww, r)
		})
	}
}

// PanicRecoveryMiddleware recovers from panics and logs stack traces
func PanicRecoveryMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get request ID from context
					requestID, _ := r.Context().Value(RequestIDKey).(string)
					
					// Log the panic with stack trace
					logger.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("request_id", requestID),
						zap.String("stack", string(debug.Stack())),
					)
					
					// Return 500 Internal Server Error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// PrometheusMiddleware instruments HTTP requests with Prometheus metrics
func PrometheusMiddleware(metricsRegistry *metrics.Registry) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create a response writer wrapper to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			
			// Process the request
			next.ServeHTTP(ww, r)
			
			// Record metrics after request completion
			duration := time.Since(start)
			
			// Get the route pattern from chi router context
			route := getRoutePattern(r)
			
			// Record the HTTP request metrics
			metricsRegistry.RecordHTTPRequest(r.Method, route, ww.Status(), duration)
		})
	}
}

// BearerTokenAuthMiddleware validates bearer token for admin routes
func BearerTokenAuthMiddleware(adminToken string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}
			
			// Check if it starts with "Bearer "
			const bearerPrefix = "Bearer "
			if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
				http.Error(w, "Invalid authorization format. Expected 'Bearer <token>'", http.StatusUnauthorized)
				return
			}
			
			// Extract token
			token := authHeader[len(bearerPrefix):]
			if token != adminToken {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			
			// Token is valid, proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// ErrorInjectionMiddleware injects errors based on toggle configuration
func ErrorInjectionMiddleware(errorToggle interface{}) func(next http.Handler) http.Handler {
	// Type assertion to get the actual ErrorToggle
	toggle, ok := errorToggle.(interface {
		ShouldInjectError() (bool, int)
	})
	if !ok {
		// If type assertion fails, return a no-op middleware
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if we should inject an error
			if shouldInject, statusCode := toggle.ShouldInjectError(); shouldInject {
				http.Error(w, "Injected error for testing", statusCode)
				return
			}
			
			// No error injection, proceed normally
			next.ServeHTTP(w, r)
		})
	}
}

// getRoutePattern extracts the route pattern from chi router context
func getRoutePattern(r *http.Request) string {
	// Try to get the route pattern from chi context
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if rctx.RoutePattern() != "" {
			return rctx.RoutePattern()
		}
	}
	
	// Fallback to the request path if no pattern is found
	return r.URL.Path
}