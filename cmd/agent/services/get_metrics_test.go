package services

import (
	"testing"

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

func TestGetMetrics(t *testing.T) {
	s := storage.NewMemStorage()
	GetMetrics(s, neсMetrics)

	for _, metricName := range neсMetrics {
		if _, ok := s.Gauges[metricName]; !ok {
			t.Errorf("Metric %v not in runtime.", metricName)
		}
	}

	if _, ok := s.Gauges["RandomValue"]; !ok {
		t.Errorf("Metric RandomValue not exist.")
	}

	if _, ok := s.Counters["PollCount"]; !ok {
		t.Errorf("Metric PollCount not exist.")
	}
}
