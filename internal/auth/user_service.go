package auth

import (
	"fmt"

	"tracelock/internal/models"
)

type UserService struct {
	auth *UserAuth
}

func NewUserService(auth *UserAuth) *UserService {
	return &UserService{auth: auth}
}

func (s *UserService) Register(name, email, password string) error {
	return s.auth.Register(name, email, password)
}

func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.auth.Authenticate(email, password)
	if err != nil {
		if err == ErrInvalidCredentials {
			return nil, err
		}
		return nil, fmt.Errorf("authenticate user %s: %w", email, err)
	}
	return user, nil
}

func (s *UserService) VerifyUser(id int) (*models.User, error) {
	user, err := s.auth.VerifyUser(id)
	if err != nil {
		if err == ErrUserNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("verify user %d: %w", id, err)
	}
	return user, nil
}
