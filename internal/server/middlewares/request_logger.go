package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger возвращает middleware для логирования входящих HTTP-запросов.
// Логирует следующие параметры:
//   - URL запроса
//   - HTTP метод
//   - Время обработки запроса
//
// Пример использования:
//  router := gin.Default()
//  logger, _ := zap.NewProduction()
//  router.Use(RequestLogger(logger))
//
// Пример вывода лога:
//  {"level":"info","ts":1630000000,"msg":"Request Info: ","URL":"/path","Method":"GET","Duration time":123456}
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
