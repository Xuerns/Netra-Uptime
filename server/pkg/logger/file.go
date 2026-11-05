package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"netra-uptime/config"
)

// SetupFileLogger redirects standard log output and Gin engine output
// to a netra.log file in the executable's directory.
// Returns a cleanup function that flushes and closes the file handle.
func SetupFileLogger() (func(), error) {
	logDir := config.GetExecutableDir()
	logPath := filepath.Clean(filepath.Join(logDir, "netra.log"))

	// #nosec G302,G304 -- netra.log is an internal log file with restrictive 0600 permissions located in the executable directory
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return func() {}, fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}

	log.SetOutput(file)
	gin.DefaultWriter = file
	gin.DefaultErrorWriter = file

	cleanup := func() {
		_ = file.Sync()
		_ = file.Close()
	}

	return cleanup, nil
}
