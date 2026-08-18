package auth

import (
	"errors"
	"fmt"
	"time"

	"tracelock/internal/models"
)

type UserService struct {
	auth UserRepository
	jwt  *JWTService
}

func NewUserService(auth UserRepository, j *JWTService) *UserService {
	return &UserService{auth: auth, jwt: j}
}

func (s *UserService) Register(name, email, password string) error {
	return s.auth.Register(name, email, password)
}

func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.auth.Authenticate(email, password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil, err
		}
		return nil, fmt.Errorf("authenticate user %s: %w", email, err)
	}
	return user, nil
}

func (s *UserService) VerifyUser(id int) (*models.User, error) {
	user, err := s.auth.VerifyUser(id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("verify user %d: %w", id, err)
	}
	return user, nil
}

// save refresh token
func (s *UserService) SaveRefreshToken(userID int, token string, expiresAt time.Time) error {
	tokenHash := HashRefreshToken(token)
	return s.auth.SaveRefreshToken(userID, tokenHash, expiresAt)
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

func (s *UserService) ResetAdminPassword(email, password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return s.auth.ResetAdminPassword(email, password)
}

// admin duties; update roles and list users
func (s *UserService) UpdateRole(userID int, role string) error {
	if role != "admin" && role != "user" {
		return ErrInvalidRole
	}
	return s.auth.UpdateRole(userID, role)
}

func (s *UserService) ListUsers() ([]*models.User, error) {
	return s.auth.ListUsers()
}

// give a user a new access token
func (s *UserService) RefreshAccessToken(token string) (string, error) {
	//validate token and get userID in one query
	tokenHash := HashRefreshToken(token)
	userID, err := s.auth.ValidateAndGetUserIDFromRefreshToken(tokenHash)
	if err != nil {
		return "", err
	}

	//verify user exists
	user, err := s.auth.VerifyUser(userID)
	if err != nil {
		return "", err
	}
	return s.jwt.GenerateToken(user)
}

func (s *UserService) Logout(token string) error {
	tokenHash := HashRefreshToken(token)
	return s.auth.RevokeRefreshToken(tokenHash)
}

func (s *UserService) DeleteExpiredTokens() error {
	return s.auth.DeleteExpiredTokens()
}

func (s *UserService) UnlockAccount(userID int) error {
	return s.auth.UnlockAccount(userID)
}

// DeleteUser permanently removes a user — admin action.
// Blocks deletion of admin accounts for safety.
func (s *UserService) DeleteUser(userID int) error {
	user, err := s.auth.VerifyUser(userID)
	if err != nil {
		return err
	}
	if user.Role == "admin" {
		return ErrCannotDeleteAdmin
	}
	return s.auth.DeleteUser(userID)
}
