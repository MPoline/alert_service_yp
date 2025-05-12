package middlewares

import "github.com/gin-gonic/gin"

func FilterContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		ct := c.ContentType()
		if ct != "application/json" && ct != "text/html" {
			c.Header("X-Gzip-Skip", "true")
		} else {
			c.Writer.Header().Del("Content-Length")
		}
		c.Next()
	}
}
