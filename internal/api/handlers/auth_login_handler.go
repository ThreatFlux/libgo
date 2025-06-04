package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/auth/jwt"
	userservice "github.com/threatflux/libgo/internal/auth/user"
	usermodels "github.com/threatflux/libgo/internal/models/user"
	"github.com/threatflux/libgo/pkg/logger"
)

// AuthHandler handles authentication-related requests.
type AuthHandler struct {
	userService  userservice.Service
	jwtGenerator jwt.Generator
	logger       logger.Logger
	tokenExpiry  time.Duration
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	userService userservice.Service,
	jwtGenerator jwt.Generator,
	logger logger.Logger,
	tokenExpiry time.Duration,
) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		jwtGenerator: jwtGenerator,
		logger:       logger,
		tokenExpiry:  tokenExpiry,
	}
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response.
type LoginResponse struct {
	// Time fields (24 bytes, must be first for alignment).
	ExpiresAt time.Time `json:"expiresAt"`
	// Pointer fields (8 bytes).
	User *usermodels.User `json:"user"`
	// String fields (16 bytes).
	Token string `json:"token"`
}

// Login handles user login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request",
			logger.String("error", err.Error()))
		HandleError(c, ErrInvalidInput)
		return
	}

	// Authenticate the user.
	u, err := h.userService.Authenticate(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Warn("Authentication failed",
			logger.String("username", req.Username),
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Generate JWT token.
	token, err := h.jwtGenerator.GenerateWithExpiration(u, h.tokenExpiry)
	if err != nil {
		h.logger.Error("Failed to generate token",
			logger.String("userId", u.ID),
			logger.Error(err))
		HandleError(c, ErrInternalError)
		return
	}

	// Calculate expiration time.
	expiresAt := time.Now().Add(h.tokenExpiry)

	// Create response.
	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      u,
	}

	h.logger.Info("User logged in successfully",
		logger.String("userId", u.ID),
		logger.String("username", u.Username))

	c.JSON(http.StatusOK, response)
}

// RefreshRequest represents a token refresh request.
type RefreshRequest struct {
	Token string `json:"token" binding:"required"`
}

// Refresh handles token refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleError(c, ErrInvalidInput)
		return
	}

	// Parse the token.
	claims, err := h.jwtGenerator.Parse(req.Token)
	if err != nil {
		h.logger.Warn("Invalid token for refresh",
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Get the user.
	u, err := h.userService.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		h.logger.Warn("Failed to get user for token refresh",
			logger.String("userId", claims.UserID),
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Generate a new token.
	newToken, err := h.jwtGenerator.GenerateWithExpiration(u, h.tokenExpiry)
	if err != nil {
		h.logger.Error("Failed to generate new token",
			logger.String("userId", u.ID),
			logger.Error(err))
		HandleError(c, ErrInternalError)
		return
	}

	// Calculate expiration time.
	expiresAt := time.Now().Add(h.tokenExpiry)

	// Create response.
	response := LoginResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
		User:      u,
	}

	h.logger.Info("Token refreshed successfully",
		logger.String("userId", u.ID),
		logger.String("username", u.Username))

	c.JSON(http.StatusOK, response)
}
