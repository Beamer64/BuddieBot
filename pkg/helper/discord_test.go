package helper

import (
	"strings"
	"testing"
)

// TestIsBotOwner locks in the owner-gate's contract:
//   - matching ID → true
//   - non-matching ID → false
//   - empty user ID → false (defensive — never grant owner on a missing ID)
//   - nil / empty owner slice → false (unpopulated config is safe-by-default)
func TestIsBotOwner(t *testing.T) {
	owners := []string{"100", "200", "300"}
	cases := []struct {
		userID string
		want   bool
	}{
		{"100", true},
		{"200", true},
		{"300", true},
		{"400", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsBotOwner(c.userID, owners); got != c.want {
			t.Errorf("IsBotOwner(%q, %v) = %v, want %v", c.userID, owners, got, c.want)
		}
	}

	// Empty / nil owner lists: nobody is owner. Defensive default for fresh
	// configs that haven't populated the list yet.
	if IsBotOwner("100", nil) {
		t.Error("expected false with nil owner list")
	}
	if IsBotOwner("100", []string{}) {
		t.Error("expected false with empty owner list")
	}
}

func TestTruncateForEmbed(t *testing.T) {
	t.Run("short string unchanged", func(t *testing.T) {
		in := "hello world"
		got := truncateForEmbed(in)
		if got != in {
			t.Errorf("got %q, want %q", got, in)
		}
	})

	t.Run("exact-length string unchanged", func(t *testing.T) {
		in := strings.Repeat("a", maxEmbedFieldValue)
		got := truncateForEmbed(in)
		if got != in {
			t.Errorf("exact-length input was modified: got len %d, want %d", len(got), len(in))
		}
	})

	t.Run("over-length string is truncated with suffix", func(t *testing.T) {
		in := strings.Repeat("a", maxEmbedFieldValue+50)
		got := truncateForEmbed(in)
		if len(got) > maxEmbedFieldValue {
			t.Errorf("output exceeds max length: got %d, want <= %d", len(got), maxEmbedFieldValue)
		}
		if !strings.HasSuffix(got, "...") {
			t.Errorf("expected truncation suffix, got %q", got[len(got)-10:])
		}
	})
}

func TestTailForEmbed(t *testing.T) {
	t.Run("short string unchanged", func(t *testing.T) {
		in := "hello world"
		got := tailForEmbed(in)
		if got != in {
			t.Errorf("got %q, want %q", got, in)
		}
	})

	t.Run("exact-length string unchanged", func(t *testing.T) {
		in := strings.Repeat("a", maxEmbedFieldValue)
		got := tailForEmbed(in)
		if got != in {
			t.Errorf("exact-length input was modified: got len %d, want %d", len(got), len(in))
		}
	})

	t.Run("over-length string is tail-truncated with prefix", func(t *testing.T) {
		in := strings.Repeat("a", maxEmbedFieldValue) + "TAIL"
		got := tailForEmbed(in)
		if len(got) > maxEmbedFieldValue {
			t.Errorf("output exceeds max length: got %d, want <= %d", len(got), maxEmbedFieldValue)
		}
		if !strings.HasPrefix(got, "...") {
			t.Errorf("expected truncation prefix, got %q", got[:10])
		}
		if !strings.HasSuffix(got, "TAIL") {
			t.Errorf("expected to keep tail of input, got %q", got[len(got)-10:])
		}
	})
}
