package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"microtransit-go/internal/models"
)

// Store holds the DB connection pool
type Store struct {
	pool *pgxpool.Pool
}

// New creates pool with LOW RAM config
func New(connStr string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	// WHY 5 connections only?
	// Default is 20-30. Each connection = ~5MB RAM.
	// 5 conns = 25MB, can handle 500 buses because our worker.go queues them.
	// This is how we stay low RAM.
	cfg.MaxConns = 5
	cfg.MinConns = 1

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	return &Store{pool: pool}, nil
}

// SaveLocation = INSERT or UPDATE latest location
func (s *Store) SaveLocation(ctx context.Context, loc models.Location) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO locations(bus_id, lat, lng, speed)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (bus_id) 
		DO UPDATE SET lat=$2, lng=$3, speed=$4, updated_at=NOW()
	`, loc.BusID, loc.Lat, loc.Lng, loc.Speed)
	return err
}

// GetAllLatest = for the map
func (s *Store) GetAllLatest(ctx context.Context) ([]models.Location, error) {
	rows, err := s.pool.Query(ctx, `SELECT bus_id, lat, lng, speed FROM locations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locs []models.Location
	for rows.Next() {
		var l models.Location
		rows.Scan(&l.BusID, &l.Lat, &l.Lng, &l.Speed)
		locs = append(locs, l)
	}
	return locs, nil
}