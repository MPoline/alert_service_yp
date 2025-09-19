package api

import (
	"fmt"
	"os"

	"github.com/MPoline/alert_service_yp/internal/logging"
	"github.com/MPoline/alert_service_yp/internal/server/middlewares"
	"github.com/MPoline/alert_service_yp/internal/server/services"
	"github.com/gin-gonic/gin"
)

type API struct {
	serviceHandler *services.ServiceHandler
}

func NewAPI(serviceHandler *services.ServiceHandler) *API {
	return &API{
		serviceHandler: serviceHandler,
	}
}

func (a *API) InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	a.registerMiddlewares(router)
	a.registerRoutes(router)

	return router
}

func (a *API) registerMiddlewares(r *gin.Engine) {
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

func (a *API) registerRoutes(r *gin.Engine) {
	r.GET("/ping", a.serviceHandler.CheckDBConnection)
	r.GET("/", a.serviceHandler.GetAllMetrics)
	r.GET("/value/", a.serviceHandler.GetMetricFromJSON)
	r.GET("/value/:type/:name", a.serviceHandler.GetMetricFromURL)

	updateGroup := r.Group("/")
	updateGroup.Use(middlewares.TrustedSubnetMiddleware())
	{
		updateGroup.POST("/update/", a.serviceHandler.UpdateMetricFromJSON)
		updateGroup.POST("/updates/", a.serviceHandler.UpdateSliceOfMetrics)
		updateGroup.POST("/update/:type/:name/:value", a.serviceHandler.UpdateMetricFromURL)
	}
}
