package voice_chat

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
)

func TestFriendlyPlayError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"not in voice", ErrNotInVoice, "Join a voice channel first."},
		{"no track found", ErrNoTrackFound, "Couldn't resolve audio (invalid URL or unavailable video)."},
		{"queue full", ErrQueueFull, "Queue is full (100 tracks max)."},
		{"voice timeout", ErrVoiceTimeout, "Voice connection didn't establish — try again."},
		{"track failed", ErrTrackFailed, "Couldn't play that track."},
		{"unknown error", errors.New("some random error"), "Failed to start playback."},
		// Wrapped errors must still hit the sentinel via errors.Is.
		{"wrapped sentinel", fmt.Errorf("context: %w", ErrNotInVoice), "Join a voice channel first."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FriendlyPlayError(tc.err); got != tc.want {
				t.Errorf("FriendlyPlayError(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}

func TestIsUserFacingError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"ErrNotInVoice", ErrNotInVoice, true},
		{"ErrNoTrackFound", ErrNoTrackFound, true},
		{"ErrQueueFull", ErrQueueFull, true},
		{"ErrVoiceTimeout", ErrVoiceTimeout, true},
		{"ErrTrackFailed", ErrTrackFailed, true},
		{"wrapped ErrTrackFailed", fmt.Errorf("wrap: %w", ErrTrackFailed), true},
		{"ErrAlreadyPlaying (not in the friendly set)", ErrAlreadyPlaying, false},
		{"ErrNothingPlaying (not in the friendly set)", ErrNothingPlaying, false},
		{"random error", errors.New("boom"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsUserFacingError(tc.err); got != tc.want {
				t.Errorf("IsUserFacingError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestBriefExceptionReason(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"whitespace only", "   \n\t  ", ""},
		{"single line", "track unavailable", "track unavailable"},
		{"trims surrounding whitespace", "  hello  ", "hello"},
		{"keeps only first line on LF", "first line\nsecond line", "first line"},
		{"keeps only first line on CRLF", "first line\r\nsecond line", "first line"},
		{"caps at 200 chars with ellipsis", strings.Repeat("a", 250), strings.Repeat("a", 200) + "…"},
		{"exactly-200 char line passes through", strings.Repeat("b", 200), strings.Repeat("b", 200)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := briefExceptionReason(tc.in); got != tc.want {
				t.Errorf("briefExceptionReason(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// trackWithTitle builds a minimal lavalink.Track that satisfies the
// fields FormatPlayResult reads (only Info.Title).
func trackWithTitle(title string) lavalink.Track {
	return lavalink.Track{Info: lavalink.TrackInfo{Title: title}}
}

func TestFormatPlayResult_SingleTrack(t *testing.T) {
	t.Run("now playing", func(t *testing.T) {
		got := FormatPlayResult(PlayResult{Title: "Song A"}, "$resume-queue")
		want := "Now playing: Song A"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("added to queue", func(t *testing.T) {
		got := FormatPlayResult(PlayResult{Title: "Song A", Queued: true, Position: 3}, "$resume-queue")
		want := "Added to queue: Song A (position 3)"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("while stopped hint", func(t *testing.T) {
		got := FormatPlayResult(PlayResult{Title: "Song A", Queued: true, Position: 1, WhileStopped: true}, "/audio resume-queue")
		want := "Added to queue: Song A (position 1). Use /audio resume-queue to start playback."
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestFormatPlayResult_Playlist(t *testing.T) {
	t.Run("fresh start — first plays, rest queued", func(t *testing.T) {
		r := PlayResult{
			Title: "Track 1",
			Playlist: &PlaylistInfo{
				Name:         "My Mix",
				TotalTracks:  10,
				QueuedTracks: 9,
			},
		}
		got := FormatPlayResult(r, "$resume-queue")
		// Expect two lines: now playing first track, then "queued N more".
		if !strings.HasPrefix(got, "Now playing: Track 1") {
			t.Errorf("expected to start with 'Now playing: Track 1', got %q", got)
		}
		if !strings.Contains(got, "Queued 9 more from My Mix") {
			t.Errorf("expected playlist-name detail, got %q", got)
		}
	})

	t.Run("in-session — all queued", func(t *testing.T) {
		r := PlayResult{
			Title:    "Track 1",
			Queued:   true,
			Position: 5,
			Playlist: &PlaylistInfo{
				Name:         "My Mix",
				TotalTracks:  10,
				QueuedTracks: 10,
			},
		}
		got := FormatPlayResult(r, "$resume-queue")
		want := "Added 10 tracks from My Mix to the queue (starting at position 5)"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("missed tracks shown when queue overflowed (fresh start)", func(t *testing.T) {
		r := PlayResult{
			Title: "Track 1",
			Playlist: &PlaylistInfo{
				Name:         "Big Mix",
				TotalTracks:  100,
				QueuedTracks: 50, // first plays + 50 queued = 51; 49 missed
			},
		}
		got := FormatPlayResult(r, "$resume-queue")
		if !strings.Contains(got, "queue full — 49 not added") {
			t.Errorf("expected overflow note, got %q", got)
		}
	})

	t.Run("missed tracks shown when queue overflowed (in-session)", func(t *testing.T) {
		r := PlayResult{
			Title:    "Track 1",
			Queued:   true,
			Position: 90,
			Playlist: &PlaylistInfo{
				Name:         "Big Mix",
				TotalTracks:  20,
				QueuedTracks: 11, // 9 missed
			},
		}
		got := FormatPlayResult(r, "$resume-queue")
		if !strings.Contains(got, "queue full — 9 not added") {
			t.Errorf("expected overflow note, got %q", got)
		}
	})

	t.Run("empty playlist name falls back to 'playlist'", func(t *testing.T) {
		r := PlayResult{
			Title: "Track 1",
			Playlist: &PlaylistInfo{
				Name:         "",
				TotalTracks:  3,
				QueuedTracks: 2,
			},
		}
		got := FormatPlayResult(r, "$resume-queue")
		if !strings.Contains(got, "from playlist") {
			t.Errorf("expected fallback name 'playlist', got %q", got)
		}
	})

	t.Run("while-stopped hint appended to playlist message", func(t *testing.T) {
		r := PlayResult{
			Title:        "Track 1",
			Queued:       true,
			Position:     1,
			WhileStopped: true,
			Playlist: &PlaylistInfo{
				Name:         "My Mix",
				TotalTracks:  5,
				QueuedTracks: 5,
			},
		}
		got := FormatPlayResult(r, "/audio resume-queue")
		if !strings.HasSuffix(got, "Use /audio resume-queue to start playback.") {
			t.Errorf("expected resume-queue hint suffix, got %q", got)
		}
	})
}

// Just a smoke check that PlaylistInfo's TrackInfo plumbing compiles —
// catches us if disgolink ever renames Track.Info.Title.
func TestPlayResult_PlaylistFieldShape(t *testing.T) {
	tr := trackWithTitle("hello")
	if tr.Info.Title != "hello" {
		t.Fatalf("trackWithTitle plumbing broken: got %q", tr.Info.Title)
	}
}
