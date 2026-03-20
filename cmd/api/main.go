package main

import (
	"log"
	"net/http"
	"tracelock/internal/auth"
	"tracelock/internal/config"
	"tracelock/internal/db"
	"tracelock/internal/auth/httpapi"
	"tracelock/internal/auth/service"
)

func main() {

	cfg := config.Load()

	auth.NewJWTService(cfg.JWTSecret)

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	userAuth := auth.NewUserAuth(database)
	userService := service.NewUserService(userAuth)
	jwtService := auth.NewJWTService(cfg.JWTSecret)

	handler := httpapi.New(userService, jwtService)

	log.Println("Tracelock API running on:" + cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal(err)
	}
}
