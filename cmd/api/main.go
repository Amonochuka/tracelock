package main

import (
	"log"
	"net/http"

	"tracelock/internal/access"
	"tracelock/internal/auth"
	"tracelock/internal/config"
	"tracelock/internal/db"
	"tracelock/internal/httpdir"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// auth
	userAuth := auth.NewUserAuth(database)
	userService := auth.NewUserService(userAuth)
	jwtService := auth.NewJWTService(cfg.JWTSecret)

	// access
	zoneRepo := access.NewZoneRepo(database)
	zoneService := access.NewZoneService(zoneRepo)

	handler := httpdir.New(userService, jwtService, zoneService)

	log.Println("Tracelock API running on: " + cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal(err)
	}
}
