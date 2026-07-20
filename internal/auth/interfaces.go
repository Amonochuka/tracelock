package auth

import (
	"time"
	"tracelock/internal/models"
)

type UserRepository interface {
	Register(name, email, password string) error
	Authenticate(email, password string) (*models.User, error)
	VerifyUser(id int) (*models.User, error)
	AdminExists() (bool, error)
	RegisterAdmin(name, email, password string) error
	UpdateRole(userID int, role string) error
	ListUsers() ([]*models.User, error)
	SaveRefreshToken(userID int, token string, expiresAt time.Time) error
	GetRefreshToken(token string) error
	RevokeRefreshToken(token string) error
	GetUserIDFromRefreshToken(token string) (int, error)
	DeleteExpiredTokens() error
	IncrementFailedAttempts(email string) error
	LockAccount(email string) error
	ResetFailedAttempts(email string) error
	IsAccountLocked(email string) (bool, error)
	UnlockAccount(userID int) error
}
