package middleware

import (
	"net/http"
	"strings"

	"gin-api-1/internal/env"

	"github.com/gin-gonic/gin"
)

const (
	allowCredentialsValue = "true"
	allowedMethodsValue   = "GET, POST, PATCH, PUT, DELETE, OPTIONS"
	allowedHeadersValue   = "Authorization, Content-Type, X-Requested-With"
	exposedHeadersValue   = "Content-Length"
	maxAgeValue           = "86400"
)

func allowedOrigins() []string {
	origins := env.GetEnvString("CORS_ALLOWED_ORIGINS", "")
	if origins == "" {
		return []string{
			"http://localhost:3000",
			"http://localhost:3700",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3700",
		}
	}

	var result []string
	for _, origin := range strings.Split(origins, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func CORS() gin.HandlerFunc {
	allowed := allowedOrigins()

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		allowedOrigin := ""
		for _, o := range allowed {
			if o == origin {
				allowedOrigin = o
				break
			}
		}

		if allowedOrigin == "" {
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", allowedMethodsValue)
		c.Header("Access-Control-Allow-Headers", allowedHeadersValue)
		c.Header("Access-Control-Allow-Credentials", allowCredentialsValue)
		c.Header("Access-Control-Expose-Headers", exposedHeadersValue)

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Max-Age", maxAgeValue)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}