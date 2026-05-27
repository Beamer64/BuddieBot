package slash

import (
	"strings"
	"testing"

	"github.com/Beamer64/BuddieBot/pkg/helper"
)

func TestTuuckPageTopLevel(t *testing.T) {
	embed, components, ok := tuuckPage("", 0)
	if !ok {
		t.Fatal("top-level scope should be ok")
	}

	// Page 0 shows up to perPage top-level commands, one field each.
	want := tuuckFieldsPerPage
	if len(Commands) < want {
		want = len(Commands)
	}
	if len(embed.Fields) != want {
		t.Errorf("page 0 shows %d fields, want %d", len(embed.Fields), want)
	}

	// Buttons appear only when there's more than one page.
	multiPage := len(Commands) > tuuckFieldsPerPage
	if multiPage && len(components) == 0 {
		t.Error("expected pagination buttons for a multi-page list")
	}
	if !multiPage && len(components) != 0 {
		t.Error("did not expect buttons for a single-page list")
	}
}

func TestTuuckPageCommandScope(t *testing.T) {
	// A real command with subcommands renders its entries as fields.
	embed, _, ok := tuuckPage("audio", 0)
	if !ok {
		t.Fatal("audio scope should be ok")
	}
	if len(embed.Fields) == 0 {
		t.Error("expected audio subcommands as fields")
	}

	// A type-choice command lists its choices.
	if embed, _, ok := tuuckPage("daily", 0); !ok || len(embed.Fields) == 0 {
		t.Errorf("expected daily choices as fields (ok=%v, fields=%d)", ok, len(embed.Fields))
	}

	// Unknown command → not ok.
	if _, _, ok := tuuckPage("definitely-not-a-command", 0); ok {
		t.Error("unknown command scope should return ok=false")
	}
}

// A command-scoped page surfaces its curated example in the embed description,
// so the example shows on every page rather than as a paginated entry.
func TestTuuckPageCommandScopeExample(t *testing.T) {
	embed, _, ok := tuuckPage("audio", 0)
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
