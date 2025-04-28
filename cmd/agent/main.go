package main

import (
	"fmt"
	"time"

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
)

func main() {
	f.ParseFlags()
	pollInterval := time.Duration(f.FlagPollInterval) * time.Second
	reportInterval := time.Duration(f.FlagReportInterval) * time.Second

	memStorage := storage.NewMemStorage()

	// Цикл обновления метрик
	go func() {
		for {
			storage.GetMetrics(memStorage, neсMetrics)

			fmt.Println("Gauges:")
			for key, value := range memStorage.Gauges {
				fmt.Printf("%s: %v\n", key, value)
			}

			fmt.Println("Counters:")
			for key, value := range memStorage.Counters {
				fmt.Printf("%s: %v\n", key, value)
			}

			time.Sleep(pollInterval)
		}
	}()

	// Цикл отправки метрик
	go func() {
		for {
			URLStorage := storage.CreateURL(memStorage)
			storage.SendMetrics(URLStorage)
			time.Sleep(reportInterval)
		}
	}()

	// Бесконечный цикл для работы программы
	select {}
}
