//go:build !windows

package browser

import (
	"os/exec"
	"runtime"
)

// Open launches the target URL in the default web browser on Unix / macOS.
func Open(url string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", url)
	} else {
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
