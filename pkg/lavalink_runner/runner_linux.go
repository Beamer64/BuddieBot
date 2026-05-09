//go:build linux

package lavalink_runner

import (
	"os/exec"
	"syscall"
)

// prepareCmd asks the kernel to send SIGKILL to the child if our process
// dies. Works for any death path — graceful, panic, or SIGKILL.
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
