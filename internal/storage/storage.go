package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
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

func (s *MemStorage) SaveToFile(filePath string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	data := struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}{
		Gauges:   s.Gauges,
		Counters: s.Counters,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marsal data: %v\n", err)
		return err
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error open file: %v\n", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Printf("Error write file: %v\n", err)
		return err
	}

	return nil
}

func (s *MemStorage) LoadFromFile(filePath string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error open file: %v\n", err)
		return err
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error read file: %v\n", err)
		return err
	}

	var data struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		log.Printf("Error unmarshal JSON: %v\n", err)
		return err
	}

	s.Gauges = data.Gauges
	s.Counters = data.Counters

	return nil
}
