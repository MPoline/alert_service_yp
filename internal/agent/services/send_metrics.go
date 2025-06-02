package services

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"time"

	"github.com/MPoline/alert_service_yp/internal/models"
	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var (
	serverURL = "http://localhost:8080/updates"
	m         models.Metrics
)

func CreateMetrics(s *storage.MemStorage) (metricsStorage []models.Metrics) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for gaugeName, gaugeValue := range s.Gauges {
		m = models.Metrics{
			ID:    gaugeName,
			MType: "gauge",
			Value: &gaugeValue,
		}
		metricsStorage = append(metricsStorage, m)
	}

	for counterName, counterValue := range s.Counters {
		m = models.Metrics{
			ID:    counterName,
			MType: "counter",
			Delta: &counterValue,
		}
		metricsStorage = append(metricsStorage, m)
	}
	return
}

func SendMetrics(s *storage.MemStorage, metricsStorage []models.Metrics) {
	client := resty.New()

	batch := map[string][]models.Metrics{"metrics": metricsStorage}
	intervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	jsonBody, err := json.Marshal(batch)
	if err != nil {
		zap.L().Error("Failed to marshal batch of metrics: ", zap.Error(err))
		return
	}

	var buff bytes.Buffer
	gz := gzip.NewWriter(&buff)
	defer gz.Close()

	_, err = gz.Write(jsonBody)
	if err != nil {
		zap.L().Info("Failed to compress data: ", zap.Error(err))
		return
	}
	gz.Close()
	compressedData := buff.Bytes()

	for attempt, interval := range intervals {
		req := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(compressedData)

		resp, err := req.Post(serverURL)
		if err != nil || resp.IsError() {
			zap.L().Warn("Sending metrics failed on attempt", zap.Int("attempt", attempt+1),
				zap.Duration("interval", interval))
			time.Sleep(interval)
			continue
		}
		break
	}

}
