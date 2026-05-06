package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/order-gateway/internal/client"
	"github.com/gin-gonic/gin"
)

// cacheEntry holds the authenticated user data and its expiration
type cacheEntry struct {
	participantID string
	expiresAt     time.Time
}

var (
	// sync.Map is highly optimized for concurrent read-heavy workloads
	authCache sync.Map
	cacheTTL  = 60 * time.Second
)

func Auth(registry *client.RegistryClient, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing x-api-key header"})
			return
		}

		// 1. THE FAST PATH: Check blazing-fast local memory cache first
		if val, ok := authCache.Load(apiKey); ok {
			entry := val.(cacheEntry)
			if time.Now().Before(entry.expiresAt) {
				// Cache Hit! Zero network overhead.
				c.Set("participant_id", entry.participantID)
				c.Next()
				return
			}
			// Cache expired, purge it and fall through to network call
			authCache.Delete(apiKey)
		}

		// 2. THE SLOW PATH: Pay the network tax (gRPC call)
		// Note: Ensure this matches the exact return signature of your client.RegistryClient
		participantID, isActive, err := registry.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			log.Error("auth verification failed", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal auth error"})
			return
		}

		if !isActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or inactive api key"})
			return
		}

		// 3. Save to cache for the next 60 seconds
		authCache.Store(apiKey, cacheEntry{
			participantID: participantID,
			expiresAt:     time.Now().Add(cacheTTL),
		})

		c.Set("participant_id", participantID)
		c.Next()
	}
}
