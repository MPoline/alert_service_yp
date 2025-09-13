package middlewares

import (
	"net"
	"net/http"
	"strings"

	"github.com/MPoline/alert_service_yp/internal/server/flags"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TrustedSubnetMiddleware проверяет, что IP адрес из заголовка X-Real-IP
// находится в доверенной подсети. Если подсеть не задана, пропускает все запросы.
func TrustedSubnetMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if flags.FlagTrustedSubnet == "" {
			c.Next()
			return
		}

		realIP := c.GetHeader("X-Real-IP")
		if realIP == "" {
			zap.L().Warn("X-Real-IP header is missing",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "X-Real-IP header required",
			})
			return
		}

		if strings.Contains(realIP, ":") {
			if host, _, err := net.SplitHostPort(realIP); err == nil {
				realIP = host
			}
		}

		ip := net.ParseIP(realIP)
		if ip == nil {
			zap.L().Warn("Invalid IP address in X-Real-IP header",
				zap.String("ip", realIP),
				zap.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid IP address",
			})
			return
		}

		_, trustedNet, err := net.ParseCIDR(flags.FlagTrustedSubnet)
		if err != nil {
			zap.L().Error("Invalid trusted subnet CIDR",
				zap.String("subnet", flags.FlagTrustedSubnet),
				zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Server configuration error",
			})
			return
		}

		if !trustedNet.Contains(ip) {
			zap.L().Warn("IP address not in trusted subnet",
				zap.String("ip", realIP),
				zap.String("trusted_subnet", flags.FlagTrustedSubnet),
				zap.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied - IP not in trusted subnet",
			})
			return
		}

		zap.L().Debug("IP address allowed",
			zap.String("ip", realIP),
			zap.String("trusted_subnet", flags.FlagTrustedSubnet),
			zap.String("path", c.Request.URL.Path))

		c.Next()
	}
}
