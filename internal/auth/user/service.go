package user

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/threatflux/libgo/internal/errors"
	"github.com/threatflux/libgo/internal/models/user"
	"github.com/threatflux/libgo/pkg/logger"
)

// UserService implements Service interface for user management
type UserService struct {
	users  map[string]*user.User // Map of user ID to user
	byName map[string]string     // Map of username to user ID
	mu     sync.RWMutex
	logger logger.Logger
}

// NewUserService creates a new UserService
func NewUserService(logger logger.Logger) *UserService {
	return &UserService{
		users:  make(map[string]*user.User),
		byName: make(map[string]string),
		logger: logger,
	}
}

// getUserByUsernameInternal gets a user by username with password (for internal use)
func (s *UserService) getUserByUsernameInternal(username string) (*user.User, error) {
	id, ok := s.byName[username]
	if !ok {
		return nil, errors.WrapWithCode(errors.New("user not found"), errors.ErrNotFound, "getting user by username")
	}

	u, ok := s.users[id]
	if !ok {
		// This should never happen, but just in case
		return nil, errors.WrapWithCode(errors.New("user mapping inconsistent"), errors.ErrNotFound, "getting user by username")
	}

	// Return the actual user (with password) for internal operations
	return u, nil
}

// Authenticate implements Service.Authenticate
func (s *UserService) Authenticate(ctx context.Context, username, password string) (*user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Get user by username (internal method that includes password)
	u, err := s.getUserByUsernameInternal(username)
	if err != nil {
		// Don't expose whether a user exists or not
		return nil, errors.WrapWithCode(err, errors.ErrInvalidCredentials, "invalid username or password")
	}

	// Check if user is active
	if !u.Active {
		return nil, errors.WrapWithCode(errors.New("user account is inactive"), errors.ErrInvalidCredentials, "user account is inactive")
	}

	// Verify the password
	if !VerifyPassword(password, u.Password) {
		return nil, errors.WrapWithCode(errors.New("invalid username or password"), errors.ErrInvalidCredentials, "invalid username or password")
	}

	// Create a copy of the user without the password
	return u.Clone(), nil
}

// GetByID implements Service.GetByID
func (s *UserService) GetByID(ctx context.Context, id string) (*user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return nil, errors.WrapWithCode(errors.New("user not found"), errors.ErrNotFound, "getting user by ID")
	}

	// Return a copy to prevent modification
	return u.Clone(), nil
}

// GetByUsername implements Service.GetByUsername
func (s *UserService) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.byName[username]
	if !ok {
		return nil, errors.WrapWithCode(errors.New("user not found"), errors.ErrNotFound, "getting user by username")
	}

	u, ok := s.users[id]
	if !ok {
		// This should never happen, but just in case
		return nil, errors.WrapWithCode(errors.New("user mapping inconsistent"), errors.ErrNotFound, "getting user by username")
	}

	// Return a copy to prevent modification
	return u.Clone(), nil
}

// HasPermission implements Service.HasPermission
func (s *UserService) HasPermission(ctx context.Context, userID string, permission string) (bool, error) {
	u, err := s.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if the user has the permission
	return user.UserHasPermission(u.Roles, permission), nil
}

// Create implements Service.Create
func (s *UserService) Create(ctx context.Context, username, password, email string, roles []string) (*user.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if username already exists
	if _, ok := s.byName[username]; ok {
		return nil, errors.WrapWithCode(errors.New("username already exists"), errors.ErrAlreadyExists, "creating user")
	}

	// Validate roles
	for _, role := range roles {
		if !user.IsValidRole(role) {
			return nil, errors.WrapWithCode(fmt.Errorf("invalid role: %s", role), errors.ErrInvalidParameter, "creating user")
		}
	}

	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, errors.Wrap(err, "hashing password")
	}

	// For integration testing, if this is the admin user, use a fixed UUID
	var newUser *user.User
	if username == "admin" {
		// Use a fixed UUID for the admin user for predictable testing
		fixedAdminUUID := "11111111-2222-3333-4444-555555555555"
		newUser = user.NewUserWithID(fixedAdminUUID, username, hashedPassword, email, roles)
	} else {
		// Create a new user with a random UUID
		newUser = user.NewUser(username, hashedPassword, email, roles)
	}

	// Store the user (with password)
	s.users[newUser.ID] = newUser
	s.byName[newUser.Username] = newUser.ID

	// Return a copy without the password
	return newUser.Clone(), nil
}

// Update implements Service.Update
func (s *UserService) Update(ctx context.Context, userID string, updateFunc func(*user.User) error) (*user.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the user
	u, ok := s.users[userID]
	if !ok {
		return nil, errors.WrapWithCode(errors.New("user not found"), errors.ErrNotFound, "updating user")
	}

	// Create a copy for updating
	copy := u.Clone()
	copy.Password = u.Password // Restore password as it's not included in Clone

	// Apply the update function
	if err := updateFunc(copy); err != nil {
		return nil, errors.Wrap(err, "updating user")
	}

	// Update username mapping if it changed
	if copy.Username != u.Username {
		// Check if the new username is already taken by another user
		if existingID, ok := s.byName[copy.Username]; ok && existingID != userID {
			return nil, errors.WrapWithCode(errors.New("username already exists"), errors.ErrAlreadyExists, "updating user")
		}

		// Remove old mapping and add new one
		delete(s.byName, u.Username)
		s.byName[copy.Username] = userID
	}

	// Update timestamp
	copy.UpdatedAt = time.Now().UTC()

	// Save the updated user
	s.users[userID] = copy

	// Return a copy without the password
	return copy.Clone(), nil
}

// Delete implements Service.Delete
func (s *UserService) Delete(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the user
	u, ok := s.users[userID]
	if !ok {
		return errors.WrapWithCode(errors.New("user not found"), errors.ErrNotFound, "deleting user")
	}

	// Remove from maps
	delete(s.users, userID)
	delete(s.byName, u.Username)

	return nil
}

// List implements Service.List
func (s *UserService) List(ctx context.Context) ([]*user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a slice to hold all users
	users := make([]*user.User, 0, len(s.users))
	for _, u := range s.users {
		// Add a copy of each user without password
		users = append(users, u.Clone())
	}

	return users, nil
}

// SetPassword updates a user's password
func (s *UserService) SetPassword(ctx context.Context, userID string, password string) (*user.User, error) {
	return s.Update(ctx, userID, func(u *user.User) error {
		// Hash the new password
		hashedPassword, err := HashPassword(password)
		if err != nil {
			return errors.Wrap(err, "hashing password")
		}

		// Set the new password
		u.SetPassword(hashedPassword)
		return nil
	})
}

// LoadUser loads an existing user into the service
// This is primarily for testing or initial data loading
func (s *UserService) LoadUser(u *user.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user ID already exists
	if _, ok := s.users[u.ID]; ok {
		return errors.WrapWithCode(errors.New("user ID already exists"), errors.ErrAlreadyExists, "loading user")
	}

	// Check if username already exists
	if _, ok := s.byName[u.Username]; ok {
		return errors.WrapWithCode(errors.New("username already exists"), errors.ErrAlreadyExists, "loading user")
	}

	// Store the user
	s.users[u.ID] = u
	s.byName[u.Username] = u.ID

	return nil
}

// InitializeDefaultUsers creates default users from configuration
func (s *UserService) InitializeDefaultUsers(ctx context.Context, defaultUsers []struct {
	Username string
	Password string
	Email    string
	Roles    []string
}) error {
	s.logger.Info("Initializing default users", logger.Int("count", len(defaultUsers)))

	for _, defaultUser := range defaultUsers {
		// Check if user already exists
		_, err := s.GetByUsername(ctx, defaultUser.Username)
		if err == nil {
			s.logger.Info("Default user already exists, skipping",
				logger.String("username", defaultUser.Username))
			continue
		}

		// Only proceed if the error is "user not found"
		if !errors.Is(err, errors.ErrNotFound) {
			return errors.Wrap(err, "checking if default user exists")
		}

		// Create the user
		s.logger.Info("Creating default user",
			logger.String("username", defaultUser.Username),
			logger.String("email", defaultUser.Email),
		)

		_, err = s.Create(ctx, defaultUser.Username, defaultUser.Password, defaultUser.Email, defaultUser.Roles)
		if err != nil {
			return errors.Wrap(err, "creating default user")
		}
	}

	return nil
}
