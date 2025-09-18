package services

import (
	"sync"
	"time"

	"github.com/MPoline/alert_service_yp/internal/models"
	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

var (
	m models.Metrics
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
