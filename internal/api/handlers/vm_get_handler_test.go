package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threatflux/libgo/internal/models/vm"
	vmservice "github.com/threatflux/libgo/internal/vm"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	mockvm "github.com/threatflux/libgo/test/mocks/vm"
	"go.uber.org/mock/gomock"
)

func TestVMHandler_GetVM(t *testing.T) {
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
	router.GET("/vms/:name", handler.GetVM)

	// Test cases
	tests := []struct {
		name             string
		vmName           string
		mockSetup        func()
		expectedStatus   int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name:   "Valid VM retrieval",
			vmName: "test-vm",
			mockSetup: func() {
				mockVMManager.EXPECT().Get(gomock.Any(), "test-vm").Return(&vm.VM{
					Name:   "test-vm",
					UUID:   "12345678-1234-1234-1234-123456789012",
					Status: vm.VMStatusRunning,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response GetVMResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, "test-vm", response.VM.Name)
				assert.Equal(t, "12345678-1234-1234-1234-123456789012", response.VM.UUID)
				assert.Equal(t, vm.VMStatusRunning, response.VM.Status)
			},
		},
		{
			name:   "VM not found",
			vmName: "non-existent-vm",
			mockSetup: func() {
				mockVMManager.EXPECT().Get(gomock.Any(), "non-existent-vm").Return(nil, vmservice.ErrVMNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, http.StatusNotFound, response.Status)
				assert.Equal(t, "NOT_FOUND", response.Code)
			},
		},
		{
			name:   "Internal error",
			vmName: "test-vm",
			mockSetup: func() {
				mockVMManager.EXPECT().Get(gomock.Any(), "test-vm").Return(nil, errors.New("internal error"))
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
			req, err := http.NewRequest(http.MethodGet, "/vms/"+tc.vmName, nil)
			require.NoError(t, err)

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
