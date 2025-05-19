package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		durationTime := time.Since(startTime)

		logger.Info("Request Info: ",
			zap.String("URL", c.Request.URL.Path),
			zap.String("Method", c.Request.Method),
			zap.Duration("Duration time", durationTime),
		)
	}

}
