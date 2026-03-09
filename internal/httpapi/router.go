package httpapi

import (
	"database/sql"
	"net/http"
	"os"
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

	jwtservice := auth.NewJWTService(os.Getenv("JWT_SECRET"))
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtservice))

		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			user := auth.GetUserClaims(r)
			w.Write([]byte("Hello user ID: " + strconv.Itoa(user.UserID) + " your role is: " + user.Role + "\n"))
		})
		r.Get("/testjwt", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("JWT middleware works!" + "\n"))
		})

		r.Get("/me", auth.MeHandler(db))

		r.With(auth.RequireRole("admin")).Get("/admin/ping", (http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("admin ok" + "\n"))
		})))

	})

	return r
}
