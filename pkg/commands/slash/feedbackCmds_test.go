package slash

import (
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

// TestFeedbackChannelFor exercises the routing table. The function takes
// channel IDs directly (not the full config) so testing is a plain pure-
// function exercise — no Configs construction needed.
func TestFeedbackChannelFor(t *testing.T) {
	const suggest, bug = "SUGGEST-CHAN", "BUG-CHAN"

	cases := []struct {
		category string
		want     string
	}{
		{"feature", suggest},
		{"bug", bug},
		{"other", suggest},
		// Unknown category — Discord's Choices list normally prevents this,
		// but the default branch must still route somewhere sensible.
		{"unrecognized", suggest},
	}
	for _, c := range cases {
		if got := feedbackChannelFor(c.category, suggest, bug); got != c.want {
			t.Errorf("feedbackChannelFor(%q) = %q, want %q", c.category, got, c.want)
		}
	}
}

// TestFeedbackChannelFor_EmptyConfig confirms that when a channel ID isn't
// configured (empty string in YAML), the route returns "" so the handler
// can surface a clear user-facing error rather than silently posting nowhere.
func TestFeedbackChannelFor_EmptyConfig(t *testing.T) {
	if got := feedbackChannelFor("bug", "SUGGEST", ""); got != "" {
		t.Errorf("expected empty when bug channel not configured, got %q", got)
	}
	if got := feedbackChannelFor("feature", "", "BUG"); got != "" {
		t.Errorf("expected empty when suggestion channel not configured, got %q", got)
	}
}

func TestFeedbackTitleAndColor(t *testing.T) {
	cases := []struct {
		category string
		wantHas  string // substring expected in title
		wantHex  int    // expected color
	}{
		{"feature", "Feature", 0x4ADE80},
		{"bug", "Bug", 0xEF4444},
		{"other", "Other", 0x6B7280},
		{"unrecognized", "Feedback", 0x6B7280},
	}
	for _, c := range cases {
		title, color := feedbackTitleAndColor(c.category)
		if !strings.Contains(title, c.wantHas) {
			t.Errorf("title for %q = %q, want substring %q", c.category, title, c.wantHas)
		}
		if color != c.wantHex {
			t.Errorf("color for %q = 0x%X, want 0x%X", c.category, color, c.wantHex)
		}
	}
}

func TestCategoryLabel(t *testing.T) {
	cases := []struct{ category, want string }{
		{"feature", "feature suggestion"},
		{"bug", "bug report"},
		{"other", "feedback"},
		{"anything-else", "feedback"},
	}
	for _, c := range cases {
		if got := categoryLabel(c.category); got != c.want {
			t.Errorf("categoryLabel(%q) = %q, want %q", c.category, got, c.want)
		}
	}
}

// TestFeedbackEmbed covers the embed shape and the guild/DM fallback. We
// don't pin the wording of every field — that's the kind of test that breaks
// on harmless copy edits — but we DO pin the structural invariants: author
// present, body shows verbatim, user ID appears both as mention and raw, and
// nil originGuild renders the "(direct message)" sentinel.
func TestFeedbackEmbed(t *testing.T) {
	submitter := &discordgo.User{ID: "12345", Username: "TestUser", Avatar: "abc"}
	guild := &discordgo.Guild{ID: "g1", Name: "Test Guild"}
	body := "The /image filter blur command throws an error when run against an animated avatar."

	embed := feedbackEmbed("bug", body, submitter, guild)

	if embed.Author == nil || embed.Author.Name != "TestUser" {
		t.Errorf("expected author name TestUser, got %+v", embed.Author)
	}
	if embed.Description != body {
		t.Errorf("expected description to equal body verbatim, got %q", embed.Description)
	}

	// "From" field must contain both <@id> mention and raw id (for log lookups).
	fields := byFieldName(embed.Fields)
	if from, ok := fields["From"]; !ok {
		t.Error("missing From field")
	} else {
		if !strings.Contains(from, "<@12345>") {
			t.Errorf("From field missing mention: %q", from)
		}
		if !strings.Contains(from, "`12345`") {
			t.Errorf("From field missing raw ID: %q", from)
		}
	}
	if g, ok := fields["Guild"]; !ok {
		t.Error("missing Guild field")
	} else if !strings.Contains(g, "Test Guild") || !strings.Contains(g, "g1") {
		t.Errorf("Guild field missing name/ID: %q", g)
	}

	// Bug should pick the red color + bug-themed title.
	if !strings.Contains(embed.Title, "Bug") {
		t.Errorf("bug title should mention Bug, got %q", embed.Title)
	}
	if embed.Color != 0xEF4444 {
		t.Errorf("bug color = 0x%X, want 0xEF4444", embed.Color)
	}

	// Timestamp present and parseable.
	if _, err := time.Parse(time.RFC3339, embed.Timestamp); err != nil {
		t.Errorf("timestamp not RFC3339: %q (err=%v)", embed.Timestamp, err)
	}
}

// TestFeedbackEmbedDMFallback verifies the (direct message) marker shows
// when originGuild is nil — important context for the maintainer reading
// the destination channel.
func TestFeedbackEmbedDMFallback(t *testing.T) {
	submitter := &discordgo.User{ID: "1", Username: "U"}
	embed := feedbackEmbed("feature", "A long enough body to pass validation no problem.", submitter, nil)

	fields := byFieldName(embed.Fields)
	if g := fields["Guild"]; !strings.Contains(g, "direct message") {
		t.Errorf("Guild field should mark DM context, got %q", g)
	}
}

// TestFormatRetryAfter checks the unit-rounding heuristic: short waits show
// seconds, long waits round to minutes (to avoid "59m 33s" noise).
func TestFormatRetryAfter(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{45 * time.Second, "45s"},
		{90 * time.Second, "1m30s"},
		{91 * time.Second, "2m0s"},
		{59 * time.Minute, "59m0s"},
	}
	for _, c := range cases {
		if got := formatRetryAfter(c.d); got != c.want {
			t.Errorf("formatRetryAfter(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

// byFieldName indexes embed fields by their Name for assertion convenience.
func byFieldName(fs []*discordgo.MessageEmbedField) map[string]string {
	out := map[string]string{}
	for _, f := range fs {
		out[f.Name] = f.Value
	}
	return out
}
