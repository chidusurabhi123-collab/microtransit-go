package handler

import (
	"encoding/json"
	"net/http"
	"microtransit-go/internal/store"
)

type RoomHandler struct{ Store *store.Store }

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name string `json:"name"` }
	json.NewDecoder(r.Body).Decode(&in)
	room, err := h.Store.CreateRoom(in.Name)
	if err != nil { http.Error(w, err.Error(), 500); return }
	json.NewEncoder(w).Encode(map[string]string{"room_id": room.ID, "admin_token": room.AdminToken})
}

func (h *RoomHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("room_id")
	room, err := h.Store.GetRoom(id)
	if err != nil { http.Error(w, "not found", 404); return }
	json.NewEncoder(w).Encode(room)
}