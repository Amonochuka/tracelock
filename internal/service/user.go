package service

import (
	"tracelock/internal/auth"
)

type UserService struct {
	auth *auth.UserAuth
}

func NewService(auth *auth.UserAuth) *UserService {
	return &UserService{auth: auth}
}

func (s *UserService) Register(name, email, password string) error {
	return s.auth.Register(name, email, password)
}

func (s *UserService) Authenticate(email, password string) (*auth.User, error) {
	return s.auth.Authenticate(email, password)
}

func (s *UserService) VerifyUser(ID int) (*auth.User, error) {
	return s.auth.VerifyUser(ID)
}
