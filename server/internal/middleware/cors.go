package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupCORS() gin.HandlerFunc {
	origin := strings.TrimSpace(os.Getenv("ALLOWED_ORIGIN"))
	if origin == "" {
		port := strings.TrimSpace(os.Getenv("PORT"))
		if port == "" {
			port = "27346"
		}
		origin = "http://localhost:" + port
	}

	return cors.New(cors.Config{
		AllowOrigins:     []string{origin},
		AllowMethods:     []string{"POST", "PUT", "GET", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
