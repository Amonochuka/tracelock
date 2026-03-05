package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"tracelock/internal/httpa"
)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// same email and password in DB ?
func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			httpa.WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}

		if req.Email == "" || req.Password == "" {
			httpa.WriteError(w, http.StatusBadRequest, "must provide name and email")
			return
		}

		user, err := Authenticate(db, req.Email, req.Password)
		if err != nil {
			httpa.WriteError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		token, err := GenerateToken(user)
		if err != nil {
			httpa.WriteError(w, http.StatusInternalServerError, "could not generate")
			return
		}

		httpa.WriteJSON(w, http.StatusOK, map[string]string{
			"token": token,
		})
	}
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			httpa.WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}

		if req.Name == "" || req.Email == "" || req.Password == "" {
			httpa.WriteError(w, http.StatusBadRequest, "all fields are required")
			return
		}

		if err := Register(db, req.Name, req.Email, req.Password); err != nil {
			httpa.WriteError(w, http.StatusInternalServerError, "could not register user: "+err.Error())
			return
		}
		httpa.WriteJSON(w, http.StatusCreated, map[string]string{
			"message": "user registered successfully",
		})
	}
}

func MeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserContextKey).(int)
		if !ok {
			httpa.WriteError(w, http.StatusUnauthorized, "unauthorized access!")
			return
		}

		user, err := VerifyUser(db, userID)
		if err != nil {
			httpa.WriteError(w, http.StatusInternalServerError, "could not fetch user")
		}

		httpa.WriteJSON(w, http.StatusOK, user)
	}
}
