package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
)

func GetMetric(s *storage.MemStorage, c *gin.Context) {

	var (
		req   storage.Metrics
		resp  storage.Metrics
		found bool
	)

	if c.GetHeader("Content-Type") != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"Error": "Only Content-Type: application/json are allowed"})
		return
	}

	data, _ := io.ReadAll(c.Request.Body)

	err := json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid input data: %v", err)})
		return
	}

	switch req.MType {
	case "gauge":
		if val, ok := s.GetGauge(req.ID); ok {
			fmt.Println("Found gauge: ", val)
			resp.Value = &val
			found = true
		}
	case "counter":
		if val, ok := s.GetCounter(req.ID); ok {
			fmt.Println("Found counter: ", val)
			resp.Delta = &val
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

	resp.ID = req.ID
	resp.MType = req.MType

	respBytes, err := json.Marshal(resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to encode response"})
		return
	}

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, string(respBytes))
}
