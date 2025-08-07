package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MPoline/alert_service_yp/internal/server/middlewares"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ExampleGZipCompress() {
	router := gin.Default()

	router.Use(middlewares.GZipCompress())

	router.GET("/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello, Gzip!"})
	})

	_ = router.Run(":8080")
}

func ExampleGZipDecompress() {
	router := gin.Default()

	router.Use(middlewares.GZipDecompress())

	router.POST("/data", func(c *gin.Context) {
		var jsonData map[string]interface{}
		if err := c.ShouldBindJSON(&jsonData); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"received": jsonData})
	})

	_ = router.Run(":8080")
}

func TestGzipMiddlewares(t *testing.T) {
	t.Run("GZipCompress", func(t *testing.T) {
		router := gin.Default()
		router.Use(middlewares.GZipCompress())
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "test response")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		router.ServeHTTP(w, req)

		if w.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Gzip compression not applied")
		}
	})
}

func ExampleRequestLogger() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	router := gin.Default()
	router.Use(middlewares.RequestLogger(logger))

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/test")
	if err != nil {
		// handle error
		return
	}
	defer resp.Body.Close()
}

func ExampleResponseLogger() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	router := gin.Default()
	router.Use(middlewares.ResponseLogger(logger))

	router.POST("/data", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/data", "application/json", nil)
	if err != nil {
		// handle error
		return
	}
	defer resp.Body.Close()
}

func TestMiddlewares(t *testing.T) {
	t.Run("RequestLogger", func(t *testing.T) {
		logger := zap.NewNop()
		router := gin.Default()
		router.Use(middlewares.RequestLogger(logger))
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "test")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, req)
	})

	t.Run("ResponseLogger", func(t *testing.T) {
		logger := zap.NewNop()
		router := gin.Default()
		router.Use(middlewares.ResponseLogger(logger))
		router.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "test")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, req)
	})
}
