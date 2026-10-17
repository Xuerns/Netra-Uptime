package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders applies OWASP-recommended HTTP response security headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME-sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking by forbidding embedding in frames
		c.Header("X-Frame-Options", "DENY")

		// Legacy XSS filter protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Protect referrer leakage
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Disable unused browser hardware features
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		// Content Security Policy for API/backend
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com data:; img-src 'self' data: blob: https:; connect-src 'self' ws: wss:; frame-ancestors 'none';")

		// HTTP Strict Transport Security if TLS or behind HTTPS reverse proxy
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Next()
	}
}
