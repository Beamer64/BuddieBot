package slash

import (
	"strings"
	"testing"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/commands/prefix"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func TestTuuckPageTopLevel(t *testing.T) {
	embed, components, ok := tuuckPage("", "$", 0)
	if !ok {
		t.Fatal("top-level scope should be ok")
	}

	// Page 0 shows up to perPage entries — top-level slash commands plus the
	// visible prefix commands (PrefCmdsList minus NoShowCmds), one field each.
	visiblePrefix := 0
	for name := range prefix.PrefCmdsList {
		hidden := false
		for _, h := range prefix.NoShowCmds {
			if h == name {
				hidden = true
				break
			}
		}
		if !hidden {
			visiblePrefix++
		}
	}
	totalTop := len(Commands) + visiblePrefix
	want := tuuckFieldsPerPage
	if totalTop < want {
		want = totalTop
	}
	if len(embed.Fields) != want {
		t.Errorf("page 0 shows %d fields, want %d", len(embed.Fields), want)
	}

	// Buttons appear only when there's more than one page.
	multiPage := totalTop > tuuckFieldsPerPage
	if multiPage && len(components) == 0 {
		t.Error("expected pagination buttons for a multi-page list")
	}
	if !multiPage && len(components) != 0 {
		t.Error("did not expect buttons for a single-page list")
	}
}

// TestTuuckPageTopLevelHonoursPrefix confirms the guild's current prefix is
// stamped onto prefix-command entries (so a guild that changed it sees the
// right character in the help list, not the default $).
func TestTuuckPageTopLevelHonoursPrefix(t *testing.T) {
	embed, _, ok := tuuckPage("", "!", 0)
	if !ok {
		t.Fatal("top-level scope should be ok")
	}

	// Pick a known-visible prefix command and confirm it shows with the
	// passed-in prefix. "palindrome" is in PrefCmdsList and not in NoShowCmds.
	// Names are wrapped in backticks for monospace rendering in Discord.
	want := "`!palindrome`"
	found := false
	for _, f := range embed.Fields {
		if f.Name == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a field named %q on top-level page (custom prefix should stamp through)", want)
	}

	// And the inverse: with the default "$" prefix, the field name uses "$".
	embed, _, _ = tuuckPage("", "$", 0)
	want = "`$palindrome`"
	found = false
	for _, f := range embed.Fields {
		if f.Name == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a field named %q with default prefix", want)
	}
}

// TestTuuckListItemsHidesNoShowCmds locks in the privacy/visibility rule:
// anything in prefix.NoShowCmds must never appear in the top-level list,
// regardless of prefix.
func TestTuuckListItemsHidesNoShowCmds(t *testing.T) {
	_, items, ok := tuuckListItems("", "$")
	if !ok {
		t.Fatal("top-level scope should be ok")
	}
	for _, hidden := range prefix.NoShowCmds {
		needle := "`$" + hidden + "`"
		for _, it := range items {
			if it.name == needle {
				t.Errorf("hidden command %q leaked into top-level list", needle)
			}
		}
	}
}

// TestTuuckListItemsTopLevelMentionShape locks in the leaf-vs-non-leaf
// rendering: parent commands (with SubCommand/SubCommandGroup options)
// render as plain "/name" because Discord can't make a clickable pill out
// of those, and leaf commands fall through helper.CommandMention.
func TestTuuckListItemsTopLevelMentionShape(t *testing.T) {
	_, items, ok := tuuckListItems("", "$")
	if !ok {
		t.Fatal("top-level scope should be ok")
	}

	// Build a fast lookup keyed by either the leaf form "/name" (plain) or
	// the literal name (the mention call returns "/name" when no ID is
	// registered, which is what tests see).
	byName := map[string]string{}
	for _, it := range items {
		byName[it.name] = it.value
	}

	// /audio has subcommands — must appear as plain "/audio" wrapped in
	// monospace backticks for Discord rendering.
	if _, ok := byName["`/audio`"]; !ok {
		t.Errorf("expected top-level entry `/audio` (non-leaf, plain text)")
	}
	// /daily is leaf — helper.CommandMention with no registered ID yields
	// "/daily" (no backticks since it's a mention fallback, not a plain
	// non-leaf path). Either way it should be present in the lookup.
	if _, ok := byName["/daily"]; !ok {
		t.Errorf("expected top-level entry /daily (leaf, mention fallback)")
	}
}

// TestTuuckListItemsDrilledUsesMentionPaths confirms the drilled view emits
// the full subcommand path through helper.CommandMention so Discord renders
// each subcommand as a clickable pill. Without registered command IDs (as
// in tests), CommandMention falls back to "/<spec> <opt>", which is what
// the assertion checks for.
func TestTuuckListItemsDrilledUsesMentionPaths(t *testing.T) {
	_, items, ok := tuuckListItems("audio", "$")
	if !ok {
		t.Fatal("audio scope should be ok")
	}
	want := "/audio play"
	found := false
	for _, it := range items {
		if it.name == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected drilled-in entry %q (full subcommand mention path)", want)
	}
}

func TestTuuckPageCommandScope(t *testing.T) {
	// A real command with subcommands renders its entries as fields. The
	// guildPrefix arg isn't consulted in a command-scoped page, so any value
	// works here.
	embed, _, ok := tuuckPage("audio", "$", 0)
	if !ok {
		t.Fatal("audio scope should be ok")
	}
	if len(embed.Fields) == 0 {
		t.Error("expected audio subcommands as fields")
	}

	// A type-choice command lists its choices.
	if embed, _, ok := tuuckPage("daily", "$", 0); !ok || len(embed.Fields) == 0 {
		t.Errorf("expected daily choices as fields (ok=%v, fields=%d)", ok, len(embed.Fields))
	}

	// Unknown command → not ok.
	if _, _, ok := tuuckPage("definitely-not-a-command", "$", 0); ok {
		t.Error("unknown command scope should return ok=false")
	}
}

// A command-scoped page surfaces its curated example in the embed description,
// so the example shows on every page rather than as a paginated entry.
func TestTuuckPageCommandScopeExample(t *testing.T) {
	embed, _, ok := tuuckPage("audio", "$", 0)
	if !ok {
		t.Fatal("audio scope should be ok")
	}
	if ex := helper.CommandExamples["audio"]; ex != "" && !strings.Contains(embed.Description, ex) {
		t.Errorf("description %q missing example %q", embed.Description, ex)
	}
}

func TestSplitScopePage(t *testing.T) {
	cases := []struct {
		id        string
		wantScope string
		wantPage  int
	}{
		{"tuuck-page::0", "", 0},
		{"tuuck-page:image:2", "image", 2},
		{"tuuck-page:rate-this:1", "rate-this", 1},
	}
	for _, c := range cases {
		scope, page := splitScopePage(c.id)
		if scope != c.wantScope || page != c.wantPage {
			t.Errorf("splitScopePage(%q) = (%q, %d), want (%q, %d)", c.id, scope, page, c.wantScope, c.wantPage)
		}
	}
}

// /feedback tests ————————————————————————————————————————————————————————————

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
