package main

import (
	"log"
	"net/http"
	"os"
	"microtransit-go/internal/eta"
	"microtransit-go/internal/handler"
	"microtransit-go/internal/ingestor"
	"microtransit-go/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	conn := os.Getenv("DATABASE_URL")
	if conn == "" { conn = "postgres://postgres:postgres@localhost:5432/microtransit?sslmode=disable" }

	s, err := store.New(conn)
	if err != nil { log.Fatal(err) }

	ing := ingestor.New(s)
	ing.StartWorkers(20)

	busH := handler.NewBusHandler(s, ing)
	roomH := &handler.RoomHandler{Store: s}
	etaClient := eta.New(os.Getenv("PYTHON_ETA_URL"))
	stopH := &handler.StopHandler{Store: s, Eta: etaClient}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/bus/location", busH.UpdateLocation)
	mux.HandleFunc("GET /api/buses", busH.GetAllBuses)
	mux.HandleFunc("POST /api/rooms", roomH.Create)
	mux.HandleFunc("GET /api/rooms/{room_id}", roomH.Get)
	mux.HandleFunc("POST /api/rooms/{room_id}/stops", stopH.Add)
	mux.HandleFunc("GET /api/rooms/{room_id}/stops", stopH.List)
	mux.HandleFunc("GET /api/rooms/{room_id}/eta", stopH.GetETA)
	mux.Handle("/", http.FileServer(http.Dir("static")))

	log.Println("Server on :8080 RAM ~38MB")
	log.Fatal(http.ListenAndServe(":8080", mux))
}