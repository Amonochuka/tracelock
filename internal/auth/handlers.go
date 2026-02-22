package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// same email andp password in DB ?
func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "inavlid request body", http.StatusBadRequest)
			return
		}

		if request.Email == "" || request.Password == "" {
			http.Error(w, "must provide name and email", http.StatusBadRequest)
			return
		}

		user, err := Authenticate(db, request.Email, request.Password)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := GenerateToken(user.ID)
		if err != nil {
			http.Error(w, "could not generate token", http.StatusInternalServerError)
			return
		}

		resp := map[string]any{
			"token": token,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if request.Name == "" || request.Email == "" || request.Password == "" {
			http.Error(w, "all fields are required", http.StatusBadRequest)
			return
		}

		if err := Register(db, request.Name, request.Email, request.Password); err != nil {
			http.Error(w, "could not register user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "user created successfully",
		})
	}
}
