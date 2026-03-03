package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ubothub/backend/pkg/token"
)

// JWTAuth returns a middleware that validates JWT bearer tokens from the
// Authorization header, checks the token blacklist in Redis, and injects
// the authenticated user's ID and role into the request context.
func JWTAuth(tokenMgr *token.Manager, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    10003,
				"message": "missing authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    10003,
				"message": "invalid authorization format",
			})
			return
		}

		claims, err := tokenMgr.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    11003,
				"message": "invalid or expired token",
			})
			return
		}

		// Check token blacklist.
		blacklistKey := "jwt:blacklist:" + claims.ID
		exists, err := rdb.Exists(context.Background(), blacklistKey).Result()
		if err == nil && exists > 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    11004,
				"message": "token revoked",
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// BlacklistToken adds a JWT token ID to the Redis blacklist with a TTL
// matching the token's remaining validity period.
func BlacklistToken(rdb *redis.Client, jti string, expiry time.Time) error {
	ttl := time.Until(expiry)
	if ttl <= 0 {
		return nil
	}
	return rdb.Set(context.Background(), "jwt:blacklist:"+jti, "1", ttl).Err()
}
