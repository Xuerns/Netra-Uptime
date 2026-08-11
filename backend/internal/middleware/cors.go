package middleware

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupCORS() gin.HandlerFunc {
	Origin := os.Getenv("ALLOWED_ORIGIN")

	return cors.New(cors.Config{
		AllowOrigins:     []string{Origin},
		AllowMethods:     []string{"POST", "PUT", "GET", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
