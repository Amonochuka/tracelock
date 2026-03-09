package main

import (
	"log"
	"net/http"
	"os"
	"tracelock/internal/auth"
	"tracelock/internal/config"
	"tracelock/internal/db"
	"tracelock/internal/httpapi"
)

func main() {

	cfg := config.Load()

	auth.NewJWTService(cfg.JWTSecret)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	handler := httpapi.New(database)

	log.Println("Tracelock API running on:" + cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal(err)
	}
}
