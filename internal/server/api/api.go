// Package api предоставляет HTTP API сервера метрик.
//
// Пакет содержит:
// - Инициализацию роутера Gin
// - Регистрацию middleware
// - Маршрутизацию запросов
package api

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/MPoline/alert_service_yp/internal/logging"
	"github.com/MPoline/alert_service_yp/internal/server/middlewares"
	"github.com/MPoline/alert_service_yp/internal/server/services"
	"github.com/gin-gonic/gin"
)

// InitRouter создает и настраивает роутер Gin с middleware и обработчиками запросов.
//
// Возвращает:
//   - *gin.Engine: настроенный роутер
//
// Регистрирует следующие эндпоинты:
//   - GET  /ping          - проверка подключения к БД
//   - GET  /              - получение всех метрик
//   - GET  /value/        - получение метрики в формате JSON
//   - POST /update/       - обновление метрики в формате JSON
//   - POST /updates/      - массовое обновление метрик
//   - GET  /value/:type/:name - получение метрики через URL
//   - POST /update/:type/:name/:value - обновление метрики через URL
//
// Пример использования:
//
//	router := api.InitRouter()
//	router.Run(":8080")
func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	registerMiddlewares(router)

	router.GET("/ping", services.CheckDBConnection)
	router.GET("/", services.GetAllMetrics)
	router.GET("/value/", services.GetMetricFromJSON)
	router.POST("/update/", services.UpdateMetricFromJSON)
	router.POST("/updates/", services.UpdateSliceOfMetrics)
	router.GET("/value/:type/:name", services.GetMetricFromURL)
	router.POST("/update/:type/:name/:value", services.UpdateMetricFromURL)

	return router
}

// InitDecryption инициализирует расшифровку с предоставленным приватным ключом
func InitDecryption(privateKey *rsa.PrivateKey) {
	services.InitDecryption(privateKey)
}

// registerMiddlewares регистрирует middleware для роутера:
//   - GZip сжатие ответов
//   - GZip распаковка запросов
//   - Логирование входящих запросов
//   - Логирование исходящих ответов
//
// Параметры:
//   - r *gin.Engine: роутер Gin
func registerMiddlewares(r *gin.Engine) {
	logger, err := logging.InitLog()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
	}
	defer logger.Sync()

	r.Use(middlewares.GZipDecompress())
	r.Use(middlewares.DecryptMiddleware())
	r.Use(middlewares.GZipCompress())
	r.Use(middlewares.RequestLogger(logger))
	r.Use(middlewares.ResponseLogger(logger))
}
