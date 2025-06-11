package services

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"sync"
	"time"

	"github.com/MPoline/alert_service_yp/internal/agent/flags"
	"github.com/MPoline/alert_service_yp/internal/hasher"
	"github.com/MPoline/alert_service_yp/internal/models"
	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

var (
	serverURL = "http://localhost:8080/updates"
	m         models.Metrics
)

func addNewMetrics() (float64, float64, []float64) {
	vmStat, _ := mem.VirtualMemory()
	CPUutilization, _ := cpu.Percent(time.Second, true)

	totalMemoryMB := float64(vmStat.Total) / (1024.0 * 1024.0)
	freeMemoryMB := float64(vmStat.Free) / (1024.0 * 1024.0)

	return totalMemoryMB, freeMemoryMB, CPUutilization
}

func CreateMetrics(s *storage.MemStorage) (metricsStorage []models.Metrics) {
	var wg sync.WaitGroup
	resultCh := make(chan models.Metrics, len(s.Gauges)+len(s.Counters)+3)

	wg.Add(3)

	go func() {
		defer wg.Done()
		s.Mu.Lock()
		defer s.Mu.Unlock()

		for gaugeName, gaugeValue := range s.Gauges {
			m = models.Metrics{
				ID:    gaugeName,
				MType: "gauge",
				Value: &gaugeValue,
			}
			resultCh <- m
		}
	}()

	go func() {
		defer wg.Done()
		s.Mu.Lock()
		defer s.Mu.Unlock()

		for counterName, counterValue := range s.Counters {
			m = models.Metrics{
				ID:    counterName,
				MType: "counter",
				Delta: &counterValue,
			}
			resultCh <- m
		}
	}()

	go func() {
		defer wg.Done()
		totalMemory, freeMemory, cpuUtilizations := addNewMetrics()

		metrics := []struct {
			id    string
			value interface{}
		}{
			{"TotalMemory", totalMemory},
			{"FreeMemory", freeMemory},
			{"CPUutilization1", cpuUtilizations},
		}

		for _, newMetric := range metrics {
			var m models.Metrics

			switch v := newMetric.value.(type) {
			case float64:
				m = models.Metrics{
					ID:    newMetric.id,
					MType: "gauge",
					Value: &v,
				}
			case []float64:
				m = models.Metrics{
					ID:    newMetric.id,
					MType: "gauge",
					Value: &v[0],
				}
			default:
				zap.L().Error("Unsupported type")
			}
			resultCh <- m
		}
	}()

	wg.Wait()
	close(resultCh)

	for metric := range resultCh {
		metricsStorage = append(metricsStorage, metric)
	}
	return metricsStorage
}

func SendMetrics(s *storage.MemStorage, metricsStorage []models.Metrics) {
	zap.L().Info("Start SendMetrics")
	client := resty.New()

	batch := map[string][]models.Metrics{"metrics": metricsStorage}
	intervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	jsonBody, err := json.Marshal(batch)
	if err != nil {
		zap.L().Error("Failed to marshal batch of metrics: ", zap.Error(err))
		return
	}

	h := hasher.InitHasher("SHA256")
	hash, err := h.CalculateHash(jsonBody, []byte(flags.FlagKey))
	if err != nil {
		zap.L().Error("Failed calculate sha256: ", zap.Error(err))
		return
	}
	hashStr := base64.StdEncoding.EncodeToString(hash)
	zap.L().Info("hash request: ", zap.String("hashStr", hashStr))

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
			SetHeader("HashSHA256", base64.StdEncoding.EncodeToString(hash)).
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
