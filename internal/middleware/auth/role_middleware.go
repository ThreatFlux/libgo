package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	userservice "github.com/threatflux/libgo/internal/auth/user"
	apierrors "github.com/threatflux/libgo/internal/errors"
	"github.com/threatflux/libgo/internal/models/user"
	"github.com/threatflux/libgo/pkg/logger"
)

// Define context keys
const UserContextKey = "user"

// RoleMiddleware implements role-based access control for Gin
type RoleMiddleware struct {
	userService userservice.Service
	logger      logger.Logger
}

// NewRoleMiddleware creates a new RoleMiddleware
func NewRoleMiddleware(userService userservice.Service, logger logger.Logger) *RoleMiddleware {
	return &RoleMiddleware{
		userService: userService,
		logger:      logger,
	}
}

// RequireRole checks if the user has the specified role
func (m *RoleMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user from the context
		userObj, exists := c.Get(UserContextKey)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Convert the user to the correct type
		u, ok := userObj.(*user.User)
		if !ok {
			m.logger.Error("Failed to get user from context",
				logger.String("type", "User"))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Internal server error",
			})
			return
		}

		// Check if the user has the required role
		if !u.HasRole(role) {
			m.handleForbidden(c, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// RequireAnyRole checks if the user has any of the specified roles
func (m *RoleMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user from the context
		userObj, exists := c.Get(UserContextKey)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Convert the user to the correct type
		u, ok := userObj.(*user.User)
		if !ok {
			m.logger.Error("Failed to get user from context",
				logger.String("type", "User"))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Internal server error",
			})
			return
		}

		// Check if the user has any of the required roles
		if !u.HasAnyRole(roles...) {
			m.handleForbidden(c, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// RequirePermission checks if the user has the specified permission
func (m *RoleMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user from the context
		userObj, exists := c.Get(UserContextKey)
		if !exists {
			m.handleUnauthorized(c, "Authentication required")
			return
		}

		// Convert the user to the correct type
		u, ok := userObj.(*user.User)
		if !ok {
			m.logger.Error("Failed to get user from context",
				logger.String("type", "User"))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Internal server error",
			})
			return
		}

		// Check if the user has the permission via their roles
		hasPermission, err := m.userService.HasPermission(c.Request.Context(), u.ID, permission)
		if err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				m.handleUnauthorized(c, "User not found")
				return
			}
			m.logger.Error("Error checking user permission",
				logger.String("userId", u.ID),
				logger.String("permission", permission),
				logger.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Internal server error",
			})
			return
		}

		if !hasPermission {
			m.handleForbidden(c, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// handleUnauthorized handles unauthorized requests
func (m *RoleMiddleware) handleUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"status":  http.StatusUnauthorized,
		"code":    "UNAUTHORIZED",
		"message": message,
	})
}

// handleForbidden handles forbidden requests
func (m *RoleMiddleware) handleForbidden(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"status":  http.StatusForbidden,
		"code":    "FORBIDDEN",
		"message": message,
	})
}
