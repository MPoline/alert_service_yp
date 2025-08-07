// Package models содержит основные структуры данных и валидацию для работы с метриками.
//
// Пакет предоставляет:
// - Модели данных для хранения метрик
// - Функции валидации метрик
// - Стандартные ошибки валидации
package models

import "errors"

// SliceMetrics представляет коллекцию метрик для batch-обработки.
// Используется для массового обновления/чтения метрик.
//
// Пример JSON:
//
//	{
//	  "metrics": [
//	    {"id": "metric1", "type": "gauge", "value": 123.45},
//	    {"id": "metric2", "type": "counter", "delta": 42}
//	  ]
//	}
type SliceMetrics struct {
	Metrics []Metrics `json:"metrics"`
}

// Metrics представляет отдельную метрику системы.
// Поля:
//   - ID: уникальное имя метрики (обязательное поле)
//   - MType: тип метрики - "gauge" или "counter" (обязательное поле)
//   - Delta: значение для counter-метрик (опциональное, должно быть nil для gauge)
//   - Value: значение для gauge-метрик (опциональное, должно быть nil для counter)
//
// Примеры JSON:
//
//	{"id": "temperature", "type": "gauge", "value": 23.5}
//	{"id": "requests", "type": "counter", "delta": 10}
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// Стандартные ошибки валидации метрик
var (
	// ErrInvalidMetricName возвращается при пустом имени метрики
	ErrInvalidMetricName = errors.New("InvalidMetricName")

	// ErrInvalidMetricType возвращается при недопустимом типе метрики
	ErrInvalidMetricType = errors.New("InvalidMetricType")

	// ErrInvalidCounterValue возвращается при некорректных значениях для counter
	ErrInvalidCounterValue = errors.New("InvalidCounterValue")

	// ErrInvalidGaugeValue возвращается при некорректных значениях для gauge
	ErrInvalidGaugeValue = errors.New("InvalidGaugeValue")
)

// IsValid проверяет корректность метрики.
//
// Правила валидации:
//   - ID не должно быть пустым
//   - MType должен быть "gauge" или "counter"
//   - Для gauge-метрик должно быть задано Value и не должно быть Delta
//   - Для counter-метрик должно быть задано Delta и не должно быть Value
//
// Возвращает:
//   - bool: true если метрика валидна
//   - error: конкретная ошибка валидации
//
// Пример использования:
//
//	metric := Metrics{ID: "temp", MType: "gauge", Value: 23.5}
//	if valid, err := metric.IsValid(); !valid {
//	    log.Fatal(err)
//	}
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
