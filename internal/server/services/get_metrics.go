package services

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/MPoline/alert_service_yp/internal/hasher"
	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetMetricFromJSON обрабатывает запрос на получение метрики в JSON-формате с проверкой подписи.
//
// Эндпоинт: GET /value/
//
// Логика работы:
//  1. Читает тело запроса
//  2. Проверяет подпись HMAC-SHA256 (если включено)
//  3. Десериализует JSON в структуру Metrics
//  4. Получает метрику из хранилища
//  5. Возвращает метрику в JSON-формате с подписью
//
// Формат запроса:
//
//	{
//	  "id": "metricName",
//	  "type": "gauge|counter"
//	}
//
// Формат ответа:
//
//	{
//	  "id": "metricName",
//	  "type": "gauge|counter",
//	  "value": 123.45,    // для gauge
//	  "delta": 42         // для counter
//	}
//
// Заголовки:
//   - HashSHA256: подпись тела запроса (обязательный)
//
// Возможные ответы:
//   - 200 OK: успешное получение метрики
//   - 400 Bad Request: неверный формат запроса, ошибка проверки подписи
//   - 404 Not Found: метрика не найдена
//   - 500 Internal Server Error: ошибка сервера
//
// Пример использования:
//
//	Запрос:
//	  GET /value/
//	  Headers:
//	    Content-Type: application/json
//	    HashSHA256: <base64-encoded-hmac-sha256>
//	  Body:
//	    {"id":"alloc","type":"gauge"}
//
//	Ответ:
//	  {"id":"alloc","type":"gauge","value":123.456}
func (h *ServiceHandler) GetMetricFromJSON(c *gin.Context) {
	var (
		req  models.Metrics
		resp models.Metrics
	)

	ctx := c.Request.Context()

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		zap.L().Error("Error in read request: ", zap.Error(err))
		return
	}
	hs := hasher.InitHasher("SHA256")
	hash, err := hs.CalculateHash(data, []byte(h.key))
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
		zap.L().Error("Signature hash does not match")
		return
	}

	err = json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid input data: %v", err)})
		zap.L().Error("Error in unmarshal request: ", zap.Error(err))
		return
	}

	resp, err = h.storage.GetMetric(ctx, req.MType, req.ID)
	if err != nil {
		if err.Error() == "MetricNotFound" {
			c.JSON(http.StatusNotFound, gin.H{"Error": "MetricNotFound"})
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "BadRequest"})
			return
		}
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to encode response"})
		zap.L().Error("Failed to encode response: ", zap.Error(err))
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("HashSHA256", base64.StdEncoding.EncodeToString(hash))
	c.String(http.StatusOK, string(respBytes))
}

// GetMetricFromURL обрабатывает запрос на получение метрики через URL параметры.
//
// Эндпоинт: GET /value/:type/:name
//
// Логика работы:
//  1. Извлекает тип и имя метрики из URL
//  2. Получает метрику из хранилища
//  3. Возвращает значение метрики в текстовом формате
//
// Параметры URL:
//   - type: тип метрики (gauge или counter)
//   - name: имя метрики
//
// Возможные ответы:
//   - 200 OK: успешное получение метрики
//     Тело: строковое значение метрики
//   - 400 Bad Request: неверный тип метрики
//   - 404 Not Found: метрика не найдена или не указано имя
//
// Примеры:
//
//	Запрос:
//	  GET /value/gauge/alloc
//
//	Ответ:
//	  123.456
//
//	Запрос:
//	  GET /value/counter/pollCount
//
//	Ответ:
//	  42
func (h *ServiceHandler) GetMetricFromURL(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	ctx := c.Request.Context()

	if metricName == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric name is required"})
		zap.L().Info("Metric name is required")
		return
	}

	resp, err := h.storage.GetMetric(ctx, metricType, metricName)
	if err != nil {
		if err.Error() == "Unknown" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Unknown metric"})
			return
		}
		if err.Error() == "NotFound" {
			c.JSON(http.StatusNotFound, gin.H{"Error": "Metric not found"})
			return
		}
	}

	c.Header("Content-Type", "text/plain")
	if resp.MType == "gauge" {
		c.String(http.StatusOK, strconv.FormatFloat(*resp.Value, 'f', -1, 64))
	}
	if resp.MType == "counter" {
		c.String(http.StatusOK, strconv.FormatInt(*resp.Delta, 10))
	}
}
