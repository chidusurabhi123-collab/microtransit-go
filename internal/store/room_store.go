package store

import (
	"context"
	"microtransit-go/internal/models"
)

func (s *Store) CreateRoom(name string) (*models.Room, error) {
	id := name
	ctx := context.Background()
	var r models.Room
	err := s.pool.QueryRow(ctx, `SELECT id,name,admin_token FROM rooms WHERE id=$1`, id).Scan(&r.ID, &r.Name, &r.AdminToken)
	if err == nil {
		return &r, nil // already exists, return REAL token
	}
	token := genToken()
	tag, err := s.pool.Exec(ctx, `INSERT INTO rooms(id,name,admin_token) VALUES($1,$2,$3) ON CONFLICT(id) DO NOTHING`, id, name, token)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		s.pool.QueryRow(ctx, `SELECT id,name,admin_token FROM rooms WHERE id=$1`, id).Scan(&r.ID, &r.Name, &r.AdminToken)
		return &r, nil
	}
	return &models.Room{ID: id, Name: name, AdminToken: token}, nil
}

func (s *Store) GetRoom(id string) (*models.Room, error) {
	var r models.Room
	err := s.pool.QueryRow(context.Background(), `SELECT id,name,admin_token FROM rooms WHERE id=$1`, id).Scan(&r.ID, &r.Name, &r.AdminToken)
	return &r, err
}