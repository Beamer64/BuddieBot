// Command pull-deploy fetches the latest BuddieBot release from GitHub and
// applies it to /opt/buddiebot.
//
// Designed as a manual-only replacement for the self-hosted GitHub runner
// deploy flow. The server initiates everything via HTTPS to api.github.com —
// no inbound ports, no runner registration, no code from pull requests ever
// running on this box.
//
// Layout produced matches scripts/deploy.sh exactly, so the existing systemd
// unit + config file setup keep working without change:
//
//	/opt/buddiebot/
//	├── current -> builds/<utc-ts>-<short-sha>/
//	├── builds/
//	│   ├── <newest>/
//	│   │   ├── buddiebot
//	│   │   └── config_files/config.yaml -> /opt/buddiebot/config.yaml
//	│   ├── <previous>/
//	│   └── ...
//	├── config.yaml                  (shared, server-managed)
//	├── current-version              (release tag of the live build)
//	└── bin/pull-deploy              (this binary)
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	repoOwner = "Beamer64"
	repoName  = "BuddieBot"
	apiBase   = "https://api.github.com"
	userAgent = "buddiebot-pull-deploy/1"
)

type config struct {
	buildRoot      string
	binAsset       string
	sha256Asset    string
	keep           int
	force          bool
	dryRun         bool
	healthDeadline time.Duration
	readyMarker    string
	service        string
}

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		log.Fatalf("pull-deploy: %v", err)
	}
}

func parseFlags() config {
	var c config
	flag.StringVar(&c.buildRoot, "build-root", "/opt/buddiebot", "base directory where the bot is installed")
	flag.StringVar(&c.binAsset, "binary-asset", "buddiebot", "name of the binary asset in the release")
	flag.StringVar(&c.sha256Asset, "sha256-asset", "buddiebot.sha256", "name of the sha256 asset in the release")
	flag.IntVar(&c.keep, "keep", 3, "number of build directories to keep after pruning")
	flag.BoolVar(&c.force, "force", false, "redeploy even if the current version already matches the latest release")
	flag.BoolVar(&c.dryRun, "dry-run", false, "stage the new release but skip the symlink swap and service restart")
	flag.DurationVar(&c.healthDeadline, "health-deadline", 30*time.Second, "max time to wait for the bot to log its ready marker after restart")
	flag.StringVar(&c.readyMarker, "ready-marker", "Logged in as", "journal substring that indicates a healthy startup")
	flag.StringVar(&c.service, "service", "buddiebot", "name of the systemd unit to restart")
	flag.Parse()
	return c
}

func run(cfg config) error {
	log.Printf("starting (build-root=%s, dry-run=%v, force=%v)", cfg.buildRoot, cfg.dryRun, cfg.force)

	rel, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("fetch latest release: %w", err)
	}
	log.Printf("latest release: %s (%d assets)", rel.TagName, len(rel.Assets))

	current := readCurrentVersion(cfg.buildRoot)
	if rel.TagName == current && !cfg.force {
		log.Printf("already at %s; nothing to do (use --force to redeploy)", current)
		return nil
	}
	log.Printf("deploying %s (current=%q)", rel.TagName, current)

	binURL, shaURL, err := findAssets(rel, cfg.binAsset, cfg.sha256Asset)
	if err != nil {
		return err
	}

	buildID := buildIDFor(rel.TagName, time.Now())
	buildDir := filepath.Join(cfg.buildRoot, "builds", buildID)
	if err := os.MkdirAll(filepath.Join(buildDir, "config_files"), 0o755); err != nil {
		return fmt.Errorf("mkdir build dir: %w", err)
	}
	log.Printf("staging into %s", buildDir)

	binPath := filepath.Join(buildDir, "buddiebot")
	if err := downloadTo(binURL, binPath); err != nil {
		return fmt.Errorf("download binary: %w", err)
	}

	expected, err := fetchHash(shaURL)
	if err != nil {
		return fmt.Errorf("fetch sha256: %w", err)
	}
	actual, err := sha256File(binPath)
	if err != nil {
		return fmt.Errorf("hash binary: %w", err)
	}
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch (expected %s, got %s) — refusing to deploy", expected, actual)
	}
	log.Printf("checksum verified (%s…)", actual[:16])

	if err := os.Chmod(binPath, 0o755); err != nil {
		return fmt.Errorf("chmod binary: %w", err)
	}

	// Symlink the build's config_files/config.yaml to the shared, server-managed
	// /opt/buddiebot/config.yaml — exactly what scripts/deploy.sh does, so the
	// systemd unit (with WorkingDirectory=/opt/buddiebot/current) finds it.
	configTarget := filepath.Join(cfg.buildRoot, "config.yaml")
	if _, err := os.Stat(configTarget); err != nil {
		return fmt.Errorf("server-managed %s missing; set it up once before deploy: %w", configTarget, err)
	}
	if err := os.Symlink(configTarget, filepath.Join(buildDir, "config_files", "config.yaml")); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("symlink config: %w", err)
	}

	if cfg.dryRun {
		log.Printf("dry-run: staged at %s — not swapping 'current' or restarting %s", buildDir, cfg.service)
		return nil
	}

	if err := atomicSwingSymlink(buildDir, filepath.Join(cfg.buildRoot, "current")); err != nil {
		return fmt.Errorf("swing 'current' symlink: %w", err)
	}
	log.Printf("swung 'current' to %s", buildDir)

	log.Printf("restarting %s.service", cfg.service)
	restartedAt := time.Now()
	if err := restartService(cfg.service); err != nil {
		return fmt.Errorf("restart service: %w", err)
	}

	if err := waitForReady(cfg.service, cfg.readyMarker, restartedAt, cfg.healthDeadline); err != nil {
		return fmt.Errorf(`%w

The new binary is staged at %s and 'current' points at it, but the readiness
check did not see the marker. To roll back to the previous build:

    ls /opt/buddiebot/builds/      # pick the previous directory
    ln -sfn /opt/buddiebot/builds/<previous>/ /opt/buddiebot/current.new
    mv -Tf /opt/buddiebot/current.new /opt/buddiebot/current
    sudo systemctl restart %s
    echo "<previous-tag>" | sudo tee /opt/buddiebot/current-version`, err, buildDir, cfg.service)
	}
	log.Printf("ready marker observed; deploy healthy")

	if err := os.WriteFile(filepath.Join(cfg.buildRoot, "current-version"), []byte(rel.TagName), 0o644); err != nil {
		return fmt.Errorf("write current-version: %w", err)
	}

	if err := pruneOldBuilds(filepath.Join(cfg.buildRoot, "builds"), cfg.keep); err != nil {
		log.Printf("warning: prune failed: %v", err)
	}

	log.Printf("deploy %s complete", rel.TagName)
	return nil
}

func fetchLatestRelease() (*release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", apiBase, repoOwner, repoName)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("GitHub API %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if r.TagName == "" {
		return nil, errors.New("response missing tag_name")
	}
	return &r, nil
}

// findAssets locates the named binary and checksum assets in a release. Returns
// a friendly error if either is missing — keeps the caller from confusing
// "release published with wrong asset names" with a genuine network failure.
func findAssets(r *release, binName, shaName string) (binURL, shaURL string, err error) {
	for _, a := range r.Assets {
		switch a.Name {
		case binName:
			binURL = a.BrowserDownloadURL
		case shaName:
			shaURL = a.BrowserDownloadURL
		}
	}
	if binURL == "" {
		return "", "", fmt.Errorf("release %s is missing binary asset %q", r.TagName, binName)
	}
	if shaURL == "" {
		return "", "", fmt.Errorf("release %s is missing checksum asset %q", r.TagName, shaName)
	}
	return binURL, shaURL, nil
}

func readCurrentVersion(buildRoot string) string {
	b, err := os.ReadFile(filepath.Join(buildRoot, "current-version"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// buildIDFor matches scripts/deploy.sh's naming: a UTC timestamp prefix (so
// directories sort chronologically for pruning) plus the release's short SHA.
func buildIDFor(tag string, now time.Time) string {
	short := strings.TrimPrefix(tag, "release-")
	return now.UTC().Format("20060102150405") + "-" + short
}

func downloadTo(url, dest string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := (&http.Client{Timeout: 5 * time.Minute}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}
	// Write to a temp file in the same directory, then atomic-rename. Avoids
	// a partial-download being visible at the final path if we crash.
	tmp, err := os.CreateTemp(filepath.Dir(dest), filepath.Base(dest)+".partial-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		_ = os.Remove(tmpName)
	}()
	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, dest)
}

func fetchHash(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return "", err
	}
	return parseShaLine(string(b))
}

// parseShaLine extracts a hex digest from `sha256sum`-style output:
//
//	"abcdef…1234  buddiebot"
//
// Also accepts a bare 64-char hex digest. Returns the digest lowercased.
func parseShaLine(s string) (string, error) {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, " \t"); i > 0 {
		s = s[:i]
	}
	if len(s) != 64 {
		return "", fmt.Errorf("expected 64-hex-char sha256, got %d chars", len(s))
	}
	if _, err := hex.DecodeString(s); err != nil {
		return "", fmt.Errorf("not hex: %w", err)
	}
	return strings.ToLower(s), nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// atomicSwingSymlink points linkPath at target via the temp-symlink + rename
// idiom. On Linux rename(2) is atomic for symlinks; matches deploy.sh exactly.
func atomicSwingSymlink(target, linkPath string) error {
	tmp := linkPath + ".new"
	_ = os.Remove(tmp)
	if err := os.Symlink(target, tmp); err != nil {
		return err
	}
	return os.Rename(tmp, linkPath)
}

func restartService(name string) error {
	return runQuiet("sudo", "systemctl", "restart", name)
}

func runQuiet(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// waitForReady polls journalctl every second until the bot logs `marker` (the
// "Logged in as …" line from ReadyHandler) or the deadline expires. Bails out
// early if the service crashes before reaching the marker, so we don't waste
// the full deadline on a dead process.
func waitForReady(service, marker string, restartedAt time.Time, deadline time.Duration) error {
	deadlineAt := time.Now().Add(deadline)
	since := restartedAt.Add(-2 * time.Second).Format("2006-01-02 15:04:05")
	for time.Now().Before(deadlineAt) {
		out, err := exec.Command("journalctl", "-u", service, "--since", since, "--no-pager", "-o", "cat").Output()
		if err == nil && strings.Contains(string(out), marker) {
			return nil
		}
		if !isServiceActive(service) {
			return fmt.Errorf("service %s is not active after restart", service)
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("did not observe ready marker %q within %s", marker, deadline)
}

func isServiceActive(name string) bool {
	out, _ := exec.Command("systemctl", "is-active", name).Output()
	state := strings.TrimSpace(string(out))
	return state == "active" || state == "activating"
}

// pruneOldBuilds keeps the `keep` newest directories under buildsDir (sorted by
// name — the YYYYMMDDHHMMSS- prefix makes that chronological) and removes the
// rest. Matches scripts/deploy.sh's pruning behaviour.
func pruneOldBuilds(buildsDir string, keep int) error {
	entries, err := os.ReadDir(buildsDir)
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) <= keep {
		return nil
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	for _, old := range names[keep:] {
		log.Printf("pruning %s", old)
		if err := os.RemoveAll(filepath.Join(buildsDir, old)); err != nil {
			return err
		}
	}
	return nil
}
