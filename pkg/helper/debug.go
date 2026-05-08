package helper

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// IsLaunchedByDebugger determines if the application is being run by the
// Delve debugger. Requires the gops executable on PATH; see
// https://github.com/google/gops. Windows-specific: looks for "\\dlv.exe" in
// the parent process info.
func IsLaunchedByDebugger() bool {
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	if err == nil && strings.Contains(string(gopsOut), "\\dlv.exe") {
		return true
	}
	return false
}
