package services

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetAllMetrics обрабатывает запрос на получение всех метрик в HTML-формате.
//
// Эндпоинт: GET /
//
// Логика работы:
//  1. Проверяет корректность URL (должен быть строго "/")
//  2. Получает все метрики из хранилища
//  3. Форматирует метрики в HTML-список
//  4. Возвращает HTML-страницу с метриками
//
// Формат ответа:
// HTML страница с двумя разделами:
//   - Gauge метрики (со значениями float64)
//   - Counter метрики (со значениями int64)
//
// Возможные ответы:
//   - 200 OK: успешное получение метрик
//     Content-Type: text/html
//     Тело: HTML страница с метриками
//   - 404 Not Found:
//   - Если URL не соответствует "/"
//     Тело: {"Error": "Method not found"}
//   - Если возникла ошибка при получении метрик
//     Тело: HTML страница с сообщением об ошибке
//
// Пример HTML ответа:
//
//	<html>
//	  <body>
//	    <h1>Metrics</h1>
//	    <ul>
//	      <li>Gauge: alloc - 123.456</li>
//	      <li>Counter: pollCount - 42</li>
//	    </ul>
//	  </body>
//	</html>
func (h *ServiceHandler) GetAllMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	if len(c.Request.URL.String()) != 1 {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Method not found"})
		zap.L().Info("Method not found")
		return
	}

	var sb strings.Builder
	sb.WriteString("<html><body><h1>Metrics</h1><ul>")

	metrics, err := h.storage.GetAllMetrics(ctx)

	if err != nil {
		sb.WriteString("<li> Error getting metric </li>")
		sb.WriteString("</ul></body></html>")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusNotFound, sb.String())
	} else {
		for _, metric := range metrics {
			if metric.MType == "gauge" {
				sb.WriteString(fmt.Sprintf("<li> Gauge: %s - %f</li>", metric.ID, *metric.Value))
			}
			if metric.MType == "counter" {
				sb.WriteString(fmt.Sprintf("<li> Counter: %s - %d</li>", metric.ID, *metric.Delta))
			}
		}
		sb.WriteString("</ul></body></html>")
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, sb.String())
	}
}
