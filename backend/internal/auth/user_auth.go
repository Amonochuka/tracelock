package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"tracelock/internal/models"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserAuth struct {
	db *sql.DB
}

const (
	// MaxFailedAttempts is the number of wrong passwords before an account locks.
	MaxFailedAttempts = 5
	// AccountLockDuration is how long an account stays locked.
	AccountLockDuration = 15 * time.Minute
)

func NewUserAuth(db *sql.DB) UserRepository {
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
		`SELECT id, name, email, password_hash, role, failed_attempts, locked_until, created_at 
		 FROM users WHERE email=$1 AND deleted_at IS NULL`, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.FailedAttempts,
		&user.LockedUntil, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	// check if account is locked; compare in UTC so a naive/aware
	// timestamp mismatch can never extend the lock window
	if user.LockedUntil != nil {
		if time.Now().UTC().Before(user.LockedUntil.UTC()) {
			return nil, ErrAccountLocked
		}
		// lock expired; clear it and give the user a fresh attempt budget,
		// otherwise a single wrong password would instantly re-lock
		_ = u.ResetFailedAttempts(email)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// wrong password; increment atomically and read back the real count
		attempts, err := u.IncrementFailedAttempts(email)
		if err != nil {
			return nil, fmt.Errorf("incrementing failed attempts: %w", err)
		}

		// lock after too many failed attempts
		if attempts >= MaxFailedAttempts {
			_ = u.LockAccount(email)
			return nil, ErrAccountLocked
		}
		return nil, ErrInvalidCredentials
	}

	// successful login;reset counter
	_ = u.ResetFailedAttempts(email)
	return user, nil
}

func (u *UserAuth) VerifyUser(id int) (*models.User, error) {
	user := &models.User{}
	err := u.db.QueryRow(
		"SELECT id, name, email, role, failed_attempts, locked_until, created_at FROM users WHERE id=$1 AND deleted_at IS NULL", id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.FailedAttempts, &user.LockedUntil, &user.CreatedAt)
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
	err := u.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE role = 'admin' AND deleted_at IS NULL)").Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking admin exists: %w", err)
	}
	return exists, nil
}

// now register an admin
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

func (u *UserAuth) ResetAdminPassword(email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	var userID int
	err = u.db.QueryRow(`
		UPDATE users
		SET password_hash = $1, failed_attempts = 0, locked_until = NULL
		WHERE email = $2 AND role = 'admin' AND deleted_at IS NULL
		RETURNING id`, string(hash), email).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("resetting admin password: %w", err)
	}

	if _, err := u.db.Exec("UPDATE refresh_tokens SET revoked = true WHERE user_id = $1", userID); err != nil {
		return fmt.Errorf("revoking admin refresh tokens: %w", err)
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
	rows, err := u.db.Query("SELECT id, name, email, role, created_at FROM users WHERE deleted_at IS NULL ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("listing users:%w", err)
	}
	defer rows.Close()
	var users []*models.User
	for rows.Next() {
		usr := &models.User{}
		if err := rows.Scan(&usr.ID, &usr.Name, &usr.Email, &usr.Role, &usr.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning users :%w", err)
		}
		users = append(users, usr) // ← was missing; caused empty list
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}
	return users, nil
}

// save refresh token
func (u *UserAuth) SaveRefreshToken(userID int, tokenHash string, expiresAt time.Time) error {
	_, err := u.db.Exec(
		"INSERT INTO refresh_tokens(user_id, token_hash, expires_at) VALUES($1,$2,$3)",
		userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("saving refresh token: %w", err)
	}
	return nil
}

// ValidateAndGetUserIDFromRefreshToken validates the token in one query and returns the associated user ID.
// Returns the user ID if the token is valid, not revoked, and not expired.
func (u *UserAuth) ValidateAndGetUserIDFromRefreshToken(tokenHash string) (int, error) {
	var userID int
	var revoked bool
	var expiresAt time.Time
	err := u.db.QueryRow(`
		SELECT user_id, revoked, expires_at FROM refresh_tokens 
		WHERE token_hash = $1
	`, tokenHash).Scan(&userID, &revoked, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrTokenNotFound
		}
		return 0, fmt.Errorf("validating refresh token: %w", err)
	}
	if revoked {
		return 0, ErrTokenRevoked
	}
	if time.Now().After(expiresAt) {
		return 0, ErrTokenExpired
	}
	return userID, nil
}

// revoke the refresh token
func (u *UserAuth) RevokeRefreshToken(tokenHash string) error {
	res, err := u.db.Exec("UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1", tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh tokens: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrTokenNotFound
	}
	return nil
}

// DeleteExpiredTokens removes all expired refresh tokens from the DB.
func (u *UserAuth) DeleteExpiredTokens() error {
	res, err := u.db.Exec(`DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return fmt.Errorf("delete expired tokens: %w", err)
	}
	rows, _ := res.RowsAffected()
	log.Printf("token cleanup: deleted %d expired refresh tokens", rows)
	return nil
}

// IncrementFailedAttempts increments the failed login counter for a user
// and returns the updated count (atomic — safe under concurrent logins).
func (u *UserAuth) IncrementFailedAttempts(email string) (int, error) {
	var attempts int
	err := u.db.QueryRow(`
		UPDATE users SET failed_attempts = failed_attempts + 1
		WHERE email = $1
		RETURNING failed_attempts`, email).Scan(&attempts)
	if err != nil {
		return 0, fmt.Errorf("increment failed attempts: %w", err)
	}
	return attempts, nil
}

// LockAccount locks a user account for AccountLockDuration.
// The expiry is computed in the app (UTC) instead of NOW() so the
// database session timezone can never skew the lock window.
func (u *UserAuth) LockAccount(email string) error {
	lockedUntil := time.Now().UTC().Add(AccountLockDuration)
	_, err := u.db.Exec(`
		UPDATE users SET locked_until = $1
		WHERE email = $2`, lockedUntil, email)
	if err != nil {
		return fmt.Errorf("lock account: %w", err)
	}
	return nil
}

// ResetFailedAttempts resets the failed counter and clears the lock on successful login.
func (u *UserAuth) ResetFailedAttempts(email string) error {
	_, err := u.db.Exec(`
		UPDATE users SET failed_attempts = 0, locked_until = NULL
		WHERE email = $1`, email)
	if err != nil {
		return fmt.Errorf("reset failed attempts: %w", err)
	}
	return nil
}

// IsAccountLocked checks if a user account is currently locked.
func (u *UserAuth) IsAccountLocked(email string) (bool, error) {
	var lockedUntil *time.Time
	err := u.db.QueryRow(`
		SELECT locked_until FROM users WHERE email = $1 AND deleted_at IS NULL`, email).Scan(&lockedUntil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrUserNotFound
		}
		return false, fmt.Errorf("check account lock: %w", err)
	}
	if lockedUntil != nil && time.Now().UTC().Before(lockedUntil.UTC()) {
		return true, nil
	}
	return false, nil
}

// UnlockAccount clears the lock and resets failed attempts — admin action.
func (u *UserAuth) UnlockAccount(userID int) error {
	res, err := u.db.Exec(`
		UPDATE users SET failed_attempts = 0, locked_until = NULL
		WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("unlock account: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// DeleteUser permanently removes a user and all their associated data (now soft deletes).
func (u *UserAuth) DeleteUser(userID int) error {
	res, err := u.db.Exec(`UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	_, _ = u.db.Exec("UPDATE refresh_tokens SET revoked = true WHERE user_id = $1", userID)
	return nil
}
