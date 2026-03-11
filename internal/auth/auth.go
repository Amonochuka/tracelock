package auth

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// how does this struct fit in ?
type User struct {
	ID           int
	Name         string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    string
}

// register a new user
func Register(db *sql.DB, name, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	//insert into the DB
	_, err = db.Exec(
		"INSERT INTO users(name, email, password_hash) VALUES($1,$2,$3)",
		name, email, string(hash),
	)
	if err != nil {
		return err
	}
	return nil
}

// Authenticate user(tobe used for verifying logins

func Authenticate(db *sql.DB, email, password string) (*User, error) {
	user := &User{}
	row := db.QueryRow("SELECT id, name, email, password_hash, role FROM users WHERE email=$1", email)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		return nil, errors.New("User not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	return user, nil
}

func VerifyUser(db *sql.DB, ID int) (*User, error) {
	user := &User{}
	err := db.QueryRow("SELECT id, name, email, role, created_at FROM users WHERE id = $1", ID).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
