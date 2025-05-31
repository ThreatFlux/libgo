package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/threatflux/libgo/internal/config"
)

// Error definitions.
var (
	ErrTokenExpired = errors.New("token has expired")
	ErrInvalidToken = errors.New("invalid token")
)

// Validator defines interface for JWT token validation.
type Validator interface {
	// Validate validates a JWT token.
	Validate(tokenString string) (*Claims, error)

	// ValidateWithClaims validates a token and populates the claims.
	ValidateWithClaims(tokenString string, claims jwt.Claims) error
}

// JWTValidator implements Validator.
type JWTValidator struct {
	secretKey     []byte
	publicKey     *rsa.PublicKey
	algorithm     jwt.SigningMethod
	issuer        string
	audience      []string
	signingMethod string
}

// NewJWTValidator creates a new JWTValidator.
func NewJWTValidator(config config.AuthConfig) *JWTValidator {
	var algorithm jwt.SigningMethod
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

	return &JWTValidator{
		secretKey:     []byte(config.JWTSecretKey),
		publicKey:     publicKey, // Will be nil if not using RSA
		algorithm:     algorithm,
		issuer:        config.Issuer,
		audience:      audience,
		signingMethod: config.SigningMethod,
	}
}

// Validate implements Validator.Validate.
func (v *JWTValidator) Validate(tokenString string) (*Claims, error) {
	claims := &Claims{}
	err := v.ValidateWithClaims(tokenString, claims)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// ValidateWithClaims implements Validator.ValidateWithClaims.
func (v *JWTValidator) ValidateWithClaims(tokenString string, claims jwt.Claims) error {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if token.Method.Alg() != v.algorithm.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Use the appropriate key based on signing method
		if v.signingMethod == SigningMethodRS256 || v.signingMethod == SigningMethodRS384 || v.signingMethod == SigningMethodRS512 {
			if v.publicKey == nil {
				return nil, fmt.Errorf("RSA public key not set")
			}
			return v.publicKey, nil
		}
		return v.secretKey, nil
	})

	// Check for parsing errors
	if err != nil {
		// Check if token is expired
		if errors.Is(err, jwt.ErrTokenExpired) {
			return fmt.Errorf("%w: %v", ErrTokenExpired, err)
		}
		// Handle other validation errors
		return fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Check if the token is valid
	if !token.Valid {
		return fmt.Errorf("%w: token is invalid", ErrInvalidToken)
	}

	return nil
}
