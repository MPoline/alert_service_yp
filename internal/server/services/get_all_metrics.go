package services

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetAllMetrics(c *gin.Context) {

	if len(c.Request.URL.String()) != 1 {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Method not found"})
		zap.L().Info("Method not found")
		return
	}

	var sb strings.Builder
	sb.WriteString("<html><body><h1>Metrics</h1><ul>")

	metrics, err := storage.MetricStorage.GetAllMetrics()

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
