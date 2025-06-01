package user

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	username := "testuser"
	password := "hashedpassword"
	email := "test@example.com"
	roles := []string{RoleAdmin}

	user := NewUser(username, password, email, roles)

	// Check basic properties
	if user.Username != username {
		t.Errorf("Expected username to be %s, got %s", username, user.Username)
	}

	if user.Password != password {
		t.Errorf("Expected password to be %s, got %s", password, user.Password)
	}

	if user.Email != email {
		t.Errorf("Expected email to be %s, got %s", email, user.Email)
	}

	if len(user.Roles) != len(roles) {
		t.Errorf("Expected roles length to be %d, got %d", len(roles), len(user.Roles))
	}

	if !user.Active {
		t.Error("Expected user to be active")
	}

	// Check UUID format
	if len(user.ID) != 36 {
		t.Errorf("Expected ID to be a UUID (length 36), got length %d", len(user.ID))
	}

	// Check timestamps
	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestUser_HasRole(t *testing.T) {
	user := &User{
		Roles: []string{RoleAdmin, RoleOperator},
	}

	tests := []struct {
		name string
		role string
		want bool
	}{
		{"Has admin role", RoleAdmin, true},
		{"Has operator role", RoleOperator, true},
		{"Does not have viewer role", RoleViewer, false},
		{"Does not have non-existent role", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := user.HasRole(tt.role); got != tt.want {
				t.Errorf("User.HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_HasAnyRole(t *testing.T) {
	user := &User{
		Roles: []string{RoleAdmin},
	}

	tests := []struct {
		name  string
		roles []string
		want  bool
	}{
		{"Has one of the roles", []string{RoleViewer, RoleAdmin}, true},
		{"Has the only role", []string{RoleAdmin}, true},
		{"Doesn't have any of the roles", []string{RoleViewer, RoleOperator}, false},
		{"Empty roles list", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := user.HasAnyRole(tt.roles...); got != tt.want {
				t.Errorf("User.HasAnyRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_HasAllRoles(t *testing.T) {
	user := &User{
		Roles: []string{RoleAdmin, RoleOperator},
	}

	tests := []struct {
		name  string
		roles []string
		want  bool
	}{
		{"Has all roles", []string{RoleAdmin, RoleOperator}, true},
		{"Has only one of two", []string{RoleAdmin, RoleViewer}, false},
		{"Has none of the roles", []string{RoleViewer}, false},
		{"Empty roles list", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := user.HasAllRoles(tt.roles...); got != tt.want {
				t.Errorf("User.HasAllRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_Clone(t *testing.T) {
	original := &User{
		ID:        "test-id",
		Username:  "testuser",
		Password:  "hashedpassword",
		Email:     "test@example.com",
		Roles:     []string{RoleAdmin, RoleOperator},
		Active:    true,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	clone := original.Clone()

	// Check that the clone has the same values except password
	if clone.ID != original.ID {
		t.Errorf("Clone.ID = %v, want %v", clone.ID, original.ID)
	}

	if clone.Username != original.Username {
		t.Errorf("Clone.Username = %v, want %v", clone.Username, original.Username)
	}

	if clone.Password != "" {
		t.Errorf("Clone.Password should be empty, got %v", clone.Password)
	}

	if clone.Email != original.Email {
		t.Errorf("Clone.Email = %v, want %v", clone.Email, original.Email)
	}

	if !clone.Active {
		t.Error("Clone should be active")
	}

	if !clone.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("Clone.CreatedAt = %v, want %v", clone.CreatedAt, original.CreatedAt)
	}

	if !clone.UpdatedAt.Equal(original.UpdatedAt) {
		t.Errorf("Clone.UpdatedAt = %v, want %v", clone.UpdatedAt, original.UpdatedAt)
	}

	// Check that roles are deep copied
	if len(clone.Roles) != len(original.Roles) {
		t.Errorf("Clone.Roles length = %v, want %v", len(clone.Roles), len(original.Roles))
	}

	// Verify it's a deep copy by modifying the clone
	clone.Roles[0] = "modified"
	if original.Roles[0] == "modified" {
		t.Error("Modifying clone.Roles should not affect original.Roles")
	}
}

func TestUser_AddRole(t *testing.T) {
	user := &User{
		Roles:     []string{RoleAdmin},
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	// Add a new role
	originalTime := user.UpdatedAt
	user.AddRole(RoleOperator)

	if len(user.Roles) != 2 {
		t.Errorf("Expected roles length to be 2, got %d", len(user.Roles))
	}

	if !user.HasRole(RoleOperator) {
		t.Errorf("Expected user to have role %s", RoleOperator)
	}

	if !user.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should have been updated")
	}

	// Add a role that the user already has
	user.UpdatedAt = time.Now().Add(-1 * time.Hour)
	originalTime = user.UpdatedAt
	originalRolesLen := len(user.Roles)

	user.AddRole(RoleAdmin)

	if len(user.Roles) != originalRolesLen {
		t.Errorf("Adding an existing role should not change roles length, got %d, expected %d",
			len(user.Roles), originalRolesLen)
	}

	// No change should occur to UpdatedAt
	if !user.UpdatedAt.Equal(originalTime) {
		t.Error("UpdatedAt should not have been updated when adding an existing role")
	}
}

func TestUser_RemoveRole(t *testing.T) {
	user := &User{
		Roles:     []string{RoleAdmin, RoleOperator, RoleViewer},
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	// Remove an existing role
	originalTime := user.UpdatedAt
	user.RemoveRole(RoleOperator)

	if len(user.Roles) != 2 {
		t.Errorf("Expected roles length to be 2, got %d", len(user.Roles))
	}

	if user.HasRole(RoleOperator) {
		t.Errorf("Expected user to not have role %s anymore", RoleOperator)
	}

	if !user.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should have been updated")
	}

	// Remove a role that the user doesn't have
	user.UpdatedAt = time.Now().Add(-1 * time.Hour)
	originalTime = user.UpdatedAt
	originalRolesLen := len(user.Roles)

	user.RemoveRole("nonexistent")

	if len(user.Roles) != originalRolesLen {
		t.Errorf("Removing a non-existent role should not change roles length")
	}

	// No change should occur to UpdatedAt
	if !user.UpdatedAt.Equal(originalTime) {
		t.Error("UpdatedAt should not have been updated when removing a non-existent role")
	}
}

func TestUser_SetActive(t *testing.T) {
	user := &User{
		Active:    true,
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	// Change active status
	originalTime := user.UpdatedAt
	user.SetActive(false)

	if user.Active {
		t.Error("Expected user to be inactive")
	}

	if !user.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should have been updated")
	}

	// Set active to the same value
	user.UpdatedAt = time.Now().Add(-1 * time.Hour)
	originalTime = user.UpdatedAt

	user.SetActive(false)

	// No change should occur to UpdatedAt
	if !user.UpdatedAt.Equal(originalTime) {
		t.Error("UpdatedAt should not have been updated when setting active to the same value")
	}
}

func TestUser_SetPassword(t *testing.T) {
	user := &User{
		Password:  "oldpassword",
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	// Change password
	originalTime := user.UpdatedAt
	user.SetPassword("newpassword")

	if user.Password != "newpassword" {
		t.Errorf("Expected password to be 'newpassword', got %s", user.Password)
	}

	if !user.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should have been updated")
	}
}

func TestUser_SetEmail(t *testing.T) {
	user := &User{
		Email:     "old@example.com",
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	// Change email
	originalTime := user.UpdatedAt
	user.SetEmail("new@example.com")

	if user.Email != "new@example.com" {
		t.Errorf("Expected email to be 'new@example.com', got %s", user.Email)
	}

	if !user.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should have been updated")
	}

	// Set email to the same value
	user.UpdatedAt = time.Now().Add(-1 * time.Hour)
	originalTime = user.UpdatedAt

	user.SetEmail("new@example.com")

	// No change should occur to UpdatedAt
	if !user.UpdatedAt.Equal(originalTime) {
		t.Error("UpdatedAt should not have been updated when setting email to the same value")
	}
}
