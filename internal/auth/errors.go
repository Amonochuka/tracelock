package auth

import "errors"

var (
	ErrTokenInvalidMethod = errors.New("invalid jwt signing method")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email/password")
	ErrInvalidRole        = errors.New("role must be 'user' or 'admin'")
	ErrAdminExists        = errors.New("admin already exists")
)
