package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/threatflux/libgo/internal/models/vm"
)

// TestVMLifecycle tests the complete VM lifecycle through the API
func TestVMLifecycle(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get API base URL from environment
	baseURL, err := getBaseURL()
	require.NoError(t, err)

	// Get authentication token
	token, err := getAuthToken(baseURL)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Generate a unique VM name for the test
	vmName := fmt.Sprintf("test-vm-%d", time.Now().Unix())

	// Test VM lifecycle
	var vmUUID string

	// Define VM creation parameters
	vmParams := map[string]interface{}{
		"name": vmName,
		"cpu": map[string]interface{}{
			"count": 1,
		},
		"memory": map[string]interface{}{
			"sizeBytes": 512 * 1024 * 1024, // 512 MB
		},
		"disk": map[string]interface{}{
			"sizeBytes": 1 * 1024 * 1024 * 1024, // 1 GB
			"format":    "qcow2",
		},
	}

	// Step 1: Create VM
	t.Run("Create VM", func(t *testing.T) {
		// Create request
		reqBody, err := json.Marshal(vmParams)
		require.NoError(t, err)

		// Prepare request
		req, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s/api/v1/vms", baseURL),
			bytes.NewBuffer(reqBody),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Parse response
		var createResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err)

		// Verify VM info in response
		vm, ok := createResp["vm"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, vmName, vm["name"])
		assert.Equal(t, "stopped", vm["status"])

		// Save UUID for later tests
		vmUUID, ok = vm["uuid"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, vmUUID)
	})

	// Step 2: Get VM details
	t.Run("Get VM details", func(t *testing.T) {
		// Prepare request
		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s/api/v1/vms/%s", baseURL, vmName),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var getResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err)

		// Verify VM info in response
		vm, ok := getResp["vm"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, vmName, vm["name"])
		assert.Equal(t, vmUUID, vm["uuid"])
	})

	// Step 3: Start VM
	t.Run("Start VM", func(t *testing.T) {
		// Prepare request
		req, err := http.NewRequest(
			http.MethodPut,
			fmt.Sprintf("%s/api/v1/vms/%s/start", baseURL, vmName),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Wait for VM to start
		time.Sleep(5 * time.Second)

		// Verify VM status
		status, err := getVMStatus(baseURL, token, vmName)
		require.NoError(t, err)
		assert.Equal(t, string(vm.VMStatusRunning), status)
	})

	// Step 4: Stop VM
	t.Run("Stop VM", func(t *testing.T) {
		// Prepare request
		req, err := http.NewRequest(
			http.MethodPut,
			fmt.Sprintf("%s/api/v1/vms/%s/stop", baseURL, vmName),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Wait for VM to stop
		time.Sleep(5 * time.Second)

		// Verify VM status
		status, err := getVMStatus(baseURL, token, vmName)
		require.NoError(t, err)
		assert.Equal(t, string(vm.VMStatusStopped), status)
	})

	// Step 5: List VMs
	t.Run("List VMs", func(t *testing.T) {
		// Prepare request
		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s/api/v1/vms", baseURL),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var listResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&listResp)
		require.NoError(t, err)

		// Verify VMs in response
		vms, ok := listResp["vms"].([]interface{})
		assert.True(t, ok)

		// Find our VM in the list
		found := false
		for _, v := range vms {
			vm, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			if vm["name"] == vmName {
				found = true
				break
			}
		}
		assert.True(t, found, "Created VM not found in VM list")
	})

	// Step 6: Delete VM
	t.Run("Delete VM", func(t *testing.T) {
		// Prepare request
		req, err := http.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("%s/api/v1/vms/%s", baseURL, vmName),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify VM is deleted
		time.Sleep(2 * time.Second)
		req, err = http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s/api/v1/vms/%s", baseURL, vmName),
			nil,
		)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		// Send request
		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return 404 Not Found
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// getAuthToken authenticates and returns a JWT token
func getAuthToken(baseURL string) (string, error) {
	// Create login request
	reqBody, err := json.Marshal(map[string]string{
		"username": "admin",
		"password": "password", // Use proper credentials from test setup
	})
	if err != nil {
		return "", err
	}

	// Send request
	resp, err := http.Post(
		fmt.Sprintf("%s/login", baseURL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	// Parse response
	var loginResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		return "", err
	}

	// Extract token
	token, ok := loginResp["token"].(string)
	if !ok || token == "" {
		return "", fmt.Errorf("token not found in response")
	}

	return token, nil
}

// getVMStatus gets the current status of a VM
func getVMStatus(baseURL, token, vmName string) (string, error) {
	// Prepare request
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/vms/%s", baseURL, vmName),
		nil,
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get VM failed with status code: %d", resp.StatusCode)
	}

	// Parse response
	var getResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&getResp)
	if err != nil {
		return "", err
	}

	// Extract status
	vm, ok := getResp["vm"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("VM data not found in response")
	}

	status, ok := vm["status"].(string)
	if !ok {
		return "", fmt.Errorf("status not found in VM data")
	}

	return status, nil
}
