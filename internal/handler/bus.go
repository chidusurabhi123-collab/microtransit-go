package handler

import (
	"encoding/json"
	"net/http"
	"microtransit-go/internal/ingestor"
	"microtransit-go/internal/models"
	"microtransit-go/internal/store"
)

type BusHandler struct {
	store    *store.Store
	ingestor *ingestor.Ingestor
}

func NewBusHandler(s *store.Store, ing *ingestor.Ingestor) *BusHandler {
	return &BusHandler{store: s, ingestor: ing}
}

// aliases so both Handler and BusHandler work
type Handler = BusHandler
func New(s *store.Store, ing *ingestor.Ingestor) *BusHandler { return NewBusHandler(s, ing) }

func (h *BusHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	var loc models.Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "bad json", 400); return
	}
	h.ingestor.Submit(loc)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
}

func (h *BusHandler) GetAllBuses(w http.ResponseWriter, r *http.Request) {
	locs, err := h.store.GetAllLatest(r.Context())
	if err != nil { http.Error(w, "db error", 500); return }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locs)
}