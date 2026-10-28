package handler

import (
	"encoding/json"
	"net/http"

	"microtransit-go/internal/ingestor"
	"microtransit-go/internal/models"
	"microtransit-go/internal/store"
)

type Handler struct {
	store    *store.Store
	ingestor *ingestor.Ingestor
}

func New(s *store.Store, ing *ingestor.Ingestor) *Handler {
	return &Handler{store: s, ingestor: ing}
}

// POST /api/bus/location
// Body: {"bus_id":"MYS-101","lat":12.2958,"lng":76.6394,"speed":30}
func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "use POST", 405)
		return
	}

	var loc models.Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "bad json", 400)
		return
	}

	// WHY NOT h.store.SaveLocation here?
	// Because if 500 buses POST at same time, 500 DB writes = crash.
	// We just put in queue and return in 1ms. Worker saves later.
	h.ingestor.Submit(loc)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
}

// GET /api/buses -> returns all latest locations for the map
func (h *Handler) GetAllBuses(w http.ResponseWriter, r *http.Request) {
	locs, err := h.store.GetAllLatest(r.Context())
	if err != nil {
		http.Error(w, "db error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locs)
}