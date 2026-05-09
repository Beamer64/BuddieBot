//go:build !linux && !windows

package lavalink_runner

import "os/exec"

// macOS and others have no built-in equivalent of Pdeathsig or Job
// Objects. Graceful stop still works; a hard-killed bot may leave Java
// orphaned. Acceptable for a dev-only path; document a manual cleanup
// command in the README.
func prepareCmd(cmd *exec.Cmd) error {
	return nil
}

func attachReaper(cmd *exec.Cmd) error {
	return nil
}
