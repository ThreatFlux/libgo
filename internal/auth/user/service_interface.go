package user

import (
	"context"

	"github.com/threatflux/libgo/internal/models/user"
)

// DefaultUserConfig represents configuration for a default user.
type DefaultUserConfig struct {
	Username string
	Password string
	Email    string
	Roles    []string
}

// Service defines interface for user management.
type Service interface {
	// Authenticate authenticates a user
	Authenticate(ctx context.Context, username, password string) (*user.User, error)

	// GetByID gets a user by ID
	GetByID(ctx context.Context, id string) (*user.User, error)

	// GetByUsername gets a user by username
	GetByUsername(ctx context.Context, username string) (*user.User, error)

	// HasPermission checks if a user has a permission
	HasPermission(ctx context.Context, userID string, permission string) (bool, error)

	// Create creates a new user
	Create(ctx context.Context, username, password, email string, roles []string) (*user.User, error)

	// Update updates an existing user
	Update(ctx context.Context, userID string, updateFunc func(*user.User) error) (*user.User, error)

	// Delete deletes a user
	Delete(ctx context.Context, userID string) error

	// List lists all users
	List(ctx context.Context) ([]*user.User, error)

	// LoadUser loads an existing user into the service
	LoadUser(u *user.User) error

	// InitializeDefaultUsers creates default users from configuration
	InitializeDefaultUsers(ctx context.Context, defaultUsers []DefaultUserConfig) error
}
