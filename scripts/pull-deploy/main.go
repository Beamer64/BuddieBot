// Command pull-deploy fetches the latest BuddieBot release from GitHub and
// applies it to /opt/buddiebot.
//
// Designed as a manual-only replacement for the self-hosted GitHub runner
// deploy flow. The server initiates everything via HTTPS to api.github.com —
// no inbound ports, no runner registration, no code from pull requests ever
// running on this box.
//
// Layout produced matches scripts/deploy.sh's historical layout exactly, so
// the existing systemd unit + config file setup keep working without change:
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
	"bytes"
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

	// bannerBar is the heavy box-drawing rule used to bracket each run's
	// output. Makes consecutive deploys easy to tell apart in the journal.
	bannerBar = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
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
	// systemd already timestamps every line it captures from us; emitting Go's
	// LstdFlags on top would double up the time prefix. Plain lines look clean
	// in journalctl output.
	log.SetFlags(0)

	cfg := parseFlags()
	started := time.Now()
	openingBanner(cfg, started)

	err := run(cfg)
	closingBanner(err, time.Since(started))

	if err != nil {
		os.Exit(1)
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

// openingBanner / closingBanner / banner / step / note are the visual layer
// for the journal output. All consecutive runs land in the same journal, so a
// heavy rule between them is the difference between "this is clearly a fresh
// deploy" and "wait, where did the previous one end?".

func openingBanner(cfg config, t time.Time) {
	banner(
		fmt.Sprintf("pull-deploy · %s · pid %d", t.UTC().Format("2006-01-02 15:04:05 UTC"), os.Getpid()),
		fmt.Sprintf("build-root=%s  force=%v  dry-run=%v", cfg.buildRoot, cfg.force, cfg.dryRun),
	)
}

func closingBanner(err error, dur time.Duration) {
	status := "OK"
	if err != nil {
		log.Printf("\nERROR: %v", err)
		status = "FAILED"
	}
	banner(fmt.Sprintf("pull-deploy · %s · %s", status, dur.Round(time.Millisecond)))
}

func banner(lines ...string) {
	log.Println(bannerBar)
	for _, l := range lines {
		log.Println(" " + l)
	}
	log.Println(bannerBar)
}

// step prints a top-level action ("==> staging release-XXX ...").
func step(format string, args ...any) {
	log.Printf("==> "+format, args...)
}

// note prints a sub-step / detail line, indented under the preceding step.
func note(format string, args ...any) {
	log.Printf("    "+format, args...)
}

func run(cfg config) error {
	rel, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("fetch latest release: %w", err)
	}
	step("latest release: %s (%d assets)", rel.TagName, len(rel.Assets))

	current := readCurrentVersion(cfg.buildRoot)
	if rel.TagName == current && !cfg.force {
		step("current version: %s", current)
		note("nothing to do (use --force to redeploy)")
		return nil
	}
	switch {
	case current == "":
		step("current version: (none) → deploying %s", rel.TagName)
	case cfg.force && current == rel.TagName:
		step("current version: %s → forcing redeploy", current)
	default:
		step("current version: %s → upgrading to %s", current, rel.TagName)
	}

	binURL, shaURL, err := findAssets(rel, cfg.binAsset, cfg.sha256Asset)
	if err != nil {
		return err
	}

	buildID := buildIDFor(rel.TagName, time.Now())
	buildDir := filepath.Join(cfg.buildRoot, "builds", buildID)
	if err := os.MkdirAll(filepath.Join(buildDir, "config_files"), 0o755); err != nil {
		return fmt.Errorf("mkdir build dir: %w", err)
	}
	step("staging %s → %s", rel.TagName, buildDir)

	binPath := filepath.Join(buildDir, "buddiebot")
	if err := downloadTo(binURL, binPath); err != nil {
		return fmt.Errorf("download binary: %w", err)
	}
	if info, err := os.Stat(binPath); err == nil {
		note("downloaded binary (%s)", humanSize(info.Size()))
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
	note("checksum verified: %s…", actual[:16])

	if err := os.Chmod(binPath, 0o755); err != nil {
		return fmt.Errorf("chmod binary: %w", err)
	}

	// Symlink the build's config_files/config.yaml to the shared, server-managed
	// /opt/buddiebot/config.yaml — so the buddiebot.service unit (with
	// WorkingDirectory=/opt/buddiebot/current) finds it in the expected spot.
	configTarget := filepath.Join(cfg.buildRoot, "config.yaml")
	if _, err := os.Stat(configTarget); err != nil {
		return fmt.Errorf("server-managed %s missing; set it up once before deploy: %w", configTarget, err)
	}
	if err := os.Symlink(configTarget, filepath.Join(buildDir, "config_files", "config.yaml")); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("symlink config: %w", err)
	}

	if cfg.dryRun {
		step("dry-run: skipped 'current' swap and service restart")
		note("staged at %s", buildDir)
		return nil
	}

	if err := atomicSwingSymlink(buildDir, filepath.Join(cfg.buildRoot, "current")); err != nil {
		return fmt.Errorf("swing 'current' symlink: %w", err)
	}
	step("swapped 'current' symlink → builds/%s", buildID)

	step("restarting %s.service", cfg.service)
	restartedAt := time.Now()
	if err := restartService(cfg.service); err != nil {
		return fmt.Errorf("restart service: %w", err)
	}

	step("waiting for ready marker %q (deadline %s)", cfg.readyMarker, cfg.healthDeadline)
	waited, err := waitForReady(cfg.service, cfg.readyMarker, restartedAt, cfg.healthDeadline)
	if err != nil {
		return fmt.Errorf(`%w

The new binary is staged at %s and 'current' points at it, but the readiness
check did not pass. To roll back to the previous build:

    ls /opt/buddiebot/builds/      # pick the previous directory
    ln -sfn /opt/buddiebot/builds/<previous>/ /opt/buddiebot/current.new
    mv -Tf /opt/buddiebot/current.new /opt/buddiebot/current
    sudo systemctl restart %s
    echo "<previous-tag>" | sudo tee /opt/buddiebot/current-version`, err, buildDir, cfg.service)
	}
	note("observed at t+%s", waited.Round(time.Millisecond))

	if err := os.WriteFile(filepath.Join(cfg.buildRoot, "current-version"), []byte(rel.TagName), 0o644); err != nil {
		return fmt.Errorf("write current-version: %w", err)
	}
	step("updated current-version → %s", rel.TagName)

	pruned, err := pruneOldBuilds(filepath.Join(cfg.buildRoot, "builds"), cfg.keep)
	if err != nil {
		note("warning: prune failed: %v", err)
	} else if pruned > 0 {
		step("pruned %d old build(s) (kept %d newest)", pruned, cfg.keep)
	}

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

// runJournalctl runs the bot's journal scan, returning stdout, stderr, and
// any process error separately. exec.Cmd.Output() throws stderr away, which
// hid permission warnings ("No journal files were opened …") and silently
// turned them into health-check timeouts — capturing both lets the caller
// distinguish "log line not there yet" from "I literally can't read this
// service's logs at all".
func runJournalctl(service, since string) (stdout, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer
	c := exec.Command("journalctl", "-u", service, "--since", since, "--no-pager", "-o", "cat")
	c.Stdout = &outBuf
	c.Stderr = &errBuf
	err = c.Run()
	return outBuf.String(), errBuf.String(), err
}

// detectJournalPermissionIssue returns (msg, true) when journalctl's stderr
// indicates the running user lacks permission to read the system journal for
// the requested service. The returned msg is the first non-empty stderr line,
// for context in the error.
func detectJournalPermissionIssue(stderr string) (string, bool) {
	s := strings.ToLower(stderr)
	for _, needle := range []string{
		"no journal files were opened",
		"insufficient permissions",
		"you are currently not seeing messages from other users",
	} {
		if strings.Contains(s, needle) {
			for _, line := range strings.Split(stderr, "\n") {
				if trimmed := strings.TrimSpace(line); trimmed != "" {
					return trimmed, true
				}
			}
			return strings.TrimSpace(stderr), true
		}
	}
	return "", false
}

// waitForReady polls journalctl every second until the bot logs `marker` (the
// "Logged in as …" line from ReadyHandler) or the deadline expires. Bails
// early on three conditions:
//
//   - The service has dropped to inactive — there's nothing to wait for.
//   - journalctl stderr indicates a permission issue (e.g. the running user
//     isn't in systemd-journal) — fail with a clear, actionable error instead
//     of timing out silently.
//
// Returns the wait duration so the caller can log "observed at t+1.2s".
func waitForReady(service, marker string, restartedAt time.Time, deadline time.Duration) (time.Duration, error) {
	startedWaiting := time.Now()
	deadlineAt := startedWaiting.Add(deadline)
	since := restartedAt.Add(-2 * time.Second).Format("2006-01-02 15:04:05")

	for time.Now().Before(deadlineAt) {
		stdout, stderr, _ := runJournalctl(service, since)

		if msg, ok := detectJournalPermissionIssue(stderr); ok {
			return time.Since(startedWaiting), fmt.Errorf(`cannot read %s logs from journalctl: %s

This usually means the user running pull-deploy isn't in the
'systemd-journal' group. Fix by adding the user (most likely
'buddiebot') to that group:

    sudo usermod -aG systemd-journal buddiebot

…then re-run the deploy. Group membership takes effect on the next
process spawn, so 'sudo systemctl start buddiebot-deploy.service'
picks it up automatically`, service, msg)
		}
		if strings.Contains(stdout, marker) {
			return time.Since(startedWaiting), nil
		}
		if !isServiceActive(service) {
			return time.Since(startedWaiting), fmt.Errorf("service %s is not active after restart", service)
		}
		time.Sleep(1 * time.Second)
	}
	return time.Since(startedWaiting), fmt.Errorf("did not observe ready marker %q within %s", marker, deadline)
}

func isServiceActive(name string) bool {
	out, _ := exec.Command("systemctl", "is-active", name).Output()
	state := strings.TrimSpace(string(out))
	return state == "active" || state == "activating"
}

// pruneOldBuilds keeps the `keep` newest directories under buildsDir (sorted
// by name — the YYYYMMDDHHMMSS- prefix makes that chronological) and removes
// the rest. Returns the number of directories actually pruned.
func pruneOldBuilds(buildsDir string, keep int) (int, error) {
	entries, err := os.ReadDir(buildsDir)
	if err != nil {
		return 0, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) <= keep {
		return 0, nil
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	pruned := 0
	for _, old := range names[keep:] {
		if err := os.RemoveAll(filepath.Join(buildsDir, old)); err != nil {
			return pruned, err
		}
		pruned++
	}
	return pruned, nil
}

// humanSize formats a byte count in binary-prefixed units (KiB/MiB/…). Used
// only for the "downloaded binary (12.3 MiB)" note — never for arithmetic.
func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
