package models

import "time"

type Room struct {
  ID string `json:"id"`
  Name string `json:"name"`
  AdminToken string `json:"-"` // never send to viewer
  CreatedAt time.Time `json:"created_at"`
}