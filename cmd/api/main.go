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

func main() {
	godotenv.Load() // loads .env file automatically

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
			time.Sleep(24 * time.Hour)
			if err := userService.DeleteExpiredTokens(); err != nil {
				log.Printf("token cleanup failed: %v", err)
			}
		}
	}()

	// access
	zoneRepo := access.NewZoneRepo(database)
	// create hub and start it

	hub := access.NewHub()
	go hub.Run()
	// pass hub to zone service
	zoneService := access.NewZoneService(zoneRepo, hub)

	// device management
	deviceRepo := access.NewDeviceRepo(database)
	deviceService := access.NewDeviceService(deviceRepo)

	//credentials
	credentialRepo := access.NewCredentialRepo(database)
	credentialService := access.NewCredentialService(credentialRepo)

	//biometrics
	biometricService := access.NewBiometricService(credentialRepo, deviceRepo, zoneService, userAuth, jwtService)

	handler := httpdir.New(userService, jwtService, zoneService, deviceService, credentialService, biometricService)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

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
