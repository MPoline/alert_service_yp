package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResponseLogger возвращает middleware для логирования исходящих HTTP-ответов.
// Логирует следующие параметры:
//   - HTTP статус код
//   - Размер контента в байтах
//
// Пример использования:
//  router := gin.Default()
//  logger, _ := zap.NewProduction()
//  router.Use(ResponseLogger(logger))
//
// Пример вывода лога:
//  {"level":"info","ts":1630000000,"msg":"Response Info","Status code":200,"Content size":42}
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
