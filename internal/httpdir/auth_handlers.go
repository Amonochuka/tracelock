package httpdir

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"tracelock/internal/auth"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() error {
	if len(strings.TrimSpace(r.Name)) < 2 {
		return errors.New("name must be at least two characters")
	}
	if !strings.Contains(r.Email, "@") {
		return errors.New("invalid email")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func (l *LoginRequest) Validate() error {
	if !strings.Contains(l.Email, "@") {
		return errors.New("invalid email")
	}
	if len(l.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func RegisterHandler(s *auth.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req RegisterRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if err := req.Validate(); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := s.Register(req.Name, req.Email, req.Password); err != nil {
			if errors.Is(err, auth.ErrEmailExists) {
				WriteError(w, http.StatusConflict, "email already exists")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not register user")
			return
		}

		WriteJSON(w, http.StatusCreated, map[string]string{
			"message": "user registered successfully",
		})
	}
}

func LoginHandler(s *auth.UserService, j *auth.JWTService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req LoginRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if err := req.Validate(); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		user, err := s.Authenticate(req.Email, req.Password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				WriteError(w, http.StatusUnauthorized, "invalid credentials")
				return
			}
			WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		token, err := j.GenerateToken(user)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not generate token")
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{
			"token": token,
		})
	}
}

func MeHandler(s *auth.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetUserClaims(r)
		if claims == nil {
			WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		user, err := s.VerifyUser(claims.UserID)
		if err != nil {
			if errors.Is(err, auth.ErrUserNotFound) {
				WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not fetch user")
			return
		}

		WriteJSON(w, http.StatusOK, UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		})
	}
}
