//go:build !linux && !windows

package lavalink_runner

import "os/exec"

// No Pdeathsig / Job Object equivalent on macOS. Graceful Stop works;
// hard-killed bot can orphan Java (dev-only path, README documents cleanup).
func prepareCmd(cmd *exec.Cmd) error {
	return nil
}

func attachReaper(cmd *exec.Cmd) error {
	return nil
}
