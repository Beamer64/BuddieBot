// Package lavalink_runner spawns a local Lavalink Java service as a child
// process for development. Production deployments run Lavalink as a
// separately-managed systemd unit (or similar) and skip this package
// entirely.
//
// On graceful shutdown the runner kills the Java child via Stop(). For
// hard kills (force-quit, panic, SIGKILL), per-platform OS hooks ensure
// the Java child dies with the bot: Pdeathsig on Linux, Job Objects on
// Windows. macOS has no equivalent built-in mechanism — graceful stop
// works there but a hard-killed bot can leave Java orphaned.
package lavalink_runner

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Runner manages a child Java process running Lavalink.
type Runner struct {
	cmd *exec.Cmd
}

// Start spawns `java -Xmx512M -jar <jarPath>` from workDir, then polls
// readyURL (sending password as the Authorization header — Lavalink v4
// requires it) until it returns 200 OK or timeout elapses. On failure
// the child is killed. Paths are normalized to absolute so the spawned
// Java process sees them correctly regardless of caller cwd.
func Start(jarPath, workDir, readyURL, password string, timeout time.Duration) (*Runner, error) {
	absJar, err := filepath.Abs(jarPath)
	if err != nil {
		return nil, fmt.Errorf("resolve jar path: %w", err)
	}
	absDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve work dir: %w", err)
	}

	if _, err := os.Stat(absJar); err != nil {
		return nil, fmt.Errorf("lavalink jar not found at %s: %w", absJar, err)
	}
	if _, err := exec.LookPath("java"); err != nil {
		return nil, fmt.Errorf("java executable not found in PATH (install JDK 17+): %w", err)
	}

	cmd := exec.Command("java", "-Xmx512M", "-jar", absJar)
	cmd.Dir = absDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := prepareCmd(cmd); err != nil {
		return nil, fmt.Errorf("prepare cmd: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start lavalink: %w", err)
	}
	if err := attachReaper(cmd); err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("attach reaper: %w", err)
	}

	if err := waitReady(readyURL, password, timeout); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, err
	}

	return &Runner{cmd: cmd}, nil
}

func waitReady(url, password string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("build readiness request: %w", err)
		}
		req.Header.Set("Authorization", password)

		resp, err := client.Do(req)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return errors.New("lavalink did not become ready before timeout")
}

// Stop kills the Java child. Idempotent.
func (r *Runner) Stop() {
	if r == nil || r.cmd == nil || r.cmd.Process == nil {
		return
	}
	_ = r.cmd.Process.Kill()
	_ = r.cmd.Wait()
}
