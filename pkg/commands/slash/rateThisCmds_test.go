package slash

import (
	"strings"
	"testing"
)

func TestFormatRating_StandardRatings(t *testing.T) {
	const user = "<@!12345>"
	const value = 42

	for name, want := range standardRatings {
		t.Run(name, func(t *testing.T) {
			gotTitle, gotDesc := formatRating(name, user, value)
			if gotTitle != want.Title {
				t.Errorf("title mismatch for %q: got %q, want %q", name, gotTitle, want.Title)
			}
			expectedDesc := user + "'s " + want.ScoreLabel + " Score: 42/100"
			if gotDesc != expectedDesc {
				t.Errorf("desc mismatch for %q: got %q, want %q", name, gotDesc, expectedDesc)
			}
		})
	}
}

func TestFormatRating_Schmeat(t *testing.T) {
	gotTitle, gotDesc := formatRating("schmeat", "<@!12345>", 5)
	if gotTitle != "Schmeat Size" {
		t.Errorf("got title %q, want \"Schmeat Size\"", gotTitle)
	}
	if !strings.Contains(gotDesc, "Thang Thangin'") {
		t.Errorf("desc missing custom phrase, got %q", gotDesc)
	}
	// Size 5 → "C=====8" (5 equals between C and 8).
	if !strings.Contains(gotDesc, "C=====8") {
		t.Errorf("desc missing expected ASCII strip for size 5, got %q", gotDesc)
	}
}

func TestFormatRating_Unknown(t *testing.T) {
	gotTitle, gotDesc := formatRating("not-a-real-rating", "user", 1)
	if gotTitle != "" || gotDesc != "" {
		t.Errorf("expected empty strings for unknown rating, got (%q, %q)", gotTitle, gotDesc)
	}
}
