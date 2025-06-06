package main

import (
	"sync"
	"time"

	"github.com/MPoline/alert_service_yp/internal/agent/flags"
	"github.com/MPoline/alert_service_yp/internal/agent/services"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"go.uber.org/zap"
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
	flags.ParseFlags()
	pollInterval := time.Duration(flags.FlagPollInterval) * time.Second
	reportInterval := time.Duration(flags.FlagReportInterval) * time.Second
	wg.Add(2)

	// Сбор метрик
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for range ticker.C {
			services.GetMetrics(memStorage, neсMetrics)

			for key, value := range memStorage.Gauges {
				zap.L().Info("Gauges: ", zap.String("key", key), zap.Float64("value", value))
			}

			for key, value := range memStorage.Counters {
				zap.L().Info("Counters: ", zap.String("key", key), zap.Int64("value", value))
			}
		}
	}()

	// Отправка метрик
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()

		for range ticker.C {
			metricStorage := services.CreateMetrics(memStorage)
			services.SendMetrics(memStorage, metricStorage)
		}
	}()

	wg.Wait()
}
