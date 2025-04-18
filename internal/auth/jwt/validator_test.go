package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/internal/errors"
	"github.com/threatflux/libgo/internal/models/user"
)

func TestNewJWTValidator(t *testing.T) {
	// Test with HMAC algorithm
	hmacConfig := config.AuthConfig{
		JWTSecretKey:    "test-secret",
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "HS256",
	}

	hmacValidator := NewJWTValidator(hmacConfig)
	if hmacValidator.algorithm.Alg() != "HS256" {
		t.Errorf("Expected algorithm to be HS256, got %s", hmacValidator.algorithm.Alg())
	}

	if string(hmacValidator.secretKey) != "test-secret" {
		t.Errorf("Expected secretKey to be %q, got %q", "test-secret", string(hmacValidator.secretKey))
	}

	if hmacValidator.issuer != "test-issuer" {
		t.Errorf("Expected issuer to be %q, got %q", "test-issuer", hmacValidator.issuer)
	}

	if len(hmacValidator.audience) != 1 || hmacValidator.audience[0] != "test-audience" {
		t.Errorf("Expected audience to be [%q], got %v", "test-audience", hmacValidator.audience)
	}

	// Test with different signing method
	rsaConfig := config.AuthConfig{
		JWTSecretKey:    "test-secret",
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "RS256",
	}

	rsaValidator := NewJWTValidator(rsaConfig)
	if rsaValidator.algorithm.Alg() != "RS256" {
		t.Errorf("Expected algorithm to be RS256, got %s", rsaValidator.algorithm.Alg())
	}

	// Test with invalid signing method (should default to HS256)
	invalidConfig := config.AuthConfig{
		JWTSecretKey:    "test-secret",
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "INVALID",
	}

	invalidValidator := NewJWTValidator(invalidConfig)
	if invalidValidator.algorithm.Alg() != "HS256" {
		t.Errorf("Expected algorithm to be HS256 (default), got %s", invalidValidator.algorithm.Alg())
	}

	// Test with empty audience
	emptyAudienceConfig := config.AuthConfig{
		JWTSecretKey:    "test-secret",
		Issuer:          "test-issuer",
		Audience:        "",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "HS256",
	}

	emptyAudienceValidator := NewJWTValidator(emptyAudienceConfig)
	if len(emptyAudienceValidator.audience) != 0 {
		t.Errorf("Expected empty audience, got %v", emptyAudienceValidator.audience)
	}
}

func TestJWTValidator_Validate(t *testing.T) {
	// Create a validator and generator for testing
	secretKey := "test-secret"
	config := config.AuthConfig{
		JWTSecretKey:    secretKey,
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "HS256",
	}

	validator := NewJWTValidator(config)
	generator := NewJWTGenerator(config)

	// Create a test user
	testUser := &user.User{
		ID:       "test-id",
		Username: "testuser",
		Roles:    []string{user.RoleAdmin},
		Active:   true,
	}

	// Generate a valid token
	validToken, err := generator.Generate(testUser)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test validating a valid token
	claims, err := validator.Validate(validToken)
	if err != nil {
		t.Errorf("Failed to validate valid token: %v", err)
	}

	if claims == nil {
		t.Fatal("Claims are nil for valid token")
	}

	if claims.UserID != testUser.ID {
		t.Errorf("Expected UserID to be %q, got %q", testUser.ID, claims.UserID)
	}

	// Test validating an invalid token (manually create a token with wrong signature)
	invalidToken := validToken + "invalid"
	_, err = validator.Validate(invalidToken)
	if err == nil {
		t.Error("Expected error when validating invalid token, got nil")
	}

	// The error should be wrapped with ErrInvalidToken
	if !errors.Is(err, errors.ErrInvalidToken) {
		t.Errorf("Expected error to be ErrInvalidToken, got %v", err)
	}

	// Test validating a token with different algorithm
	// Create a token with a different algorithm
	wrongAlgClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			Subject:   testUser.ID,
			Audience:  jwt.ClaimStrings{"test-audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:   testUser.ID,
		Username: testUser.Username,
		Roles:    testUser.Roles,
	}

	// Here we use HS512, but our validator expects HS256
	wrongAlgToken := jwt.NewWithClaims(jwt.SigningMethodHS512, wrongAlgClaims)
	wrongAlgTokenString, err := wrongAlgToken.SignedString([]byte(secretKey))
	if err != nil {
		t.Fatalf("Failed to create token with wrong algorithm: %v", err)
	}

	_, err = validator.Validate(wrongAlgTokenString)
	if err == nil {
		t.Error("Expected error when validating token with wrong algorithm, got nil")
	}

	// Test validating an expired token
	// Create a token that's already expired
	expiredClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			Subject:   testUser.ID,
			Audience:  jwt.ClaimStrings{"test-audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
		UserID:   testUser.ID,
		Username: testUser.Username,
		Roles:    testUser.Roles,
	}

	expiredToken := jwt.NewWithClaims(validator.algorithm, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString([]byte(secretKey))
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = validator.Validate(expiredTokenString)
	if err == nil {
		t.Error("Expected error when validating expired token, got nil")
	}

	// The error should be wrapped with ErrTokenExpired
	if !errors.Is(err, errors.ErrTokenExpired) {
		t.Errorf("Expected error to be ErrTokenExpired, got %v", err)
	}
}

func TestJWTValidator_ValidateWithClaims(t *testing.T) {
	// Create a validator and generator for testing
	secretKey := "test-secret"
	config := config.AuthConfig{
		JWTSecretKey:    secretKey,
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "HS256",
	}

	validator := NewJWTValidator(config)
	generator := NewJWTGenerator(config)

	// Create a test user
	testUser := &user.User{
		ID:       "test-id",
		Username: "testuser",
		Roles:    []string{user.RoleAdmin},
		Active:   true,
	}

	// Generate a valid token
	validToken, err := generator.Generate(testUser)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test validating with custom claims
	customClaims := &Claims{}
	err = validator.ValidateWithClaims(validToken, customClaims)
	if err != nil {
		t.Errorf("Failed to validate valid token with custom claims: %v", err)
	}

	if customClaims.UserID != testUser.ID {
		t.Errorf("Expected UserID to be %q, got %q", testUser.ID, customClaims.UserID)
	}

	if customClaims.Username != testUser.Username {
		t.Errorf("Expected Username to be %q, got %q", testUser.Username, customClaims.Username)
	}

	// Test validating with a completely different claims structure
	// First, let's create a token with a custom claims structure
	type CustomTestClaims struct {
		jwt.RegisteredClaims
		CustomField string `json:"customField"`
	}

	originalClaims := &CustomTestClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			Subject:   testUser.ID,
			Audience:  jwt.ClaimStrings{"test-audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		CustomField: "test-value",
	}

	customToken := jwt.NewWithClaims(validator.algorithm, originalClaims)
	customTokenString, err := customToken.SignedString([]byte(secretKey))
	if err != nil {
		t.Fatalf("Failed to create token with custom claims: %v", err)
	}

	// Now validate with the same claims structure
	validateClaims := &CustomTestClaims{}
	err = validator.ValidateWithClaims(customTokenString, validateClaims)
	if err != nil {
		t.Errorf("Failed to validate valid token with matching custom claims: %v", err)
	}

	if validateClaims.CustomField != originalClaims.CustomField {
		t.Errorf("Expected CustomField to be %q, got %q",
			originalClaims.CustomField, validateClaims.CustomField)
	}

	// Test validating a token with incorrect claims structure
	// This should still work as the standard JWT claims will be validated,
	// but the custom fields won't be accessible
	standardClaims := &jwt.RegisteredClaims{}
	err = validator.ValidateWithClaims(customTokenString, standardClaims)
	if err != nil {
		t.Errorf("Failed to validate token with standard claims: %v", err)
	}

	// The standard claims should be populated
	if standardClaims.Subject != testUser.ID {
		t.Errorf("Expected Subject to be %q, got %q", testUser.ID, standardClaims.Subject)
	}

	// But trying to validate with claims that lack required fields should fail parsing
	_, err = validator.Validate(customTokenString)
	if err == nil {
		t.Error("Expected error when validating token with mismatched claims structure, got nil")
	}
}
