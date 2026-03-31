package auth

import (
	"database/sql"
	"errors"
	"fmt"

	"tracelock/internal/models"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserAuth struct {
	db *sql.DB
}

func NewUserAuth(db *sql.DB) *UserAuth {
	return &UserAuth{db: db}
}

func (u *UserAuth) Register(name, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	_, err = u.db.Exec(
		"INSERT INTO users(name, email, password_hash) VALUES($1,$2,$3)",
		name, email, string(hash),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrEmailExists
		}
		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}

func (u *UserAuth) Authenticate(email, password string) (*models.User, error) {
	user := &models.User{}

	err := u.db.QueryRow(
		"SELECT id, name, email, password_hash, role FROM users WHERE email=$1", email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (u *UserAuth) VerifyUser(id int) (*models.User, error) {
	user := &models.User{}

	err := u.db.QueryRow(
		"SELECT id, name, email, role, created_at FROM users WHERE id=$1", id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by id: %w", err)
	}

	return user, nil
}
