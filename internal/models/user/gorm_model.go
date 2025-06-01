package user

import (
	"encoding/json"
	"time"
)

// GormUser represents the database model for a user
type GormUser struct {
	ID        string `gorm:"primaryKey"`
	Username  string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	RolesJSON string `gorm:"column:roles;type:text;not null;default:'[]'"`
	Active    bool   `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name for the GormUser model
func (GormUser) TableName() string {
	return "users"
}

// GetRoles returns the roles as a string slice
func (gu *GormUser) GetRoles() ([]string, error) {
	// Handle common error cases
	if gu.RolesJSON == "" || gu.RolesJSON == "null" {
		return []string{}, nil
	}

	// If the data is just a string like "admin" instead of ["admin"],
	// wrap it in an array
	if gu.RolesJSON == "admin" {
		return []string{"admin"}, nil
	}

	// Try to unmarshal as an array of strings
	var roles []string
	if err := json.Unmarshal([]byte(gu.RolesJSON), &roles); err != nil {
		// If we can't parse it, return an empty array instead of an error
		return []string{}, nil
	}
	return roles, nil
}

// SetRoles sets the roles from a string slice
func (gu *GormUser) SetRoles(roles []string) error {
	if len(roles) == 0 {
		gu.RolesJSON = "[]"
		return nil
	}

	rolesJSON, err := json.Marshal(roles)
	if err != nil {
		return err
	}
	gu.RolesJSON = string(rolesJSON)
	return nil
}

// ToUser converts a GormUser to a domain User model
func (gu *GormUser) ToUser() *User {
	roles, err := gu.GetRoles()
	if err != nil {
		roles = []string{} // Default to empty array on error
	}

	return &User{
		ID:        gu.ID,
		Username:  gu.Username,
		Password:  gu.Password,
		Email:     gu.Email,
		Roles:     roles,
		Active:    gu.Active,
		CreatedAt: gu.CreatedAt,
		UpdatedAt: gu.UpdatedAt,
	}
}

// FromUser creates a GormUser from a domain User model
func FromUser(u *User) *GormUser {
	gu := &GormUser{
		ID:        u.ID,
		Username:  u.Username,
		Password:  u.Password,
		Email:     u.Email,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	// Set roles using the helper method
	if err := gu.SetRoles(u.Roles); err != nil {
		gu.RolesJSON = "[]" // Default to empty array on error
	}

	return gu
}
