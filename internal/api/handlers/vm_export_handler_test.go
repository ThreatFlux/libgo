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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	customErrors "github.com/threatflux/libgo/internal/errors"
	"github.com/threatflux/libgo/internal/export"
)

func TestExportHandler_ExportVM(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVMManager := vm_mocks.NewMockManager(ctrl)
	mockExportManager := export_mocks.NewMockManager(ctrl)
	mockLogger := logger_mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Create handler
	handler := NewExportHandler(mockVMManager, mockExportManager, mockLogger)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/vms/:name/export", handler.ExportVM)

	t.Run("Successful export", func(t *testing.T) {
		// Create test job
		testJob := &export.Job{
			ID:        "test-job-id",
			VMName:    "test-vm",
			Format:    "qcow2",
			Status:    export.StatusPending,
			Progress:  0,
			StartTime: time.Now(),
		}

		// Create test request
		reqBody := map[string]interface{}{
			"format":   "qcow2",
			"fileName": "test-export.qcow2",
			"options": map[string]string{
				"compression": "9",
			},
		}
		reqJSON, err := json.Marshal(reqBody)
		require.NoError(t, err)

		// Set up expectations
		mockExportManager.EXPECT().
			CreateExportJob(gomock.Any(), "test-vm", gomock.Any()).
			DoAndReturn(func(ctx context.Context, vmName string, params export.Params) (*export.Job, error) {
				// Verify parameters
				assert.Equal(t, "qcow2", params.Format)
				assert.Equal(t, "test-export.qcow2", params.FileName)
				assert.Equal(t, "9", params.Options["compression"])
				return testJob, nil
			})

		// Create request
		req, err := http.NewRequest(http.MethodPost, "/vms/test-vm/export", bytes.NewBuffer(reqJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusAccepted, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify job in response
		job, ok := response["job"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "test-job-id", job["id"])
		assert.Equal(t, "test-vm", job["vmName"])
		assert.Equal(t, "qcow2", job["format"])
		assert.Equal(t, "pending", job["status"])
	})

	t.Run("VM not found", func(t *testing.T) {
		// Create test request
		reqBody := map[string]interface{}{
			"format": "qcow2",
		}
		reqJSON, err := json.Marshal(reqBody)
		require.NoError(t, err)

		// Set up expectations
		mockExportManager.EXPECT().
			CreateExportJob(gomock.Any(), "non-existent-vm", gomock.Any()).
			Return(nil, errors.New("VM not found: non-existent-vm"))

		// Create request
		req, err := http.NewRequest(http.MethodPost, "/vms/non-existent-vm/export", bytes.NewBuffer(reqJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Invalid format", func(t *testing.T) {
		// Create test request with invalid format
		reqBody := map[string]interface{}{
			"format": "invalid-format",
		}
		reqJSON, err := json.Marshal(reqBody)
		require.NoError(t, err)

		// Create request
		req, err := http.NewRequest(http.MethodPost, "/vms/test-vm/export", bytes.NewBuffer(reqJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unsupported format error", func(t *testing.T) {
		// Create test request
		reqBody := map[string]interface{}{
			"format": "qcow2",
		}
		reqJSON, err := json.Marshal(reqBody)
		require.NoError(t, err)

		// Set up expectations
		mockExportManager.EXPECT().
			CreateExportJob(gomock.Any(), "test-vm", gomock.Any()).
			Return(nil, customErrors.ErrUnsupportedFormat)

		// Create request
		req, err := http.NewRequest(http.MethodPost, "/vms/test-vm/export", bytes.NewBuffer(reqJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestExportHandler_GetExportStatus(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVMManager := vm_mocks.NewMockManager(ctrl)
	mockExportManager := export_mocks.NewMockManager(ctrl)
	mockLogger := logger_mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Create handler
	handler := NewExportHandler(mockVMManager, mockExportManager, mockLogger)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/exports/:id", handler.GetExportStatus)

	t.Run("Get existing job", func(t *testing.T) {
		// Create test job
		testJob := &export.Job{
			ID:        "test-job-id",
			VMName:    "test-vm",
			Format:    "qcow2",
			Status:    export.StatusRunning,
			Progress:  50,
			StartTime: time.Now(),
		}

		// Set up expectations
		mockExportManager.EXPECT().
			GetJob(gomock.Any(), "test-job-id").
			Return(testJob, nil)

		// Create request
		req, err := http.NewRequest(http.MethodGet, "/exports/test-job-id", nil)
		require.NoError(t, err)

		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify job in response
		job, ok := response["job"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "test-job-id", job["id"])
		assert.Equal(t, "test-vm", job["vmName"])
		assert.Equal(t, "running", job["status"])
		assert.Equal(t, float64(50), job["progress"])
	})

	t.Run("Job not found", func(t *testing.T) {
		// Set up expectations
		mockExportManager.EXPECT().
			GetJob(gomock.Any(), "non-existent-job").
			Return(nil, customErrors.ErrExportJobNotFound)

		// Create request
		req, err := http.NewRequest(http.MethodGet, "/exports/non-existent-job", nil)
		require.NoError(t, err)

		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestExportHandler_CancelExport(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVMManager := vm_mocks.NewMockManager(ctrl)
	mockExportManager := export_mocks.NewMockManager(ctrl)
	mockLogger := logger_mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Create handler
	handler := NewExportHandler(mockVMManager, mockExportManager, mockLogger)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/exports/:id", handler.CancelExport)

	t.Run("Cancel existing job", func(t *testing.T) {
		// Set up expectations
		mockExportManager.EXPECT().
			CancelJob(gomock.Any(), "test-job-id").
			Return(nil)

		// Create request
		req, err := http.NewRequest(http.MethodDelete, "/exports/test-job-id", nil)
		require.NoError(t, err)

		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Job not found", func(t *testing.T) {
		// Set up expectations
		mockExportManager.EXPECT().
			CancelJob(gomock.Any(), "non-existent-job").
			Return(customErrors.ErrExportJobNotFound)

		// Create request
		req, err := http.NewRequest(http.MethodDelete, "/exports/non-existent-job", nil)
		require.NoError(t, err)

		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestExportHandler_ListExports(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVMManager := vm_mocks.NewMockManager(ctrl)
	mockExportManager := export_mocks.NewMockManager(ctrl)
	mockLogger := logger_mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Create handler
	handler := NewExportHandler(mockVMManager, mockExportManager, mockLogger)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/exports", handler.ListExports)

	t.Run("List all jobs", func(t *testing.T) {
		// Create test jobs
		testJobs := []*export.Job{
			{
				ID:        "job-1",
				VMName:    "vm-1",
				Format:    "qcow2",
				Status:    export.StatusCompleted,
				Progress:  100,
				StartTime: time.Now().Add(-10 * time.Minute),
				EndTime:   time.Now().Add(-5 * time.Minute),
			},
			{
				ID:        "job-2",
				VMName:    "vm-2",
				Format:    "vmdk",
				Status:    export.StatusRunning,
				Progress:  50,
				StartTime: time.Now().Add(-3 * time.Minute),
			},
		}

		// Set up expectations
		mockExportManager.EXPECT().
			ListJobs(gomock.Any()).
			Return(testJobs, nil)

		// Create request
		req, err := http.NewRequest(http.MethodGet, "/exports", nil)
		require.NoError(t, err)

		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify jobs in response
		jobs, ok := response["jobs"].([]interface{})
		require.True(t, ok)
		assert.Len(t, jobs, 2)

		// Check job details
		job1 := jobs[0].(map[string]interface{})
		job2 := jobs[1].(map[string]interface{})

		// Verify jobs (order not guaranteed)
		jobMap := make(map[string]map[string]interface{})
		jobMap[job1["id"].(string)] = job1
		jobMap[job2["id"].(string)] = job2

		j1 := jobMap["job-1"]
		assert.Equal(t, "vm-1", j1["vmName"])
		assert.Equal(t, "completed", j1["status"])
		assert.Equal(t, float64(100), j1["progress"])

		j2 := jobMap["job-2"]
		assert.Equal(t, "vm-2", j2["vmName"])
		assert.Equal(t, "running", j2["status"])
		assert.Equal(t, float64(50), j2["progress"])
	})

	t.Run("List jobs error", func(t *testing.T) {
		// Set up expectations
		mockExportManager.EXPECT().
			ListJobs(gomock.Any()).
			Return(nil, errors.New("database error"))

		// Create request
		req, err := http.NewRequest(http.MethodGet, "/exports", nil)
		require.NoError(t, err)

		// Perform request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
