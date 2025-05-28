package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetMetricFromJSON(c *gin.Context) {
	var (
		req  models.Metrics
		resp models.Metrics
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

	resp, err = storage.MetricStorage.GetMetric(req.MType, req.ID)
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
	c.String(http.StatusOK, string(respBytes))
}

func GetMetricFromURL(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	if metricName == "" {
		c.JSON(http.StatusNotFound, gin.H{"Error": "Metric name is required"})
		zap.L().Info("Metric name is required")
		return
	}

	resp, err := storage.MetricStorage.GetMetric(metricType, metricName)
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
