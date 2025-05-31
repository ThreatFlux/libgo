package user

import "errors"

// Error definitions
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrDuplicateUsername  = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
	ErrEmailRequired      = errors.New("email is required")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidRole        = errors.New("invalid role")
	ErrPasswordRequired   = errors.New("password is required")
)
