package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"monitoring-dashboard-automation/internal/config"
	httphandler "monitoring-dashboard-automation/internal/http"
	"monitoring-dashboard-automation/internal/metrics"

	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize metrics
	metricsRegistry := metrics.NewRegistry()

	// Initialize HTTP router
	router := httphandler.NewRouter(cfg, logger, metricsRegistry)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Perform graceful shutdown
	if err := gracefulShutdown(ctx, server, metricsRegistry, logger); err != nil {
		logger.Error("Graceful shutdown failed", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}

// gracefulShutdown handles the graceful shutdown process
func gracefulShutdown(ctx context.Context, server *http.Server, metricsRegistry *metrics.Registry, logger *zap.Logger) error {
	// Start shutdown process
	shutdownComplete := make(chan error, 1)
	
	go func() {
		// Wait for in-flight work jobs to complete
		logger.Info("Waiting for in-flight work jobs to complete...")
		
		// Check for in-flight jobs periodically
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				// Timeout reached, force shutdown
				shutdownComplete <- ctx.Err()
				return
			case <-ticker.C:
				inflightJobs := metricsRegistry.GetInflightJobs()
				if inflightJobs == 0 {
					logger.Info("All work jobs completed")
					break
				}
				logger.Info("Waiting for work jobs to complete", zap.Float64("inflight_jobs", inflightJobs))
			}
			
			// Break out of the for loop when no inflight jobs
			if metricsRegistry.GetInflightJobs() == 0 {
				break
			}
		}
		
		// Shutdown HTTP server
		logger.Info("Shutting down HTTP server...")
		if err := server.Shutdown(ctx); err != nil {
			shutdownComplete <- err
			return
		}
		
		// Flush metrics
		logger.Info("Flushing metrics...")
		if err := metricsRegistry.Flush(); err != nil {
			logger.Warn("Failed to flush metrics", zap.Error(err))
		}
		
		shutdownComplete <- nil
	}()
	
	// Wait for shutdown to complete or timeout
	select {
	case err := <-shutdownComplete:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func initLogger(level string) (*zap.Logger, error) {
	var config zap.Config
	
	switch level {
	case "debug":
		config = zap.NewDevelopmentConfig()
	case "production":
		config = zap.NewProductionConfig()
	default:
		config = zap.NewDevelopmentConfig()
	}

	return config.Build()
}