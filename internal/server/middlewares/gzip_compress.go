// Package middlewares предоставляет HTTP middleware для обработки сжатия Gzip.
package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	gzipWriterPool = sync.Pool{
		New: func() interface{} {
			w, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression)
			return w
		},
	}

	gzipReaderPool = sync.Pool{
		New: func() interface{} {
			return new(gzip.Reader)
		},
	}
)

// gzipWriter оборачивает gin.ResponseWriter для прозрачного сжатия Gzip.
// Реализует интерфейс io.Writer.
type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
	buf    *bytes.Buffer
}

func (gzw *gzipWriter) Write(b []byte) (int, error) {
	return gzw.writer.Write(b)
}

func (gzw *gzipWriter) Close() {
	gzw.writer.Close()
	gzipWriterPool.Put(gzw.writer)
}

// GZipDecompress возвращает middleware для распаковки входящих Gzip-запросов.
// Проверяет заголовок Content-Encoding: gzip и автоматически распаковывает тело запроса.
//
// Пример использования:
//  router := gin.Default()
//  router.Use(GZipDecompress())
func GZipDecompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.EqualFold(c.GetHeader("Content-Encoding"), "gzip") {
			c.Next()
			return
		}

		gzr := gzipReaderPool.Get().(*gzip.Reader)
		defer gzipReaderPool.Put(gzr)

		if err := gzr.Reset(c.Request.Body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error": "Failed to decode compressed request"})
			zap.L().Error("Failed to decode compressed request", zap.Error(err))
			return
		}
		defer gzr.Close()

		body, err := io.ReadAll(gzr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"Error": "Failed to read compressed request"})
			zap.L().Error("Failed to read compressed request", zap.Error(err))
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Request.ContentLength = int64(len(body))
		c.Next()
	}
}

// GZipCompress возвращает middleware для сжатия исходящих ответов в Gzip.
// Сжимает только ответы с Content-Type: application/json или text/html.
// Проверяет заголовок Accept-Encoding запроса на наличие gzip.
//
// Пример использования:
//  router := gin.Default()
//  router.Use(GZipCompress())
func GZipCompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.GetHeader("Accept-Encoding")
		contentType := c.GetHeader("Content-Type")

		if !strings.Contains(acceptEncoding, "gzip") ||
			!(strings.Contains(contentType, "application/json") ||
				strings.Contains(contentType, "text/html")) {
			c.Next()
			return
		}

		gzw := gzipWriterPool.Get().(*gzip.Writer)
		gzipWriter := &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gzw,
		}
		gzipWriter.writer.Reset(c.Writer)
		defer gzipWriter.Close()

		c.Writer = gzipWriter
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Header("Content-Type", contentType)
		c.Writer.Header().Del("Content-Length")

		c.Next()
	}
}
