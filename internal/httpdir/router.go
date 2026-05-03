package httpdir

import (
	"net/http"
	"strconv"

	"tracelock/internal/access"
	"tracelock/internal/auth"
	"tracelock/internal/httpdir/middleware"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/go-chi/chi/v5"
)

func New(s *auth.UserService, jwtService *auth.JWTService, zoneService *access.ZoneService) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.RequestID)

	// Public
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	limiter := middleware.NewRateLimiter(5) // 5 requests per minute

	r.With(limiter.Middleware).Post("/register", RegisterHandler(s))
	r.With(limiter.Middleware).Post("/login", LoginHandler(s, jwtService))
	r.Post("/bootstrap", BootStrapHandler(s))
	r.Post("/logout", LogoutHandler(s))
	r.Post("/refresh", RefreshHandler(s))

	// Authenticated
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtService))

		r.Get("/me", MeHandler(s))
		// Me routes; users can see their own data
		r.Get("/me/events", MeEventsHandler(zoneService))
		r.Get("/me/access", MeAccessHandler(zoneService))

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

		// User routes

		// Admin only
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("admin"))

			r.Get("/admin/ping", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("admin ok\n"))
			})

			// User management
			r.Get("/admin/users", ListUsersHandler(s))
			r.Put("/admin/users/{id}/role", UpdateRoleHandler(s))
			r.Get("/users/{id}/events", ListUserEventsHandler(zoneService))
			r.Get("/users/{id}/access", ListUserAccessHandler(zoneService))

			// Zone management
			r.Post("/admin/zones", CreateZoneHandler(zoneService))
			r.Put("/admin/zones/{id}", UpdateZoneHandler(zoneService))
			r.Delete("/admin/zones/{id}", DeleteZoneHandler(zoneService))
			r.Get("/admin/zones/{id}/users", ListZoneUsersHandler(zoneService))
			r.Get("/zones/{id}/events", ListZoneEventsHandler(zoneService))
			r.Get("/admin/zones/{id}/verify-chain", VerifyChainHandler(zoneService))

			// Access control
			r.Post("/admin/access", GrantAccessHandler(zoneService))
			r.Delete("/admin/access", RevokeAccessHandler(zoneService))
		})
	})

	return r
}
