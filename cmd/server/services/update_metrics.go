package services

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
)

func UpdateMetric(s *storage.MemStorage, c *gin.Context) {

	// if c.GetHeader("Content-Type") != "text/plain" {
	// 	c.JSON(http.StatusUnsupportedMediaType, gin.H{"Error": "Only Content-Type: text/plain are allowed!"})
	// 	return
	// }

	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	if metricType == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric type is required"})
		return
	}
	if metricName == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric name is required"})
		return
	}
	if metricValue == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric value is required"})
		return
	}

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid gauge value"})
			return
		}
		s.SetGauge(metricName, value)
		newValue, checkValue := s.GetGauge(metricName)
		fmt.Println("newValue: ", newValue)
		fmt.Println("checkValue: ", checkValue)
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid counter value"})
			return
		}
		s.IncrementCounter(metricName, value)
		newValue, checkValue := s.GetCounter(metricName)
		fmt.Println("newValue: ", newValue)
		fmt.Println("checkValue: ", checkValue)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Unknown metric type"})
	}
	
	c.Header("Content-Type", c.GetHeader("Content-Type"))
	c.Header("Content-Length", c.GetHeader("Content-Length"))
	c.Header("Date", time.Now().UTC().Format(http.TimeFormat))
	c.Status(http.StatusOK)
}
