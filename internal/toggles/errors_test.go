package toggles

import (
	"testing"
)

func TestNewErrorToggle(t *testing.T) {
	toggle := NewErrorToggle()
	
	if toggle == nil {
		t.Fatal("NewErrorToggle() returned nil")
	}
	
	enabled, rate, statusCode := toggle.GetConfig()
	if enabled {
		t.Errorf("Expected enabled to be false, got %v", enabled)
	}
	if rate != 0.0 {
		t.Errorf("Expected rate to be 0.0, got %v", rate)
	}
	if statusCode != 500 {
		t.Errorf("Expected status code to be 500, got %v", statusCode)
	}
}

func TestErrorToggle_SetConfig(t *testing.T) {
	toggle := NewErrorToggle()
	
	// Test setting configuration
	toggle.SetConfig(true, 0.5, 503)
	
	enabled, rate, statusCode := toggle.GetConfig()
	if !enabled {
		t.Errorf("Expected enabled to be true, got %v", enabled)
	}
	if rate != 0.5 {
		t.Errorf("Expected rate to be 0.5, got %v", rate)
	}
	if statusCode != 503 {
		t.Errorf("Expected status code to be 503, got %v", statusCode)
	}
}

func TestErrorToggle_ShouldInjectError_Disabled(t *testing.T) {
	toggle := NewErrorToggle()
	
	// When disabled, should never inject errors
	for i := 0; i < 100; i++ {
		shouldInject, statusCode := toggle.ShouldInjectError()
		if shouldInject {
			t.Errorf("Expected no error injection when disabled, but got shouldInject=true")
		}
		if statusCode != 0 {
			t.Errorf("Expected status code to be 0 when disabled, got %v", statusCode)
		}
	}
}

func TestErrorToggle_ShouldInjectError_EnabledZeroRate(t *testing.T) {
	toggle := NewErrorToggle()
	toggle.SetConfig(true, 0.0, 500)
	
	// With rate 0.0, should never inject errors
	for i := 0; i < 100; i++ {
		shouldInject, statusCode := toggle.ShouldInjectError()
		if shouldInject {
			t.Errorf("Expected no error injection with rate 0.0, but got shouldInject=true")
		}
		if statusCode != 0 {
			t.Errorf("Expected status code to be 0 with rate 0.0, got %v", statusCode)
		}
	}
}

func TestErrorToggle_ShouldInjectError_EnabledFullRate(t *testing.T) {
	toggle := NewErrorToggle()
	toggle.SetConfig(true, 1.0, 502)
	
	// With rate 1.0, should always inject errors
	for i := 0; i < 100; i++ {
		shouldInject, statusCode := toggle.ShouldInjectError()
		if !shouldInject {
			t.Errorf("Expected error injection with rate 1.0, but got shouldInject=false")
		}
		if statusCode != 502 {
			t.Errorf("Expected status code to be 502, got %v", statusCode)
		}
	}
}

func TestErrorToggle_ShouldInjectError_EnabledPartialRate(t *testing.T) {
	toggle := NewErrorToggle()
	toggle.SetConfig(true, 0.5, 503)
	
	// With rate 0.5, should inject errors approximately 50% of the time
	// We'll run many iterations and check that we get some errors but not all
	injectedCount := 0
	totalCount := 1000
	
	for i := 0; i < totalCount; i++ {
		shouldInject, statusCode := toggle.ShouldInjectError()
		if shouldInject {
			injectedCount++
			if statusCode != 503 {
				t.Errorf("Expected status code to be 503, got %v", statusCode)
			}
		}
	}
	
	// With rate 0.5, we expect roughly 500 injections out of 1000
	// Allow for some variance due to randomness (between 30% and 70%)
	expectedMin := int(float64(totalCount) * 0.3)
	expectedMax := int(float64(totalCount) * 0.7)
	
	if injectedCount < expectedMin || injectedCount > expectedMax {
		t.Errorf("Expected error injection count to be between %d and %d, got %d", 
			expectedMin, expectedMax, injectedCount)
	}
}

func TestErrorToggle_ConcurrentAccess(t *testing.T) {
	toggle := NewErrorToggle()
	
	// Test concurrent access to ensure thread safety
	done := make(chan bool, 2)
	
	// Goroutine 1: continuously set config
	go func() {
		for i := 0; i < 100; i++ {
			toggle.SetConfig(true, 0.5, 500+i%100)
		}
		done <- true
	}()
	
	// Goroutine 2: continuously check if should inject error
	go func() {
		for i := 0; i < 100; i++ {
			toggle.ShouldInjectError()
		}
		done <- true
	}()
	
	// Wait for both goroutines to complete
	<-done
	<-done
	
	// If we get here without panicking, the concurrent access test passed
}