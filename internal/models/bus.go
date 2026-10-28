package models

// Location is what a bus phone sends every 3 sec
type Location struct {
    BusID string  `json:"bus_id"`
    Lat   float64 `json:"lat"`
    Lng   float64 `json:"lng"`
    Speed float64 `json:"speed"`
}