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
	vmservice "github.com/threatflux/libgo/internal/vm"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	mockvm "github.com/threatflux/libgo/test/mocks/vm"
	"go.uber.org/mock/gomock"
)

func TestVMHandler_DeleteVM(t *testing.T) {
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
	router.DELETE("/vms/:name", handler.DeleteVM)

	// Test cases
	tests := []struct {
		name             string
		vmName           string
		queryParams      string
		mockSetup        func()
		expectedStatus   int
		validateResponse func(t *testing.T, body []byte)
	}{
		{
			name:        "Valid VM deletion",
			vmName:      "test-vm",
			queryParams: "",
			mockSetup: func() {
				mockVMManager.EXPECT().Delete(gomock.Any(), "test-vm").Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response DeleteVMResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "VM deleted successfully", response.Message)
			},
		},
		{
			name:        "Force VM deletion",
			vmName:      "test-vm",
			queryParams: "?force=true",
			mockSetup: func() {
				mockVMManager.EXPECT().Delete(gomock.Any(), "test-vm").Return(nil)
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response DeleteVMResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.True(t, response.Success)
				assert.Equal(t, "VM deleted successfully", response.Message)
			},
		},
		{
			name:        "VM not found",
			vmName:      "non-existent-vm",
			queryParams: "",
			mockSetup: func() {
				mockVMManager.EXPECT().Delete(gomock.Any(), "non-existent-vm").Return(vmservice.ErrVMNotFound)
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
			name:        "VM in use",
			vmName:      "in-use-vm",
			queryParams: "",
			mockSetup: func() {
				mockVMManager.EXPECT().Delete(gomock.Any(), "in-use-vm").Return(vmservice.ErrVMInUse)
			},
			expectedStatus: http.StatusForbidden,
			validateResponse: func(t *testing.T, body []byte) {
				var response ErrorResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, http.StatusForbidden, response.Status)
				assert.Equal(t, "FORBIDDEN", response.Code)
			},
		},
		{
			name:        "Internal error",
			vmName:      "test-vm",
			queryParams: "",
			mockSetup: func() {
				mockVMManager.EXPECT().Delete(gomock.Any(), "test-vm").Return(errors.New("internal error"))
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
			req, err := http.NewRequest(http.MethodDelete, "/vms/"+tc.vmName+tc.queryParams, nil)
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
