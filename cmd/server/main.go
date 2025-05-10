package main

import (
	"fmt"

	"github.com/MPoline/alert_service_yp/cmd/server/middlewares"
	services "github.com/MPoline/alert_service_yp/cmd/server/services"
	flags "github.com/MPoline/alert_service_yp/internal/server"
	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	memStorage = storage.NewMemStorage()
	logger     *zap.Logger
)

func updateMetricsHandler(c *gin.Context) {
	services.UpdateMetric(memStorage, c)
}

func getMetricsHandler(c *gin.Context) {
	services.GetMetric(memStorage, c)
}

func getAllMetricsHandler(c *gin.Context) {
	services.GetAllMetrics(memStorage, c)
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

	router.GET("/", getAllMetricsHandler)
	router.GET("/value/:type/:name", getMetricsHandler)
	router.POST("/update/:type/:name/:value", updateMetricsHandler)

	flags.ParseFlags()
	fmt.Println("Running server on", flags.FlagRunAddr)
	router.Run(flags.FlagRunAddr)
}
