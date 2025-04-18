package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/internal/models/user"
)

func TestNewJWTGenerator(t *testing.T) {
	// Test with HMAC algorithm
	hmacConfig := config.AuthConfig{
		JWTSecretKey:    "test-secret",
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "HS256",
	}

	hmacGen := NewJWTGenerator(hmacConfig)
	if hmacGen.algorithm.Alg() != "HS256" {
		t.Errorf("Expected algorithm to be HS256, got %s", hmacGen.algorithm.Alg())
	}

	if string(hmacGen.secretKey) != "test-secret" {
		t.Errorf("Expected secretKey to be %q, got %q", "test-secret", string(hmacGen.secretKey))
	}

	if hmacGen.issuer != "test-issuer" {
		t.Errorf("Expected issuer to be %q, got %q", "test-issuer", hmacGen.issuer)
	}

	if len(hmacGen.audience) != 1 || hmacGen.audience[0] != "test-audience" {
		t.Errorf("Expected audience to be [%q], got %v", "test-audience", hmacGen.audience)
	}

	if hmacGen.expiresIn != 15*time.Minute {
		t.Errorf("Expected expiresIn to be %v, got %v", 15*time.Minute, hmacGen.expiresIn)
	}

	// Test with different signing method
	rsaConfig := config.AuthConfig{
		JWTSecretKey:    "test-secret",
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "RS256",
	}

	rsaGen := NewJWTGenerator(rsaConfig)
	if rsaGen.algorithm.Alg() != "RS256" {
		t.Errorf("Expected algorithm to be RS256, got %s", rsaGen.algorithm.Alg())
	}

	// Test with invalid signing method (should default to HS256)
	invalidConfig := config.AuthConfig{
		JWTSecretKey:    "test-secret",
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "INVALID",
	}

	invalidGen := NewJWTGenerator(invalidConfig)
	if invalidGen.algorithm.Alg() != "HS256" {
		t.Errorf("Expected algorithm to be HS256 (default), got %s", invalidGen.algorithm.Alg())
	}

	// Test with empty audience
	emptyAudienceConfig := config.AuthConfig{
		JWTSecretKey:    "test-secret",
		Issuer:          "test-issuer",
		Audience:        "",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "HS256",
	}

	emptyAudienceGen := NewJWTGenerator(emptyAudienceConfig)
	if len(emptyAudienceGen.audience) != 0 {
		t.Errorf("Expected empty audience, got %v", emptyAudienceGen.audience)
	}
}

func TestJWTGenerator_Generate(t *testing.T) {
	// Create a test generator with HMAC
	generator := NewJWTGenerator(config.AuthConfig{
		JWTSecretKey:    "test-secret",
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "HS256",
	})

	// Create a test user
	testUser := &user.User{
		ID:       "test-id",
		Username: "testuser",
		Roles:    []string{user.RoleAdmin},
		Active:   true,
	}

	// Generate a token
	token, err := generator.Generate(testUser)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be a non-empty string
	if token == "" {
		t.Error("Generated token is empty")
	}

	// Parse and validate the token
	claims, err := generator.Parse(token)
	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}

	// Check claims
	if claims.UserID != testUser.ID {
		t.Errorf("Expected UserID to be %q, got %q", testUser.ID, claims.UserID)
	}

	if claims.Username != testUser.Username {
		t.Errorf("Expected Username to be %q, got %q", testUser.Username, claims.Username)
	}

	if len(claims.Roles) != len(testUser.Roles) {
		t.Errorf("Expected Roles length to be %d, got %d", len(testUser.Roles), len(claims.Roles))
	}

	if claims.Issuer != "test-issuer" {
		t.Errorf("Expected Issuer to be %q, got %q", "test-issuer", claims.Issuer)
	}

	aud := []string(claims.Audience)
	if len(aud) != 1 || aud[0] != "test-audience" {
		t.Errorf("Expected Audience to be [%q], got %v", "test-audience", aud)
	}

	// Check that expiration is in the future (approximately 15 minutes)
	now := time.Now()
	expTime := claims.ExpiresAt.Time
	expectedExp := now.Add(15 * time.Minute)
	tolerance := 2 * time.Second

	diff := expTime.Sub(expectedExp)
	if diff < -tolerance || diff > tolerance {
		t.Errorf("Expiration time is not within expected range. Got %v, expected around %v (diff: %v)",
			expTime, expectedExp, diff)
	}
}

func TestJWTGenerator_GenerateWithExpiration(t *testing.T) {
	// Create a test generator
	generator := NewJWTGenerator(config.AuthConfig{
		JWTSecretKey:    "test-secret",
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute, // Default, should be overridden
		SigningMethod:   "HS256",
	})

	// Create a test user
	testUser := &user.User{
		ID:       "test-id",
		Username: "testuser",
		Roles:    []string{user.RoleAdmin},
		Active:   true,
	}

	// Generate a token with custom expiration
	customExpiration := 5 * time.Minute
	token, err := generator.GenerateWithExpiration(testUser, customExpiration)
	if err != nil {
		t.Fatalf("Failed to generate token with custom expiration: %v", err)
	}

	// Parse and validate the token
	claims, err := generator.Parse(token)
	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}

	// Check that expiration matches the custom value
	now := time.Now()
	expTime := claims.ExpiresAt.Time
	expectedExp := now.Add(customExpiration)
	tolerance := 2 * time.Second

	diff := expTime.Sub(expectedExp)
	if diff < -tolerance || diff > tolerance {
		t.Errorf("Expiration time is not within expected range. Got %v, expected around %v (diff: %v)",
			expTime, expectedExp, diff)
	}
}

func TestJWTGenerator_Parse(t *testing.T) {
	// Create a generator for testing
	secretKey := "test-secret"
	generator := NewJWTGenerator(config.AuthConfig{
		JWTSecretKey:    secretKey,
		Issuer:          "test-issuer",
		Audience:        "test-audience",
		TokenExpiration: 15 * time.Minute,
		SigningMethod:   "HS256",
	})

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

	// Test parsing a valid token
	claims, err := generator.Parse(validToken)
	if err != nil {
		t.Errorf("Failed to parse valid token: %v", err)
	}

	if claims == nil {
		t.Fatal("Claims are nil for valid token")
	}

	if claims.UserID != testUser.ID {
		t.Errorf("Expected UserID to be %q, got %q", testUser.ID, claims.UserID)
	}

	// Test parsing an invalid token (manually create a token with wrong signature)
	invalidToken := validToken + "invalid"
	_, err = generator.Parse(invalidToken)
	if err == nil {
		t.Error("Expected error when parsing invalid token, got nil")
	}

	// Test parsing a token with different algorithm
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

	// Here we use HS512, but our generator expects HS256
	wrongAlgToken := jwt.NewWithClaims(jwt.SigningMethodHS512, wrongAlgClaims)
	wrongAlgTokenString, err := wrongAlgToken.SignedString([]byte(secretKey))
	if err != nil {
		t.Fatalf("Failed to create token with wrong algorithm: %v", err)
	}

	_, err = generator.Parse(wrongAlgTokenString)
	if err == nil {
		t.Error("Expected error when parsing token with wrong algorithm, got nil")
	}

	// Test parsing an expired token
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

	expiredToken := jwt.NewWithClaims(generator.algorithm, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString([]byte(secretKey))
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = generator.Parse(expiredTokenString)
	if err == nil {
		t.Error("Expected error when parsing expired token, got nil")
	}
}
