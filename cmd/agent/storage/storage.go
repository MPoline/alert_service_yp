package storage

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"net/http"
	"reflect"
	"runtime"
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
	for _, URL := range URLStorage {
		req, err := http.NewRequest("POST", URL, bytes.NewBuffer([]byte("")))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		req.Header.Set("Content-Type", "text/plain")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Error response:", resp.Status)
		}
	}
}
