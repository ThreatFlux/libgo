package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/threatflux/libgo/internal/config"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
)

// MockVMManagerWithSnapshots extends the mock to include snapshot methods
type MockVMManagerWithSnapshots struct {
	mock.Mock
}

func (m *MockVMManagerWithSnapshots) Create(ctx context.Context, params vmmodels.VMParams) (*vmmodels.VM, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vmmodels.VM), args.Error(1)
}

func (m *MockVMManagerWithSnapshots) Get(ctx context.Context, name string) (*vmmodels.VM, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vmmodels.VM), args.Error(1)
}

func (m *MockVMManagerWithSnapshots) List(ctx context.Context) ([]*vmmodels.VM, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*vmmodels.VM), args.Error(1)
}

func (m *MockVMManagerWithSnapshots) Delete(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockVMManagerWithSnapshots) Start(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockVMManagerWithSnapshots) Stop(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockVMManagerWithSnapshots) Restart(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockVMManagerWithSnapshots) CreateSnapshot(ctx context.Context, vmName string, params vmmodels.SnapshotParams) (*vmmodels.Snapshot, error) {
	args := m.Called(ctx, vmName, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vmmodels.Snapshot), args.Error(1)
}

func (m *MockVMManagerWithSnapshots) ListSnapshots(ctx context.Context, vmName string, opts vmmodels.SnapshotListOptions) ([]*vmmodels.Snapshot, error) {
	args := m.Called(ctx, vmName, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*vmmodels.Snapshot), args.Error(1)
}

func (m *MockVMManagerWithSnapshots) GetSnapshot(ctx context.Context, vmName string, snapshotName string) (*vmmodels.Snapshot, error) {
	args := m.Called(ctx, vmName, snapshotName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vmmodels.Snapshot), args.Error(1)
}

func (m *MockVMManagerWithSnapshots) DeleteSnapshot(ctx context.Context, vmName string, snapshotName string) error {
	args := m.Called(ctx, vmName, snapshotName)
	return args.Error(0)
}

func (m *MockVMManagerWithSnapshots) RevertSnapshot(ctx context.Context, vmName string, snapshotName string) error {
	args := m.Called(ctx, vmName, snapshotName)
	return args.Error(0)
}

func TestCreateSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		vmName         string
		requestBody    interface{}
		setupMock      func(*MockVMManagerWithSnapshots)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "successful snapshot creation",
			vmName: "test-vm",
			requestBody: vmmodels.SnapshotParams{
				Name:          "test-snapshot",
				Description:   "Test snapshot",
				IncludeMemory: true,
				Quiesce:       false,
			},
			setupMock: func(m *MockVMManagerWithSnapshots) {
				snapshot := &vmmodels.Snapshot{
					Name:        "test-snapshot",
					Description: "Test snapshot",
					State:       vmmodels.SnapshotStateRunning,
					CreatedAt:   time.Now(),
					IsCurrent:   true,
					HasMemory:   true,
					HasDisk:     true,
				}
				m.On("CreateSnapshot", mock.Anything, "test-vm", mock.Anything).Return(snapshot, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"snapshot": map[string]interface{}{
					"name":        "test-snapshot",
					"description": "Test snapshot",
					"state":       "running",
					"is_current":  true,
					"has_memory":  true,
					"has_disk":    true,
				},
			},
		},
		{
			name:           "missing VM name",
			vmName:         "",
			requestBody:    vmmodels.SnapshotParams{Name: "test-snapshot"},
			setupMock:      func(m *MockVMManagerWithSnapshots) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request body",
			vmName:         "test-vm",
			requestBody:    "invalid",
			setupMock:      func(m *MockVMManagerWithSnapshots) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing snapshot name",
			vmName:         "test-vm",
			requestBody:    vmmodels.SnapshotParams{},
			setupMock:      func(m *MockVMManagerWithSnapshots) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "snapshot creation error",
			vmName:      "test-vm",
			requestBody: vmmodels.SnapshotParams{Name: "test-snapshot"},
			setupMock: func(m *MockVMManagerWithSnapshots) {
				m.On("CreateSnapshot", mock.Anything, "test-vm", mock.Anything).
					Return(nil, errors.New("failed to create snapshot"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockVM := new(MockVMManagerWithSnapshots)
			tt.setupMock(mockVM)

			log, _ := logger.NewZapLogger(config.LoggingConfig{Level: "debug"})
			handler := NewVMHandler(mockVM, log)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/vms/"+tt.vmName+"/snapshots", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Create gin context
			router := gin.New()
			router.POST("/api/v1/vms/:name/snapshots", handler.CreateSnapshot)
			router.ServeHTTP(rec, req)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Verify response body if expected
			if tt.expectedBody != nil && rec.Code == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Check snapshot fields exist
				snapshot, ok := response["snapshot"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "test-snapshot", snapshot["name"])
				assert.Equal(t, "Test snapshot", snapshot["description"])
			}

			// Verify mock expectations
			mockVM.AssertExpectations(t)
		})
	}
}

func TestListSnapshots(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		vmName         string
		queryParams    map[string]string
		setupMock      func(*MockVMManagerWithSnapshots)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "successful list snapshots",
			vmName:      "test-vm",
			queryParams: map[string]string{},
			setupMock: func(m *MockVMManagerWithSnapshots) {
				snapshots := []*vmmodels.Snapshot{
					{
						Name:      "snapshot1",
						State:     vmmodels.SnapshotStateShutoff,
						CreatedAt: time.Now(),
					},
					{
						Name:      "snapshot2",
						State:     vmmodels.SnapshotStateRunning,
						CreatedAt: time.Now(),
					},
				}
				m.On("ListSnapshots", mock.Anything, "test-vm", mock.Anything).Return(snapshots, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:        "list with metadata",
			vmName:      "test-vm",
			queryParams: map[string]string{"include_metadata": "true"},
			setupMock: func(m *MockVMManagerWithSnapshots) {
				m.On("ListSnapshots", mock.Anything, "test-vm", mock.Anything).Return([]*vmmodels.Snapshot{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "missing VM name",
			vmName:         "",
			queryParams:    map[string]string{},
			setupMock:      func(m *MockVMManagerWithSnapshots) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "list error",
			vmName:      "test-vm",
			queryParams: map[string]string{},
			setupMock: func(m *MockVMManagerWithSnapshots) {
				m.On("ListSnapshots", mock.Anything, "test-vm", mock.Anything).
					Return(nil, errors.New("failed to list snapshots"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockVM := new(MockVMManagerWithSnapshots)
			tt.setupMock(mockVM)

			log, _ := logger.NewZapLogger(config.LoggingConfig{Level: "debug"})
			handler := NewVMHandler(mockVM, log)

			// Create request
			req := httptest.NewRequest("GET", "/api/v1/vms/"+tt.vmName+"/snapshots", nil)

			// Add query parameters
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			// Create response recorder
			rec := httptest.NewRecorder()

			// Create gin context
			router := gin.New()
			router.GET("/api/v1/vms/:name/snapshots", handler.ListSnapshots)
			router.ServeHTTP(rec, req)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Verify response body if successful
			if rec.Code == http.StatusOK {
				var response ListSnapshotsResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, response.Count)
				assert.Equal(t, tt.expectedCount, len(response.Snapshots))
			}

			// Verify mock expectations
			mockVM.AssertExpectations(t)
		})
	}
}
