package main

import (
	"fmt"
	"sync"
	"time"

	services "github.com/MPoline/alert_service_yp/cmd/agent/services"
	f "github.com/MPoline/alert_service_yp/internal/agent"
	storage "github.com/MPoline/alert_service_yp/internal/storage"
)

var (
	neсMetrics = []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction",
		"GCSys", "HeapAlloc", "HeapIdle", "HeapInuse",
		"HeapObjects", "HeapReleased", "HeapSys", "LastGC",
		"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse",
		"MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
		"StackSys", "Sys", "TotalAlloc",
	}
	memStorage = storage.NewMemStorage()
	wg         sync.WaitGroup
)

func main() {
	f.ParseFlags()
	pollInterval := time.Duration(f.FlagPollInterval) * time.Second
	reportInterval := time.Duration(f.FlagReportInterval) * time.Second
	wg.Add(2)

	// Сбор метрик
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for range ticker.C {
			services.GetMetrics(memStorage, neсMetrics)

			fmt.Println("Gauges:")
			for key, value := range memStorage.Gauges {
				fmt.Printf("%s: %v\n", key, value)
			}

			fmt.Println("Counters:")
			for key, value := range memStorage.Counters {
				fmt.Printf("%s: %v\n", key, value)
			}
		}
	}()

	// Отправка метрик
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()

		for range ticker.C {
			URLStorage := services.CreateURLS(memStorage)
			services.SendMetrics(memStorage, URLStorage)
		}
	}()

	wg.Wait()
}
