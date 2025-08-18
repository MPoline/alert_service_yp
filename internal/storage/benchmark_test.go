package storage_test

import (
	"testing"

	"github.com/MPoline/alert_service_yp/internal/storage"
)

func BenchmarkMemoryStorage_SetGauge(b *testing.B) {
	ms := storage.NewMemStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.SetGauge("testMetric", 1.23)
	}
}

func BenchmarkMemoryStorage_GetGauge(b *testing.B) {
	ms := storage.NewMemStorage()
	ms.SetGauge("testMetric", 1.23)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.GetGauge("testMetric")
	}
}

func BenchmarkMemoryStorage_SetCounter(b *testing.B) {
	ms := storage.NewMemStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.IncrementCounter("testCounter", 1)
	}
}

func BenchmarkMemoryStorage_GetCounter(b *testing.B) {
	ms := storage.NewMemStorage()
	ms.IncrementCounter("testCounter", 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.GetCounter("testCounter")
	}
}
