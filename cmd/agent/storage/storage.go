package storage

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
)

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

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

	serverURL := "http://localhost:8080/update"

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
	nRetries := 3

	for _, URL := range URLStorage {
		nAttempts := 0
		for nAttempts < nRetries {
			resp, err := client.R().SetHeader("Content-Type", "text/plain").Post(URL)

			if err != nil {
				fmt.Println("Error sending request:", err)
				nAttempts++
				time.Sleep(2 * time.Second)
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
