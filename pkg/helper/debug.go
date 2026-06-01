package helper

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// IsLaunchedByDebugger reports whether the process should use dev/test
// configuration. Returns true when either:
//   - the BUDDIEBOT_FORCE_DEV env var is set to a non-empty value, or
//   - the parent process is dlv.exe (running under the Go debugger).
//
// The env var is the escape hatch for "plain Run from the IDE without
// attaching a debugger" workflows — without it, a non-Delve launch defaults
// to PROD config, which is a footgun (prod token, prod Lavalink, prod DB
// path) when iterating locally. Set BUDDIEBOT_FORCE_DEV=1 in the IDE's Run
// configuration to land on Test config without paying the dlv attach cost.
func IsLaunchedByDebugger() bool {
	if os.Getenv("BUDDIEBOT_FORCE_DEV") != "" {
		return true
	}
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	if err == nil && strings.Contains(string(gopsOut), "\\dlv.exe") {
		return true
	}
	return false
}
