package services

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/MPoline/alert_service_yp/internal/hasher"
	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/server/flags"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UpdateMetricFromJSON обрабатывает запрос на обновление метрики в JSON-формате с проверкой подписи.
//
// Эндпоинт: POST /update/
//
// Логика работы:
//  1. Читает и проверяет тело запроса
//  2. Валидирует подпись HMAC-SHA256
//  3. Обновляет метрику в хранилище
//  4. Возвращает обновленную метрику с заголовками
//
// Формат запроса:
//
//	{
//	  "id": "metricName",
//	  "type": "gauge|counter",
//	  "value": 123.45,    // для gauge
//	  "delta": 42         // для counter
//	}
//
// Заголовки:
//   - HashSHA256: обязательная подпись тела запроса (base64)
//   - Content-Type: application/json
//
// Возможные ответы:
//   - 200 OK: успешное обновление
//     Тело: обновленная метрика в JSON
//   - 400 Bad Request: неверный формат, ошибка валидации
//   - 404 Not Found: обязательные параметры отсутствуют
//   - 500 Internal Server Error: ошибка сервера
//
// Пример:
//
//	Запрос:
//	  POST /update/
//	  Headers:
//	    Content-Type: application/json
//	    HashSHA256: <base64-hmac-sha256>
//	  Body:
//	    {"id":"alloc","type":"gauge","value":123.45}
//
//	Ответ:
//	  {"id":"alloc","type":"gauge","value":123.45}
func UpdateMetricFromJSON(c *gin.Context) {

	var (
		req models.Metrics
	)

	ctx := c.Request.Context()

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		zap.L().Error("Error read request: ", zap.Error(err))
	}

	err = json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid input data: %v", err)})
		zap.L().Error("Error unmarshal request: ", zap.Error(err))
		return
	}

	h := hasher.InitHasher("SHA256")
	hash, err := h.CalculateHash(data, []byte(flags.FlagKey))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Failed calculate sha256"})
		zap.L().Error("Failed calculate sha256: ", zap.Error(err))
		return
	}

	hashFromHeader, err := base64.StdEncoding.DecodeString(c.Request.Header.Get("HashSHA256"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Failed to decode hash"})
		zap.L().Error("Failed to decode hash: ", zap.Error(err))
		return
	}

	if !(hmac.Equal(hash, hashFromHeader)) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Signature hash does not match"})
		zap.L().Error("Signature hash does not match: ", zap.Error(err))
		return
	}

	err = storage.MetricStorage.UpdateMetric(ctx, req)

	if err != nil {
		if errors.Is(err, models.ErrInvalidMetricName) || errors.Is(err, models.ErrInvalidMetricType) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		if errors.Is(err, models.ErrInvalidCounterValue) || errors.Is(err, models.ErrInvalidGaugeValue) {
			c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	respBytes, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to encode response"})
		zap.L().Error("Failed to encode response: ", zap.Error(err))
		return
	}

	c.Header("Content-Type", c.GetHeader("Content-Type"))
	c.Header("Content-Length", c.GetHeader("Content-Length"))
	c.Header("Date", time.Now().UTC().Format(http.TimeFormat))
	c.Header("HashSHA256", base64.StdEncoding.EncodeToString(hash))
	c.String(http.StatusOK, string(respBytes))
}

// UpdateMetricFromURL обрабатывает запрос на обновление метрики через URL параметры.
//
// Эндпоинт: POST /update/:type/:name/:value
//
// Логика работы:
//  1. Извлекает параметры из URL
//  2. Парсит значение в соответствии с типом метрики
//  3. Обновляет метрику в хранилище
//
// Параметры URL:
//   - type: "gauge" или "counter"
//   - name: имя метрики
//   - value: значение (float64 для gauge, int64 для counter)
//
// Возможные ответы:
//   - 200 OK: успешное обновление
//   - 400 Bad Request: неверный тип или значение
//   - 404 Not Found: отсутствуют обязательные параметры
//
// Примеры:
//
//	Запрос:
//	  POST /update/gauge/alloc/123.45
//
//	Запрос:
//	  POST /update/counter/pollCount/1
func UpdateMetricFromURL(c *gin.Context) {

	var req models.Metrics

	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	ctx := c.Request.Context()

	if metricType == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric type is required"})
		zap.L().Info("Metric type is required")
		return
	}
	if metricName == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric name is required"})
		zap.L().Info("Metric name is required")
		return
	}
	if metricValue == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric value is required"})
		zap.L().Info("Metric value is required")
		return
	}

	if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid gauge value"})
			zap.L().Info("Invalid gauge value: ", zap.Error(err))
			return
		}
		req = models.Metrics{
			ID:    metricName,
			MType: metricType,
			Value: &value,
		}
	}

	if metricType == "counter" {
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid counter value"})
			zap.L().Info("Invalid counter value: ", zap.Error(err))
			return
		}
		req = models.Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: &delta,
		}
	}

	err := storage.MetricStorage.UpdateMetric(ctx, req)
	if err != nil {
		if errors.Is(err, models.ErrInvalidMetricType) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "InvalidMetricType"})
			return
		}
		if errors.Is(err, models.ErrInvalidCounterValue) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "InvalidCounterValue"})
			return
		}
		if errors.Is(err, models.ErrInvalidGaugeValue) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "InvalidGaugeValue"})
			return
		}
		if errors.Is(err, models.ErrInvalidMetricName) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "InvalidMetricName"})
			return
		}
	}

	c.Header("Content-Type", c.GetHeader("Content-Type"))
	c.Header("Content-Length", c.GetHeader("Content-Length"))
	c.Header("Date", time.Now().UTC().Format(http.TimeFormat))
	c.Status(http.StatusOK)
}
