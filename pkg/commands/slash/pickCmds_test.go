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
