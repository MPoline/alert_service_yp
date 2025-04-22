package storage

import (
	"testing"
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
	memStorage := NewMemStorage()
	GetMetrics(memStorage, neсMetrics)

	for _, metricName := range neсMetrics {
		if _, ok := memStorage.Gauges[metricName]; !ok {
			t.Errorf("Metric %v not in runtime.", metricName)
		}
	}

	if _, ok := memStorage.Gauges["RandomValue"]; !ok {
		t.Errorf("Metric RandomValue not exist.")
	}

	if _, ok := memStorage.Counters["PollCount"]; !ok {
		t.Errorf("Metric PollCount not exist.")
	}
}

func TestCreateURL(t *testing.T) {
	memStorage1 := MemStorage{
		Gauges:   map[string]float64{"TestGauge": 0.123},
		Counters: map[string]int64{"TestCounter": 1},
	}
	tests := []struct {
		name       string
		memStorage *MemStorage
		want       []string
	}{
		{
			name: "Test create URL", memStorage: &memStorage1, want: []string{"http://localhost:8080/update/gauge/TestGauge/0.123000", "http://localhost:8080/update/counter/TestCounter/1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urls := CreateURL(test.memStorage)

			for _, url := range urls {
				found := false
				for _, wantUrl := range test.want {
					if wantUrl == url {
						found = true
					}
				}
				if !found {
					t.Errorf("Unexpecter URL %v", url)
				}
			}
		})

	}
}
