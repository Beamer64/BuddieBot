package helper

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// IsLaunchedByDebugger reports whether the parent process is dlv.exe.
func IsLaunchedByDebugger() bool {
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	if err == nil && strings.Contains(string(gopsOut), "\\dlv.exe") {
		return true
	}
	return false
}
