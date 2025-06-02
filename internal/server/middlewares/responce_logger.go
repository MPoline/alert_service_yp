package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ResponseLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		statusCode := c.Writer.Status()
		contentSize := c.Writer.Size()

		logger.Info("Response Info",
			zap.Int("Status code", statusCode),
			zap.Int("Content size", contentSize),
		)
	}
}
