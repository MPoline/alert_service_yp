package models

import "errors"

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m Metrics) IsValid() (bool, error) {
	if m.ID == "" {
		return false, errors.New("InvalidMetricName")
	}

	if m.MType == "gauge" {
		if m.Delta == nil && m.Value != nil {
			return true, nil
		}
		return false, errors.New("InvalidGaugeValue")
	}

	if m.MType == "counter" {
		if m.Value == nil && m.Delta != nil {
			return true, nil
		}
		return false, errors.New("InvalidCounterValue")
	}

	return false, errors.New("InvalidMetricType")
}
