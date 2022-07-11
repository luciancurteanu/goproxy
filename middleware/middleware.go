package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Whitelist returns a middleware to only accept
// requests from authorized addresses
func Whitelist(addresses []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		remote := c.ClientIP()
		for _, address := range addresses {
			if remote == address {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":  "unauthorized",
			"status": http.StatusUnauthorized,
		})
	}
}

// Headers returns a middleware to set default headers
// on HTTP responses
func Headers(headers map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for header, value := range headers {
			c.Header(header, value)
		}
		c.Next()
	}
}
