package slash

import (
	"testing"
)

func TestDrawCistLines_Dimensions(t *testing.T) {
	img := drawCistLines(false, "1685")
	if img.Bounds().Dx() != 200 || img.Bounds().Dy() != 200 {
		t.Errorf("expected 200×200 canvas, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestDrawCistLines_VerticalBarDrawn(t *testing.T) {
	// Center vertical line goes from (100,20) to (100,180). The midpoint
	// (100,100) must be on it regardless of which digits were drawn.
	img := drawCistLines(false, "0000")
	if img.RGBAAt(100, 100).A == 0 {
		t.Error("center vertical line missing — pixel at (100,100) is transparent")
	}
}

func TestDrawCistLines_NegativeAddsSlash(t *testing.T) {
	// The negative slash runs horizontally from (60,100) to (140,100).
	// Pick a pixel that's on the slash but off the vertical bar (x≠100)
	// so the test isolates the slash specifically.
	pos := drawCistLines(false, "0005")
	neg := drawCistLines(true, "0005")
	if pos.RGBAAt(70, 100).A != 0 {
		t.Error("positive variant unexpectedly drew at (70,100)")
	}
	if neg.RGBAAt(70, 100).A == 0 {
		t.Error("negative variant missing slash at (70,100)")
	}
}

func TestDrawCistLines_ZeroDigitsDoNotDrawStrayPixels(t *testing.T) {
	// "0000" should produce only the vertical bar. The top-left pixel
	// (0,0) is far from any glyph and must remain untouched — the
	// pre-refactor code drew a stray pixel there for missing digit
	// entries, which this test guards against regressing.
	img := drawCistLines(false, "0000")
	if img.RGBAAt(0, 0).A != 0 {
		t.Error("stray pixel at (0,0) — zero digits should not draw")
	}
}
