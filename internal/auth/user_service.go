package auth

import (
	"fmt"
	"time"

	"tracelock/internal/models"
)

type UserService struct {
	auth *UserAuth
	jwt  *JWTService
}

func NewUserService(auth *UserAuth, j *JWTService) *UserService {
	return &UserService{auth: auth, jwt: j}
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

// save refersh token
func (s *UserService) SaveRefreshToken(userID int, token string, expiresAt time.Time) error {
	return s.auth.SaveRefreshToken(userID, token, expiresAt)
}

// bootstrap admin
func (s *UserService) BootStrapAdmin(name, email, password string) error {
	exists, err := s.auth.AdminExists()
	if err != nil {
		return err
	}
	if exists {
		return ErrAdminExists
	}
	return s.auth.RegisterAdmin(name, email, password)
}

// admin duties; update roles and list users
func (s *UserService) UpdateRole(userID int, role string) error {
	if role != "admin" && role != "user" {
		return ErrInvalidRole
	}
	return s.UpdateRole(userID, role)
}

func (s *UserService) ListUsers() ([]*models.User, error) {
	return s.auth.ListUsers()
}

// give a user a new access token
func (s *UserService) RefreshAccessToken(token string) (string, error) {
	//get the refresh token
	if err := s.auth.GetRefreshToken(token); err != nil {
		return "", err
	}

	//get userID associated to a refersh token
	userID, err := s.auth.GetUserIDFromRefreshToken(token)
	if err != nil {
		return "", err
	}

	//verify user
	user, err := s.auth.VerifyUser(userID)
	if err != nil {
		return "", err
	}
	return s.jwt.GenerateToken(user)
}

func (s *UserService) Logout(token string) error {
	return s.auth.RevokeRefreshToken(token)
}
