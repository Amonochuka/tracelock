package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

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

// register admin account, but first check if an admin exists
func (u *UserAuth) AdminExists() (bool, error) {
	var exists bool
	err := u.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE role = 'admin')").Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking admin exists: %w", err)
	}
	return exists, nil
}

// now regsiter an admin
func (u *UserAuth) RegisterAdmin(name, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	_, err = u.db.Exec("INSERT INTO users(name, email, password_hash, role)VALUES($1, $2, $3, 'admin')",
		name, email, string(hash))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrEmailExists
		}
		return fmt.Errorf("inserting admin: %w", err)
	}
	return nil
}

// admin duty; update
func (u *UserAuth) UpdateRole(userID int, role string) error {
	res, err := u.db.Exec("UPDATE users SET role = $1 WHERE id = $2", role, userID)
	if err != nil {
		return fmt.Errorf("updating user role: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// admin duty:list all users
func (u *UserAuth) ListUsers() ([]*models.User, error) {
	rows, err := u.db.Query("SELECT id, name, email, role, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("listing users:%w", err)
	}
	defer rows.Close()
	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning users :%w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

// save refersh token
func (u *UserAuth) SaveRefreshToken(userID int, token string, expiresAt time.Time) error {
	_, err := u.db.Exec(
		"INSERT INTO refresh_tokens(user_id, token, expires_at) VALUES($1,$2,$3)",
		userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("saving refresh token: %w", err)
	}
	return nil
}

// get the refresh token
func (u *UserAuth) GetRefreshToken(token string) error {
	var revoked bool
	var expiresAt time.Time
	err := u.db.QueryRow(`SELECT revoked, expires_at from refresh_tokens 
				WHERE token = $1`, token).Scan(&revoked, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTokenNotFound
		}
		return fmt.Errorf("get refresh token: %w", err)
	}
	if revoked {
		return ErrTokenRevoked
	}
	if time.Now().After(expiresAt) {
		return ErrTokenExpired
	}
	return nil
}

// revoke the refresh token
func (u *UserAuth) RevokeRefreshToken(token string) error {
	res, err := u.db.Exec("UPDATE refresh_tokens SET revoked = true WHERE token = $1", token)
	if err != nil {
		return fmt.Errorf("revoke refresh tokens: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrTokenNotFound
	}
	return nil
}

// get user using a certain refresh token
func (u *UserAuth) GetUserIDFromRefreshToken(token string) (int, error) {
	var userID int
	err := u.db.QueryRow("SELECT user_id FROM refresh_tokens WHERE token = $1", token).Scan(&token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrTokenNotFound
		}
		return 0, fmt.Errorf("getting user from refresh token: %w", err)
	}
	return userID, nil
}
