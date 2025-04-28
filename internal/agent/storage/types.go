package storage

import (
	"sync"
)

type MemStorage struct {
	mu       sync.Mutex
	Gauges   map[string]float64
	Counters map[string]int64
}
