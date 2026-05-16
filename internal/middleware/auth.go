package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"meta-prompt/internal/store"
)

func APIKeyAuth(apiKeyStore *store.APIKeyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing api key"})
			return
		}

		apiKey, err := apiKeyStore.ValidateKey(key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}

		c.Set("api_key_id", apiKey.ID)
		c.Set("user_id", apiKey.UserID)
		c.Next()
	}
}
