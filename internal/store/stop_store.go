package store

import (
	"context"
	"microtransit-go/internal/geo"
	"microtransit-go/internal/models"
)

func (s *Store) AddStop(roomID, name string, lat, lng float64) (*models.Stop, error) {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var lastLat, lastLng float64
	var lastOrder int
	err = tx.QueryRow(ctx, `SELECT lat,lng,stop_order FROM stops WHERE room_id=$1 ORDER BY stop_order DESC LIMIT 1 FOR UPDATE`, roomID).Scan(&lastLat, &lastLng, &lastOrder)

	order := 1
	distFromPrev := 0.0
	if err == nil {
		order = lastOrder + 1
		distFromPrev = geo.Haversine(lastLat, lastLng, lat, lng)
		tx.Exec(ctx, `UPDATE stops SET dist_to_final = dist_to_final + $1 WHERE room_id=$2`, distFromPrev, roomID)
	}

	var st models.Stop
	err = tx.QueryRow(ctx, `INSERT INTO stops(room_id,name,lat,lng,stop_order,dist_from_prev,dist_to_final) VALUES($1,$2,$3,$4,$5,$6,0) RETURNING id,room_id,name,lat,lng,stop_order,dist_from_prev,dist_to_final`,
		roomID, name, lat, lng, order, distFromPrev).Scan(&st.ID, &st.RoomID, &st.Name, &st.Lat, &st.Lng, &st.Order, &st.DistFromPrev, &st.DistToFinal)
	if err != nil {
		return nil, err
	}
	tx.Commit(ctx)
	return &st, nil
}

func (s *Store) GetStops(roomID string) ([]models.Stop, error) {
	rows, err := s.pool.Query(context.Background(), `SELECT id,room_id,name,lat,lng,stop_order,dist_from_prev,dist_to_final FROM stops WHERE room_id=$1 ORDER BY stop_order`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stops []models.Stop
	for rows.Next() {
		var st models.Stop
		rows.Scan(&st.ID, &st.RoomID, &st.Name, &st.Lat, &st.Lng, &st.Order, &st.DistFromPrev, &st.DistToFinal)
		stops = append(stops, st)
	}
	return stops, nil
}