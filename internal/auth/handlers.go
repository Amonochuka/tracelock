package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"tracelock/internal/httpa"
)

type requestRegister struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

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

		token, err := GenerateToken(user)
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
		var req requestRegister
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
		UserClaims, ok := r.Context().Value(UserContextKey).(*UserClaims)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		uuid := UserClaims.UserID

		var user struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		}

		err := db.QueryRow(`SELECT id, name, email, 
		role FROM users WHERE id =$1`, uuid).Scan(&user.ID, &user.Name, &user.Email, &user.Role)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}
