package services

// import (
// 	"testing"

// 	storage "github.com/MPoline/alert_service_yp/internal/storage"
// )

// func TestCreateURL(t *testing.T) {
// 	memStorage1 := storage.MemStorage{
// 		Gauges:   map[string]float64{"TestGauge": 0.123},
// 		Counters: map[string]int64{"TestCounter": 1},
// 	}
// 	tests := []struct {
// 		name       string
// 		memStorage *storage.MemStorage
// 		want       []string
// 	}{
// 		{
// 			name: "Test create URL", memStorage: &memStorage1, want: []string{"http://localhost:8080/update/gauge/TestGauge/0.123000", "http://localhost:8080/update/counter/TestCounter/1"},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			metricStorage := CreateMetrics(test.memStorage)

// 			for _, metric := range metricStorage {
// 				found := false
// 				for _, wantMetric := range test.want {
// 					if wantMetric == metric {
// 						found = true
// 					}
// 				}
// 				if !found {
// 					t.Errorf("Unexpecter URL %v", metric)
// 				}
// 			}
// 		})

// 	}
// }
