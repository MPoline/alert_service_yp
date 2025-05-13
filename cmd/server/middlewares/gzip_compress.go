package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	io.Writer
	writer *gzip.Writer
}

func (gzw *gzipWriter) Write(b []byte) (int, error) {
	return gzw.writer.Write(b)
}

func GZipDecompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.EqualFold(c.GetHeader("Content-Encoding"), "gzip") {
			gzr, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Failed to decode compressed request"})
				return
			}
			defer gzr.Close()

			body, err := io.ReadAll(gzr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Failed to read compressed request"})
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}

		c.Next()
	}
}

func GZipCompress() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") || !(c.ContentType() == "application/json" || c.ContentType() == "text/html") {
			c.Next()
			return
		}

		c.Writer.Header().Set("Content-Encoding", "gzip")

		// Создаем gzip.Writer поверх существующего response writer
		gzw, err := gzip.NewWriterLevel(c.Writer, gzip.DefaultCompression)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Failed to decode compressed request."})
			return
		}
		defer gzw.Close()

		// Перехватываем оригинальный писатель
		c.Writer = &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gzw,
		}
		c.Writer.Header().Del("Content-Length")
		c.Next()
	}
}
