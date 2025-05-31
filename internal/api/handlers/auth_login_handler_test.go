package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threatflux/libgo/internal/auth/jwt"
	"github.com/threatflux/libgo/internal/auth/user"
	usermodels "github.com/threatflux/libgo/internal/models/user"
	mocks_auth "github.com/threatflux/libgo/test/mocks/auth"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

func TestAuthHandler_Login(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks_auth.NewMockUserService(ctrl)
	mockJWTGenerator := mocks_auth.NewMockGenerator(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Create handler
	tokenExpiry := 15 * time.Minute
	handler := NewAuthHandler(mockUserService, mockJWTGenerator, mockLogger, tokenExpiry)

	// Create router
	router := gin.New()
	router.POST("/login", handler.Login)

	// Test data
	validUser := &usermodels.User{
		ID:       "user123",
		Username: "testuser",
		Roles:    []string{"admin"},
		Email:    "test@example.com",
		Active:   true,
	}

	validToken := "valid.jwt.token"

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func()
		expectedStatus int
		checkResponse  func(t *testing.T, response *httptest.ResponseRecorder)
	}{
		{
			name: "Valid credentials",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "password123",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					Authenticate(gomock.Any(), "testuser", "password123").
					Return(validUser, nil)

				mockJWTGenerator.EXPECT().
					GenerateWithExpiration(validUser, tokenExpiry).
					Return(validToken, nil)

				mockLogger.EXPECT().
					Info(gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp LoginResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, validToken, resp.Token)
				assert.NotEmpty(t, resp.ExpiresAt)
				assert.Equal(t, validUser.ID, resp.User.ID)
				assert.Equal(t, validUser.Username, resp.User.Username)
			},
		},
		{
			name: "Missing username",
			requestBody: map[string]interface{}{
				"password": "password123",
			},
			setupMocks: func() {
				mockLogger.EXPECT().
					Warn(gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusBadRequest, resp.Status)
				assert.Equal(t, "INVALID_INPUT", resp.Code)
			},
		},
		{
			name: "Missing password",
			requestBody: map[string]interface{}{
				"username": "testuser",
			},
			setupMocks: func() {
				mockLogger.EXPECT().
					Warn(gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusBadRequest, resp.Status)
				assert.Equal(t, "INVALID_INPUT", resp.Code)
			},
		},
		{
			name: "Invalid credentials",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "wrongpassword",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					Authenticate(gomock.Any(), "testuser", "wrongpassword").
					Return(nil, user.ErrInvalidCredentials)

				mockLogger.EXPECT().
					Warn(gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusUnauthorized, resp.Status)
				assert.Equal(t, "UNAUTHORIZED", resp.Code)
			},
		},
		{
			name: "Inactive user",
			requestBody: map[string]interface{}{
				"username": "inactive",
				"password": "password123",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					Authenticate(gomock.Any(), "inactive", "password123").
					Return(nil, user.ErrUserInactive)

				mockLogger.EXPECT().
					Warn(gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusForbidden, resp.Status)
				assert.Equal(t, "FORBIDDEN", resp.Code)
			},
		},
		{
			name: "Token generation error",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "password123",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					Authenticate(gomock.Any(), "testuser", "password123").
					Return(validUser, nil)

				mockJWTGenerator.EXPECT().
					GenerateWithExpiration(validUser, tokenExpiry).
					Return("", errors.New("token generation error"))

				mockLogger.EXPECT().
					Error(gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusInternalServerError, resp.Status)
				assert.Equal(t, "INTERNAL_SERVER_ERROR", resp.Code)
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.setupMocks()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			resp := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(resp, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, resp.Code)

			// Check response
			tt.checkResponse(t, resp)
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks_auth.NewMockUserService(ctrl)
	mockJWTGenerator := mocks_auth.NewMockGenerator(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Create handler
	tokenExpiry := 15 * time.Minute
	handler := NewAuthHandler(mockUserService, mockJWTGenerator, mockLogger, tokenExpiry)

	// Create router
	router := gin.New()
	router.POST("/refresh", handler.Refresh)

	// Test data
	validToken := "valid.jwt.token"
	newToken := "new.jwt.token"
	validClaims := &jwt.Claims{
		UserID:   "user123",
		Username: "testuser",
		Roles:    []string{"admin"},
	}
	validUser := &usermodels.User{
		ID:       "user123",
		Username: "testuser",
		Roles:    []string{"admin"},
		Email:    "test@example.com",
		Active:   true,
	}

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func()
		expectedStatus int
		checkResponse  func(t *testing.T, response *httptest.ResponseRecorder)
	}{
		{
			name: "Valid token refresh",
			requestBody: map[string]interface{}{
				"token": validToken,
			},
			setupMocks: func() {
				mockJWTGenerator.EXPECT().
					Parse(validToken).
					Return(validClaims, nil)

				mockUserService.EXPECT().
					GetByID(gomock.Any(), validClaims.UserID).
					Return(validUser, nil)

				mockJWTGenerator.EXPECT().
					GenerateWithExpiration(validUser, tokenExpiry).
					Return(newToken, nil)

				mockLogger.EXPECT().
					Info(gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp LoginResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, newToken, resp.Token)
				assert.NotEmpty(t, resp.ExpiresAt)
				assert.Equal(t, validUser.ID, resp.User.ID)
			},
		},
		{
			name:           "Missing token",
			requestBody:    map[string]interface{}{},
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusBadRequest, resp.Status)
				assert.Equal(t, "INVALID_INPUT", resp.Code)
			},
		},
		{
			name: "Invalid token",
			requestBody: map[string]interface{}{
				"token": "invalid.token",
			},
			setupMocks: func() {
				mockJWTGenerator.EXPECT().
					Parse("invalid.token").
					Return(nil, jwt.ErrInvalidToken)

				mockLogger.EXPECT().
					Warn(gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusUnauthorized, resp.Status)
				assert.Equal(t, "UNAUTHORIZED", resp.Code)
			},
		},
		{
			name: "Expired token",
			requestBody: map[string]interface{}{
				"token": "expired.token",
			},
			setupMocks: func() {
				mockJWTGenerator.EXPECT().
					Parse("expired.token").
					Return(nil, jwt.ErrTokenExpired)

				mockLogger.EXPECT().
					Warn(gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusUnauthorized, resp.Status)
				assert.Equal(t, "UNAUTHORIZED", resp.Code)
			},
		},
		{
			name: "User not found",
			requestBody: map[string]interface{}{
				"token": validToken,
			},
			setupMocks: func() {
				mockJWTGenerator.EXPECT().
					Parse(validToken).
					Return(validClaims, nil)

				mockUserService.EXPECT().
					GetByID(gomock.Any(), validClaims.UserID).
					Return(nil, user.ErrUserNotFound)

				mockLogger.EXPECT().
					Warn(gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusInternalServerError, // Mapped from user.ErrUserNotFound
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusInternalServerError, resp.Status)
			},
		},
		{
			name: "Token generation error",
			requestBody: map[string]interface{}{
				"token": validToken,
			},
			setupMocks: func() {
				mockJWTGenerator.EXPECT().
					Parse(validToken).
					Return(validClaims, nil)

				mockUserService.EXPECT().
					GetByID(gomock.Any(), validClaims.UserID).
					Return(validUser, nil)

				mockJWTGenerator.EXPECT().
					GenerateWithExpiration(validUser, tokenExpiry).
					Return("", errors.New("token generation error"))

				mockLogger.EXPECT().
					Error(gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var resp ErrorResponse
				err := json.Unmarshal(response.Body.Bytes(), &resp)
				require.NoError(t, err)

				assert.Equal(t, http.StatusInternalServerError, resp.Status)
				assert.Equal(t, "INTERNAL_SERVER_ERROR", resp.Code)
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			tt.setupMocks()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			resp := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(resp, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, resp.Code)

			// Check response
			tt.checkResponse(t, resp)
		})
	}
}
