package logging

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/threatflux/libgo/pkg/logger"
)

func TestRequestLogger(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockLoggerWithFields := logger.NewMockLogger(ctrl)

	// Expect the logger to be called with fields
	mockLogger.EXPECT().
		WithFields(gomock.Any()).
		Return(mockLoggerWithFields).
		AnyTimes()

	// Set up different expectations for different status codes
	mockLoggerWithFields.EXPECT().
		WithFields(gomock.Any()).
		Return(mockLoggerWithFields).
		AnyTimes()

	mockLoggerWithFields.EXPECT().
		Info(gomock.Eq("Request handled")).
		AnyTimes()

	mockLoggerWithFields.EXPECT().
		Warn(gomock.Eq("Client error")).
		AnyTimes()

	mockLoggerWithFields.EXPECT().
		Error(gomock.Eq("Server error")).
		AnyTimes()

	// Create router with middleware
	router := gin.New()

	config := Config{
		SkipPaths:          []string{"/health", "/metrics"},
		MaxBodyLogSize:     1024,
		IncludeRequestBody: true,
	}

	router.Use(RequestLogger(mockLogger, config))

	// Add test routes
	router.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.GET("/client-error", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Bad request"})
	})

	router.GET("/server-error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Server error"})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	router.POST("/with-body", func(c *gin.Context) {
		var body map[string]interface{}
		_ = c.BindJSON(&body)
		c.JSON(http.StatusOK, body)
	})

	// Test cases
	tests := []struct {
		name           string
		path           string
		method         string
		body           string
		expectStatus   int
		requestHeaders map[string]string
	}{
		{
			name:         "Success response",
			path:         "/success",
			method:       "GET",
			expectStatus: http.StatusOK,
		},
		{
			name:         "Client error response",
			path:         "/client-error",
			method:       "GET",
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Server error response",
			path:         "/server-error",
			method:       "GET",
			expectStatus: http.StatusInternalServerError,
		},
		{
			name:         "Skipped path",
			path:         "/health",
			method:       "GET",
			expectStatus: http.StatusOK,
		},
		{
			name:         "With request body",
			path:         "/with-body",
			method:       "POST",
			body:         `{"key":"value"}`,
			expectStatus: http.StatusOK,
		},
		{
			name:         "With request ID",
			path:         "/success",
			method:       "GET",
			expectStatus: http.StatusOK,
			requestHeaders: map[string]string{
				"X-Request-ID": "test-request-id",
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var req *http.Request
			if tt.body != "" {
				req, _ = http.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest(tt.method, tt.path, nil)
			}

			// Add request headers
			for k, v := range tt.requestHeaders {
				req.Header.Set(k, v)
			}

			// Create recorder
			rec := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(rec, req)

			// Check status code
			assert.Equal(t, tt.expectStatus, rec.Code)

			// Check for request ID header in response
			if tt.requestHeaders["X-Request-ID"] != "" {
				assert.Equal(t, tt.requestHeaders["X-Request-ID"], rec.Header().Get("X-Request-ID"))
			} else {
				// Should have generated a UUID
				assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
			}

			// For requests with body, check if body is preserved
			if tt.body != "" {
				var requestBody map[string]interface{}
				var responseBody map[string]interface{}

				err := json.Unmarshal([]byte(tt.body), &requestBody)
				assert.NoError(t, err)

				err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
				assert.NoError(t, err)

				// Response should echo the request for this test endpoint
				assert.Equal(t, requestBody, responseBody)
			}
		})
	}
}

func TestBodyLogWriter(t *testing.T) {
	// Test the bodyLogWriter
	testCases := []struct {
		name          string
		input         string
		maxSize       int
		expectedSize  int
		expectedBytes int
	}{
		{
			name:          "Small body",
			input:         "small body",
			maxSize:       100,
			expectedSize:  10,
			expectedBytes: 10,
		},
		{
			name:          "Body equal to max",
			input:         strings.Repeat("a", 10),
			maxSize:       10,
			expectedSize:  10,
			expectedBytes: 10,
		},
		{
			name:          "Body larger than max",
			input:         strings.Repeat("a", 20),
			maxSize:       10,
			expectedSize:  10,
			expectedBytes: 20,
		},
		{
			name:          "Zero max size",
			input:         "body",
			maxSize:       0,
			expectedSize:  0,
			expectedBytes: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a buffer for the original writer
			originalBuf := &bytes.Buffer{}

			// Create the body log writer
			bodyBuf := &bytes.Buffer{}
			blw := &bodyLogWriter{
				ResponseWriter: &mockResponseWriter{originalBuf},
				bodyBuffer:     bodyBuf,
				maxSize:        tc.maxSize,
			}

			// Write to the body log writer
			n, err := blw.Write([]byte(tc.input))

			// Check results
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedBytes, n)
			assert.Equal(t, tc.input, originalBuf.String())
			assert.Equal(t, tc.expectedSize, bodyBuf.Len())
			if tc.maxSize > 0 && len(tc.input) <= tc.maxSize {
				assert.Equal(t, tc.input, bodyBuf.String())
			}
		})
	}
}

// Mock response writer for testing
type mockResponseWriter struct {
	buf *bytes.Buffer
}

func (m *mockResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return m.buf.Write(b)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	// Do nothing
}

func (m *mockResponseWriter) Status() int {
	return http.StatusOK
}

func (m *mockResponseWriter) Size() int {
	return m.buf.Len()
}

func (m *mockResponseWriter) WriteString(s string) (int, error) {
	return m.buf.WriteString(s)
}
