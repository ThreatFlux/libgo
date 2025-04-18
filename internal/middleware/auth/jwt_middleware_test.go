package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/wroersma/libgo/internal/auth/jwt"
	"github.com/wroersma/libgo/internal/models/user"
	"github.com/wroersma/libgo/test/mocks/auth"
	"github.com/wroersma/libgo/test/mocks/logger"
)

// Setup test environment
func setupTest(t *testing.T) (
	*gomock.Controller,
	*mocks_auth.MockValidator,
	*mocks_auth.MockService,
	*mocks_logger.MockLogger,
	*JWTMiddleware,
) {
	ctrl := gomock.NewController(t)
	mockValidator := mocks_auth.NewMockValidator(ctrl)
	mockUserService := mocks_auth.NewMockService(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	middleware := NewJWTMiddleware(mockValidator, mockUserService, mockLogger)

	// Configure gin to test mode
	gin.SetMode(gin.TestMode)

	return ctrl, mockValidator, mockUserService, mockLogger, middleware
}

func TestJWTMiddleware_Authenticate_ValidToken(t *testing.T) {
	ctrl, mockValidator, mockUserService, mockLogger, middleware := setupTest(t)
	defer ctrl.Finish()

	// Create test router
	router := gin.New()
	router.Use(middleware.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup expectations
	mockClaims := &jwt.Claims{
		UserID:   "test-user-id",
		Username: "testuser",
		Roles:    []string{"admin"},
	}

	testUser := &user.User{
		ID:       "test-user-id",
		Username: "testuser",
		Active:   true,
		Roles:    []string{"admin"},
	}

	mockValidator.EXPECT().
		Validate("valid-token").
		Return(mockClaims, nil)

	mockUserService.EXPECT().
		GetByID(gomock.Any(), "test-user-id").
		Return(testUser, nil)

	// Create test request with token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTMiddleware_Authenticate_MissingToken(t *testing.T) {
	ctrl, _, _, mockLogger, middleware := setupTest(t)
	defer ctrl.Finish()

	// Create test router
	router := gin.New()
	router.Use(middleware.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Expect logging call
	mockLogger.EXPECT().
		Warn(gomock.Any(), gomock.Any()).
		AnyTimes()

	// Create test request without token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTMiddleware_Authenticate_InvalidTokenFormat(t *testing.T) {
	ctrl, _, _, mockLogger, middleware := setupTest(t)
	defer ctrl.Finish()

	// Create test router
	router := gin.New()
	router.Use(middleware.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Expect logging call
	mockLogger.EXPECT().
		Warn(gomock.Any(), gomock.Any()).
		AnyTimes()

	// Test cases for invalid token formats
	testCases := []struct {
		name          string
		authorization string
	}{
		{
			name:          "No bearer prefix",
			authorization: "invalid-token",
		},
		{
			name:          "Wrong format",
			authorization: "Basic invalid-token",
		},
		{
			name:          "Extra parts",
			authorization: "Bearer token extra-part",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test request with invalid token format
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tc.authorization)
			w := httptest.NewRecorder()

			// Perform the request
			router.ServeHTTP(w, req)

			// Assert the response
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestJWTMiddleware_Authenticate_TokenValidationFails(t *testing.T) {
	ctrl, mockValidator, _, mockLogger, middleware := setupTest(t)
	defer ctrl.Finish()

	// Create test router
	router := gin.New()
	router.Use(middleware.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup expectations
	mockValidator.EXPECT().
		Validate("invalid-token").
		Return(nil, errors.New("token validation failed"))

	// Expect logging call
	mockLogger.EXPECT().
		Warn(gomock.Any(), gomock.Any()).
		AnyTimes()

	// Create test request with invalid token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTMiddleware_Authenticate_UserNotFound(t *testing.T) {
	ctrl, mockValidator, mockUserService, mockLogger, middleware := setupTest(t)
	defer ctrl.Finish()

	// Create test router
	router := gin.New()
	router.Use(middleware.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup expectations
	mockClaims := &jwt.Claims{
		UserID:   "test-user-id",
		Username: "testuser",
		Roles:    []string{"admin"},
	}

	mockValidator.EXPECT().
		Validate("valid-token").
		Return(mockClaims, nil)

	mockUserService.EXPECT().
		GetByID(gomock.Any(), "test-user-id").
		Return(nil, errors.New("user not found"))

	// Expect logging call
	mockLogger.EXPECT().
		Warn(gomock.Any(), gomock.Any()).
		AnyTimes()

	// Create test request with token for non-existent user
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTMiddleware_Authenticate_InactiveUser(t *testing.T) {
	ctrl, mockValidator, mockUserService, mockLogger, middleware := setupTest(t)
	defer ctrl.Finish()

	// Create test router
	router := gin.New()
	router.Use(middleware.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup expectations
	mockClaims := &jwt.Claims{
		UserID:   "test-user-id",
		Username: "testuser",
		Roles:    []string{"admin"},
	}

	inactiveUser := &user.User{
		ID:       "test-user-id",
		Username: "testuser",
		Active:   false, // Inactive user
		Roles:    []string{"admin"},
	}

	mockValidator.EXPECT().
		Validate("valid-token").
		Return(mockClaims, nil)

	mockUserService.EXPECT().
		GetByID(gomock.Any(), "test-user-id").
		Return(inactiveUser, nil)

	// Expect logging call
	mockLogger.EXPECT().
		Warn(gomock.Any(), gomock.Any()).
		AnyTimes()

	// Create test request with token for inactive user
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTMiddleware_Authorize_ValidPermission(t *testing.T) {
	ctrl, mockValidator, mockUserService, mockLogger, middleware := setupTest(t)
	defer ctrl.Finish()

	// Create test router with authentication and authorization
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate successful authentication
		mockClaims := &jwt.Claims{
			UserID:   "test-user-id",
			Username: "testuser",
			Roles:    []string{"admin"},
		}
		c.Set("claims", mockClaims)
		c.Next()
	})
	router.Use(middleware.Authorize("create"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup expectations
	mockUserService.EXPECT().
		HasPermission(gomock.Any(), "test-user-id", "create").
		Return(true, nil)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTMiddleware_Authorize_MissingAuthentication(t *testing.T) {
	ctrl, _, _, mockLogger, middleware := setupTest(t)
	defer ctrl.Finish()

	// Create test router with only authorization
	router := gin.New()
	router.Use(middleware.Authorize("create"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Expect logging call
	mockLogger.EXPECT().
		Warn(gomock.Any(), gomock.Any()).
		AnyTimes()

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTMiddleware_Authorize_PermissionCheckFails(t *testing.T) {
	ctrl, _, mockUserService, mockLogger, middleware := setupTest(t)
	defer ctrl.Finish()

	// Create test router with authentication and authorization
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate successful authentication
		mockClaims := &jwt.Claims{
			UserID:   "test-user-id",
			Username: "testuser",
			Roles:    []string{"admin"},
		}
		c.Set("claims", mockClaims)
		c.Next()
	})
	router.Use(middleware.Authorize("create"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup expectations
	mockUserService.EXPECT().
		HasPermission(gomock.Any(), "test-user-id", "create").
		Return(false, errors.New("permission check failed"))

	// Expect logging calls
	mockLogger.EXPECT().
		Error(gomock.Any(), gomock.Any()).
		AnyTimes()
	mockLogger.EXPECT().
		Warn(gomock.Any(), gomock.Any()).
		AnyTimes()

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestJWTMiddleware_Authorize_InsufficientPermissions(t *testing.T) {
	ctrl, _, mockUserService, mockLogger, middleware := setupTest(t)
	defer ctrl.Finish()

	// Create test router with authentication and authorization
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate successful authentication
		mockClaims := &jwt.Claims{
			UserID:   "test-user-id",
			Username: "testuser",
			Roles:    []string{"viewer"}, // viewer doesn't have create permission
		}
		c.Set("claims", mockClaims)
		c.Next()
	})
	router.Use(middleware.Authorize("create"))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Setup expectations
	mockUserService.EXPECT().
		HasPermission(gomock.Any(), "test-user-id", "create").
		Return(false, nil) // no error, but no permission

	// Expect logging call
	mockLogger.EXPECT().
		Warn(gomock.Any(), gomock.Any()).
		AnyTimes()

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert the response
	assert.Equal(t, http.StatusForbidden, w.Code)
}
