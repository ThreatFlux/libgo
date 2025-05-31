package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vm_models "github.com/threatflux/libgo/internal/models/vm"
	vmservice "github.com/threatflux/libgo/internal/vm"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	mockvm "github.com/threatflux/libgo/test/mocks/vm"
	"go.uber.org/mock/gomock"
)

func TestVMHandler_CreateVM(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVMManager := mockvm.NewMockManager(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Expect logger methods to be called
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewVMHandler(mockVMManager, mockLogger)

	// Setup router
	router := gin.New()
	router.POST("/vms", handler.CreateVM)

	// Test cases
	tests := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func()
		expectedStatus   int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name: "Valid VM creation",
			requestBody: vm_models.VMParams{
				Name: "test-vm",
				CPU: vm_models.CPUParams{
					Count: 2,
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024, // 2 GB
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024, // 10 GB
					Format:    "qcow2",
				},
			},
			mockSetup: func() {
				mockVMManager.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&vm_models.VM{
					Name: "test-vm",
					UUID: "12345678-1234-1234-1234-123456789012",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, body []byte) {
				var response CreateVMResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "test-vm", response.VM.Name)
				assert.Equal(t, "12345678-1234-1234-1234-123456789012", response.VM.UUID)
			},
		},
		{
			name: "Invalid request body",
			requestBody: map[string]string{
				"invalid": "params",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, http.StatusBadRequest, response.Status)
				assert.Equal(t, "INVALID_INPUT", response.Code)
			},
		},
		{
			name: "Invalid VM parameters",
			requestBody: vm_models.VMParams{
				Name: "test-vm",
				CPU: vm_models.CPUParams{
					Count: 0, // Invalid CPU count
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024,
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024,
					Format:    "qcow2",
				},
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, http.StatusBadRequest, response.Status)
			},
		},
		{
			name: "VM already exists",
			requestBody: vm_models.VMParams{
				Name: "existing-vm",
				CPU: vm_models.CPUParams{
					Count: 2,
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024,
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024,
					Format:    "qcow2",
				},
			},
			mockSetup: func() {
				mockVMManager.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, vmservice.ErrVMAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			validateResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, http.StatusConflict, response.Status)
				assert.Equal(t, "RESOURCE_CONFLICT", response.Code)
			},
		},
		{
			name: "Internal error",
			requestBody: vm_models.VMParams{
				Name: "test-vm",
				CPU: vm_models.CPUParams{
					Count: 2,
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024,
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024,
					Format:    "qcow2",
				},
			},
			mockSetup: func() {
				mockVMManager.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, http.StatusInternalServerError, response.Status)
				assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks for this test case
			tc.mockSetup()

			// Create request
			jsonData, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/vms", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tc.expectedStatus, w.Code)
			tc.validateResponse(t, w.Body.Bytes())
		})
	}
}

func TestVMHandler_validateCreateParams(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVMManager := mockvm.NewMockManager(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)
	handler := NewVMHandler(mockVMManager, mockLogger)

	// Test cases
	tests := []struct {
		name        string
		params      vm_models.VMParams
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Valid parameters",
			params: vm_models.VMParams{
				Name: "test-vm",
				CPU: vm_models.CPUParams{
					Count: 2,
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024, // 2 GB
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024, // 10 GB
					Format:    "qcow2",
				},
			},
			wantErr: false,
		},
		{
			name: "Empty VM name",
			params: vm_models.VMParams{
				Name: "",
				CPU: vm_models.CPUParams{
					Count: 2,
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024,
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024,
					Format:    "qcow2",
				},
			},
			wantErr:     true,
			expectedErr: ErrInvalidInput,
		},
		{
			name: "Invalid CPU count",
			params: vm_models.VMParams{
				Name: "test-vm",
				CPU: vm_models.CPUParams{
					Count: 0, // Invalid
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024,
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024,
					Format:    "qcow2",
				},
			},
			wantErr:     true,
			expectedErr: vmservice.ErrInvalidCPUCount,
		},
		{
			name: "Memory too small",
			params: vm_models.VMParams{
				Name: "test-vm",
				CPU: vm_models.CPUParams{
					Count: 2,
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 10 * 1024 * 1024, // 10 MB (too small)
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024,
					Format:    "qcow2",
				},
			},
			wantErr:     true,
			expectedErr: vmservice.ErrInvalidMemorySize,
		},
		{
			name: "Disk too small",
			params: vm_models.VMParams{
				Name: "test-vm",
				CPU: vm_models.CPUParams{
					Count: 2,
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024,
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 100 * 1024 * 1024, // 100 MB (too small)
					Format:    "qcow2",
				},
			},
			wantErr:     true,
			expectedErr: vmservice.ErrInvalidDiskSize,
		},
		{
			name: "Invalid disk format",
			params: vm_models.VMParams{
				Name: "test-vm",
				CPU: vm_models.CPUParams{
					Count: 2,
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024,
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024,
					Format:    "invalid-format", // Invalid
				},
			},
			wantErr:     true,
			expectedErr: vmservice.ErrInvalidDiskFormat,
		},
		{
			name: "Invalid network type",
			params: vm_models.VMParams{
				Name: "test-vm",
				CPU: vm_models.CPUParams{
					Count: 2,
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024,
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024,
					Format:    "qcow2",
				},
				Network: vm_models.NetParams{
					Type:   "invalid-type", // Invalid
					Source: "default",
				},
			},
			wantErr:     true,
			expectedErr: vmservice.ErrInvalidNetworkType,
		},
		{
			name: "Missing network source",
			params: vm_models.VMParams{
				Name: "test-vm",
				CPU: vm_models.CPUParams{
					Count: 2,
				},
				Memory: vm_models.MemoryParams{
					SizeBytes: 2 * 1024 * 1024 * 1024,
				},
				Disk: vm_models.DiskParams{
					SizeBytes: 10 * 1024 * 1024 * 1024,
					Format:    "qcow2",
				},
				Network: vm_models.NetParams{
					Type:   "bridge",
					Source: "", // Missing
				},
			},
			wantErr:     true,
			expectedErr: vmservice.ErrInvalidNetworkSource,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := handler.validateCreateParams(tc.params)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
