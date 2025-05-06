package main

import (
	"fmt"

	services "github.com/MPoline/alert_service_yp/cmd/server/services"
	flags "github.com/MPoline/alert_service_yp/internal/server"
	storage "github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
)

var (
	memStorage = storage.NewMemStorage()
)

// func middleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		if r.Header.Get("Content-Type") != "text/plain" {
// 			http.Error(w, "Only Content-Type:text/plain are allowed!", http.StatusUnsupportedMediaType)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

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
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/", getAllMetricsHandler)
	router.GET("/value/:type/:name", getMetricsHandler)
	router.POST("/update/:type/:name/:value", updateMetricsHandler)

	flags.ParseFlags()
	fmt.Println("Running server on", flags.FlagRunAddr)
	router.Run(flags.FlagRunAddr)
}
