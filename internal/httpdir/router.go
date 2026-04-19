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

	// Public
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Post("/register", RegisterHandler(s))
	r.Post("/login", LoginHandler(s, jwtService))
	r.Post("/bootstrap", BootStrapHandler(s))

	// Authenticated
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

		// Zone entry / exit
		r.Post("/zones/enter", EnterZoneHandler(zoneService))
		r.Post("/zones/exit", ExitZoneHandler(zoneService))

		// Zone read
		r.Get("/zones", ListZonesHandler(zoneService))
		r.Get("/zones/{id}", GetZoneHandler(zoneService))
		r.Get("/zones/{id}/events", ListZoneEventsHandler(zoneService))

		// User routes
		r.Get("/users/{id}/events", ListUserEventsHandler(zoneService))
		r.Get("/users/{id}/access", ListUserAccessHandler(zoneService))

		// Admin only
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("admin"))

			r.Get("/admin/ping", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("admin ok\n"))
			})

			// User management
			r.Get("/admin/users", ListUsersHandler(s))
			r.Put("/admin/users/{id}/role", UpdateRoleHandler(s))

			// Zone management
			r.Post("/admin/zones", CreateZoneHandler(zoneService))
			r.Put("/admin/zones/{id}", UpdateZoneHandler(zoneService))
			r.Delete("/admin/zones/{id}", DeleteZoneHandler(zoneService))
			r.Get("/admin/zones/{id}/users", ListZoneUsersHandler(zoneService))
			r.Get("/admin/zones/{id}/verify-chain", VerifyChainHandler(zoneService))

			// Access control
			r.Post("/admin/access", GrantAccessHandler(zoneService))
			r.Delete("/admin/access", RevokeAccessHandler(zoneService))
		})
	})

	return r
}
