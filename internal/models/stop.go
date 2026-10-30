package models

type Stop struct {
  ID int `json:"id"`
  RoomID string `json:"room_id"`
  Name string `json:"name"`
  Lat float64 `json:"lat"`
  Lng float64 `json:"lng"`
  Order int `json:"stop_order"`
  DistFromPrev float64 `json:"dist_from_prev"` // km
  DistToFinal float64 `json:"dist_to_final"`   // km to final
}