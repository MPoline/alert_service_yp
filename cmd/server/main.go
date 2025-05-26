package main

import (
	"fmt"
	"os"
	"time"

	"github.com/MPoline/alert_service_yp/cmd/server/middlewares"
	services "github.com/MPoline/alert_service_yp/cmd/server/services"
	"github.com/MPoline/alert_service_yp/internal/logging"
	flags "github.com/MPoline/alert_service_yp/internal/server"
	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	memStorage = storage.NewMemStorage()
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

	logger, err := logging.InitLog()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
	}
	defer logger.Sync()

	undo := zap.ReplaceGlobals(logger)
	defer undo()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(middlewares.GZipCompress())
	router.Use(middlewares.GZipDecompress())
	router.Use(middlewares.RequestLogger(logger))
	router.Use(middlewares.ResponseLogger(logger))

	router.GET("/ping", services.CheckDBConnection)
	router.GET("/", getAllMetricsHandler)
	router.GET("/value/", getMetricsJSONHandler)
	router.POST("/update/", updateMetricsJSONHandler)
	router.GET("/value/:type/:name", getMetricsURLHandler)
	router.POST("/update/:type/:name/:value", updateMetricsURLHandler)

	flags.ParseFlags()

	logger.Info(
		"Server settings",
		zap.String("Running server address: ", flags.FlagRunAddr),
		zap.String("Running database address: ", flags.FlagDatabaseDSN),
		zap.Int64("Store metrics interval: ", flags.FlagStoreInterval),
		zap.String("Store path: ", flags.FlagFileStoragePath),
		zap.Bool("Is restore: ", flags.FlagRestore),
	)

	if flags.FlagRestore {
		err := memStorage.LoadFromFile(flags.FlagFileStoragePath)
		if err != nil {
			logger.Warn("Error read from file: ", zap.Error(err))
		}
	}

	storeInterval := time.Second * time.Duration(flags.FlagStoreInterval)
	if flags.FlagStoreInterval > 0 {
		ticker := time.NewTicker(storeInterval)
		go func() {
			for range ticker.C {
				err := memStorage.SaveToFile(flags.FlagFileStoragePath)
				if err != nil {
					logger.Error("Error save metrics: ", zap.Error(err))
				}
			}
		}()
	}

	defer func() {
		err := memStorage.SaveToFile(flags.FlagFileStoragePath)
		if err != nil {
			logger.Fatal("Error last save metrics: ", zap.Error(err))
		}
	}()

	router.Run(flags.FlagRunAddr)
}
