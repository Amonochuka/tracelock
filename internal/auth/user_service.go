package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"tracelock/internal/models"

	"github.com/lib/pq"
)

type UserService struct {
	auth *UserAuth
}

func NewUserService(auth *UserAuth) *UserService {
	return &UserService{auth: auth}
}

func (s *UserService) Register(name, email, password string) error {
	err := s.auth.Register(name, email, password)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23505" {
			return fmt.Errorf("registration failed :%w", ErrEmailExists)
		}
	}
	return fmt.Errorf("service registerUser: %w", err)
}

func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.auth.Authenticate(email, password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("authentication failed :%w", ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("service authenticateUser: %w", err)
	}
	return user, nil
}

func (s *UserService) VerifyUser(ID int) (*models.User, error) {
	user, err := s.auth.VerifyUser(ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("verifyUser failed :%w", ErrUserNotFound)
		}
		return nil, fmt.Errorf("service verifyUser: %w", err)
	}
	return user, nil
}
