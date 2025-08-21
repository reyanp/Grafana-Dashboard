package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Registry wraps prometheus registry and provides metrics
type Registry struct {
	registry *prometheus.Registry
}

// NewRegistry creates a new metrics registry
func NewRegistry() *Registry {
	registry := prometheus.NewRegistry()
	
	// Register default Go metrics
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	
	return &Registry{
		registry: registry,
	}
}

// GetRegistry returns the underlying prometheus registry
func (r *Registry) GetRegistry() *prometheus.Registry {
	return r.registry
}