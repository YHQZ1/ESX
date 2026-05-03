package middleware

import (
	"net/http"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/order-gateway/internal/client"
	"github.com/gin-gonic/gin"
)

func Auth(registry *client.RegistryClient, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing x-api-key header"})
			return
		}

		participantID, isActive, err := registry.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			log.Error("failed to validate api key", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
			return
		}

		if participantID == "" || !isActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or inactive api key"})
			return
		}

		c.Set("participant_id", participantID)
		c.Next()
	}
}
