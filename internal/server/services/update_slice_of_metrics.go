package services

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MPoline/alert_service_yp/internal/hasher"
	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UpdateSliceOfMetrics обрабатывает запрос на массовое обновление метрик с проверкой подписи.
//
// Эндпоинт: POST /updates/
//
// Логика работы:
//  1. Читает тело запроса
//  2. Проверяет подпись HMAC-SHA256
//  3. Десериализует JSON в SliceMetrics
//  4. Обновляет метрики в хранилище (в транзакции)
//  5. Возвращает обновленные метрики
//
// Формат запроса:
//
//	{
//	  "metrics": [
//	    {"id": "metric1", "type": "gauge", "value": 123.45},
//	    {"id": "metric2", "type": "counter", "delta": 42}
//	  ]
//	}
//
// Заголовки:
//   - HashSHA256: обязательная подпись тела запроса (base64)
//   - Content-Type: application/json
//
// Возможные ответы:
//   - 200 OK: успешное обновление
//     Тело: обновленные метрики в JSON
//   - 400 Bad Request: неверный формат, ошибка валидации или подписи
//   - 404 Not Found: обязательные параметры отсутствуют
//   - 500 Internal Server Error: ошибка сервера
//
// Пример:
//
//	Запрос:
//	  POST /updates/
//	  Headers:
//	    Content-Type: application/json
//	    HashSHA256: <base64-hmac-sha256>
//	  Body:
//	    {
//	      "metrics": [
//	        {"id":"alloc","type":"gauge","value":123.45},
//	        {"id":"pollCount","type":"counter","delta":1}
//	      ]
//	    }
//
//	Ответ:
//	  {
//	    "metrics": [
//	      {"id":"alloc","type":"gauge","value":123.45},
//	      {"id":"pollCount","type":"counter","delta":1}
//	    ]
//	  }
func (h *ServiceHandler) UpdateSliceOfMetrics(c *gin.Context) {
	var req models.SliceMetrics

	ctx := c.Request.Context()

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to read request"})
		zap.L().Error("Error reading request body: ", zap.Error(err))
		return
	}

	// Используем ключ из ServiceHandler вместо глобального флага
	hasherInstance := hasher.InitHasher("SHA256")
	hash, err := hasherInstance.CalculateHash(data, []byte(h.key))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Failed calculate sha256"})
		zap.L().Error("Failed calculate sha256: ", zap.Error(err))
		return
	}

	zap.L().Info("===================================")
	hashStr := base64.StdEncoding.EncodeToString(hash)
	hashHeader := c.Request.Header.Get("HashSHA256")
	zap.L().Info("hash request: ", zap.String("hashStr", hashStr))
	zap.L().Info("hash from header: ", zap.String("hash from header", hashHeader))
	zap.L().Info("===================================")

	hashFromHeader, err := base64.StdEncoding.DecodeString(c.Request.Header.Get("HashSHA256"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Failed to decode hash"})
		zap.L().Error("Failed to decode hash: ", zap.Error(err))
		return
	}

	if !(hmac.Equal(hash, hashFromHeader)) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Signature hash does not match"})
		zap.L().Error("Signature hash does not match")
		return
	}

	err = json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid input data: %v", err)})
		zap.L().Error("Error unmarshalling request: ", zap.Error(err))
		return
	}

	// Используем переданное хранилище вместо глобальной переменной
	err = h.storage.UpdateSliceOfMetrics(ctx, req)

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
