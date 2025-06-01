package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	apierrors "github.com/threatflux/libgo/internal/errors"
	user_models "github.com/threatflux/libgo/internal/models/user"
	mocks_auth "github.com/threatflux/libgo/test/mocks/auth"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

func TestRoleMiddleware_RequireRole(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks_auth.NewMockService(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	middleware := NewRoleMiddleware(mockUserService, mockLogger)

	// Test users
	adminUser := &user_models.User{
		ID:        "admin123",
		Username:  "admin",
		Roles:     []string{"admin"},
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	viewerUser := &user_models.User{
		ID:        "viewer123",
		Username:  "viewer",
		Roles:     []string{"viewer"},
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add test routes with middleware
	router.GET("/admin-only", func(c *gin.Context) {
		c.Set(UserContextKey, adminUser)
		middleware.RequireRole("admin")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.GET("/admin-only-viewer", func(c *gin.Context) {
		c.Set(UserContextKey, viewerUser)
		middleware.RequireRole("admin")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.GET("/no-auth", middleware.RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.GET("/invalid-user", func(c *gin.Context) {
		c.Set(UserContextKey, "not-a-user")
		middleware.RequireRole("admin")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		setupMocks func()
	}{
		{
			name:       "Admin accessing admin-only",
			path:       "/admin-only",
			wantStatus: http.StatusOK,
			setupMocks: func() {},
		},
		{
			name:       "Viewer accessing admin-only",
			path:       "/admin-only-viewer",
			wantStatus: http.StatusForbidden,
			setupMocks: func() {},
		},
		{
			name:       "No authentication",
			path:       "/no-auth",
			wantStatus: http.StatusUnauthorized,
			setupMocks: func() {},
		},
		{
			name:       "Invalid user in context",
			path:       "/invalid-user",
			wantStatus: http.StatusInternalServerError,
			setupMocks: func() {
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any())
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req, _ := http.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestRoleMiddleware_RequireAnyRole(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks_auth.NewMockService(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	middleware := NewRoleMiddleware(mockUserService, mockLogger)

	// Test users
	adminUser := &user_models.User{
		ID:       "admin123",
		Username: "admin",
		Roles:    []string{"admin"},
		Active:   true,
	}

	operatorUser := &user_models.User{
		ID:       "operator123",
		Username: "operator",
		Roles:    []string{"operator"},
		Active:   true,
	}

	viewerUser := &user_models.User{
		ID:       "viewer123",
		Username: "viewer",
		Roles:    []string{"viewer"},
		Active:   true,
	}

	// Add test routes with middleware
	router.GET("/admin-or-operator", func(c *gin.Context) {
		c.Set(UserContextKey, adminUser)
		middleware.RequireAnyRole("admin", "operator")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.GET("/operator-only", func(c *gin.Context) {
		c.Set(UserContextKey, operatorUser)
		middleware.RequireAnyRole("operator")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.GET("/viewer-trying-admin", func(c *gin.Context) {
		c.Set(UserContextKey, viewerUser)
		middleware.RequireAnyRole("admin", "operator")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "Admin accessing admin-or-operator",
			path:       "/admin-or-operator",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Operator accessing operator-only",
			path:       "/operator-only",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Viewer trying to access admin/operator route",
			path:       "/viewer-trying-admin",
			wantStatus: http.StatusForbidden,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestRoleMiddleware_RequirePermission(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks_auth.NewMockService(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	middleware := NewRoleMiddleware(mockUserService, mockLogger)

	// Test users
	adminUser := &user_models.User{
		ID:       "admin123",
		Username: "admin",
		Roles:    []string{"admin"},
		Active:   true,
	}

	viewerUser := &user_models.User{
		ID:       "viewer123",
		Username: "viewer",
		Roles:    []string{"viewer"},
		Active:   true,
	}

	inactiveUser := &user_models.User{
		ID:       "inactive123",
		Username: "inactive",
		Roles:    []string{"admin"},
		Active:   false,
	}

	// Add test routes with middleware
	router.GET("/create-permission", func(c *gin.Context) {
		c.Set(UserContextKey, adminUser)
		middleware.RequirePermission("create")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.GET("/read-permission", func(c *gin.Context) {
		c.Set(UserContextKey, viewerUser)
		middleware.RequirePermission("read")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.GET("/inactive-user", func(c *gin.Context) {
		c.Set(UserContextKey, inactiveUser)
		middleware.RequirePermission("read")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.GET("/service-error", func(c *gin.Context) {
		c.Set(UserContextKey, adminUser)
		middleware.RequirePermission("create")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	router.GET("/user-not-found", func(c *gin.Context) {
		c.Set(UserContextKey, adminUser)
		middleware.RequirePermission("create")(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		setupMocks func()
	}{
		{
			name:       "Admin has create permission",
			path:       "/create-permission",
			wantStatus: http.StatusOK,
			setupMocks: func() {
				mockUserService.EXPECT().HasPermission(gomock.Any(), adminUser.ID, "create").Return(true, nil)
			},
		},
		{
			name:       "Viewer has read permission",
			path:       "/read-permission",
			wantStatus: http.StatusOK,
			setupMocks: func() {
				mockUserService.EXPECT().HasPermission(gomock.Any(), viewerUser.ID, "read").Return(true, nil)
			},
		},
		{
			name:       "Inactive user attempts access",
			path:       "/inactive-user",
			wantStatus: http.StatusForbidden, // Should be forbidden when permission is false
			setupMocks: func() {
				mockUserService.EXPECT().HasPermission(gomock.Any(), inactiveUser.ID, "read").Return(false, nil)
			},
		},
		{
			name:       "Service error",
			path:       "/service-error",
			wantStatus: http.StatusInternalServerError,
			setupMocks: func() {
				mockUserService.EXPECT().HasPermission(gomock.Any(), adminUser.ID, "create").Return(false, errors.New("database error"))
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		{
			name:       "User not found",
			path:       "/user-not-found",
			wantStatus: http.StatusUnauthorized,
			setupMocks: func() {
				mockUserService.EXPECT().HasPermission(gomock.Any(), adminUser.ID, "create").Return(false, apierrors.ErrNotFound)
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req, _ := http.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
