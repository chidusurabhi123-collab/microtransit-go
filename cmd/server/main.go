package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"microtransit-go/internal/handler"
	"microtransit-go/internal/ingestor"
	"microtransit-go/internal/store"
)

func main() {
	// Load .env file
	godotenv.Load()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL not set in .env")
	}

	// 1. Connect to Postgres with 5 connections (low RAM)
	s, err := store.New(connStr)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Create queue and start 5 workers
	ing := ingestor.New(s)
	ing.StartWorkers(5)
	log.Println("Started 5 workers with 1000-slot queue")

	// 3. HTTP handlers
	h := handler.New(s, ing)

	http.HandleFunc("/api/bus/location", h.UpdateLocation)
	http.HandleFunc("/api/buses", h.GetAllBuses)
	// Serve map files from ./static
	http.Handle("/", http.FileServer(http.Dir("./static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on http://localhost:%s - RAM ~20MB", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}