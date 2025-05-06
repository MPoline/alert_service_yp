package services

import (
	"fmt"
	"net/http"
	"strings"
	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
)

func GetAllMetrics(s *storage.MemStorage, c *gin.Context) {

	if len(c.Request.URL.String()) != 1 {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Method not found"})
		return
	}

	var sb strings.Builder
	sb.WriteString("<html><body><h1>Metrics</h1><ul>")

	for metricName, metricValue := range s.Gauges {
		sb.WriteString(fmt.Sprintf("<li> Gauge metrics: %s - %f</li>", metricName, metricValue))
	}

	for metricName, metricValue := range s.Counters {
		sb.WriteString(fmt.Sprintf("<li> Counter metrics: %s - %d</li>", metricName, metricValue))
	}

	sb.WriteString("</ul></body></html>")

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, sb.String())
}
