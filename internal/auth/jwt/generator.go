package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/internal/models/user"
)

// Signing method constants.
const (
	SigningMethodHS256 = "HS256"
	SigningMethodHS384 = "HS384"
	SigningMethodHS512 = "HS512"
	SigningMethodRS256 = "RS256"
	SigningMethodRS384 = "RS384"
	SigningMethodRS512 = "RS512"
	SigningMethodES256 = "ES256"
	SigningMethodES384 = "ES384"
	SigningMethodES512 = "ES512"
)

// Generator defines interface for JWT token generation.
type Generator interface {
	// Generate generates a JWT token for a user.
	Generate(user *user.User) (string, error)

	// GenerateWithExpiration generates a JWT token with specific expiration.
	GenerateWithExpiration(user *user.User, expiration time.Duration) (string, error)

	// Parse parses and validates a JWT token.
	Parse(tokenString string) (*Claims, error)
}

// JWTGenerator implements Generator.
type JWTGenerator struct {
	// Slice fields (24 bytes each)
	secretKey []byte
	audience  []string
	// Interface fields (16 bytes)
	algorithm jwt.SigningMethod
	// String fields (16 bytes each)
	issuer        string
	signingMethod string
	// Duration (8 bytes)
	expiresIn time.Duration
	// Pointer fields (8 bytes each)
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewJWTGenerator creates a new JWTGenerator.
func NewJWTGenerator(config config.AuthConfig) *JWTGenerator {
	var algorithm jwt.SigningMethod
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey

	// Choose signing method based on configuration
	switch config.SigningMethod {
	case SigningMethodHS256:
		algorithm = jwt.SigningMethodHS256
	case SigningMethodHS384:
		algorithm = jwt.SigningMethodHS384
	case SigningMethodHS512:
		algorithm = jwt.SigningMethodHS512
	case SigningMethodRS256:
		algorithm = jwt.SigningMethodRS256
	case SigningMethodRS384:
		algorithm = jwt.SigningMethodRS384
	case SigningMethodRS512:
		algorithm = jwt.SigningMethodRS512
	case SigningMethodES256:
		algorithm = jwt.SigningMethodES256
	case SigningMethodES384:
		algorithm = jwt.SigningMethodES384
	case SigningMethodES512:
		algorithm = jwt.SigningMethodES512
	default:
		// Default to HS256 for safety
		algorithm = jwt.SigningMethodHS256
	}

	// Parse audience from comma-separated string if needed
	audience := []string{config.Audience}
	if config.Audience == "" {
		audience = []string{}
	}

	return &JWTGenerator{
		secretKey:     []byte(config.JWTSecretKey),
		privateKey:    privateKey, // Will be nil if not using RSA
		publicKey:     publicKey,  // Will be nil if not using RSA
		algorithm:     algorithm,
		issuer:        config.Issuer,
		audience:      audience,
		expiresIn:     config.TokenExpiration,
		signingMethod: config.SigningMethod,
	}
}

// Generate implements Generator.Generate.
func (g *JWTGenerator) Generate(user *user.User) (string, error) {
	return g.GenerateWithExpiration(user, g.expiresIn)
}

// GenerateWithExpiration implements Generator.GenerateWithExpiration.
func (g *JWTGenerator) GenerateWithExpiration(user *user.User, expiration time.Duration) (string, error) {
	now := time.Now()
	expirationTime := now.Add(expiration)

	// Create the registered claims
	registeredClaims := jwt.RegisteredClaims{
		Issuer:    g.issuer,
		Subject:   user.ID,
		Audience:  jwt.ClaimStrings(g.audience),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	// Create the custom claims
	claims := NewClaims(user, registeredClaims)

	// Create the token
	token := jwt.NewWithClaims(g.algorithm, claims)

	// Sign the token based on the algorithm
	var tokenString string
	var err error

	if g.signingMethod == SigningMethodRS256 || g.signingMethod == SigningMethodRS384 || g.signingMethod == SigningMethodRS512 {
		if g.privateKey == nil {
			return "", fmt.Errorf("RSA private key not set")
		}
		tokenString, err = token.SignedString(g.privateKey)
	} else {
		tokenString, err = token.SignedString(g.secretKey)
	}

	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return tokenString, nil
}

// Parse implements Generator.Parse.
func (g *JWTGenerator) Parse(tokenString string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if token.Method.Alg() != g.algorithm.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Use the appropriate key based on signing method
		if g.signingMethod == SigningMethodRS256 || g.signingMethod == SigningMethodRS384 || g.signingMethod == SigningMethodRS512 {
			if g.publicKey == nil {
				return nil, fmt.Errorf("RSA public key not set")
			}
			return g.publicKey, nil
		}
		return g.secretKey, nil
	})

	// Check for parsing errors
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	// Extract the claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
