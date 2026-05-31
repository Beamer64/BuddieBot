package slash

import (
	"strings"
	"testing"

	"github.com/Beamer64/BuddieBot/pkg/commands/prefix"
	"github.com/Beamer64/BuddieBot/pkg/helper"
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
