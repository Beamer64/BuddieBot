package slash

import (
	"strings"
	"testing"

	"github.com/Beamer64/BuddieBot/pkg/database"
	"github.com/bwmarrin/discordgo"
)

func TestFormatRecentRatingLine_Standard(t *testing.T) {
	got := formatRecentRatingLine(&database.UserRating{RatingName: "neckbeard", Value: 15})
	want := "**Neck Beard**: `15/100`"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatRecentRatingLine_Schmeat(t *testing.T) {
	got := formatRecentRatingLine(&database.UserRating{RatingName: "schmeat", Value: 5})
	if !strings.HasPrefix(got, "**Schmeat**: `C") || !strings.HasSuffix(got, "8`") {
		t.Errorf("got %q, want a Schmeat ASCII line", got)
	}
	if !strings.Contains(got, "C=====8") {
		t.Errorf("expected 5 '=' for size 5, got %q", got)
	}
}

func TestFormatRecentRatings_EmptyAndJoin(t *testing.T) {
	if got := formatRecentRatings(nil); got != "—" {
		t.Errorf("empty list should render as em-dash, got %q", got)
	}
	got := formatRecentRatings([]*database.UserRating{
		{RatingName: "nerd", Value: 45},
		{RatingName: "stinky", Value: 80},
	})
	if !strings.Contains(got, "**Nerd**: `45/100`") || !strings.Contains(got, "**Stinky**: `80/100`") {
		t.Errorf("expected both lines present, got %q", got)
	}
	if !strings.Contains(got, "\n") {
		t.Errorf("expected newline between lines, got %q", got)
	}
}

func TestFormatCommandStats(t *testing.T) {
	// Happy path: total + top command both shown, command rendered as "/path".
	got := formatCommandStats(42, "image filter blur", 15)
	if !strings.Contains(got, "**Total**: `42`") {
		t.Errorf("missing total line, got %q", got)
	}
	if !strings.Contains(got, "`/image filter blur`") {
		t.Errorf("missing slash-prefixed command, got %q", got)
	}
	if !strings.Contains(got, "(`15×`)") {
		t.Errorf("missing count parenthetical, got %q", got)
	}

	// Empty top → em-dash, total still rendered.
	got = formatCommandStats(0, "", 0)
	if !strings.Contains(got, "**Total**: `0`") {
		t.Errorf("missing zero-total line, got %q", got)
	}
	if !strings.Contains(got, "**Most used**: —") {
		t.Errorf("expected em-dash placeholder, got %q", got)
	}
}

func TestBotProfileEmbed(t *testing.T) {
	target := &discordgo.User{
		ID:       "999",
		Username: "HelpfulBot",
		Bot:      true,
		Avatar:   "abc123",
	}
	embed := botProfileEmbed(target)
	if embed.Title != "HelpfulBot's Profile" {
		t.Errorf("title = %q, want %q", embed.Title, "HelpfulBot's Profile")
	}
	if !strings.Contains(embed.Description, "Bots don't have") {
		t.Errorf("description = %q, want bot-stub phrasing", embed.Description)
	}
	if embed.Thumbnail == nil || embed.Thumbnail.URL == "" {
		t.Error("expected a thumbnail URL on the bot stub")
	}
	if len(embed.Fields) != 0 {
		t.Errorf("expected no fields on the bot stub, got %d", len(embed.Fields))
	}
}

func TestMemberSinceDisplay(t *testing.T) {
	// All three accepted formats should produce a Discord <t:…:D> stamp.
	for _, c := range []struct{ name, input string }{
		{"rfc3339", "2026-05-26T21:04:50Z"},
		{"sqlite-native", "2026-05-26 21:04:50"},
		{"date-only", "2026-05-26"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := memberSinceDisplay(c.input)
			if !strings.HasPrefix(got, "<t:") || !strings.HasSuffix(got, ":D>") {
				t.Errorf("expected Discord timestamp form, got %q", got)
			}
		})
	}

	// The two same-instant inputs must produce the identical timestamp — that
	// catches off-by-timezone parse bugs.
	if a, b := memberSinceDisplay("2026-05-26T21:04:50Z"), memberSinceDisplay("2026-05-26 21:04:50"); a != b {
		t.Errorf("same instant parsed differently: rfc3339=%q sqlite=%q", a, b)
	}

	// Garbage long enough to look date-shaped → trim to the date prefix instead
	// of dumping the full ISO into the embed.
	if got := memberSinceDisplay("2026-05-26T21:04:50.123456+05:30"); got == "2026-05-26T21:04:50.123456+05:30" {
		t.Errorf("expected trimmed prefix, got raw %q", got)
	}
	if got := memberSinceDisplay("oops"); got != "oops" {
		t.Errorf("short garbage should pass through, got %q", got)
	}
}

func TestParseProfilePageID(t *testing.T) {
	cases := []struct {
		id          string
		wantTarget  string
		wantInvoker string
		wantPage    int
	}{
		{"profile-page:111:222:0", "111", "222", 0},
		{"profile-page:111:222:3", "111", "222", 3},
		{"profile-page:malformed", "", "", 0},
	}
	for _, c := range cases {
		target, invoker, page := parseProfilePageID(c.id)
		if target != c.wantTarget || invoker != c.wantInvoker || page != c.wantPage {
			t.Errorf("parseProfilePageID(%q) = (%q, %q, %d), want (%q, %q, %d)",
				c.id, target, invoker, page, c.wantTarget, c.wantInvoker, c.wantPage)
		}
	}
}
