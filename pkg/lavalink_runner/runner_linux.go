//go:build linux

package lavalink_runner

import (
	"os/exec"
	"syscall"
)

// prepareCmd: Pdeathsig=SIGKILL ensures the child dies with us on any death path.
func prepareCmd(cmd *exec.Cmd) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Pdeathsig = syscall.SIGKILL
	return nil
}

func attachReaper(cmd *exec.Cmd) error {
	return nil
}
