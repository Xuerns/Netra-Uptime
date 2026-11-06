//go:build !windows

package browser

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
)

// Open launches the target URL in the default web browser on Unix / macOS.
func Open(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("invalid browser url %q: only http/https allowed", rawURL)
	}

	safeURL := parsed.String()

	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// #nosec G204 -- safeURL is strictly validated to http/https schemes
		cmd = exec.Command("open", safeURL)
	} else {
		// #nosec G204 -- safeURL is strictly validated to http/https schemes
		cmd = exec.Command("xdg-open", safeURL)
	}
	return cmd.Start()
}
