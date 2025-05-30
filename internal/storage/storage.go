package storage

import "github.com/MPoline/alert_service_yp/internal/models"

var MetricStorage Storage

type Storage interface {
	GetAllMetrics() ([]models.Metrics, error)
	GetMetric(metricType string, metricName string) (models.Metrics, error)
	UpdateMetric(metric models.Metrics) error
	UpdateSliceOfMetrics(sliceMitrics models.SliceMetrics) error
	Close()
}

func InitStorage(storageType string) {
	if storageType == "memory" {
		MetricStorage = NewMemStorage()
		return
	}
	if storageType == "database" {
		MetricStorage = NewDBStorage()
		return
	}
}

func SaveToFile(s Storage, filePath string) error {
	if ms, ok := s.(*MemStorage); ok {
		ms.SaveToFile(filePath)
	}
	return nil
}

func LoadFromFile(s Storage, filePath string) error {
	if ms, ok := s.(*MemStorage); ok {
		ms.LoadFromFile(filePath)
	}
	return nil
}
