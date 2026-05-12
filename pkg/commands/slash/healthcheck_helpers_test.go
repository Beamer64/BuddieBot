package slash

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
)

// requireINTEGRATION skips the test unless INTEGRATION=true. Matches the
// existing pattern from pkg/web/web_test.go — default `go test ./...` is
// silent and fast; `INTEGRATION=true go test` runs live API checks.
func requireINTEGRATION(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("set INTEGRATION=true to run live API health checks")
	}
}

// requireINTEGRATIONHeavy gates slow / fragile tests (e.g. chromedp-based
// screenshot tests) behind a second env var so they don't slow down a
// normal health-check run.
func requireINTEGRATIONHeavy(t *testing.T) {
	t.Helper()
	requireINTEGRATION(t)
	if os.Getenv("INTEGRATION_HEAVY") != "true" {
		t.Skip("set INTEGRATION_HEAVY=true to also run slow / fragile health checks")
	}
}

// loadTestConfig reads config.yaml from the standard search paths,
// temporarily chdir-ing to the repo root so the existing relative-path
// search in config.ReadConfig still works regardless of which package
// the test is running in. Cached via sync.Once so we only do this work
// once per process even if every test calls it.
var (
	cfgOnce sync.Once
	cfgInst *config.Configs
	cfgErr  error
)

func loadTestConfig(t *testing.T) *config.Configs {
	t.Helper()
	cfgOnce.Do(func() {
		root, err := findRepoRoot()
		if err != nil {
			cfgErr = err
			return
		}
		original, err := os.Getwd()
		if err != nil {
			cfgErr = err
			return
		}
		if err := os.Chdir(root); err != nil {
			cfgErr = err
			return
		}
		defer func() {
			_ = os.Chdir(original)
		}()
		cfgInst, cfgErr = config.ReadConfig()
	})
	if cfgErr != nil {
		t.Fatalf("load test config: %v", cfgErr)
	}
	return cfgInst
}

// findRepoRoot walks up from cwd looking for go.mod so the test helpers
// can locate config_files / datasets regardless of which package's test
// directory we were invoked from.
func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd, nil
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return "", errors.New("repo root (go.mod) not found by walking up from " + cwd)
		}
		cwd = parent
	}
}

// requireKey skips with a clear message when the named config key is
// empty. Distinguishes "key not provisioned" (skip) from "API is broken"
// (fail) so a missing dev token doesn't masquerade as an outage.
func requireKey(t *testing.T, key, name string) {
	t.Helper()
	if key == "" {
		t.Skipf("%s not set in config.yaml — skipping health check", name)
	}
}

// rateLimit returns a release func that the test must defer-call. Holds
// a per-API mutex (so tests for commands sharing an API serialize against
// each other), and on release sleeps for rateLimitDelay() to give the
// upstream API breathing room before the next call.
//
//	release := rateLimit("dagpi")
//	defer release()
var (
	rateLimitsMu sync.Mutex
	rateLimits   = map[string]*sync.Mutex{}
)

func rateLimit(api string) func() {
	rateLimitsMu.Lock()
	m, ok := rateLimits[api]
	if !ok {
		m = &sync.Mutex{}
		rateLimits[api] = m
	}
	rateLimitsMu.Unlock()
	m.Lock()
	return func() {
		time.Sleep(rateLimitDelay())
		m.Unlock()
	}
}

func rateLimitDelay() time.Duration {
	if v := os.Getenv("API_RATE_LIMIT_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms >= 0 {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return 2 * time.Second
}

// testImageURL is a stable, publicly-fetchable URL for the bot's own logo.
// Image-processing tests pass it to Dagpi (and similar services) — those
// servers reach out from their network, so the URL must be reachable from
// the internet (a localhost httptest.Server would not be).
//
// Uses raw.githubusercontent.com because the github.com/.../blob/... URL
// serves an HTML page, not the raw image bytes.
const testImageURL = "https://raw.githubusercontent.com/Beamer64/BuddieBot/master/res/repo_imgs/BuddieBot.png"
