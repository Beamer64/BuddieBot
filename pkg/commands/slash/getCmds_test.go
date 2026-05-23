package slash

import "testing"

// TestGetFakePerson checks randomuser.me. No key required.
func TestGetFakePerson(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	defer rateLimit("randomuser")()

	person, err := callFakePersonAPI(cfg)
	if err != nil {
		t.Fatalf("callFakePersonAPI: %v", err)
	}
	if len(person.Results) == 0 || person.Results[0].Email == "" {
		t.Fatal("empty fake-person response — randomuser.me likely changed shape")
	}
}

// TestGetXkcd checks the xkcd.com random-comic scrape. No key required.
func TestGetXkcd(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	defer rateLimit("xkcd")()

	embed, err := getXkcdEmbed(cfg)
	if err != nil {
		t.Fatalf("getXkcdEmbed: %v", err)
	}
	if embed == nil || embed.Image == nil || embed.Image.URL == "https:" {
		// "https:" alone means the selector matched but found no src attribute.
		t.Fatal("empty xkcd image URL — xkcd.com DOM selector likely needs an update")
	}
}

// TestGetLandsat exercises the chromedp screenshot path. Gated behind
// INTEGRATION_HEAVY because it spawns a headless Chrome — slow and
// fragile compared to the HTTP-only tests.
func TestGetLandsat(t *testing.T) {
	requireINTEGRATIONHeavy(t)
	cfg := loadTestConfig(t)
	defer rateLimit("landsat-chromedp")()

	imgBytes, err := getLandsatImage(cfg, "TEST")
	if err != nil {
		t.Fatalf("getLandsatImage: %v", err)
	}
	if len(imgBytes) == 0 {
		t.Fatal("empty image bytes returned — chromedp screenshot may have silently failed")
	}
}
