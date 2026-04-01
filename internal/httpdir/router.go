package httpdir

import (
	"net/http"
	"strconv"

	"tracelock/internal/access"
	"tracelock/internal/auth"
	"tracelock/internal/httpdir/middleware"

	"github.com/go-chi/chi/v5"
)

func New(s *auth.UserService, jwtService *auth.JWTService, zoneService *access.ZoneService) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Post("/register", RegisterHandler(s))
	r.Post("/login", LoginHandler(s, jwtService))

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtService))

		r.Get("/me", MeHandler(s))

		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			user := auth.GetUserClaims(r)
			w.Write([]byte("Hello user ID: " + strconv.Itoa(user.UserID) + " your role is: " + user.Role + "\n"))
		})

		r.Get("/testjwt", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("JWT middleware works!\n"))
		})

		// zone routes
		r.Post("/zones/enter", EnterZoneHandler(zoneService))
		r.Post("/zones/exit", ExitZoneHandler(zoneService))

		// admin routes
		r.With(middleware.RequireRole("admin")).Get("/admin/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("admin ok\n"))
		})
	})

	return r
}
