package auth

import (
	"os/exec"
	"runtime"
)

// OpenBrowser opens the given URL in the user's default browser.
func OpenBrowser(url string) error {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	return exec.Command(cmd, url).Start()
}
