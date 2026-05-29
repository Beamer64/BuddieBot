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
