package main

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"monitoring-dashboard-automation/internal/config"
	httphandler "monitoring-dashboard-automation/internal/http"
	"monitoring-dashboard-automation/internal/metrics"

	"go.uber.org/zap/zaptest"
)

func TestGracefulShutdown(t *testing.T) {
	tests := []struct {
		name           string
		inflightJobs   int
		shutdownTimeout time.Duration
		expectError    bool
	}{
		{
			name:           "shutdown with no inflight jobs",
			inflightJobs:   0,
			shutdownTimeout: 5 * time.Second,
			expectError:    false,
		},
		{
			name:           "shutdown with inflight jobs that complete quickly",
			inflightJobs:   2,
			shutdownTimeout: 5 * time.Second,
			expectError:    false,
		},
		{
			name:           "shutdown timeout with long running jobs",
			inflightJobs:   1,
			shutdownTimeout: 100 * time.Millisecond,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test logger
			logger := zaptest.NewLogger(t)
			
			// Create metrics registry
			metricsRegistry := metrics.NewRegistry()
			
			// Create test config
			cfg := &config.Config{
				Port:       "0", // Use random port
				AdminToken: "test-token",
				LogLevel:   "debug",
			}
			
			// Create router and server
			router := httphandler.NewRouter(cfg, logger, metricsRegistry)
			server := httptest.NewServer(router)
			defer server.Close()
			
			// Simulate inflight jobs
			for i := 0; i < tt.inflightJobs; i++ {
				metricsRegistry.IncWorkJobsInflight()
			}
			
			// Start goroutines to simulate job completion
			var wg sync.WaitGroup
			for i := 0; i < tt.inflightJobs; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					// Simulate work duration based on test case
					if tt.expectError {
						time.Sleep(200 * time.Millisecond) // Longer than timeout
					} else {
						time.Sleep(50 * time.Millisecond) // Shorter than timeout
					}
					metricsRegistry.DecWorkJobsInflight()
				}()
			}
			
			// Create shutdown context
			ctx, cancel := context.WithTimeout(context.Background(), tt.shutdownTimeout)
			defer cancel()
			
			// Test graceful shutdown
			err := gracefulShutdown(ctx, server.Config, metricsRegistry, logger)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			// Wait for all goroutines to complete
			wg.Wait()
		})
	}
}

func TestGracefulShutdownWithRealServer(t *testing.T) {
	// Create test logger
	logger := zaptest.NewLogger(t)
	
	// Create metrics registry
	metricsRegistry := metrics.NewRegistry()
	
	// Create test config
	cfg := &config.Config{
		Port:       "0", // Use random port
		AdminToken: "test-token",
		LogLevel:   "debug",
	}
	
	// Create router
	router := httphandler.NewRouter(cfg, logger, metricsRegistry)
	
	// Create HTTP server
	server := &http.Server{
		Addr:    ":0",
		Handler: router,
	}
	
	// Start server in a goroutine
	serverStarted := make(chan struct{})
	serverError := make(chan error, 1)
	
	go func() {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			serverError <- err
			return
		}
		
		server.Addr = listener.Addr().String()
		close(serverStarted)
		
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverError <- err
		}
	}()
	
	// Wait for server to start or error
	select {
	case <-serverStarted:
		// Server started successfully
	case err := <-serverError:
		t.Fatalf("Server failed to start: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("Server took too long to start")
	}
	
	// Make a request to simulate inflight work
	go func() {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get("http://" + server.Addr + "/api/v1/work?ms=200")
		if err == nil {
			resp.Body.Close()
		}
	}()
	
	// Give the request time to start
	time.Sleep(50 * time.Millisecond)
	
	// Test graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := gracefulShutdown(ctx, server, metricsRegistry, logger)
	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}
}

func TestMetricsFlush(t *testing.T) {
	// Create metrics registry
	metricsRegistry := metrics.NewRegistry()
	
	// Record some metrics
	metricsRegistry.RecordHTTPRequest("GET", "/test", 200, 100*time.Millisecond)
	metricsRegistry.IncWorkJobsInflight()
	metricsRegistry.DecWorkJobsInflight()
	metricsRegistry.IncWorkFailures("test_operation")
	
	// Test flush
	err := metricsRegistry.Flush()
	if err != nil {
		t.Errorf("Metrics flush failed: %v", err)
	}
}

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		wantErr  bool
	}{
		{
			name:    "debug level",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "production level",
			level:   "production",
			wantErr: false,
		},
		{
			name:    "default level",
			level:   "invalid",
			wantErr: false, // Should default to development
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := initLogger(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("initLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if logger == nil && !tt.wantErr {
				t.Error("initLogger() returned nil logger")
			}
			if logger != nil {
				logger.Sync()
			}
		})
	}
}