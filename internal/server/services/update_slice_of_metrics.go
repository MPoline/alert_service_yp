package services

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MPoline/alert_service_yp/internal/hasher"
	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/server/flags"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func UpdateSliceOfMetrics(c *gin.Context) {
	var req models.SliceMetrics

	ctx := c.Request.Context()

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Failed to read request"})
		zap.L().Error("Error reading request body: ", zap.Error(err))
		return
	}

	h := hasher.InitHasher("SHA256")
	hash, err := h.CalculateHash(data, []byte(flags.FlagKey))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Failed calculate sha256"})
		zap.L().Error("Failed calculate sha256: ", zap.Error(err))
		return
	}
	zap.L().Info("===================================")
	hashStr := base64.StdEncoding.EncodeToString(hash)
	hashHeader := c.Request.Header.Get("HashSHA256")
	zap.L().Info("hash request: ", zap.String("hashStr", hashStr))
	zap.L().Info("hash from header: ", zap.String("hash from header", hashHeader))
	zap.L().Info("===================================")

	hashFromHeader, err := base64.StdEncoding.DecodeString(c.Request.Header.Get("HashSHA256"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Failed to decode hash"})
		zap.L().Error("Failed to decode hash: ", zap.Error(err))
		return
	}

	if !(hmac.Equal(hash, hashFromHeader)) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Signature hash does not match"})
		zap.L().Error("Signature hash does not match: ", zap.Error(err))
		return
	}

	err = json.Unmarshal(data, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Invalid input data: %v", err)})
		zap.L().Error("Error unmarshalling request: ", zap.Error(err))
		return
	}

	err = storage.MetricStorage.UpdateSliceOfMetrics(ctx, req)

	if err != nil {
		if errors.Is(err, models.ErrInvalidMetricName) || errors.Is(err, models.ErrInvalidMetricType) {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		if errors.Is(err, models.ErrInvalidCounterValue) || errors.Is(err, models.ErrInvalidGaugeValue) {
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
	c.Header("HashSHA256", base64.StdEncoding.EncodeToString(hash))
	c.String(http.StatusOK, string(respBytes))

}
