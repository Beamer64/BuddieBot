package slash

import (
	"strings"
	"testing"
)

// TestPickSteam checks the Steam GetAppList API (api.steampowered.com).
// The endpoint is public — the STEAMKEY placeholder in the URL is unused
// for this particular endpoint, so no key gate.
func TestPickSteam(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	defer rateLimit("steam")()

	gameURL, err := getSteamGame(cfg)
	if err != nil {
		t.Fatalf("getSteamGame: %v", err)
	}
	if !strings.HasPrefix(gameURL, "https://store.steampowered.com/app/") {
		t.Fatalf("unexpected steam URL %q — API likely changed shape", gameURL)
	}
}

// TestPickAlbum checks the AlbumRecommender API. No key required.
func TestPickAlbum(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	defer rateLimit("album-picker")()

	albums, err := callAlbumPickerAPI(cfg, []string{"rock"}, "")
	if err != nil {
		t.Fatalf("callAlbumPickerAPI: %v", err)
	}
	// The slash handler indexes albums[0..4] — anything fewer than that
	// would crash the live command, so use 5 as the threshold.
	if len(albums) < 5 {
		t.Fatalf("expected ≥5 albums, got %d — API likely changed shape", len(albums))
	}
	if albums[0].AlbumName == "" {
		t.Fatal("empty album name — API likely changed shape")
	}
}
