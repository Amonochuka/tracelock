package auth

import "tracelock/internal/models"

type UserService struct {
	auth *UserAuth
}

func NewUserService(auth *UserAuth) *UserService {
	return &UserService{auth: auth}
}

func (s *UserService) Register(name, email, password string) error {
	return s.auth.Register(name, email, password)
}

func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	return s.auth.Authenticate(email, password)
}

func (s *UserService) VerifyUser(ID int) (*models.User, error) {
	return s.auth.VerifyUser(ID)
}
