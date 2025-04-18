package recovery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/wroersma/libgo/pkg/logger"
)

func TestRecoveryHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockLogger := logger.NewMockLogger(ctrl)
	mockContextLogger := logger.NewMockLogger(ctrl)
	
	// Create router with middleware
	router := gin.New()
	
	// Configure recovery middleware
	config := Config{
		DisableStackTrace: false,
		DisableRecovery:   false,
	}
	
	router.Use(Handler(mockLogger, config))
	
	// Test routes
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})
	
	router.GET("/panic-with-context-logger", func(c *gin.Context) {
		c.Set("logger", mockContextLogger)
		panic("test panic with context logger")
	})
	
	router.GET("/no-panic", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	
	// Test with custom handler
	customHandlerCalled := false
	customConfig := Config{
		DisableStackTrace: true,
		DisableRecovery:   false,
		RecoveryHandler: func(c *gin.Context, err interface{}) {
			customHandlerCalled = true
			c.JSON(http.StatusServiceUnavailable, gin.H{"custom": "handler"})
		},
	}
	
	customRouter := gin.New()
	customRouter.Use(Handler(mockLogger, customConfig))
	customRouter.GET("/custom-handler", func(c *gin.Context) {
		panic("test panic with custom handler")
	})
	
	// Test with disabled recovery
	disabledRouter := gin.New()
	disabledRouter.Use(Handler(mockLogger, Config{DisableRecovery: true}))
	disabledRouter.GET("/disabled", func(c *gin.Context) {
		panic("this should crash")
	})
	
	// Test cases
	tests := []struct {
		name           string
		router         *gin.Engine
		path           string
		expectPanic    bool
		expectStatus   int
		expectResponse string
		setupMocks     func()
	}{
		{
			name:           "Route with panic",
			router:         router,
			path:           "/panic",
			expectStatus:   http.StatusInternalServerError,
			expectResponse: `{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","status":500}`,
			setupMocks: func() {
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		{
			name:           "With context logger",
			router:         router,
			path:           "/panic-with-context-logger",
			expectStatus:   http.StatusInternalServerError,
			expectResponse: `{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","status":500}`,
			setupMocks: func() {
				mockContextLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		{
			name:           "No panic",
			router:         router,
			path:           "/no-panic",
			expectStatus:   http.StatusOK,
			expectResponse: `{"status":"success"}`,
			setupMocks:     func() {},
		},
		{
			name:           "Custom handler",
			router:         customRouter,
			path:           "/custom-handler",
			expectStatus:   http.StatusServiceUnavailable,
			expectResponse: `{"custom":"handler"}`,
			setupMocks: func() {
				// No logger mocks needed since we disabled stack traces
			},
		},
		{
			name:           "Disabled recovery",
			router:         disabledRouter,
			path:           "/disabled",
			expectPanic:    true,
			expectStatus:   0, // Won't reach this point due to panic
			expectResponse: "",
			setupMocks:     func() {},
		},
	}
	
	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			
			// Create request
			req, _ := http.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()
			
			// Wrap test in a recovery function if we expect a panic
			if tt.expectPanic {
				assert.Panics(t, func() {
					tt.router.ServeHTTP(rec, req)
				})
				return
			}
			
			// Normal execution
			tt.router.ServeHTTP(rec, req)
			
			// Check response
			assert.Equal(t, tt.expectStatus, rec.Code)
			if tt.expectResponse != "" {
				assert.JSONEq(t, tt.expectResponse, rec.Body.String())
			}
			
			// Check if custom handler was called
			if tt.name == "Custom handler" {
				assert.True(t, customHandlerCalled)
			}
		})
	}
}
