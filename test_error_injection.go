package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	baseURL := "http://localhost:8080"
	adminToken := "changeme"

	fmt.Println("Testing error injection system...")

	// Test 1: Try to access admin endpoint without token (should fail)
	fmt.Println("\n1. Testing access without token...")
	if err := testWithoutToken(baseURL); err != nil {
		fmt.Printf("✓ Correctly rejected request without token: %v\n", err)
	} else {
		fmt.Println("✗ Should have rejected request without token")
		os.Exit(1)
	}

	// Test 2: Try to access admin endpoint with invalid token (should fail)
	fmt.Println("\n2. Testing access with invalid token...")
	if err := testWithInvalidToken(baseURL); err != nil {
		fmt.Printf("✓ Correctly rejected request with invalid token: %v\n", err)
	} else {
		fmt.Println("✗ Should have rejected request with invalid token")
		os.Exit(1)
	}

	// Test 3: Configure error injection with valid token (should succeed)
	fmt.Println("\n3. Testing error injection configuration...")
	if err := configureErrorInjection(baseURL, adminToken, true, 1.0, 503); err != nil {
		fmt.Printf("✗ Failed to configure error injection: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Successfully configured error injection")

	// Test 4: Test that API endpoints now return errors
	fmt.Println("\n4. Testing that API endpoints return injected errors...")
	errorCount := 0
	for i := 0; i < 10; i++ {
		resp, err := http.Get(baseURL + "/api/v1/ping")
		if err != nil {
			fmt.Printf("✗ Request failed: %v\n", err)
			continue
		}
		resp.Body.Close()
		
		if resp.StatusCode == 503 {
			errorCount++
		}
	}
	
	if errorCount >= 8 { // With rate 1.0, we should get mostly errors
		fmt.Printf("✓ Error injection working: %d/10 requests returned 503\n", errorCount)
	} else {
		fmt.Printf("✗ Error injection not working properly: only %d/10 requests returned 503\n", errorCount)
		os.Exit(1)
	}

	// Test 5: Disable error injection
	fmt.Println("\n5. Testing error injection disable...")
	if err := configureErrorInjection(baseURL, adminToken, false, 0.0, 500); err != nil {
		fmt.Printf("✗ Failed to disable error injection: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Successfully disabled error injection")

	// Test 6: Test that API endpoints now work normally
	fmt.Println("\n6. Testing that API endpoints work normally...")
	successCount := 0
	for i := 0; i < 10; i++ {
		resp, err := http.Get(baseURL + "/api/v1/ping")
		if err != nil {
			fmt.Printf("✗ Request failed: %v\n", err)
			continue
		}
		resp.Body.Close()
		
		if resp.StatusCode == 200 {
			successCount++
		}
	}
	
	if successCount >= 9 { // Should get mostly success now
		fmt.Printf("✓ Error injection disabled: %d/10 requests returned 200\n", successCount)
	} else {
		fmt.Printf("✗ Error injection not properly disabled: only %d/10 requests returned 200\n", successCount)
		os.Exit(1)
	}

	fmt.Println("\n✓ All tests passed! Error injection system is working correctly.")
}

func testWithoutToken(baseURL string) error {
	reqBody := `{"enabled": true, "rate": 0.5, "status_code": 503}`
	resp, err := http.Post(baseURL+"/api/v1/toggles/error-rate", "application/json", bytes.NewReader([]byte(reqBody)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 401 {
		return fmt.Errorf("unauthorized")
	}
	return nil
}

func testWithInvalidToken(baseURL string) error {
	reqBody := `{"enabled": true, "rate": 0.5, "status_code": 503}`
	req, err := http.NewRequest("POST", baseURL+"/api/v1/toggles/error-rate", bytes.NewReader([]byte(reqBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid-token")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 401 {
		return fmt.Errorf("unauthorized")
	}
	return nil
}

func configureErrorInjection(baseURL, token string, enabled bool, rate float64, statusCode int) error {
	config := map[string]interface{}{
		"enabled":     enabled,
		"rate":        rate,
		"status_code": statusCode,
	}
	
	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("POST", baseURL+"/api/v1/toggles/error-rate", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	
	// Give the system a moment to apply the configuration
	time.Sleep(100 * time.Millisecond)
	return nil
}