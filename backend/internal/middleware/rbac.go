package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole returns a middleware that restricts access to users with
// one of the specified roles. Must be applied after JWTAuth middleware.
func RequireRole(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    10004,
				"message": "forbidden",
			})
			return
		}

		if _, ok := roleSet[role.(string)]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    10004,
				"message": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}
