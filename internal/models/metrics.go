package models

import "errors"

type SliceMetrics struct {
	Metrics []Metrics `json:"metrics"`
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var (
	ErrInvalidMetricName   = errors.New("InvalidMetricName")
	ErrInvalidMetricType   = errors.New("InvalidMetricType")
	ErrInvalidCounterValue = errors.New("InvalidCounterValue")
	ErrInvalidGaugeValue   = errors.New("InvalidGaugeValue")
)

func (m Metrics) IsValid() (bool, error) {
	if m.ID == "" {
		return false, ErrInvalidMetricName
	}

	if m.MType == "gauge" {
		if m.Delta == nil && m.Value != nil {
			return true, nil
		}
		return false, ErrInvalidGaugeValue
	}

	if m.MType == "counter" {
		if m.Value == nil && m.Delta != nil {
			return true, nil
		}
		return false, ErrInvalidCounterValue
	}

	return false, ErrInvalidMetricType
}
