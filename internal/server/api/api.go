package api

import (
	"fmt"
	"os"

	"github.com/MPoline/alert_service_yp/internal/logging"
	"github.com/MPoline/alert_service_yp/internal/server/middlewares"
	"github.com/MPoline/alert_service_yp/internal/server/services"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	registerMiddlewares(router)

	router.GET("/ping", services.CheckDBConnection)
	router.GET("/", services.GetAllMetrics)
	router.GET("/value/", services.GetMetricFromJSON)
	router.POST("/update/", services.UpdateMetricFromJSON)
	router.GET("/value/:type/:name", services.GetMetricFromURL)
	router.POST("/update/:type/:name/:value", services.UpdateMetricFromURL)

	return router
}

func registerMiddlewares(r *gin.Engine) {
	logger, err := logging.InitLog()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
	}
	defer logger.Sync()

	r.Use(middlewares.GZipCompress())
	r.Use(middlewares.GZipDecompress())
	r.Use(middlewares.RequestLogger(logger))
	r.Use(middlewares.ResponseLogger(logger))
}
