package commands

import (
	"os/exec"
	"runtime"
)

func Open(url string) int {
	var cmd string
	args := []string{}

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "*bsd", ...
		cmd = "xdg-open"
	}

	if err := exec.Command(cmd, append(args, url)...).Start(); err != nil {
		return 1
	}
	return 0
}
