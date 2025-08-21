package toggles

import (
	"math/rand"
	"sync"
)

// ErrorToggle represents the configuration for error injection
type ErrorToggle struct {
	mu         sync.RWMutex
	Enabled    bool    `json:"enabled"`
	Rate       float64 `json:"rate"`        // 0.0 to 1.0
	StatusCode int     `json:"status_code"` // HTTP status code to return
}

// NewErrorToggle creates a new ErrorToggle with default values
func NewErrorToggle() *ErrorToggle {
	return &ErrorToggle{
		Enabled:    false,
		Rate:       0.0,
		StatusCode: 500,
	}
}

// SetConfig updates the error toggle configuration
func (et *ErrorToggle) SetConfig(enabled bool, rate float64, statusCode int) {
	et.mu.Lock()
	defer et.mu.Unlock()
	
	et.Enabled = enabled
	et.Rate = rate
	et.StatusCode = statusCode
}

// GetConfig returns the current error toggle configuration
func (et *ErrorToggle) GetConfig() (bool, float64, int) {
	et.mu.RLock()
	defer et.mu.RUnlock()
	
	return et.Enabled, et.Rate, et.StatusCode
}

// ShouldInjectError determines if an error should be injected based on the current configuration
func (et *ErrorToggle) ShouldInjectError() (bool, int) {
	et.mu.RLock()
	defer et.mu.RUnlock()
	
	if !et.Enabled {
		return false, 0
	}
	
	// Generate random number between 0.0 and 1.0
	if rand.Float64() < et.Rate {
		return true, et.StatusCode
	}
	
	return false, 0
}