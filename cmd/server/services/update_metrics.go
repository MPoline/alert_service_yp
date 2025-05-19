package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func UpdateMetricFromJSON(s *storage.MemStorage, c *gin.Context) {

	var (
		req  storage.Metrics
		resp storage.Metrics
	)

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

	switch req.MType {
	case "gauge":
		if req.Value == nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid gauge value"})
			zap.L().Info("Invalid gauge value")
			return
		}
		s.SetGauge(req.ID, *req.Value)
	case "counter":
		if req.Delta == nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid counter value"})
			zap.L().Info("Invalid counter value")
			return
		}
		s.IncrementCounter(req.ID, *req.Delta)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Unknown metric type"})
		zap.L().Info("Unknown metric type")
	}

	resp.ID = req.ID
	resp.MType = req.MType
	resp.Delta = req.Delta
	resp.Value = req.Value

	respBytes, err := json.Marshal(resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to encode response"})
		zap.L().Error("Failed to encode response: ", zap.Error(err))
		return
	}

	c.Header("Content-Type", c.GetHeader("Content-Type"))
	c.Header("Content-Length", c.GetHeader("Content-Length"))
	c.Header("Date", time.Now().UTC().Format(http.TimeFormat))
	c.String(http.StatusOK, string(respBytes))
}

func UpdateMetricFromURL(s *storage.MemStorage, c *gin.Context) {

	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

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

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid gauge value"})
			zap.L().Info("Invalid gauge value: ", zap.Error(err))
			return
		}
		s.SetGauge(metricName, value)
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid counter value"})
			zap.L().Info("Invalid counter value: ", zap.Error(err))
			return
		}
		s.IncrementCounter(metricName, value)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Unknown metric type"})
		zap.L().Info("Unknown metric type")
		return
	}

	c.Header("Content-Type", c.GetHeader("Content-Type"))
	c.Header("Content-Length", c.GetHeader("Content-Length"))
	c.Header("Date", time.Now().UTC().Format(http.TimeFormat))
	c.Status(http.StatusOK)
}
