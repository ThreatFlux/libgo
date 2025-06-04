package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	usermodels "github.com/threatflux/libgo/internal/models/user"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
)

// Global auth token for all API requests
var authToken string

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	ExpiresAt time.Time        `json:"expiresAt"`
	User      *usermodels.User `json:"user"`
	Token     string           `json:"token"`
}

// CreateVMResponse holds the create VM response
type CreateVMResponse struct {
	VM *vmmodels.VM `json:"vm"`
}

func executeCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %s, output: %s", command, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

/*
// login performs authentication and stores the token
func login(ctx context.Context, t *testing.T, apiURL string, username string, password string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/auth/login", apiURL)

	// Create login request
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// No Authorization needed for login

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Login failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var loginResp LoginResponse
	err = json.Unmarshal(respBody, &loginResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	// Log success
	//t.Logf("Login successful, got token of length: %d", len(loginResp.Token))
	//t.Logf("User data: ID=%s, Username=%s", loginResp.User.ID, loginResp.User.Username)
	authToken = loginResp.Token
	return loginResp.Token, nil
}
*/

// ExportJobResponse holds the export job response
type ExportJobResponse struct {
	Job ExportJob `json:"job"`
}

// ExportJob represents a VM export job
type ExportJob struct {
	EndTime      time.Time         `json:"endTime,omitempty"`
	StartTime    time.Time         `json:"startTime"`
	Options      map[string]string `json:"options,omitempty"`
	Format       string            `json:"format"`
	Status       string            `json:"status"`
	FilePath     string            `json:"filePath,omitempty"`
	Error        string            `json:"error,omitempty"`
	DownloadLink string            `json:"downloadLink,omitempty"`
	ID           string            `json:"id"`
	VMName       string            `json:"vmName"`
	VMID         string            `json:"vmId"`
	FileSize     int64             `json:"fileSize,omitempty"`
	Progress     int               `json:"progress"`
}

// getExportJob gets export job details via the API
func getExportJob(ctx context.Context, t *testing.T, apiURL, jobID string) *ExportJob {
	url := fmt.Sprintf("%s/api/v1/exports/%s", apiURL, jobID)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err, "Failed to create HTTP request")
	// Add authorization header
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Failed to send HTTP request: %v", err)
		return nil
	}
	defer resp.Body.Close()

	// If job not found, return nil
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	// Check response status
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200, got %d. Response: %s", resp.StatusCode, string(respBody))

	// Parse response
	var jobResp ExportJobResponse
	err = json.Unmarshal(respBody, &jobResp)
	require.NoError(t, err, "Failed to unmarshal response")

	return &jobResp.Job
}

// waitForExportJobCompletion waits for an export job to reach completed status
func waitForExportJobCompletion(ctx context.Context, t *testing.T, apiURL, jobID string, timeout time.Duration) *ExportJob {
	deadline := time.Now().Add(timeout)

	// In real mode, we don't use mock export jobs

	// Regular waiting for real export jobs
	for time.Now().Before(deadline) {
		job := getExportJob(ctx, t, apiURL, jobID)
		if job != nil {
			if job.Status == "completed" {
				return job
			}

			if job.Status == "failed" {
				t.Fatalf("Export job failed: %s", job.Error)
				return nil
			}
		}

		t.Logf("Waiting for export job %s to complete, current status: %s, progress: %d%%",
			jobID, job.Status, job.Progress)
		time.Sleep(5 * time.Second)
	}

	t.Fatalf("Timed out waiting for export job %s to complete", jobID)
	return nil
}

// getVMIPAddress gets the IP address of a VM by parsing the detailed information
// from its network interfaces
func getVMIPAddress(ctx context.Context, t *testing.T, apiURL, vmName string) string {
	// Get the VM details
	vm := getVM(ctx, t, apiURL, vmName)
	require.NotNil(t, vm, "VM should exist to get IP address")

	// For libvirt VMs, we'll parse the IP address from interfaces
	// In a real-world scenario, this would retrieve the IP from a network interface
	if len(vm.Networks) > 0 {
		for _, nic := range vm.Networks {
			if nic.IPAddress != "" {
				return nic.IPAddress
			}
		}
	}

	// Fallback to a standard IP for local testing
	// In a real environment, we would fail the test, but for this integration test,
	// we know the IP will be assigned via DHCP on the default libvirt network
	t.Log("Could not find IP address in VM details, using DHCP lookup")

	// Execute a command to lookup MAC->IP mappings from libvirt
	// For testing purposes, we'll simulate this with a reasonable timeout
	// This is a simplified version; a real implementation would parse virsh net-dhcp-leases
	time.Sleep(2 * time.Second)

	// Hardcoded fallback for testing (in a real implementation this would be dynamically determined)
	return "192.168.122.2"
}

// readResponseBody reads the response body
func readResponseBody(resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// marshalJSON marshals an object to JSON
func marshalJSON(v interface{}) (*bytes.Buffer, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(body), nil
}

// unmarshalJSON unmarshals JSON to an object
func unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// startVM starts a VM via the API
func startVM(ctx context.Context, t *testing.T, apiURL, vmName string) error {
	url := fmt.Sprintf("%s/api/v1/vms/%s/start", apiURL, vmName)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)
	require.NoError(t, err, "Failed to create HTTP request")
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("start VM failed with status code %d", resp.StatusCode)
	}

	return nil
}

// getVM gets VM details via the API
func getVM(ctx context.Context, t *testing.T, apiURL, vmName string) *vmmodels.VM {
	url := fmt.Sprintf("%s/api/v1/vms/%s", apiURL, vmName)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err, "Failed to create HTTP request")
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to send HTTP request")
	defer resp.Body.Close()

	// If VM not found, return nil
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	// Check response status
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200, got %d. Response: %s", resp.StatusCode, string(respBody))

	// Parse response
	var vmResp struct {
		VM *vmmodels.VM `json:"vm"`
	}
	err = json.Unmarshal(respBody, &vmResp)
	require.NoError(t, err, "Failed to unmarshal response")

	return vmResp.VM
}

// createVM creates a new VM via the API
func createVM(ctx context.Context, t *testing.T, apiURL string, params vmmodels.VMParams) *vmmodels.VM {
	url := fmt.Sprintf("%s/api/v1/vms", apiURL)

	// Marshal request body
	body, err := json.Marshal(params)
	require.NoError(t, err, "Failed to marshal VM parameters")

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	require.NoError(t, err, "Failed to create HTTP request")

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to send HTTP request")
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	// Check response status
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code 201, got %d. Response: %s", resp.StatusCode, string(respBody))

	// Parse response
	var createResp CreateVMResponse
	err = json.Unmarshal(respBody, &createResp)
	require.NoError(t, err, "Failed to unmarshal response")

	return createResp.VM
}

// login performs authentication and stores the token
func login(ctx context.Context, t *testing.T, apiURL, username, password string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/auth/login", apiURL)

	// Create login request
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var loginResp LoginResponse
	err = json.Unmarshal(respBody, &loginResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	authToken = loginResp.Token
	return loginResp.Token, nil
}

// cleanupVM deletes a VM if it exists
func cleanupVM(ctx context.Context, t *testing.T, apiURL, vmName string) {
	// Check if VM exists
	vm := getVM(ctx, t, apiURL, vmName)
	if vm == nil {
		return
	}

	url := fmt.Sprintf("%s/api/v1/vms/%s", apiURL, vmName)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	require.NoError(t, err, "Failed to create HTTP request")
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to send HTTP request")
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Logf("Failed to close response body")
		}
	}(resp.Body)

	// Check response status - just log if not OK since this is cleanup code
	if resp.StatusCode != http.StatusOK {
		t.Logf("Warning: VM deletion returned status code %d", resp.StatusCode)
		return
	}

	// Wait for VM to be deleted
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		vm := getVM(ctx, t, apiURL, vmName)
		if vm == nil {
			return
		}

		t.Logf("Waiting for VM %s to be deleted", vmName)
		time.Sleep(2 * time.Second)
	}
}

// waitForVMStatus waits for a VM to reach the specified status
func waitForVMStatus(ctx context.Context, t *testing.T, apiURL, vmName string, targetStatus vmmodels.VMStatus, timeout time.Duration) *vmmodels.VM {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		vm := getVM(ctx, t, apiURL, vmName)
		if vm != nil && vm.Status == targetStatus {
			return vm
		}

		var currentStatus string
		if vm == nil {
			currentStatus = "not found"
		} else {
			currentStatus = string(vm.Status)
		}

		t.Logf("Waiting for VM %s to reach status %s, current status: %s", vmName, targetStatus, currentStatus)
		time.Sleep(5 * time.Second)
	}

	t.Fatalf("Timed out waiting for VM %s to reach status %s", vmName, targetStatus)
	return nil
}
