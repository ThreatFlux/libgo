package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/threatflux/libgo/internal/models/user"
)

// Claims represents custom JWT claims.
type Claims struct {
	jwt.RegisteredClaims
	UserID   string   `json:"userId"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

// NewClaims creates a new Claims from user information and registered claims.
func NewClaims(userModel *user.User, registeredClaims jwt.RegisteredClaims) *Claims {
	return &Claims{
		RegisteredClaims: registeredClaims,
		UserID:           userModel.ID,
		Username:         userModel.Username,
		Roles:            userModel.Roles,
	}
}

// Valid implements jwt.Claims interface for the Claims type.
func (c *Claims) Valid() error {
	now := time.Now()

	// Check expiration
	if c.ExpiresAt != nil && c.ExpiresAt.Before(now) {
		return fmt.Errorf("token has expired")
	}

	// Check not before
	if c.NotBefore != nil && c.NotBefore.After(now) {
		return fmt.Errorf("token used before valid")
	}

	// Check issued at
	if c.IssuedAt != nil && c.IssuedAt.After(now.Add(time.Minute)) {
		return fmt.Errorf("token used before issued")
	}

	// Validate that required fields are present
	if c.UserID == "" {
		return fmt.Errorf("userId is required")
	}

	if c.Username == "" {
		return fmt.Errorf("username is required")
	}

	// Validate that at least one role is present
	if len(c.Roles) == 0 {
		return fmt.Errorf("at least one role is required")
	}

	// Validate that all roles are valid
	for _, role := range c.Roles {
		if !user.IsValidRole(role) {
			return fmt.Errorf("invalid role: %s", role)
		}
	}

	return nil
}

// HasPermission checks if the claims has the specified permission.
func (c *Claims) HasPermission(permission string) bool {
	return user.UserHasPermission(c.Roles, permission)
}

// HasRole checks if the claims has the specified role.
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// ToUser converts claims to a user model (without password).
func (c *Claims) ToUser() *user.User {
	return &user.User{
		ID:       c.UserID,
		Username: c.Username,
		Roles:    c.Roles,
		Active:   true, // Assuming active since we're creating from valid claims
	}
}
