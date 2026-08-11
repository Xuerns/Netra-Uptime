package static

import (
	"embed"
	"io/fs"
	"net/http"
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
