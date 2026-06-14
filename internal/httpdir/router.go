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

func New(authService *auth.UserService, jwtService *auth.JWTService, zoneService *access.ZoneService,
	deviceService *access.DeviceService, credentialService *access.CredentialService,
	biometricService *access.BiometricService) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.RequestID)

	// Public
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	limiter := middleware.NewRateLimiter(5) // 5 requests per minute

	r.With(limiter.Middleware).Post("/register", RegisterHandler(authService))
	r.With(limiter.Middleware).Post("/login", LoginHandler(authService, jwtService))
	r.Post("/bootstrap", BootStrapHandler(authService))
	r.Post("/logout", LogoutHandler(authService))
	r.Post("/refresh", RefreshHandler(authService))

	r.Post("/devices/authenticate", AuthenticateBiometricHandler(biometricService))
	// hub
	// WebSocket; live zone occupancy feed
	r.Get("/ws/zones", zoneService.GetHub().HandleWebSocket)

	//for frontend dashboard
	r.Get("/zones/occupancy", ListZoneOccupancyHandler(zoneService))

	// Authenticated
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware(jwtService))

		r.Get("/me", MeHandler(authService))
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

		// Admin only
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("admin"))

			r.Get("/admin/ping", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("admin ok\n"))
			})

			// User management
			r.Get("/admin/users", ListUsersHandler(authService))
			r.Put("/admin/users/{id}/role", UpdateRoleHandler(authService))
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

			// Device management (admin only — already inside admin group)
			r.Post("/admin/zones/{id}/devices", CreateDeviceHandler(deviceService))
			r.Get("/admin/zones/{id}/devices", ListDevicesHandler(deviceService))
			r.Get("/admin/devices/{id}", GetDeviceHandler(deviceService))
			r.Put("/admin/devices/{id}", UpdateDeviceHandler(deviceService))
			r.Patch("/admin/devices/{id}/deactivate", DeactivateDeviceHandler(deviceService))
			r.Delete("/admin/devices/{id}", DeleteDeviceHandler(deviceService))

			// Credential management
			r.Post("/admin/users/{id}/credentials", EnrollCredentialHandler(credentialService))
			r.Get("/admin/users/{id}/credentials", ListUserCredentialsHandler(credentialService))
			r.Get("/admin/users/{id}/credentials/{method}", GetCredentialHandler(credentialService))
			r.Delete("/admin/users/{id}/credentials/{method}", RevokeCredentialHandler(credentialService))

			//peak analysis
			r.Get("/admin/zones/{id}/analytics", GetZoneAnalyticsHandler(zoneService))

			//unlock a locked account
			r.Put("/admin/users/{id}/unlock", UnlockAccountHandler(authService))
		})
	})

	return r
}
