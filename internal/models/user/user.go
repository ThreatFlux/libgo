package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system.
type User struct {
	Roles     []string  `json:"roles"`
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Hashed password, not exposed in JSON.
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Active    bool      `json:"active"`
}

// NewUser creates a new user with the given username, password, and roles.
func NewUser(username, password, email string, roles []string) *User {
	now := time.Now().UTC()
	return &User{
		ID:        uuid.New().String(),
		Username:  username,
		Password:  password, // Should be already hashed before calling this function.
		Email:     email,
		Roles:     roles,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewUserWithID creates a user with a specified ID (for testing).
func NewUserWithID(id, username, password, email string, roles []string) *User {
	now := time.Now().UTC()
	return &User{
		ID:        id,
		Username:  username,
		Password:  password, // Should be already hashed before calling this function.
		Email:     email,
		Roles:     roles,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// HasRole checks if the user has a specific role.
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles.
func (u *User) HasAnyRole(roles ...string) bool {
	for _, r := range roles {
		if u.HasRole(r) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the user has all of the specified roles.
func (u *User) HasAllRoles(roles ...string) bool {
	for _, r := range roles {
		if !u.HasRole(r) {
			return false
		}
	}
	return true
}

// Clone returns a copy of the user without the password.
func (u *User) Clone() *User {
	clone := *u
	clone.Password = ""

	// Deep copy the roles slice.
	clone.Roles = make([]string, len(u.Roles))
	copy(clone.Roles, u.Roles)

	return &clone
}

// AddRole adds a role to the user if they don't already have it.
func (u *User) AddRole(role string) {
	if u.HasRole(role) {
		return
	}
	u.Roles = append(u.Roles, role)
	u.UpdatedAt = time.Now().UTC()
}

// RemoveRole removes a role from the user.
func (u *User) RemoveRole(role string) {
	for i, r := range u.Roles {
		if r == role {
			// Remove the role by replacing it with the last element and truncating.
			u.Roles[i] = u.Roles[len(u.Roles)-1]
			u.Roles = u.Roles[:len(u.Roles)-1]
			u.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

// SetActive sets the active status of the user.
func (u *User) SetActive(active bool) {
	if u.Active != active {
		u.Active = active
		u.UpdatedAt = time.Now().UTC()
	}
}

// SetPassword updates the user's password.
func (u *User) SetPassword(password string) {
	u.Password = password // Should be already hashed before calling this function.
	u.UpdatedAt = time.Now().UTC()
}

// SetEmail updates the user's email.
func (u *User) SetEmail(email string) {
	if u.Email != email {
		u.Email = email
		u.UpdatedAt = time.Now().UTC()
	}
}
