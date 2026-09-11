package static

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:build
var BuildFS embed.FS

// RegisterStaticRoutes serves the embedded frontend files and handles SPA fallback for non-API routes.
func RegisterStaticRoutes(r *gin.Engine) {
	subFS, err := fs.Sub(BuildFS, "build")
	if err != nil {
		return
	}

	fileServer := http.FileServer(http.FS(subFS))

	// Serve /_app/env.js dynamically so users can configure PUBLIC_API_URL
	// at runtime via environment variable without rebuilding the image.
	r.GET("/_app/env.js", func(c *gin.Context) {
		apiURL := os.Getenv("PUBLIC_API_URL")
		googleClientID := os.Getenv("PUBLIC_GOOGLE_CLIENT_ID")
		googleRedirectURL := os.Getenv("PUBLIC_GOOGLE_REDIRECT_URL")
		appMode := os.Getenv("PUBLIC_APP_MODE")

		js := fmt.Sprintf(
			`export const env={PUBLIC_API_URL:%q,PUBLIC_GOOGLE_CLIENT_ID:%q,PUBLIC_GOOGLE_REDIRECT_URL:%q,PUBLIC_APP_MODE:%q}`,
			apiURL, googleClientID, googleRedirectURL, appMode,
		)

		c.Header("Content-Type", "application/javascript; charset=utf-8")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.String(http.StatusOK, js)
	})

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Don't intercept API or WebSocket routes
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/ws") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
			return
		}

		// Check if file exists in static FS
		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath != "" {
			f, err := subFS.Open(cleanPath)
			if err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// SPA Fallback: Serve index.html for frontend routes
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
