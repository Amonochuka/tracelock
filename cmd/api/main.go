package main

import (
	"log"
	"net/http"
	"os"
	"tracelock/internal/db"
	"tracelock/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}

	handler := httpapi.New(database)

	log.Println("Tracelock API running on:" + port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
