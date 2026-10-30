package handler

import (
	"encoding/json"
	"net/http"
	"microtransit-go/internal/eta"
	"microtransit-go/internal/geo"
	"microtransit-go/internal/store"
)

type StopHandler struct {
	Store *store.Store
	Eta   *eta.Client
}

// POST /api/rooms/{room_id}/stops - admin only
func (h *StopHandler) Add(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	token := r.Header.Get("X-Admin-Token")
	room, err := h.Store.GetRoom(roomID)
	if err != nil || room.AdminToken != token {
		http.Error(w, "only admin/driver can add stops", 403)
		return
	}
	var in struct {
		Name string  `json:"name"`
		Lat  float64 `json:"lat"`
		Lng  float64 `json:"lng"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	st, err := h.Store.AddStop(roomID, in.Name, in.Lat, in.Lng)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(st)
}

// GET /api/rooms/{room_id}/stops
func (h *StopHandler) List(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	stops, _ := h.Store.GetStops(roomID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stops)
}

// GET /api/rooms/{room_id}/eta - this is your 【entity-Google Maps¦canonical_name=Google Maps】 logic
func (h *StopHandler) GetETA(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room_id")
	stops, _ := h.Store.GetStops(roomID)
	if len(stops) == 0 {
		json.NewEncoder(w).Encode(map[string]any{"total_km": 0, "remaining_km": 0, "next_km": 0})
		return
	}
	var total float64
	for _, s := range stops {
		total += s.DistFromPrev
	}
	buses, _ := h.Store.GetAllLatest(r.Context())
	var busLat, busLng, busSpeed float64
	var found bool
	for _, b := range buses {
		if b.BusID == roomID {
			busLat = b.Lat
			busLng = b.Lng
			busSpeed = b.Speed
			found = true
		}
	}
	if !found {
		busLat = stops[0].Lat
		busLng = stops[0].Lng
		busSpeed = 20
	}
	nextIdx := len(stops) - 1
	minD := 999999.0
	for i, s := range stops {
		d := geo.Haversine(busLat, busLng, s.Lat, s.Lng)
		if d < minD {
			minD = d
			nextIdx = i
		}
	}
	if minD < 0.05 && nextIdx < len(stops)-1 {
		nextIdx++
	}
	next := stops[nextIdx]
	distNext := geo.Haversine(busLat, busLng, next.Lat, next.Lng)
	remaining := distNext + next.DistToFinal
	etaNext, _ := h.Eta.Predict(distNext, busSpeed)
	etaFinal, _ := h.Eta.Predict(remaining, busSpeed)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"total_km": total, "remaining_km": remaining,
		"next_km": distNext, "next_name": next.Name, "next_idx": nextIdx,
		"next_eta_min": etaNext, "final_eta_min": etaFinal,
	})
}
