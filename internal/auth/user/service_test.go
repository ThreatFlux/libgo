package user

import (
	"context"
	"testing"
	"time"

	"github.com/wroersma/libgo/internal/errors"
	"github.com/wroersma/libgo/internal/models/user"
	"github.com/wroersma/libgo/pkg/logger"
)

// mockLogger implements logger.Logger interface for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...logger.Field) {}
func (m *mockLogger) Info(msg string, fields ...logger.Field)  {}
func (m *mockLogger) Warn(msg string, fields ...logger.Field)  {}
func (m *mockLogger) Error(msg string, fields ...logger.Field) {}
func (m *mockLogger) Fatal(msg string, fields ...logger.Field) {}
func (m *mockLogger) WithFields(fields ...logger.Field) logger.Logger {
	return m
}
func (m *mockLogger) WithError(err error) logger.Logger {
	return m
}
func (m *mockLogger) Sync() error {
	return nil
}

func setupUserService() *UserService {
	return NewUserService(&mockLogger{})
}

func createTestUser(t *testing.T, service *UserService) *user.User {
	// Create a test user
	username := "testuser"
	password := "password123"
	email := "test@example.com"
	roles := []string{user.RoleAdmin}

	u, err := service.Create(context.Background(), username, password, email, roles)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return u
}

func TestUserService_Create(t *testing.T) {
	service := setupUserService()
	ctx := context.Background()

	// Create a user
	username := "testuser"
	password := "password123"
	email := "test@example.com"
	roles := []string{user.RoleAdmin}

	u, err := service.Create(ctx, username, password, email, roles)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// Check the returned user
	if u.Username != username {
		t.Errorf("Created user has wrong username: got %s, want %s", u.Username, username)
	}

	if u.Email != email {
		t.Errorf("Created user has wrong email: got %s, want %s", u.Email, email)
	}

	if !u.Active {
		t.Error("Created user should be active")
	}

	if len(u.Roles) != len(roles) || u.Roles[0] != roles[0] {
		t.Errorf("Created user has wrong roles: got %v, want %v", u.Roles, roles)
	}

	// Password should not be returned
	if u.Password != "" {
		t.Error("Password should not be returned in user object")
	}

	// Try to create a user with the same username
	_, err = service.Create(ctx, username, "different-password", "other@example.com", roles)
	if err == nil {
		t.Error("Create should fail with duplicate username")
	}
	if !errors.Is(err, errors.ErrAlreadyExists) {
		t.Errorf("Create with duplicate username should return ErrAlreadyExists, got: %v", err)
	}

	// Try to create a user with an invalid role
	_, err = service.Create(ctx, "newuser", password, email, []string{"invalid-role"})
	if err == nil {
		t.Error("Create should fail with invalid role")
	}
	if !errors.Is(err, errors.ErrInvalidParameter) {
		t.Errorf("Create with invalid role should return ErrInvalidParameter, got: %v", err)
	}
}

func TestUserService_Authenticate(t *testing.T) {
	service := setupUserService()
	ctx := context.Background()

	// Create a test user
	username := "testuser"
	password := "password123"
	email := "test@example.com"
	roles := []string{user.RoleAdmin}

	_, err := service.Create(ctx, username, password, email, roles)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Authenticate with correct credentials
	u, err := service.Authenticate(ctx, username, password)
	if err != nil {
		t.Errorf("Authenticate failed with correct credentials: %v", err)
	}
	if u == nil {
		t.Fatal("Authenticate returned nil user with correct credentials")
	}
	if u.Username != username {
		t.Errorf("Authenticated user has wrong username: got %s, want %s", u.Username, username)
	}

	// Authenticate with wrong password
	_, err = service.Authenticate(ctx, username, "wrong-password")
	if err == nil {
		t.Error("Authenticate should fail with wrong password")
	}

	// Authenticate with non-existent user
	_, err = service.Authenticate(ctx, "nonexistent", password)
	if err == nil {
		t.Error("Authenticate should fail with non-existent user")
	}
	if !errors.Is(err, errors.ErrInvalidCredentials) {
		t.Errorf("Authenticate with non-existent user should return ErrInvalidCredentials, got: %v", err)
	}

	// Make the user inactive and try to authenticate
	err = service.Update(ctx, u.ID, func(u *user.User) error {
		u.SetActive(false)
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	_, err = service.Authenticate(ctx, username, password)
	if err == nil {
		t.Error("Authenticate should fail with inactive user")
	}
	if !errors.Is(err, errors.ErrInvalidCredentials) {
		t.Errorf("Authenticate with inactive user should return ErrInvalidCredentials, got: %v", err)
	}
}

func TestUserService_GetByID(t *testing.T) {
	service := setupUserService()
	u := createTestUser(t, service)
	ctx := context.Background()

	// Get existing user by ID
	retrieved, err := service.GetByID(ctx, u.ID)
	if err != nil {
		t.Errorf("GetByID failed with existing user: %v", err)
	}
	if retrieved == nil || retrieved.ID != u.ID {
		t.Errorf("GetByID returned wrong user: got ID %s, want %s", 
			retrieved.ID, u.ID)
	}

	// Get non-existent user
	_, err = service.GetByID(ctx, "nonexistent-id")
	if err == nil {
		t.Error("GetByID should fail with non-existent user")
	}
	if !errors.Is(err, errors.ErrNotFound) {
		t.Errorf("GetByID with non-existent user should return ErrNotFound, got: %v", err)
	}
}

func TestUserService_GetByUsername(t *testing.T) {
	service := setupUserService()
	u := createTestUser(t, service)
	ctx := context.Background()

	// Get existing user by username
	retrieved, err := service.GetByUsername(ctx, u.Username)
	if err != nil {
		t.Errorf("GetByUsername failed with existing user: %v", err)
	}
	if retrieved == nil || retrieved.Username != u.Username {
		t.Errorf("GetByUsername returned wrong user: got username %s, want %s", 
			retrieved.Username, u.Username)
	}

	// Get non-existent user
	_, err = service.GetByUsername(ctx, "nonexistent-user")
	if err == nil {
		t.Error("GetByUsername should fail with non-existent user")
	}
	if !errors.Is(err, errors.ErrNotFound) {
		t.Errorf("GetByUsername with non-existent user should return ErrNotFound, got: %v", err)
	}
}

func TestUserService_Update(t *testing.T) {
	service := setupUserService()
	u := createTestUser(t, service)
	ctx := context.Background()

	// Update user email
	newEmail := "updated@example.com"
	originalUpdatedAt := u.UpdatedAt

	updated, err := service.Update(ctx, u.ID, func(user *user.User) error {
		user.SetEmail(newEmail)
		return nil
	})
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	if updated.Email != newEmail {
		t.Errorf("Updated user has wrong email: got %s, want %s", updated.Email, newEmail)
	}
	if !updated.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Updated user should have a newer UpdatedAt timestamp")
	}

	// Update user with a duplicate username
	// First create another user
	_, err = service.Create(ctx, "seconduser", "password", "second@example.com", []string{user.RoleViewer})
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}

	// Try to update the original user with the second user's username
	_, err = service.Update(ctx, u.ID, func(user *user.User) error {
		user.Username = "seconduser"
		return nil
	})
	if err == nil {
		t.Error("Update should fail with duplicate username")
	}
	if !errors.Is(err, errors.ErrAlreadyExists) {
		t.Errorf("Update with duplicate username should return ErrAlreadyExists, got: %v", err)
	}

	// Update non-existent user
	_, err = service.Update(ctx, "nonexistent-id", func(user *user.User) error {
		return nil
	})
	if err == nil {
		t.Error("Update should fail with non-existent user")
	}
	if !errors.Is(err, errors.ErrNotFound) {
		t.Errorf("Update with non-existent user should return ErrNotFound, got: %v", err)
	}

	// Update user with an error from the update function
	_, err = service.Update(ctx, u.ID, func(user *user.User) error {
		return errors.New("update function error")
	})
	if err == nil {
		t.Error("Update should fail when update function returns an error")
	}
}

func TestUserService_Delete(t *testing.T) {
	service := setupUserService()
	u := createTestUser(t, service)
	ctx := context.Background()

	// Delete existing user
	err := service.Delete(ctx, u.ID)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify user is gone
	_, err = service.GetByID(ctx, u.ID)
	if err == nil {
		t.Error("User should no longer exist after deletion")
	}
	if !errors.Is(err, errors.ErrNotFound) {
		t.Errorf("GetByID after deletion should return ErrNotFound, got: %v", err)
	}

	// Delete non-existent user
	err = service.Delete(ctx, "nonexistent-id")
	if err == nil {
		t.Error("Delete should fail with non-existent user")
	}
	if !errors.Is(err, errors.ErrNotFound) {
		t.Errorf("Delete with non-existent user should return ErrNotFound, got: %v", err)
	}
}

func TestUserService_HasPermission(t *testing.T) {
	service := setupUserService()
	ctx := context.Background()

	// Create users with different roles
	adminUser, err := service.Create(ctx, "admin", "password", "admin@example.com", []string{user.RoleAdmin})
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	operatorUser, err := service.Create(ctx, "operator", "password", "operator@example.com", []string{user.RoleOperator})
	if err != nil {
		t.Fatalf("Failed to create operator user: %v", err)
	}

	viewerUser, err := service.Create(ctx, "viewer", "password", "viewer@example.com", []string{user.RoleViewer})
	if err != nil {
		t.Fatalf("Failed to create viewer user: %v", err)
	}

	// Test permissions for admin
	hasCreate, err := service.HasPermission(ctx, adminUser.ID, user.PermCreate)
	if err != nil {
		t.Errorf("HasPermission failed for admin: %v", err)
	}
	if !hasCreate {
		t.Error("Admin should have create permission")
	}

	// Test permissions for operator
	hasRead, err := service.HasPermission(ctx, operatorUser.ID, user.PermRead)
	if err != nil {
		t.Errorf("HasPermission failed for operator: %v", err)
	}
	if !hasRead {
		t.Error("Operator should have read permission")
	}

	hasCreate, err = service.HasPermission(ctx, operatorUser.ID, user.PermCreate)
	if err != nil {
		t.Errorf("HasPermission failed for operator: %v", err)
	}
	if hasCreate {
		t.Error("Operator should not have create permission")
	}

	// Test permissions for viewer
	hasRead, err = service.HasPermission(ctx, viewerUser.ID, user.PermRead)
	if err != nil {
		t.Errorf("HasPermission failed for viewer: %v", err)
	}
	if !hasRead {
		t.Error("Viewer should have read permission")
	}

	hasCreate, err = service.HasPermission(ctx, viewerUser.ID, user.PermCreate)
	if err != nil {
		t.Errorf("HasPermission failed for viewer: %v", err)
	}
	if hasCreate {
		t.Error("Viewer should not have create permission")
	}

	// Test non-existent user
	_, err = service.HasPermission(ctx, "nonexistent-id", user.PermRead)
	if err == nil {
		t.Error("HasPermission should fail with non-existent user")
	}
	if !errors.Is(err, errors.ErrNotFound) {
		t.Errorf("HasPermission with non-existent user should return ErrNotFound, got: %v", err)
	}
}

func TestUserService_List(t *testing.T) {
	service := setupUserService()
	ctx := context.Background()

	// Create some users
	_, err := service.Create(ctx, "user1", "password", "user1@example.com", []string{user.RoleAdmin})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	_, err = service.Create(ctx, "user2", "password", "user2@example.com", []string{user.RoleOperator})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// List users
	users, err := service.List(ctx)
	if err != nil {
		t.Errorf("List failed: %v", err)
	}

	// Check number of users
	if len(users) != 2 {
		t.Errorf("List returned wrong number of users: got %d, want %d", len(users), 2)
	}

	// Check that passwords are not included
	for _, u := range users {
		if u.Password != "" {
			t.Errorf("User from List contains password: %s", u.ID)
		}
	}
}

func TestUserService_SetPassword(t *testing.T) {
	service := setupUserService()
	u := createTestUser(t, service)
	ctx := context.Background()

	// Change password
	newPassword := "new-password"
	err := service.SetPassword(ctx, u.ID, newPassword)
	if err != nil {
		t.Errorf("SetPassword failed: %v", err)
	}

	// Authenticate with new password
	authenticated, err := service.Authenticate(ctx, u.Username, newPassword)
	if err != nil {
		t.Errorf("Authenticate failed after password change: %v", err)
	}
	if authenticated == nil {
		t.Fatal("Authenticate returned nil user after password change")
	}

	// Old password should no longer work
	_, err = service.Authenticate(ctx, u.Username, "password123")
	if err == nil {
		t.Error("Authenticate should fail with old password after change")
	}
}

func TestUserService_LoadUser(t *testing.T) {
	service := setupUserService()
	ctx := context.Background()

	// Create a user to load
	now := time.Now().UTC()
	loadedUser := &user.User{
		ID:        "custom-id",
		Username:  "loadeduser",
		Password:  "hashed-password",
		Email:     "loaded@example.com",
		Roles:     []string{user.RoleAdmin},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Load the user
	err := service.LoadUser(loadedUser)
	if err != nil {
		t.Errorf("LoadUser failed: %v", err)
	}

	// Verify the user was loaded
	retrieved, err := service.GetByID(ctx, loadedUser.ID)
	if err != nil {
		t.Errorf("Failed to retrieve loaded user: %v", err)
	}
	if retrieved.Username != loadedUser.Username {
		t.Errorf("Loaded user has wrong username: got %s, want %s", 
			retrieved.Username, loadedUser.Username)
	}

	// Try to load a user with duplicate ID
	duplicateID := &user.User{
		ID:        loadedUser.ID,
		Username:  "different",
		Password:  "password",
		Email:     "different@example.com",
		Roles:     []string{user.RoleViewer},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = service.LoadUser(duplicateID)
	if err == nil {
		t.Error("LoadUser should fail with duplicate ID")
	}
	if !errors.Is(err, errors.ErrAlreadyExists) {
		t.Errorf("LoadUser with duplicate ID should return ErrAlreadyExists, got: %v", err)
	}

	// Try to load a user with duplicate username
	duplicateUsername := &user.User{
		ID:        "different-id",
		Username:  loadedUser.Username,
		Password:  "password",
		Email:     "different@example.com",
		Roles:     []string{user.RoleViewer},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = service.LoadUser(duplicateUsername)
	if err == nil {
		t.Error("LoadUser should fail with duplicate username")
	}
	if !errors.Is(err, errors.ErrAlreadyExists) {
		t.Errorf("LoadUser with duplicate username should return ErrAlreadyExists, got: %v", err)
	}
}
