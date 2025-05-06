package services

import (
	"net/http"
	"strconv"

	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
)

var (
	value string
	found bool
)

func GetMetric(s *storage.MemStorage, c *gin.Context) {

	metricType := c.Param("type")
	metricName := c.Param("name")

	if metricName == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric name is required"})
		return
	}

	switch metricType {
	case "gauge":
		if val, ok := s.GetGauge(metricName); ok {
			value = strconv.FormatFloat(val, 'f', -1, 64)
			found = true
		}
	case "counter":
		if val, ok := s.GetCounter(metricName); ok {
			value = strconv.FormatInt(val, 10)
			found = true
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Unknown metric type"})
		return
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric not found"})
		return
	}

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, value)
}
