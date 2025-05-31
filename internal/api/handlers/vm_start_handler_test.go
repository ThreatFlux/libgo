package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vmservice "github.com/threatflux/libgo/internal/vm"
	mocklogger "github.com/threatflux/libgo/test/mocks/logger"
	mockvm "github.com/threatflux/libgo/test/mocks/vm"
)

func TestVMHandler_StartVM(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVMManager := mockvm.NewMockManager(ctrl)
	mockLogger := mocklogger.NewMockLogger(ctrl)

	// Expect logger methods to be called
	mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	handler := NewVMHandler(mockVMManager, mockLogger)

	// Setup router
	router := gin.New()
	router.PUT("/vms/:name/start", handler.StartVM)

	// Test cases
	tests := []struct {
		name             string
		vmName           string
		mockSetup        func()
		expectedStatus   int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name:   "Valid VM start",
			vmName: "test-vm",
			mockSetup: func() {
				mockVMManager.EXPECT().Start(gomock.Any(), "test-vm").Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response StartVMResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "VM started successfully", response.Message)
			},
		},
		{
			name:   "VM not found",
			vmName: "non-existent-vm",
			mockSetup: func() {
				mockVMManager.EXPECT().Start(gomock.Any(), "non-existent-vm").Return(vmservice.ErrVMNotFound)
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
			name:   "VM already running",
			vmName: "running-vm",
			mockSetup: func() {
				mockVMManager.EXPECT().Start(gomock.Any(), "running-vm").Return(vmservice.ErrVMInvalidState)
			},
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
			name:   "Internal error",
			vmName: "test-vm",
			mockSetup: func() {
				mockVMManager.EXPECT().Start(gomock.Any(), "test-vm").Return(errors.New("internal error"))
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
			req, err := http.NewRequest(http.MethodPut, "/vms/"+tc.vmName+"/start", nil)
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
