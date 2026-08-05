package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tracelock/internal/access"
	"tracelock/internal/auth"
	"tracelock/internal/config"
	"tracelock/internal/db"
	"tracelock/internal/httpdir"

	"github.com/joho/godotenv"
)

const (
	serverReadHeaderTimeout = 5 * time.Second
	serverReadTimeout       = 15 * time.Second
	serverWriteTimeout      = 30 * time.Second
	serverIdleTimeout       = 60 * time.Second
)

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}
}

func main() {
	// loads .env file automatically
	godotenv.Load()

	cfg := config.Load()

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// auth
	userAuth := auth.NewUserAuth(database)
	jwtService := auth.NewJWTService(cfg.JWTSecret)
	userService := auth.NewUserService(userAuth, jwtService)

	// start token cleanup job
	go func() {
		for {
			if err := userService.DeleteExpiredTokens(); err != nil {
				log.Printf("token cleanup failed: %v", err)
			}
			time.Sleep(24 * time.Hour)
		}
	}()

	if err := userService.DeleteExpiredTokens(); err != nil {
		log.Printf("initial token cleanup failed: %v", err)
	}

	// access
	zoneRepo := access.NewZoneRepo(database)

	// create hub and start it
	hub := access.NewHub(cfg.AllowedOrigin)
	go hub.Run()
	// pass hub to zone service
	zoneService := access.NewZoneService(zoneRepo, hub)

	// start stale session cleanup job
	sessionTimeout := time.Duration(cfg.SessionTimeoutHours) * time.Hour
	go func() {
		for {
			closed, err := zoneService.CleanupStaleSessions(sessionTimeout)
			if err != nil {
				log.Printf("stale session cleanup failed: %v", err)
			} else if closed > 0 {
				log.Printf("stale session cleanup: force-closed %d sessions", closed)
			}
			time.Sleep(1 * time.Hour)
		}
	}()

	// device management
	deviceRepo := access.NewDeviceRepo(database)
	deviceService := access.NewDeviceService(deviceRepo)

	//credentials
	credentialRepo := access.NewCredentialRepo(database)
	credentialService := access.NewCredentialService(credentialRepo)

	//biometrics
	biometricService := access.NewBiometricService(credentialRepo, deviceRepo, zoneService, userAuth, jwtService)

	handler := httpdir.New(userService, jwtService, zoneService, deviceService, credentialService, biometricService, cfg.DeviceAPIKey, cfg.AllowedOrigin)

	srv := newHTTPServer(":"+cfg.Port, handler)

	// run server in the background
	go func() {
		log.Println("Tracelock API running on: " + cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// wait for SIGTERM(render interruption) or SIGINT(local interruption "Ctrl + C")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("shutting down server...")

	// give in-flight requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	log.Println("server stopped cleanly")
}
