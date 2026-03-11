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

type UserAuth struct {
	db *sql.DB
}

func NewUserAuth(db *sql.DB) *UserAuth {
	return &UserAuth{db: db}
}

// register a new user
func (u *UserAuth) Register(name, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	//insert into the DB
	_, err = u.db.Exec(
		"INSERT INTO users(name, email, password_hash) VALUES($1,$2,$3)",
		name, email, string(hash),
	)
	if err != nil {
		return err
	}
	return nil
}

// Authenticate user(tobe used for verifying logins

func (u *UserAuth) Authenticate(email, password string) (*User, error) {
	user := &User{}
	row := u.db.QueryRow("SELECT id, name, email, password_hash, role FROM users WHERE email=$1", email)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		return nil, errors.New("User not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}
	return user, nil
}

func (u *UserAuth) VerifyUser(ID int) (*User, error) {
	user := &User{}
	err := u.db.QueryRow("SELECT id, name, email, role, created_at FROM users WHERE id = $1", ID).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
