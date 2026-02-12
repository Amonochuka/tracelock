package httpapi

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"tracelock/internal/auth"

	"github.com/go-chi/chi/v5"
)

func New(db *sql.DB) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))

	})

	//regsiter endpoint
	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" || req.Email == "" || req.Password == "" {
			http.Error(w, "all fields are required", http.StatusBadRequest)
			return
		}

		err := auth.Register(db, req.Name, req.Email, req.Password)
		if err != nil {
			log.Println("Register error:", err)
			http.Error(w, "could not register user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("user regsitered succesfully"))
	})

	//test JWT middleware
	r.Group(func(r chi.Router) {
		r.Use((auth.JWTMiddleware))

		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			user := auth.GetUserClaims(r)
			w.Write([]byte("Hello user ID" + strconv.Itoa(user.UserID) + "role"))
		})
	})

	return r
}
