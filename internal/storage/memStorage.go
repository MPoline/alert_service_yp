package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/server/flags"
	"go.uber.org/zap"
)

type MemStorage struct {
	Mu       sync.Mutex
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

//SERVER
//----------------------------------------------

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	value, exists := s.Gauges[name]
	return value, exists
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	value, exists := s.Counters[name]
	return value, exists
}

func (s *MemStorage) SetGauge(name string, value float64) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Gauges[name] = value
}

func (s *MemStorage) IncrementCounter(name string, value int64) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Counters[name] += value
}

func (s *MemStorage) GetAllMetrics() ([]models.Metrics, error) {
	var allMetrics []models.Metrics

	for metricName, metricValue := range s.Gauges {
		metric := models.Metrics{
			ID:    metricName,
			MType: "gauge",
			Value: &metricValue,
		}
		allMetrics = append(allMetrics, metric)
	}

	for metricName, metricValue := range s.Counters {
		metric := models.Metrics{
			ID:    metricName,
			MType: "counter",
			Delta: &metricValue,
		}
		allMetrics = append(allMetrics, metric)
	}
	return allMetrics, nil
}

func (s *MemStorage) GetMetric(metricType string, metricName string) (models.Metrics, error) {
	var found bool
	var metric models.Metrics

	switch metricType {
	case "gauge":
		if value, ok := s.GetGauge(metricName); ok {
			metric.ID = metricName
			metric.MType = metricType
			metric.Value = &value
			found = true
			zap.L().Info("Found gauge", zap.Float64("value", value))
		}
	case "counter":
		if delta, ok := s.GetCounter(metricName); ok {
			metric.ID = metricName
			metric.MType = metricType
			metric.Delta = &delta
			found = true
			zap.L().Info("Found counter", zap.Int64("delta", delta))
		}
	default:
		zap.L().Info("Unknown metric type")
		err := errors.New("Unknown")
		return metric, err
	}

	if !found {
		zap.L().Info("Metric not found")
		err := errors.New("NotFound")
		return metric, err
	}
	return metric, nil
}

func (s *MemStorage) UpdateMetric(metric models.Metrics) error {

	_, err := metric.IsValid()
	if err != nil {
		zap.L().Info("Error in Metric Parametrs: ", zap.Error(err))
		return err
	}

	switch metric.MType {
	case "gauge":
		s.SetGauge(metric.ID, *metric.Value)
	case "counter":
		s.IncrementCounter(metric.ID, *metric.Delta)
	}
	return nil
}

func (s *MemStorage) SaveToFile(filePath string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	data := struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}{
		Gauges:   s.Gauges,
		Counters: s.Counters,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marsal data: %v\n", err)
		zap.L().Error("Error marsal data: ", zap.Error(err))
		return err
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		zap.L().Error("Error open file: ", zap.Error(err))
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		zap.L().Error("Error write file: ", zap.Error(err))
		return err
	}
	return nil
}

func (s *MemStorage) Close() {
	s.SaveToFile(flags.FlagFileStoragePath)
}

func (s *MemStorage) LoadFromFile(filePath string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		zap.L().Error("Error open file: ", zap.Error(err))
		return err
	}
	defer file.Close()

	jsonData, err := io.ReadAll(file)
	if err != nil {
		zap.L().Error("Error read file: ", zap.Error(err))
		return err
	}

	var data struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		zap.L().Error("Error unmarshal JSON: ", zap.Error(err))
		return err
	}

	s.Gauges = data.Gauges
	s.Counters = data.Counters

	return nil
}

func (s *MemStorage) String() string {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	var gaugesStr, countersStr string

	for k, v := range s.Gauges {
		gaugesStr += fmt.Sprintf("%v=%v ", k, v)
	}

	for k, v := range s.Counters {
		countersStr += fmt.Sprintf("%v=%d ", k, v)
	}

	return fmt.Sprintf("Gauges(%v), Counters(%v)", gaugesStr, countersStr)
}
