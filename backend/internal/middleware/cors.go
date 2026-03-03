package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that configures Cross-Origin Resource Sharing
// headers based on the application run mode. Production restricts origins
// to explicit allowed domains; development permits all origins.
func CORS(mode string, allowedOrigins []string) gin.HandlerFunc {
	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	}

	if mode == "debug" {
		config.AllowAllOrigins = true
		config.AllowCredentials = false
	} else {
		config.AllowOrigins = allowedOrigins
	}

	return cors.New(config)
}
