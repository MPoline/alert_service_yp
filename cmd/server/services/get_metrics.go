package services

import (
	"fmt"
	"net/http"
	"strconv"

	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
)

func GetMetric(s *storage.MemStorage, c *gin.Context) {

	var (
		value string
		found bool 
	)

	fmt.Println("GetMetric start: ", s)

	metricType := c.Param("type")
	metricName := c.Param("name")

	fmt.Println("Params: ", metricType, metricName)

	if metricName == "" {
		fmt.Println("metricName == \"\"")
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric name is required"})
		return
	}

	switch metricType {
	case "gauge":
		if val, ok := s.GetGauge(metricName); ok {
			fmt.Println("Found gauge: ", val)
			value = strconv.FormatFloat(val, 'f', -1, 64)
			found = true
		}
	case "counter":
		if val, ok := s.GetCounter(metricName); ok {
			fmt.Println("Found counter: ", val)
			value = strconv.FormatInt(val, 10)
			found = true
		}
	default:
		fmt.Println("Unknown metric type")
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Unknown metric type"})
		return
	}

	fmt.Println("Found flag is: ", found)

	if !found {
		fmt.Println("Metric not found")
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric not found"})
		return
	}

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, value)
}
