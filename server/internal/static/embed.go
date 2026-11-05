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

// RegisterStaticRoutes serves the embedded client files and handles SPA fallback for non-API routes.
func RegisterStaticRoutes(r *gin.Engine) {
	subFS, err := fs.Sub(BuildFS, "build")
	if err != nil {
		return
	}

	fileServer := http.FileServer(http.FS(subFS))

	// Serve /_app/env.js dynamically so users can configure PUBLIC_API_URL
	// at runtime via environment variable without rebuilding the image.
	r.GET("/_app/env.js", func(c *gin.Context) {
		apiURL := strings.TrimSpace(os.Getenv("PUBLIC_API_URL"))
		if apiURL == "" {
			scheme := "http"
			if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
				scheme = "https"
			}
			host := c.Request.Host
			if host == "" {
				host = "localhost:27346"
			}
			apiURL = fmt.Sprintf("%s://%s/api/v1", scheme, host)
		}

		js := fmt.Sprintf(
			`export const env={PUBLIC_API_URL:%q}`,
			apiURL,
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

		// SPA Fallback: Serve index.html for client routes
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
