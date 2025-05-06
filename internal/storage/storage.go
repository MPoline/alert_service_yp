package storage

import (
	"fmt"
	"sync"
)

type MemStorage struct {
	Mu       sync.Mutex
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

//SERVER
//----------------------------------------------

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	value, exists := s.Gauges[name]
	return value, exists
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	value, exists := s.Counters[name]
	return value, exists
}

func (s *MemStorage) SetGauge(name string, value float64) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Gauges[name] = value
}

func (s *MemStorage) IncrementCounter(name string, value int64) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Counters[name] += value
}

func (s *MemStorage) String() string {
    s.Mu.Lock()
    defer s.Mu.Unlock()

    var gaugesStr, countersStr string

    for k, v := range s.Gauges {
        gaugesStr += fmt.Sprintf("%v=%v ", k, v)
    }

    for k, v := range s.Counters {
        countersStr += fmt.Sprintf("%v=%d ", k, v)
    }

    return fmt.Sprintf("Gauges(%v), Counters(%v)", gaugesStr, countersStr)
}