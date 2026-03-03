package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

const requestIDHeader = "X-Request-ID"

// RequestID generates a unique request ID for each incoming request and
// propagates it through the context and response headers for traceability.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(requestIDHeader)
		if rid == "" {
			rid = xid.New().String()
		}
		c.Set("request_id", rid)
		c.Header(requestIDHeader, rid)
		c.Next()
	}
}
