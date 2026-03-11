package main

import (
	"log"
	"net/http"
	"tracelock/internal/auth"
	"tracelock/internal/config"
	"tracelock/internal/db"
	"tracelock/internal/httpapi"
)

func main() {

	cfg := config.Load()

	auth.NewJWTService(cfg.JWTSecret)

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
