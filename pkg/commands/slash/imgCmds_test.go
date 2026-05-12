package slash

import "testing"

// All these tests hit Dagpi's image-processing endpoints with the
// publicly-reachable testImageURL. They share the "dagpi" rate-limit
// bucket with the text endpoints in getCmds_test.go, so the full Dagpi
// surface serializes — protects against quota bursts.
//
// The five tests below are representative of the three argument shapes
// Dagpi image endpoints come in: single-image (Pixelate, Ascii, Invert),
// image+text (Retromeme), and two-image (Slap). If a wider Dagpi change
// breaks them all the same way, that signal is enough — testing all 60+
// individual transforms would be quota-hostile.

// TestImgPixelate — Dagpi /image/pixelate (single-image transform).
func TestImgPixelate(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	requireKey(t, cfg.Configs.Keys.DagpiAPIkey, "DagpiAPIkey")
	defer rateLimit("dagpi")()

	out, err := cfg.Clients.Dagpi.Pixelate(testImageURL)
	if err != nil {
		t.Fatalf("Dagpi Pixelate: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("Pixelate returned empty bytes — Dagpi endpoint may have changed")
	}
}

// TestImgAscii — Dagpi /image/ascii (alternative single-image transform).
func TestImgAscii(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	requireKey(t, cfg.Configs.Keys.DagpiAPIkey, "DagpiAPIkey")
	defer rateLimit("dagpi")()

	out, err := cfg.Clients.Dagpi.Ascii(testImageURL)
	if err != nil {
		t.Fatalf("Dagpi Ascii: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("Ascii returned empty bytes — Dagpi endpoint may have changed")
	}
}

// TestImgInvert — Dagpi /image/invert.
func TestImgInvert(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	requireKey(t, cfg.Configs.Keys.DagpiAPIkey, "DagpiAPIkey")
	defer rateLimit("dagpi")()

	out, err := cfg.Clients.Dagpi.Invert(testImageURL)
	if err != nil {
		t.Fatalf("Dagpi Invert: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("Invert returned empty bytes — Dagpi endpoint may have changed")
	}
}

// TestImgRetromeme — Dagpi /image/retromeme (image + two strings).
func TestImgRetromeme(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	requireKey(t, cfg.Configs.Keys.DagpiAPIkey, "DagpiAPIkey")
	defer rateLimit("dagpi")()

	out, err := cfg.Clients.Dagpi.Retromeme(testImageURL, "TOP", "BOTTOM")
	if err != nil {
		t.Fatalf("Dagpi Retromeme: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("Retromeme returned empty bytes — Dagpi endpoint may have changed")
	}
}

// TestImgSlap — Dagpi /image/slap (two-image transform).
func TestImgSlap(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	requireKey(t, cfg.Configs.Keys.DagpiAPIkey, "DagpiAPIkey")
	defer rateLimit("dagpi")()

	out, err := cfg.Clients.Dagpi.Slap(testImageURL, testImageURL)
	if err != nil {
		t.Fatalf("Dagpi Slap: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("Slap returned empty bytes — Dagpi endpoint may have changed")
	}
}
