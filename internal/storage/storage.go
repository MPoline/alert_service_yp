package storage

import (
	"fmt"
	"math/rand/v2"
	"net/http/httputil"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type MemStorage struct {
	mu       sync.Mutex
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

var (
	serverURL = "http://localhost:8080/update"
	nRetries  = 3
)

//SERVER
//----------------------------------------------

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, exists := s.Gauges[name]
	return value, exists
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, exists := s.Counters[name]
	return value, exists
}

func (s *MemStorage) SetGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Gauges[name] = value
}

func (s *MemStorage) IncrementCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Counters[name] += value
}

//CLIENT
//----------------------------------------------

func GetMetrics(MemStorage *MemStorage, neсMetrics []string) {
	MemStorage.mu.Lock()
	defer MemStorage.mu.Unlock()

	var MemStat runtime.MemStats
	runtime.ReadMemStats(&MemStat)

	MemStatType := reflect.TypeOf(MemStat)
	MemStatValue := reflect.ValueOf(MemStat)

	for i := 0; i < MemStatType.NumField(); i++ {
		fieldName := MemStatType.Field(i).Name
		fieldValue := MemStatValue.Field(i)

		for _, metricName := range neсMetrics {
			if fieldName == metricName {
				if value, ok := fieldValue.Interface().(float64); ok {
					MemStorage.Gauges[fieldName] = value
				} else if intValue, ok := fieldValue.Interface().(int64); ok {
					MemStorage.Gauges[fieldName] = float64(intValue)
				} else if uintValue, ok := fieldValue.Interface().(uint64); ok {
					MemStorage.Gauges[fieldName] = float64(uintValue)
				} else if uint32Value, ok := fieldValue.Interface().(uint32); ok {
					MemStorage.Gauges[fieldName] = float64(uint32Value)
				}
				break
			}
		}
	}
	MemStorage.Counters["PollCount"]++
	MemStorage.Gauges["RandomValue"] = rand.Float64()
}

func CreateURL(MemStorage *MemStorage) (URLStorage []string) {
	MemStorage.mu.Lock()
	defer MemStorage.mu.Unlock()

	for key, value := range MemStorage.Gauges {
		url := fmt.Sprintf("%s/gauge/%s/%f", serverURL, key, value)
		URLStorage = append(URLStorage, url)
	}

	for key, value := range MemStorage.Counters {
		url := fmt.Sprintf("%s/counter/%s/%d", serverURL, key, value)
		URLStorage = append(URLStorage, url)
	}
	return
}

func SendMetrics(URLStorage []string) {
	client := resty.New()

	for _, URL := range URLStorage {
		nAttempts := 0
		for nAttempts < nRetries {
			req := client.R().SetHeader("Content-Type", "text/plain")
			req.Body = ""
			req.URL = URL
			req.Method = "POST"
			resp, err := req.Send()

			if err != nil {
				fmt.Println("Error sending request:", err)
				nAttempts++
				time.Sleep(2 * time.Second)
				dump, _ := httputil.DumpRequest(req.RawRequest, true)
				fmt.Printf("Оригинальный запрос:\n\n%s", dump)
				continue
			}
			if resp.IsError() {
				fmt.Println("Error response:", resp.Status())
			}
			break
		}
		if nAttempts == nRetries {
			fmt.Println("All retries failed for URL:", URL)
		}
	}
}
