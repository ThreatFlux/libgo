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
)

// TestVMExport tests the VM export functionality
func TestVMExport(t *testing.T) {
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
	vmName := fmt.Sprintf("export-test-vm-%d", time.Now().Unix())

	// Create VM for export
	t.Run("Setup - Create VM", func(t *testing.T) {
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
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// Wait for VM to be fully created
	time.Sleep(3 * time.Second)

	// Clean up VM after tests
	defer func() {
		// Delete the VM
		req, err := http.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("%s/api/v1/vms/%s", baseURL, vmName),
			nil,
		)
		if err != nil {
			t.Logf("Error creating delete request: %v", err)
			return
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Error deleting VM: %v", err)
			return
		}
		defer resp.Body.Close()
		t.Logf("VM cleanup status: %d", resp.StatusCode)
	}()

	var exportJobID string

	// Test QCOW2 export
	t.Run("Export VM to QCOW2", func(t *testing.T) {
		// Define export parameters
		exportParams := map[string]interface{}{
			"format": "qcow2",
			"options": map[string]string{
				"compression": "zlib",
			},
		}

		// Create request
		reqBody, err := json.Marshal(exportParams)
		require.NoError(t, err)

		// Prepare request
		req, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s/api/v1/vms/%s/export", baseURL, vmName),
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
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		// Parse response
		var exportResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&exportResp)
		require.NoError(t, err)

		// Verify export job in response
		job, ok := exportResp["job"].(map[string]interface{})
		assert.True(t, ok)
		
		// Save job ID for later
		exportJobID, ok = job["id"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, exportJobID)
	})

	// Poll export job status
	t.Run("Poll Export Job Status", func(t *testing.T) {
		require.NotEmpty(t, exportJobID, "Export job ID not available")

		// Wait for export to complete (with timeout)
		const maxRetries = 60 // Wait up to 5 minutes (5s per retry)
		var completed bool

		for i := 0; i < maxRetries; i++ {
			// Prepare request
			req, err := http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("%s/api/v1/exports/%s", baseURL, exportJobID),
				nil,
			)
			require.NoError(t, err)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			
			// Check response status
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				time.Sleep(5 * time.Second)
				continue
			}

			// Parse response
			var jobResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&jobResp)
			resp.Body.Close()
			require.NoError(t, err)

			// Check job status
			job, ok := jobResp["job"].(map[string]interface{})
			require.True(t, ok)

			status, ok := job["status"].(string)
			require.True(t, ok)

			t.Logf("Export job status: %s, progress: %v", status, job["progress"])

			if status == "completed" {
				completed = true
				
				// Verify output path exists
				outputPath, ok := job["outputPath"].(string)
				assert.True(t, ok)
				assert.NotEmpty(t, outputPath)
				
				break
			} else if status == "failed" {
				error, _ := job["error"].(string)
				t.Fatalf("Export job failed: %s", error)
			}

			// Wait before next check
			time.Sleep(5 * time.Second)
		}

		assert.True(t, completed, "Export job did not complete within the timeout period")
	})

	// Test listing export jobs
	t.Run("List Export Jobs", func(t *testing.T) {
		// Prepare request
		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s/api/v1/exports", baseURL),
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

		// Verify jobs in response
		jobs, ok := listResp["jobs"].([]interface{})
		assert.True(t, ok)
		
		// Find our job in the list
		found := false
		for _, j := range jobs {
			job, ok := j.(map[string]interface{})
			if !ok {
				continue
			}
			if job["id"] == exportJobID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created export job not found in jobs list")
	})
}
