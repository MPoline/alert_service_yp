package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetMetricFromJSON(s *storage.MemStorage, c *gin.Context) {

	var (
		req   storage.Metrics
		resp  storage.Metrics
		found bool
	)

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		zap.L().Error("Error in read request: ", zap.Error(err))
		return
	}

	err = json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid input data: %v", err)})
		zap.L().Error("Error in unmarshal request: ", zap.Error(err))
		return
	}

	switch req.MType {
	case "gauge":
		if val, ok := s.GetGauge(req.ID); ok {
			zap.L().Info("Found gauge", zap.Float64("value", val))
			resp.Value = &val
			found = true
		}
	case "counter":
		if val, ok := s.GetCounter(req.ID); ok {
			zap.L().Info("Found gauge", zap.Int64("value", val))
			resp.Delta = &val
			found = true
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Unknown metric type"})
		zap.L().Info("Unknown metric type")
		return
	}

	zap.L().Info("Found flag is: ", zap.Bool("flag", found))

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric not found"})
		zap.L().Info("Metric not found")
		return
	}

	resp.ID = req.ID
	resp.MType = req.MType

	respBytes, err := json.Marshal(resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to encode response"})
		zap.L().Error("Failed to encode response: ", zap.Error(err))
		return
	}

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, string(respBytes))
}

func GetMetricFromURL(s *storage.MemStorage, c *gin.Context) {

	var (
		value string
		found bool
	)

	metricType := c.Param("type")
	metricName := c.Param("name")

	if metricName == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric name is required"})
		zap.L().Info("Metric name is required")
		return
	}

	switch metricType {
	case "gauge":
		if val, ok := s.GetGauge(metricName); ok {
			zap.L().Info("Found gauge", zap.Float64("value", val))
			value = strconv.FormatFloat(val, 'f', -1, 64)
			found = true
		}
	case "counter":
		if val, ok := s.GetCounter(metricName); ok {
			zap.L().Info("Found counter: ", zap.Int64("value", val))
			value = strconv.FormatInt(val, 10)
			found = true
		}
	default:
		zap.L().Info("Unknown metric type")
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Unknown metric type"})
		return
	}

	zap.L().Info("Found flag is: ", zap.Bool("flag", found))

	if !found {
		zap.L().Info("Metric not found")
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric not found"})
		return
	}

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, value)
}
