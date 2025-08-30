package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/MPoline/alert_service_yp/internal/crypto"
	"github.com/MPoline/alert_service_yp/internal/server/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DecryptMiddleware проверяет и расшифровывает входящие зашифрованные данные
func DecryptMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isEncrypted := c.GetHeader("X-Encrypted") == "true"

		if !isEncrypted {
			c.Next()
			return
		}

		privateKey := services.GetPrivateKey()
		if privateKey == nil {
			zap.L().Error("Received encrypted data but decryption is not configured")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Decryption not configured"})
			c.Abort()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			zap.L().Error("Failed to read request body", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			c.Abort()
			return
		}
		defer c.Request.Body.Close()

		zap.L().Debug("Encrypted request received (after gzip decompress)",
			zap.Int("size", len(body)),
			zap.Int("key_size", privateKey.Size()),
			zap.Bool("is_chunk_protocol", crypto.IsChunkProtocol(body)))

		decryptedData, err := crypto.DecryptLargeData(privateKey, body)
		if err != nil {
			zap.L().Error("Failed to decrypt data",
				zap.Error(err),
				zap.Int("data_size", len(body)),
				zap.Int("key_size", privateKey.Size()))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Decryption failed: " + err.Error()})
			c.Abort()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(decryptedData))
		c.Request.ContentLength = int64(len(decryptedData))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.Header.Del("X-Encrypted")
		c.Request.Header.Del("X-Encryption-Algorithm")

		zap.L().Info("Data decrypted successfully",
			zap.Int("encrypted_size", len(body)),
			zap.Int("decrypted_size", len(decryptedData)))

		c.Next()
	}
}
