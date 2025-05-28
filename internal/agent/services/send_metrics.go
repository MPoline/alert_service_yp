package services

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"time"

	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var (
	serverURL = "http://localhost:8080/update"
	nRetries  = 3
	m         storage.Metrics
)

func CreateMetrics(s *storage.MemStorage) (metricsStorage []storage.Metrics) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for gaugeName, gaugeValue := range s.Gauges {
		m = storage.Metrics{
			ID:    gaugeName,
			MType: "gauge",
			Value: &gaugeValue,
		}
		metricsStorage = append(metricsStorage, m)
	}

	for counterName, counterValue := range s.Counters {
		m = storage.Metrics{
			ID:    counterName,
			MType: "counter",
			Delta: &counterValue,
		}
		metricsStorage = append(metricsStorage, m)
	}
	return
}

func SendMetrics(s *storage.MemStorage, metricsStorage []storage.Metrics) {
	client := resty.New()

	for _, metric := range metricsStorage {

		jsonBody, err := json.Marshal(metric)
		if err != nil {
			zap.L().Error("Failed to encode metric: ", zap.Error(err))
			return
		}

		var buff bytes.Buffer
		gz := gzip.NewWriter(&buff)
		defer gz.Close()

		_, err = gz.Write(jsonBody)
		if err != nil {
			zap.L().Info("Failed to compress data: ", zap.Error(err))
			continue
		}
		gz.Close()
		compressedData := buff.Bytes()

		nAttempts := 0
		for nAttempts < nRetries {
			req := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Content-Encoding", "gzip").
				SetBody(compressedData)

			resp, err := req.Post(serverURL)
			if err != nil || resp.IsError() {
				nAttempts++
				time.Sleep(time.Second * 2)
				continue
			}
			break
		}
	}
}
