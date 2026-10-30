package ingestor

import (
	"context"
	"log"

	"microtransit-go/internal/models"
	"microtransit-go/internal/store"
)

// Ingestor is a queue that holds GPS pings
type Ingestor struct {
	ch    chan models.Location // buffered channel = queue
	store *store.Store
}

func New(s *store.Store) *Ingestor {
	return &Ingestor{
		// WHY 1000? If DB is slow for 2 sec, we can still collect 1000 pings
		// without dropping. After 1000, handler will wait (backpressure)
		// This prevents memory from growing to infinity = low RAM
		ch:    make(chan models.Location, 30000),
		store: s,
	}
}

// Submit is called by HTTP handler - super fast, just puts in queue
func (ing *Ingestor) Submit(loc models.Location) {
	ing.ch <- loc // if queue full, this line waits until a worker frees space
}

// StartWorkers starts N goroutines that work forever
func (ing *Ingestor) StartWorkers(n int) {
	for i := 0; i < n; i++ {
		go ing.worker(i) // WHY go? goroutine = 2KB RAM vs thread = 1MB RAM
	}
}

// worker reads from queue and saves to DB
func (ing *Ingestor) worker(id int) {
	// range ch = keep reading forever until channel closed
	for loc := range ing.ch {
		err := ing.store.SaveLocation(context.Background(), loc)
		if err != nil {
			log.Printf("worker %d: failed to save %s: %v", id, loc.BusID, err)
		}
	}
}