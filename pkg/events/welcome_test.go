package events

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// channelsFixture builds a guild with one of each channel type relevant to the
// picker: a system channel, two regular text channels, a voice channel, and a
// category. Tests vary the canWrite stub to drive the picker's branches.
func channelsFixture() *discordgo.Guild {
	return &discordgo.Guild{
		ID:              "g",
		Name:            "Test Guild",
		SystemChannelID: "sys",
		Channels: []*discordgo.Channel{
			{ID: "sys", Type: discordgo.ChannelTypeGuildText},
			{ID: "general", Type: discordgo.ChannelTypeGuildText},
			{ID: "off-topic", Type: discordgo.ChannelTypeGuildText},
			{ID: "voice", Type: discordgo.ChannelTypeGuildVoice},
			{ID: "cat", Type: discordgo.ChannelTypeGuildCategory},
		},
	}
}

func TestPickWelcomeChannel_PrefersSystemChannel(t *testing.T) {
	g := channelsFixture()
	got := pickWelcomeChannel(g, func(string) bool { return true })
	if got != "sys" {
		t.Errorf("expected system channel preferred, got %q", got)
	}
}

func TestPickWelcomeChannel_FallsBackToFirstWritableText(t *testing.T) {
	g := channelsFixture()
	// System channel exists but the bot can't write there → fall through.
	canWrite := func(id string) bool { return id != "sys" }
	got := pickWelcomeChannel(g, canWrite)
	if got != "general" {
		t.Errorf("expected first writable text channel, got %q", got)
	}
}

func TestPickWelcomeChannel_SkipsNonTextChannels(t *testing.T) {
	g := channelsFixture()
	// Bot can ONLY write in voice + category — picker must still return ""
	// because those aren't text channels.
	canWrite := func(id string) bool { return id == "voice" || id == "cat" }
	got := pickWelcomeChannel(g, canWrite)
	if got != "" {
		t.Errorf("expected empty (voice/category aren't text), got %q", got)
	}
}

func TestPickWelcomeChannel_NoWritableChannelReturnsEmpty(t *testing.T) {
	g := channelsFixture()
	got := pickWelcomeChannel(g, func(string) bool { return false })
	if got != "" {
		t.Errorf("expected empty when nothing writable, got %q", got)
	}
}

func TestPickWelcomeChannel_NoSystemChannelSetUsesFirstText(t *testing.T) {
	g := channelsFixture()
	g.SystemChannelID = ""
	got := pickWelcomeChannel(g, func(string) bool { return true })
	if got != "sys" {
		// "sys" is still the first text channel by iteration order — the picker
		// shouldn't care that it WAS the system channel ID once that's blanked.
		t.Errorf("expected first text channel %q, got %q", "sys", got)
	}
}

func TestPickWelcomeChannel_NilGuildReturnsEmpty(t *testing.T) {
	if got := pickWelcomeChannel(nil, func(string) bool { return true }); got != "" {
		t.Errorf("expected empty for nil guild, got %q", got)
	}
}

// TestWelcomeEmbedShape spot-checks the embed without overfitting on phrasing —
// asserts the guild name lands in the description, all six core feature fields
// plus the two full-width ones are present, and the 3-rows-of-2 layout has
// the right number of inline spacers between pairs.
func TestWelcomeEmbedShape(t *testing.T) {
	embed := welcomeEmbed("Acme Inc.", &discordgo.User{ID: "1", Avatar: "abc"})

	if !strings.Contains(embed.Description, "Acme Inc.") {
		t.Errorf("guild name missing from description: %q", embed.Description)
	}

	wantTitles := []string{
		"🎨 Image Effects",
		"🎲 Games",
		"📅 Daily",
		"👤 Profiles & Ratings",
		"😂 Random",
		"⚙️ For the Admins",
		"🚀 Getting Started",
		"🔒 Privacy",
		"💬 Feedback",
	}
	have := map[string]bool{}
	for _, f := range embed.Fields {
		have[f.Name] = true
	}
	for _, w := range wantTitles {
		if !have[w] {
			t.Errorf("missing field %q", w)
		}
	}

	// Layout: 6 inline feature fields + 3 inline spacers + 3 full-width =
	// 12 total. Spacers force 2-per-row by occupying the 3rd slot.
	if len(embed.Fields) != 12 {
		t.Errorf("expected 12 fields (6 features + 3 spacers + 3 full-width), got %d", len(embed.Fields))
	}

	// The full-width fields must NOT be inline.
	for _, f := range embed.Fields {
		if (f.Name == "🚀 Getting Started" || f.Name == "🔒 Privacy" || f.Name == "💬 Feedback") && f.Inline {
			t.Errorf("field %q should be full-width (Inline=false)", f.Name)
		}
	}

	if embed.Thumbnail == nil || embed.Thumbnail.URL == "" {
		t.Error("expected a thumbnail URL on the welcome embed")
	}
	if embed.Footer == nil || embed.Footer.Text == "" {
		t.Error("expected a footer on the welcome embed")
	}
}

// TestWelcomeEmbedHandlesNilBotUser covers the defensive path: if session
// state is somehow missing the bot user, embed still builds (just without a
// thumbnail URL) rather than panicking.
func TestWelcomeEmbedHandlesNilBotUser(t *testing.T) {
	embed := welcomeEmbed("X", nil)
	if embed.Thumbnail == nil {
		t.Fatal("expected Thumbnail struct present even with nil bot user")
	}
	if embed.Thumbnail.URL != "" {
		t.Errorf("expected blank URL when bot user is nil, got %q", embed.Thumbnail.URL)
	}
}
