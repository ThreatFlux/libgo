package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/auth/jwt"
	"github.com/threatflux/libgo/internal/auth/user"
	"github.com/threatflux/libgo/pkg/logger"
)

// ErrInvalidToken indicates authentication failed due to invalid token.
var ErrInvalidToken = errors.New("invalid or missing authentication token")

// ErrInsufficientPermissions indicates authorization failed due to insufficient permissions.
var ErrInsufficientPermissions = errors.New("insufficient permissions for this operation")

// JWTMiddleware implements JWT authentication middleware for Gin.
type JWTMiddleware struct {
	validator   jwt.Validator
	userService user.Service
	logger      logger.Logger
}

// NewJWTMiddleware creates a new JWTMiddleware.
func NewJWTMiddleware(validator jwt.Validator, userService user.Service, logger logger.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		validator:   validator,
		userService: userService,
		logger:      logger,
	}
}

// Authenticate middleware for JWT authentication.
func (m *JWTMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.handleAuthError(c, ErrInvalidToken, "Missing authentication token")
			return
		}

		// Check if the Authorization header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			m.handleAuthError(c, ErrInvalidToken, "Invalid authorization format")
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := parts[1]

		// Validate token
		claims, err := m.validator.Validate(tokenString)
		if err != nil {
			m.handleAuthError(c, ErrInvalidToken, "Invalid token: "+err.Error())
			return
		}

		// Get user from claims and verify existence
		user, err := m.userService.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			m.handleAuthError(c, ErrInvalidToken, "User not found or inactive")
			return
		}

		// Check if user is active
		if !user.Active {
			m.handleAuthError(c, ErrInvalidToken, "User account is inactive")
			return
		}

		// Store user and claims in context for later use
		c.Set("user", user)
		c.Set("claims", claims)

		// Continue to next middleware/handler
		c.Next()
	}
}

// Authorize middleware for permission-based authorization.
func (m *JWTMiddleware) Authorize(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		claims, exists := c.Get("claims")
		if !exists {
			m.handleAuthError(c, ErrInvalidToken, "Authentication required")
			return
		}

		jwtClaims, ok := claims.(*jwt.Claims)
		if !ok {
			m.handleAuthError(c, ErrInvalidToken, "Invalid token claims")
			return
		}

		// Check if user has required permission
		hasPermission, err := m.userService.HasPermission(c.Request.Context(), jwtClaims.UserID, permission)
		if err != nil {
			m.logger.Error("Error checking permissions",
				logger.String("userId", jwtClaims.UserID),
				logger.String("permission", permission),
				logger.Error(err))
			m.handleAuthError(c, ErrInsufficientPermissions, "Permission check failed")
			return
		}

		if !hasPermission {
			m.handleAuthError(c, ErrInsufficientPermissions, "Insufficient permissions")
			return
		}

		// Continue to next middleware/handler
		c.Next()
	}
}

// handleAuthError handles authentication and authorization errors.
func (m *JWTMiddleware) handleAuthError(c *gin.Context, err error, message string) {
	statusCode := http.StatusUnauthorized
	if errors.Is(err, ErrInsufficientPermissions) {
		statusCode = http.StatusForbidden
	}

	m.logger.Warn("Authentication/authorization failed",
		logger.String("path", c.Request.URL.Path),
		logger.String("method", c.Request.Method),
		logger.String("message", message),
		logger.Error(err))

	c.AbortWithStatusJSON(statusCode, gin.H{
		"status":  statusCode,
		"code":    strings.ReplaceAll(strings.ToUpper(err.Error()), " ", "_"),
		"message": message,
	})
}
