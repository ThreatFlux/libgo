package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wroersma/libgo/internal/config"
)

// TestLogin tests the login API endpoint
func TestLogin(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get API base URL from environment
	baseURL, err := getBaseURL()
	require.NoError(t, err)

	// Test login
	t.Run("Login success", func(t *testing.T) {
		// Create login request
		reqBody, err := json.Marshal(map[string]string{
			"username": "admin",
			"password": "password", // Use proper credentials from test setup
		})
		require.NoError(t, err)

		// Send request
		resp, err := http.Post(
			fmt.Sprintf("%s/login", baseURL),
			"application/json",
			bytes.NewBuffer(reqBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var loginResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&loginResp)
		require.NoError(t, err)

		// Verify token in response
		token, ok := loginResp["token"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, token)
	})

	t.Run("Login failure", func(t *testing.T) {
		// Create login request with invalid credentials
		reqBody, err := json.Marshal(map[string]string{
			"username": "admin",
			"password": "wrong-password",
		})
		require.NoError(t, err)

		// Send request
		resp, err := http.Post(
			fmt.Sprintf("%s/login", baseURL),
			"application/json",
			bytes.NewBuffer(reqBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestHealth tests the health API endpoint
func TestHealth(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get API base URL from environment
	baseURL, err := getBaseURL()
	require.NoError(t, err)

	// Test health endpoint
	t.Run("Health check", func(t *testing.T) {
		// Send request
		resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var healthResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&healthResp)
		require.NoError(t, err)

		// Verify status in response
		status, ok := healthResp["status"].(string)
		assert.True(t, ok)
		assert.Equal(t, "UP", status)
	})
}

// getBaseURL returns the base URL for the API from environment
func getBaseURL() (string, error) {
	// Try to get base URL from environment variable
	baseURL := os.Getenv("LIBGO_API_URL")
	if baseURL != "" {
		return baseURL, nil
	}

	// Load configuration
	cfg := &config.Config{}
	loader := config.NewYAMLLoader("../../configs/config.yaml")
	if err := loader.Load(cfg); err != nil {
		return "", fmt.Errorf("loading config: %w", err)
	}

	// Use configuration to construct URL
	return fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port), nil
}

// setupAPI starts the API server for testing
func setupAPI() error {
	// Check if API is already running
	baseURL, err := getBaseURL()
	if err != nil {
		return err
	}

	// Try to connect to API
	client := http.Client{
		Timeout: 1 * time.Second,
	}
	_, err = client.Get(fmt.Sprintf("%s/health", baseURL))
	if err == nil {
		// API is already running
		return nil
	}

	// API is not running, start it
	// In a real integration test, you would start the API here
	// For this example, we assume the API is already running
	return fmt.Errorf("API is not running at %s - please start the server before running integration tests", baseURL)
}
