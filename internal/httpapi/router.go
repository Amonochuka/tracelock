package httpapi

import (
	"net/http"
	"strconv"
	"tracelock/internal/auth"
	"tracelock/internal/handlers"
	"tracelock/internal/service"

	"github.com/go-chi/chi/v5"
)

func New(s *service.UserService, jwtservice *auth.JWTService) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))

	})

	//regsiter endpoint
	r.Post("/register", handlers.RegisterHandler(s))

	//login route
	r.Post("/login", handlers.LoginHandler(s, jwtservice))

	//test JWT middleware

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtservice))

		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			user := auth.GetUserClaims(r)
			w.Write([]byte("Hello user ID: " + strconv.Itoa(user.UserID) + " your role is: " + user.Role + "\n"))
		})
		r.Get("/testjwt", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("JWT middleware works!" + "\n"))
		})

		r.Get("/me", handlers.MeHandler(s))

		r.With(auth.RequireRole("admin")).Get("/admin/ping", (http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("admin ok" + "\n"))
		})))

	})

	return r
}
