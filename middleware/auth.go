package middleware

import (
	"net/http"

	"github.com/achmadnp/peroot/model"
	"github.com/gin-gonic/gin"
)

func APIKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" {
			c.Next()
			return
		}

		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Success: false,
				Error:   "Missing X-API-Key header",
				Code:    "UNAUTHORIZED",
			})
			return
		}

		if key != apiKey {
			c.AbortWithStatusJSON(http.StatusForbidden, model.ErrorResponse{
				Success: false,
				Error:   "Invalid API key",
				Code:    "FORBIDDEN",
			})
			return
		}

		c.Next()
	}
}
