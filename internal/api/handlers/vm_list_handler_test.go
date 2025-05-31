package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/threatflux/libgo/internal/models/vm"
	"go.uber.org/mock/gomock"

	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	mock_vm "github.com/threatflux/libgo/test/mocks/vm"
)

func TestVMHandler_ListVMs(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Test cases
	testCases := []struct {
		name           string
		url            string
		setupMocks     func(mockVMManager *mock_vm.MockManager, mockLogger *mock_logger.MockLogger)
		expectedStatus int
		expectedCount  int
		expectError    bool
	}{
		{
			name: "List VMs successfully",
			url:  "/vms",
			setupMocks: func(mockVMManager *mock_vm.MockManager, mockLogger *mock_logger.MockLogger) {
				vms := []*vm.VM{
					{
						Name:   "test-vm-1",
						UUID:   uuid.NewString(),
						Status: vm.VMStatusRunning,
					},
					{
						Name:   "test-vm-2",
						UUID:   uuid.NewString(),
						Status: vm.VMStatusStopped,
					},
				}

				mockVMManager.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(vms, nil)

				mockLogger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectError:    false,
		},
		{
			name: "List VMs with pagination",
			url:  "/vms?page=2&pageSize=1",
			setupMocks: func(mockVMManager *mock_vm.MockManager, mockLogger *mock_logger.MockLogger) {
				vms := []*vm.VM{
					{
						Name:   "test-vm-1",
						UUID:   uuid.NewString(),
						Status: vm.VMStatusRunning,
					},
					{
						Name:   "test-vm-2",
						UUID:   uuid.NewString(),
						Status: vm.VMStatusStopped,
					},
				}

				mockVMManager.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(vms, nil)

				mockLogger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Total count is 2, but only 1 VM should be returned due to pagination
			expectError:    false,
		},
		{
			name: "List VMs with invalid page parameter",
			url:  "/vms?page=invalid",
			setupMocks: func(mockVMManager *mock_vm.MockManager, mockLogger *mock_logger.MockLogger) {
				// No VM list call should be made due to validation error

				mockLogger.EXPECT().
					Error(gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "List VMs with invalid pageSize parameter",
			url:  "/vms?pageSize=invalid",
			setupMocks: func(mockVMManager *mock_vm.MockManager, mockLogger *mock_logger.MockLogger) {
				// No VM list call should be made due to validation error

				mockLogger.EXPECT().
					Error(gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "VM manager returns error",
			url:  "/vms",
			setupMocks: func(mockVMManager *mock_vm.MockManager, mockLogger *mock_logger.MockLogger) {
				mockVMManager.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))

				mockLogger.EXPECT().
					Error(gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name: "Filter by VM name",
			url:  "/vms?name=test-vm",
			setupMocks: func(mockVMManager *mock_vm.MockManager, mockLogger *mock_logger.MockLogger) {
				// Use gomock.Any() for filter parameter to avoid complex matcher
				vms := []*vm.VM{
					{
						Name:   "test-vm",
						UUID:   uuid.NewString(),
						Status: vm.VMStatusRunning,
					},
				}

				mockVMManager.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(vms, nil)

				mockLogger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectError:    false,
		},
		{
			name: "Filter by VM status",
			url:  "/vms?status=running",
			setupMocks: func(mockVMManager *mock_vm.MockManager, mockLogger *mock_logger.MockLogger) {
				vms := []*vm.VM{
					{
						Name:   "test-vm-1",
						UUID:   uuid.NewString(),
						Status: vm.VMStatusRunning,
					},
				}

				mockVMManager.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(vms, nil)

				mockLogger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectError:    false,
		},
		{
			name: "Page beyond available data",
			url:  "/vms?page=10&pageSize=10",
			setupMocks: func(mockVMManager *mock_vm.MockManager, mockLogger *mock_logger.MockLogger) {
				vms := []*vm.VM{
					{
						Name:   "test-vm-1",
						UUID:   uuid.NewString(),
						Status: vm.VMStatusRunning,
					},
					{
						Name:   "test-vm-2",
						UUID:   uuid.NewString(),
						Status: vm.VMStatusStopped,
					},
				}

				mockVMManager.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(vms, nil)

				mockLogger.EXPECT().
					Info(gomock.Any(), gomock.Any()).
					AnyTimes()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Total count should still be 2
			expectError:    false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup controller and mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVMManager := mock_vm.NewMockManager(ctrl)
			mockLogger := mocks_logger.NewMockLogger(ctrl)

			// Setup handler
			handler := NewVMHandler(mockVMManager, mockLogger)

			// Set up mock expectations
			tc.setupMocks(mockVMManager, mockLogger)

			// Setup router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("logger", mockLogger)
				c.Next()
			})
			router.GET("/vms", handler.ListVMs)

			// Create request and recorder
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			// Perform request
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, w.Code)
			}

			// Check response
			if !tc.expectError {
				var response ListVMsResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				// Check total count
				if response.Count != tc.expectedCount {
					t.Errorf("Expected total count %d, got %d", tc.expectedCount, response.Count)
				}
			} else {
				var errorResponse ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}

				// Check status in error response
				if errorResponse.Status != tc.expectedStatus {
					t.Errorf("Expected error status %d, got %d", tc.expectedStatus, errorResponse.Status)
				}
			}
		})
	}
}
