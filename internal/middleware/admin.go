package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"meta-prompt/internal/store"
)

func AdminAuth(userStore *store.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		user, err := userStore.GetByID(userID)
		if err != nil || user.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
