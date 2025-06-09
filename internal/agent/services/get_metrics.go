package services

import (
	"math/rand/v2"
	"reflect"
	"runtime"

	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"go.uber.org/zap"
)

func GetMetrics(s *storage.MemStorage, neсMetrics []string) {
	zap.L().Info("Start GetMetrics")
	s.Mu.Lock()
	defer s.Mu.Unlock()

	var MemStat runtime.MemStats
	runtime.ReadMemStats(&MemStat)

	MemStatType := reflect.TypeOf(MemStat)
	MemStatValue := reflect.ValueOf(MemStat)

	for i := 0; i < MemStatType.NumField(); i++ {
		fieldName := MemStatType.Field(i).Name
		fieldValue := MemStatValue.Field(i)

		for _, metricName := range neсMetrics {
			if fieldName == metricName {
				if value, ok := fieldValue.Interface().(float64); ok {
					s.Gauges[fieldName] = value
				} else if intValue, ok := fieldValue.Interface().(int64); ok {
					s.Gauges[fieldName] = float64(intValue)
				} else if uintValue, ok := fieldValue.Interface().(uint64); ok {
					s.Gauges[fieldName] = float64(uintValue)
				} else if uint32Value, ok := fieldValue.Interface().(uint32); ok {
					s.Gauges[fieldName] = float64(uint32Value)
				}
				break
			}
		}
	}
	s.Counters["PollCount"]++
	s.Gauges["RandomValue"] = rand.Float64()
}
