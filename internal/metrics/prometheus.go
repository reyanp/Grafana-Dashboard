package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry wraps prometheus registry and provides metrics
type Registry struct {
	registry *prometheus.Registry
	
	// HTTP metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	
	// Work metrics (for future tasks)
	workJobsInflight     prometheus.Gauge
	workFailuresTotal    *prometheus.CounterVec
}

// NewRegistry creates a new metrics registry
func NewRegistry() *Registry {
	registry := prometheus.NewRegistry()
	
	// Register default Go metrics
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	
	// Create HTTP metrics
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)
	
	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
	
	// Create work metrics (for future tasks)
	workJobsInflight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "work_jobs_inflight",
			Help: "Number of work jobs currently in progress",
		},
	)
	
	workFailuresTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "work_failures_total",
			Help: "Total number of work operation failures",
		},
		[]string{"operation"},
	)
	
	// Register HTTP metrics
	registry.MustRegister(httpRequestsTotal)
	registry.MustRegister(httpRequestDuration)
	
	// Register work metrics
	registry.MustRegister(workJobsInflight)
	registry.MustRegister(workFailuresTotal)
	
	return &Registry{
		registry:            registry,
		httpRequestsTotal:   httpRequestsTotal,
		httpRequestDuration: httpRequestDuration,
		workJobsInflight:    workJobsInflight,
		workFailuresTotal:   workFailuresTotal,
	}
}

// GetRegistry returns the underlying prometheus registry
func (r *Registry) GetRegistry() *prometheus.Registry {
	return r.registry
}

// GetHandler returns the Prometheus HTTP handler
func (r *Registry) GetHandler() http.Handler {
	return promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{})
}

// RecordHTTPRequest records metrics for an HTTP request
func (r *Registry) RecordHTTPRequest(method, route string, statusCode int, duration time.Duration) {
	status := strconv.Itoa(statusCode)
	
	r.httpRequestsTotal.WithLabelValues(method, route, status).Inc()
	r.httpRequestDuration.WithLabelValues(method, route).Observe(duration.Seconds())
}

// IncWorkJobsInflight increments the work jobs inflight gauge
func (r *Registry) IncWorkJobsInflight() {
	r.workJobsInflight.Inc()
}

// DecWorkJobsInflight decrements the work jobs inflight gauge
func (r *Registry) DecWorkJobsInflight() {
	r.workJobsInflight.Dec()
}

// IncWorkFailures increments the work failures counter
func (r *Registry) IncWorkFailures(operation string) {
	r.workFailuresTotal.WithLabelValues(operation).Inc()
}