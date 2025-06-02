package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func UpdateSliceOfMetrics(c *gin.Context) {
	var req models.SliceMetrics

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to read request"})
		zap.L().Error("Error reading request body: ", zap.Error(err))
	}

	err = json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid input data: %v", err)})
		zap.L().Error("Error unmarshalling request: ", zap.Error(err))
		return
	}

	err = storage.MetricStorage.UpdateSliceOfMetrics(req)

	if err != nil {
		if err.Error() == "InvalidMetricName" || err.Error() == "InvalidMetricType" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		if err.Error() == "InvalidCounterValue" || err.Error() == "InvalidGaugeValue" {
			c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	respBytes, err := json.Marshal(req)
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
