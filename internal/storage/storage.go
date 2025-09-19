package storage

import (
	"context"

	"github.com/MPoline/alert_service_yp/internal/models"
)

type Storage interface {
	GetAllMetrics(ctx context.Context) ([]models.Metrics, error)
	GetMetric(ctx context.Context, metricType string, metricName string) (models.Metrics, error)
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	UpdateSliceOfMetrics(ctx context.Context, sliceMitrics models.SliceMetrics) error
	Close()
}

// NewStorage создает и возвращает экземпляр хранилища указанного типа
func NewStorage(storageType string) Storage {
	switch storageType {
	case "memory":
		return NewMemStorage()
	case "database":
		return NewDBStorage()
	default:
		return nil
	}
}

// SaveToFile сохраняет данные в файл (только для MemStorage)
func SaveToFile(s Storage, filePath string) error {
	if ms, ok := s.(*MemStorage); ok {
		return ms.SaveToFile(filePath)
	}
	return nil
}

// LoadFromFile загружает данные из файла (только для MemStorage)
func LoadFromFile(s Storage, filePath string) error {
	if ms, ok := s.(*MemStorage); ok {
		return ms.LoadFromFile(filePath)
	}
	return nil
}
