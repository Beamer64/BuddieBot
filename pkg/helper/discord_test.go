package helper

import (
	"strings"
	"testing"
)

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
