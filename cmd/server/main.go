package main

import (
	"fmt"

	"github.com/MPoline/alert_service_yp/cmd/server/middlewares"
	services "github.com/MPoline/alert_service_yp/cmd/server/services"
	flags "github.com/MPoline/alert_service_yp/internal/server"
	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	memStorage = storage.NewMemStorage()
	logger     *zap.Logger
)

func updateMetricsJSONHandler(c *gin.Context) {
	services.UpdateMetricFromJSON(memStorage, c)
}

func getMetricsJSONHandler(c *gin.Context) {
	services.GetMetricFromJSON(memStorage, c)
}

func getAllMetricsHandler(c *gin.Context) {
	services.GetAllMetrics(memStorage, c)
}

func updateMetricsURLHandler(c *gin.Context) {
	services.UpdateMetricFromURL(memStorage, c)
}

func getMetricsURLHandler(c *gin.Context) {
	services.GetMetricFromURL(memStorage, c)
}

func main() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		fmt.Println("Logger initialization error", err)
		panic(err)
	}
	defer logger.Sync()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(middlewares.RequestLogger(logger))
	router.Use(middlewares.ResponseLogger(logger))

	router.Use(middlewares.FilterContentType())
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.GET("/", getAllMetricsHandler)
	router.GET("/value/", getMetricsJSONHandler)
	router.POST("/update/", updateMetricsJSONHandler)
	router.GET("/value/:type/:name", getMetricsURLHandler)
	router.POST("/update/:type/:name/:value", updateMetricsURLHandler)

	flags.ParseFlags()
	fmt.Println("Running server on", flags.FlagRunAddr)
	router.Run(flags.FlagRunAddr)
}
