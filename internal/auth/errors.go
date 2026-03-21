package auth

import "errors"

var (
	ErrTokenInvalidMethod = errors.New("invalid jwt signing method")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrEmailExists        = errors.New("email already exists")
)
