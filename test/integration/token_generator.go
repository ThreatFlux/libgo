package integration

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateTestToken creates a valid JWT token for testing purposes using direct JWT library methods
func GenerateTestToken() (string, error) {
	// Use the exact same secret key as in test-config.yaml
	secretKey := []byte("test-secret-key-for-jwt-token-generation")

	// Use a fixed UUID for the admin user, ensuring it matches exactly what's in the database
	// SQLite might be stripping a digit, so we use the format as it appears in the DB
	adminID := "1111111-2222-3333-4444-555555555555"

	// Get current time and expiry
	now := time.Now()
	expiry := now.Add(24 * time.Hour)

	// Create token directly with a map of claims - this avoids any potential mismatch with
	// the structure expected by the server validation code
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// Standard JWT claims
		"sub": adminID,                   // Subject (user ID)
		"exp": expiry.Unix(),             // Expiration time
		"iat": now.Unix(),                // Issued at time
		"nbf": now.Unix(),                // Not before time

		// Custom claims that match what the server expects
		"userId":   adminID,              // Must match the subject
		"username": "admin",              // Must match username in test-config.yaml
		"roles":    []string{"admin"},    // Must match roles in test-config.yaml
	})

	// Sign and return the token
	return token.SignedString(secretKey)
}
