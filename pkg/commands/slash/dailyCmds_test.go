package slash

import "testing"

// TestDailyAdvice checks the AdviceSlip API (adviceslip.com).
// No key required.
func TestDailyAdvice(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	defer rateLimit("adviceslip")()

	embed, err := getDailyAdviceEmbed(cfg)
	if err != nil {
		t.Fatalf("getDailyAdviceEmbed: %v", err)
	}
	if embed == nil || embed.Description == "" {
		t.Fatal("empty advice embed — AdviceSlip API likely changed shape")
	}
}

// TestDailyHoroscope checks the horoscope.com page scrape.
// No key required. Scraping is more fragile than a real API — if this
// breaks, the DOM selector in scrapeHoroscope probably needs updating.
func TestDailyHoroscope(t *testing.T) {
	requireINTEGRATION(t)
	defer rateLimit("horoscope-scrape")()

	horoscope, err := scrapeHoroscope("1") // aries
	if err != nil {
		t.Fatalf("scrapeHoroscope: %v", err)
	}
	if horoscope == "" {
		t.Fatal("empty horoscope — horoscope.com DOM selector likely needs an update")
	}
}
