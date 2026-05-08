package slash

import (
	"strings"
	"testing"
)

func TestGetRateTitleAndDesc_StandardRatings(t *testing.T) {
	const user = "<@!12345>"
	const score = "42"

	for name, want := range standardRatings {
		t.Run(name, func(t *testing.T) {
			gotTitle, gotDesc := getRateTitleAndDesc(name, user, score)
			if gotTitle != want.Title {
				t.Errorf("title mismatch for %q: got %q, want %q", name, gotTitle, want.Title)
			}
			expectedDesc := user + "'s " + want.ScoreLabel + " Score: " + score + "/100"
			if gotDesc != expectedDesc {
				t.Errorf("desc mismatch for %q: got %q, want %q", name, gotDesc, expectedDesc)
			}
		})
	}
}

func TestGetRateTitleAndDesc_Schmeat(t *testing.T) {
	gotTitle, gotDesc := getRateTitleAndDesc("schmeat", "<@!12345>", "ignored")
	if gotTitle != "Schmeat Size" {
		t.Errorf("got title %q, want \"Schmeat Size\"", gotTitle)
	}
	if !strings.Contains(gotDesc, "Thang Thangin'") {
		t.Errorf("desc missing custom phrase, got %q", gotDesc)
	}
}

func TestGetRateTitleAndDesc_Unknown(t *testing.T) {
	gotTitle, gotDesc := getRateTitleAndDesc("not-a-real-rating", "user", "1")
	if gotTitle != "" || gotDesc != "" {
		t.Errorf("expected empty strings for unknown rating, got (%q, %q)", gotTitle, gotDesc)
	}
}
