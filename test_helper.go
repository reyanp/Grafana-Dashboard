package main

import (
	"os"
	"testing"
)

// TestMain is the entry point for all tests in this package
func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("LOG_LEVEL", "error") // Reduce log noise during tests
	
	// Run tests
	code := m.Run()
	
	// Clean up
	os.Exit(code)
}