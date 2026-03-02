package httpapi

import (
	"database/sql"
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
	r.Post("/register", auth.RegisterHandler(db))

	//login route
	r.Post("/login", auth.LoginHandler(db))

	//test JWT middleware
	r.Group(func(r chi.Router) {
		r.Use((auth.JWTMiddleware))

		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			user := auth.GetUserClaims(r)
			w.Write([]byte("Hello user ID" + strconv.Itoa(user.UserID) + "role"))
		})
		r.Get("/testjwt", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("JWT middleware works!"))
		})

		r.Get("/me", auth.MeHandler(db))

		r.With(auth.RequireRole("admin")).Get("/admin/ping", (http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("admin ok"))
		})))

	})

	return r
}
