//go:build windows

package browser

import (
	"fmt"
	"net/url"
	"os/exec"
	"syscall"
)

// Open launches the target URL in the default web browser without popping a command window.
func Open(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("invalid browser url %q: only http/https allowed", rawURL)
	}

	safeURL := parsed.String()

	// #nosec G204 -- safeURL is strictly validated to http/https schemes and launched with HideWindow
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", safeURL)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		// #nosec G204 -- fallback execution with validated safeURL
		fallback := exec.Command("cmd", "/c", "start", "", safeURL)
		fallback.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		return fallback.Start()
	}
	return nil
}
