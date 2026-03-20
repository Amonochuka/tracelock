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

// validation methods for registering and logins
func (r *RegisterRequest) Validate() error {
	if len(r.Name) < 2 {
		return errors.New("name must be atleast two charcaters")
	}
	if !strings.Contains(r.Email, "@") {
		return errors.New("invalid email")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be atleast 8 characters")
	}
	return nil
}

func (l *LoginRequest) Validate() error {
	if !strings.Contains(l.Email, "@") {
		return errors.New("invalid error")
	}
	if len(l.Password) < 8 {
		return errors.New("password must be atleast 8 characters")
	}
	return nil
}

// same email and password in DB ?
func LoginHandler(s *auth.UserService, j *auth.JWTService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}

		if err := req.Validate(); err != nil {
			WriteError(w, http.StatusBadRequest, "must provide name and email")
			return
		}

		user, err := s.Authenticate(req.Email, req.Password)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		token, err := j.GenerateToken(user)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not generate")
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{
			"token": token,
		})
	}
}

func RegisterHandler(s *auth.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}

		if err := req.Validate(); err != nil {
			WriteError(w, http.StatusBadRequest, "all fields are required")
			return
		}

		if err := s.Register(req.Name, req.Email, req.Password); err != nil {
			WriteError(w, http.StatusInternalServerError, "could not register user: "+err.Error())
			return
		}
		WriteJSON(w, http.StatusCreated, map[string]string{
			"message": "user registered successfully",
		})
	}
}

func MeHandler(s *auth.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetUserClaims(r)
		if claims == nil {
			WriteError(w, http.StatusUnauthorized, "unauthorized access!")
			return
		}

		user, err := s.VerifyUser(claims.UserID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not fetch user")
			return
		}

		WriteJSON(w, http.StatusOK, user)
	}
}
