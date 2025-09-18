package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MPoline/alert_service_yp/internal/agent/flags"
	"github.com/MPoline/alert_service_yp/internal/crypto"
	"github.com/MPoline/alert_service_yp/internal/hasher"
	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type HTTPClient struct {
	serverURL       string
	metricProcessor *MetricProcessor
	client          *resty.Client
}

func NewHTTPClient() *HTTPClient {
	var pubKey *rsa.PublicKey
	if flags.FlagCryptoKey != "" {
		var err error
		pubKey, err = crypto.LoadPublicKey(flags.FlagCryptoKey)
		if err != nil {
			zap.L().Error("Failed to load public key for HTTP client", zap.Error(err))
		}
	}

	metricProcessor := NewMetricProcessor(pubKey, flags.FlagKey)

	return &HTTPClient{
		serverURL:       "http://" + flags.FlagRunAddr + "/updates",
		metricProcessor: metricProcessor,
		client:          resty.New().SetTimeout(5 * time.Second),
	}
}

func (c *HTTPClient) HealthCheck() error {
	endpoints := []string{
		"http://" + flags.FlagRunAddr + "/",
		"http://" + flags.FlagRunAddr + "/ping",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var lastErr error
	for _, endpoint := range endpoints {
		resp, err := c.client.R().
			SetContext(ctx).
			Head(endpoint)

		if err == nil && resp.StatusCode() < 500 {
			zap.L().Debug("HTTP health check successful",
				zap.String("url", endpoint),
				zap.Int("status", resp.StatusCode()))
			return nil
		}

		if err != nil {
			lastErr = fmt.Errorf("endpoint %s: %w", endpoint, err)
		} else {
			lastErr = fmt.Errorf("endpoint %s returned status: %d", endpoint, resp.StatusCode())
		}

		zap.L().Debug("HTTP health check failed for endpoint",
			zap.String("endpoint", endpoint),
			zap.Error(lastErr))
	}

	return fmt.Errorf("all HTTP health checks failed: %w", lastErr)
}

func (c *HTTPClient) Close() {
	c.client.GetClient().CloseIdleConnections()
	zap.L().Debug("HTTP client connections closed")
}

// SendMetrics отправляет метрики на сервер через HTTP
func (c *HTTPClient) SendMetrics(memStorage *storage.MemStorage, metrics []models.Metrics, realIP string) {
	zap.L().Info("Start SendMetrics", zap.String("real_ip", realIP))

	intervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	jsonBody, err := json.Marshal(map[string][]models.Metrics{"metrics": metrics})
	if err != nil {
		zap.L().Error("Failed to marshal batch of metrics: ", zap.Error(err))
		return
	}

	h := hasher.InitHasher("SHA256")
	hash, err := h.CalculateHash(jsonBody, []byte(flags.FlagKey))
	if err != nil {
		zap.L().Error("Failed calculate sha256: ", zap.Error(err))
		return
	}
	hashStr := base64.StdEncoding.EncodeToString(hash)
	zap.L().Info("hash request: ", zap.String("hashStr", hashStr))

	var requestData []byte
	var contentType string
	publicKey := c.metricProcessor.pubKey

	if publicKey != nil {
		encryptedData, err := crypto.EncryptLargeData(publicKey, jsonBody)
		if err != nil {
			zap.L().Error("Failed to encrypt data: ", zap.Error(err))
			return
		}
		requestData = encryptedData
		contentType = "application/octet-stream"
		zap.L().Info("Data encrypted with chunk protocol",
			zap.Int("original_size", len(jsonBody)),
			zap.Int("encrypted_size", len(encryptedData)),
			zap.Int("key_size", publicKey.Size()))
	} else {
		requestData = jsonBody
		contentType = "application/json"
		zap.L().Debug("Encryption disabled, using plain JSON")
	}

	var buff bytes.Buffer
	gz := gzip.NewWriter(&buff)
	_, err = gz.Write(requestData)
	if err != nil {
		zap.L().Error("Failed to compress data: ", zap.Error(err))
		return
	}
	if err := gz.Close(); err != nil {
		zap.L().Error("Failed to close gzip writer: ", zap.Error(err))
		return
	}
	compressedData := buff.Bytes()

	headers := map[string]string{
		"Content-Type":     contentType,
		"Content-Encoding": "gzip",
		"HashSHA256":       base64.StdEncoding.EncodeToString(hash),
		"X-Real-IP":        realIP,
	}

	if publicKey != nil {
		headers["X-Encrypted"] = "true"
		headers["X-Encryption-Algorithm"] = "RSA-OAEP"
		headers["X-Encryption-Mode"] = "chunk-protocol"
	}

	zap.L().Debug("Request prepared",
		zap.Int("json_size", len(jsonBody)),
		zap.Int("request_size", len(requestData)),
		zap.Int("compressed_size", len(compressedData)),
		zap.Bool("encrypted", publicKey != nil),
		zap.String("real_ip", realIP))

	for attempt, interval := range intervals {
		req := c.client.R().
			SetHeaders(headers).
			SetBody(compressedData)

		zap.L().Debug("Sending request attempt",
			zap.Int("attempt", attempt+1),
			zap.Int("total_attempts", len(intervals)),
			zap.String("real_ip", realIP))

		resp, err := req.Post(c.serverURL)
		if err != nil {
			zap.L().Warn("Sending metrics failed on attempt",
				zap.Int("attempt", attempt+1),
				zap.Duration("interval", interval),
				zap.Error(err),
				zap.Bool("encrypted", publicKey != nil))
			time.Sleep(interval)
			continue
		}

		if resp.IsError() {
			zap.L().Warn("Server returned error on attempt",
				zap.Int("attempt", attempt+1),
				zap.Int("status", resp.StatusCode()),
				zap.String("response", resp.String()),
				zap.Bool("encrypted", publicKey != nil))
			time.Sleep(interval)
			continue
		}

		zap.L().Info("Metrics sent successfully",
			zap.Int("attempt", attempt+1),
			zap.Int("metrics_count", len(metrics)),
			zap.Int("compressed_size", len(compressedData)),
			zap.Bool("encrypted", publicKey != nil),
			zap.String("server_response", resp.String()),
			zap.String("real_ip", realIP))
		return
	}

	zap.L().Error("All attempts to send metrics failed",
		zap.Int("metrics_count", len(metrics)),
		zap.Bool("encrypted", publicKey != nil))
}
