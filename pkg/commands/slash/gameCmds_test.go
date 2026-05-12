package slash

import "testing"

// TestPlayCoinFlip exercises the only network-touching /play subcommand.
// Hits Tenor (g.tenor.com) via api.RequestGifURL. Requires TenorAPIkey.
func TestPlayCoinFlip(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	requireKey(t, cfg.Configs.Keys.TenorAPIkey, "TenorAPIkey")
	defer rateLimit("tenor")()

	embed, err := getCoinFlipEmbed(cfg)
	if err != nil {
		t.Fatalf("getCoinFlipEmbed: %v", err)
	}
	if embed == nil || embed.Image == nil || embed.Image.URL == "" {
		t.Fatal("empty coin-flip gif URL — Tenor API likely changed shape")
	}
}
