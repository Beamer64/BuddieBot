package helper

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// IsLaunchedByDebugger determines if the application is being run by the Delve debugger.
func IsLaunchedByDebugger() bool {
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	if err == nil && strings.Contains(string(gopsOut), "\\dlv.exe") {
		return true
	}
	return false
}
