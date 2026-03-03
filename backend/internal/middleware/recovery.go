package middleware

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns a middleware that recovers from panics, logs the error,
// and returns a 500 response. Broken pipe errors are handled separately
// to avoid logging noise from client disconnects.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)

				if brokenPipe {
					logger.Warn("broken pipe",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					c.Error(err.(error))
					c.Abort()
					return
				}

				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("request", string(httpRequest)),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    10001,
					"message": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
