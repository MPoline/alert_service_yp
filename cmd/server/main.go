package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	storage "github.com/MPoline/alert_service_yp/cmd/server/memstorage"
	"github.com/gin-gonic/gin"

	"net/http"
)

var memStorage = storage.NewMemStorage()

// func middleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		if r.Header.Get("Content-Type") != "text/plain" {
// 			http.Error(w, "Only Content-Type:text/plain are allowed!", http.StatusUnsupportedMediaType)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

func updateMetric(c *gin.Context) {

	if c.GetHeader("Content-Type") != "text/plain" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"Error": "Only Content-Type: text/plain are allowed!"})
		return
	}

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

	contentType := c.GetHeader("Content-Type")
	contentLength := c.GetHeader("Content-Length")
	date := time.Now().UTC().Format(http.TimeFormat)

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid gauge value"})
			return
		}
		memStorage.SetGauge(metricName, value)
		newValue, checkValue := memStorage.GetGauge(metricName)
		fmt.Println("newValue: ", newValue)
		fmt.Println("checkValue: ", checkValue)

		c.Header("Content-Type", contentType)
		c.Header("Content-Length", contentLength)
		c.Header("Date", date)
		c.Status(http.StatusOK)

	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid counter value"})
			return
		}
		memStorage.IncrementCounter(metricName, value)
		newValue, checkValue := memStorage.GetCounter(metricName)
		fmt.Println("newValue: ", newValue)
		fmt.Println("checkValue: ", checkValue)

		c.Header("Content-Type", contentType)
		c.Header("Content-Length", contentLength)
		c.Header("Date", date)
		c.Status(http.StatusOK)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Unknown metric type"})
	}

}

func getMetric(c *gin.Context) {

	metricType := c.Param("type")
	metricName := c.Param("name")

	if metricName == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric name is required"})
		return
	}

	var value string
	var found bool
	switch metricType {
	case "gauge":
		if val, ok := memStorage.GetGauge(metricName); ok {
			value = fmt.Sprintf("%f", val)
			found = true
		}
	case "counter":
		if val, ok := memStorage.GetCounter(metricName); ok {
			value = fmt.Sprintf("%d", val)
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

func getAllMetrics(c *gin.Context) {

	if len(c.Request.URL.String()) != 1 {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Method not found"})
		return
	}

	var sb strings.Builder
	sb.WriteString("<html><body><h1>Metrics</h1><ul>")

	for metricName, metricValue := range memStorage.Gauges {
		sb.WriteString(fmt.Sprintf("<li> Gauge metrics: %s - %f</li>", metricName, metricValue))
	}

	for metricName, metricValue := range memStorage.Counters {
		sb.WriteString(fmt.Sprintf("<li> Counter metrics: %s - %d</li>", metricName, metricValue))
	}

	sb.WriteString("</ul></body></html>")

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, sb.String())
}

func main() {
	router := gin.Default()

	router.GET("/", getAllMetrics)
	router.GET("/value/:type/:name", getMetric)
	router.POST("/update/:type/:name/:value", updateMetric)

	fmt.Println("Starting server on :8080")

	router.Run(":8080")
}
