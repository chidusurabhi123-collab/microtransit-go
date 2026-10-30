package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"microtransit-go/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(connStr string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 20
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func genToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Store) SaveLocation(ctx context.Context, loc models.Location) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO bus_locations(bus_id, lat, lng, speed) VALUES($1,$2,$3,$4)
		 ON CONFLICT(bus_id) DO UPDATE SET lat=$2, lng=$3, speed=$4, updated_at=NOW()`,
		loc.BusID, loc.Lat, loc.Lng, loc.Speed)
	return err
}

func (s *Store) GetAllLatest(ctx context.Context) ([]models.Location, error) {
	rows, err := s.pool.Query(ctx, `SELECT bus_id, lat, lng, speed FROM bus_locations LIMIT 5000`)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []models.Location
	for rows.Next() {
		var l models.Location
		rows.Scan(&l.BusID, &l.Lat, &l.Lng, &l.Speed)
		out = append(out, l)
	}
	return out, nil
}