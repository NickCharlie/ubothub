package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CSRF returns a middleware that validates CSRF tokens for state-changing requests.
// Requests with a valid Authorization Bearer header are exempt because JWT-based
// auth is inherently CSRF-safe (the token is not auto-attached by the browser).
// For cookie-based flows, it uses the double-submit cookie pattern: the server sets
// a random token in a cookie, and the client must send the same value in the
// X-CSRF-Token header.
// The secure parameter controls whether the cookie requires HTTPS (set false for local HTTP dev).
func CSRF(secure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Safe methods don't need CSRF protection.
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			ensureCSRFCookie(c, secure)
			c.Next()
			return
		}

		// Bearer token auth is inherently CSRF-safe.
		if strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
			c.Next()
			return
		}

		// Validate double-submit: cookie value must match header value.
		cookieToken, err := c.Cookie("_csrf")
		if err != nil || cookieToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    40300,
				"message": "CSRF token missing",
			})
			return
		}

		headerToken := c.GetHeader("X-CSRF-Token")
		if headerToken == "" || headerToken != cookieToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    40300,
				"message": "CSRF token mismatch",
			})
			return
		}

		c.Next()
	}
}

// ensureCSRFCookie sets a CSRF cookie if one doesn't already exist.
func ensureCSRFCookie(c *gin.Context, secure bool) {
	if _, err := c.Cookie("_csrf"); err != nil {
		token := generateCSRFToken()
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("_csrf", token, 86400, "/", "", secure, false)
	}
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
