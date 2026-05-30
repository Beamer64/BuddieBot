package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseShaLine(t *testing.T) {
	hex64 := strings.Repeat("a", 64)
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{hex64 + "  buddiebot", hex64, false},           // sha256sum text mode
		{hex64 + " *buddiebot", hex64, false},           // sha256sum binary mode
		{strings.ToUpper(hex64), hex64, false},          // uppercase normalised down
		{"  " + hex64 + "\t buddiebot\n", hex64, false}, // surrounding whitespace
		{hex64, hex64, false},                           // bare digest
		{"too short", "", true},                         // wrong length
		{strings.Repeat("z", 64), "", true},             // not hex
		{"", "", true},                                  // empty
	}
	for _, c := range cases {
		got, err := parseShaLine(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("parseShaLine(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
			continue
		}
		if !c.wantErr && got != c.want {
			t.Errorf("parseShaLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFindAssets(t *testing.T) {
	r := &release{
		TagName: "release-abc1234",
		Assets: []asset{
			{Name: "buddiebot", BrowserDownloadURL: "https://example.test/bin"},
			{Name: "buddiebot.sha256", BrowserDownloadURL: "https://example.test/sha"},
			{Name: "ignored", BrowserDownloadURL: "https://example.test/other"},
		},
	}
	bin, sha, err := findAssets(r, "buddiebot", "buddiebot.sha256")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bin != "https://example.test/bin" || sha != "https://example.test/sha" {
		t.Errorf("got bin=%q sha=%q", bin, sha)
	}

	if _, _, err := findAssets(r, "missing", "buddiebot.sha256"); err == nil {
		t.Error("expected error for missing binary asset")
	}
	if _, _, err := findAssets(r, "buddiebot", "missing.sha256"); err == nil {
		t.Error("expected error for missing checksum asset")
	}
}

func TestDetectJournalPermissionIssue(t *testing.T) {
	cases := []struct {
		name      string
		stderr    string
		wantFound bool
	}{
		{
			name:      "ubuntu-style permissions warning",
			stderr:    "No journal files were opened due to insufficient permissions.",
			wantFound: true,
		},
		{
			name:      "hint variant some distros emit",
			stderr:    "Hint: You are currently not seeing messages from other users and the system.\n      Users in groups 'adm', 'systemd-journal', 'wheel' can see all messages.",
			wantFound: true,
		},
		{
			name:      "case insensitive match",
			stderr:    "INSUFFICIENT PERMISSIONS to read journal.",
			wantFound: true,
		},
		{
			name:      "unrelated journalctl error is not a permissions issue",
			stderr:    "Failed to seek to head: Invalid argument",
			wantFound: false,
		},
		{name: "clean stderr", stderr: "", wantFound: false},
		{name: "noise only", stderr: "\n   \n", wantFound: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			msg, ok := detectJournalPermissionIssue(c.stderr)
			if ok != c.wantFound {
				t.Errorf("detectJournalPermissionIssue(%q) ok=%v, want %v", c.stderr, ok, c.wantFound)
			}
			if c.wantFound && strings.TrimSpace(msg) == "" {
				t.Errorf("flagged but returned empty msg (input %q)", c.stderr)
			}
			if !c.wantFound && msg != "" {
				t.Errorf("not flagged but msg=%q", msg)
			}
		})
	}
}

func TestHumanSize(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"}, // fractional formatting
		{1024 * 1024, "1.0 MiB"},
		{13 * 1024 * 1024, "13.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
		{2 * 1024 * 1024 * 1024 * 1024, "2.0 TiB"},
	}
	for _, c := range cases {
		got := humanSize(c.in)
		if got != c.want {
			t.Errorf("humanSize(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBuildIDFor(t *testing.T) {
	now := time.Date(2026, 5, 29, 14, 30, 0, 0, time.UTC)
	got := buildIDFor("release-abc1234", now)
	want := "20260529143000-abc1234"
	if got != want {
		t.Errorf("buildIDFor: got %q, want %q", got, want)
	}

	// Falls back gracefully if the tag has no "release-" prefix.
	got = buildIDFor("v1.2.3", now)
	if !strings.HasSuffix(got, "-v1.2.3") {
		t.Errorf("buildIDFor: untagged input should pass through; got %q", got)
	}
}
