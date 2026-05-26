// Package voice_chat plays audio in Discord voice channels via Lavalink.
// Lavalink owns the voice WebSocket (DAVE/E2EE); the bot just forwards
// voice state events from discordgo and tells Lavalink what to play.
package voice_chat

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

var (
	ErrNotInVoice   = errors.New("user is not in a voice channel")
	ErrNoTrackFound = errors.New("no playable track at that URL")
	// ErrAlreadyPlaying is returned by ResumeQueue when a track is already playing.
	ErrAlreadyPlaying = errors.New("already playing in this guild")
	// ErrVoiceTimeout means Discord didn't deliver VOICE_STATE / VOICE_SERVER —
	// retrying the join usually fixes it.
	ErrVoiceTimeout   = errors.New("voice connection didn't establish")
	ErrQueueFull      = errors.New("queue is full")
	ErrNothingPlaying = errors.New("nothing is playing")
	// ErrTrackFailed is distinct from ErrVoiceTimeout because retrying is
	// pointless — the track itself is the problem.
	ErrTrackFailed = errors.New("track failed to play")
)

// FriendlyPlayError maps a Player error to a short user-facing message.
// ErrTrackFailed stays brief because OnTrackException already announced the detail.
func FriendlyPlayError(err error) string {
	switch {
	case errors.Is(err, ErrNotInVoice):
		return "Join a voice channel first."
	case errors.Is(err, ErrNoTrackFound):
		return "Couldn't resolve audio (invalid URL or unavailable video)."
	case errors.Is(err, ErrQueueFull):
		return "Queue is full (100 tracks max)."
	case errors.Is(err, ErrVoiceTimeout):
		return "Voice connection didn't establish — try again."
	case errors.Is(err, ErrTrackFailed):
		return "Couldn't play that track."
	default:
		return "Failed to start playback."
	}
}

// IsUserFacingError gates error-channel logging — true for routine user mistakes,
// false for genuine bugs.
func IsUserFacingError(err error) bool {
	return errors.Is(err, ErrNotInVoice) ||
		errors.Is(err, ErrNoTrackFound) ||
		errors.Is(err, ErrQueueFull) ||
		errors.Is(err, ErrVoiceTimeout) ||
		errors.Is(err, ErrTrackFailed)
}

// FormatPlayResult renders a PlayResult as a user-facing message.
// resumeCmd is the suggested resume command when r.WhileStopped is set.
func FormatPlayResult(r PlayResult, resumeCmd string) string {
	var b strings.Builder
	if r.Playlist != nil {
		name := r.Playlist.Name
		if name == "" {
			name = "playlist"
		}
		if r.Queued {
			fmt.Fprintf(&b, "Added %d tracks from %s to the queue (starting at position %d)",
				r.Playlist.QueuedTracks, name, r.Position)
		} else {
			fmt.Fprintf(&b, "Now playing: %s", r.Title)
			if r.Playlist.QueuedTracks > 0 {
				fmt.Fprintf(&b, "\nQueued %d more from %s", r.Playlist.QueuedTracks, name)
			}
		}
		// In the fresh-start case the first track is playing (not queued),
		// so subtract one when computing missed.
		played := 0
		if !r.Queued {
			played = 1
		}
		if missed := r.Playlist.TotalTracks - r.Playlist.QueuedTracks - played; missed > 0 {
			fmt.Fprintf(&b, " (queue full — %d not added)", missed)
		}
	} else if r.Queued {
		fmt.Fprintf(&b, "Added to queue: %s (position %d)", r.Title, r.Position)
	} else {
		fmt.Fprintf(&b, "Now playing: %s", r.Title)
	}
	if r.WhileStopped {
		fmt.Fprintf(&b, ". Use %s to start playback.", resumeCmd)
	}
	return b.String()
}

const (
	// voiceConnectTimeout: per-attempt wait for Lavalink Connected before retrying.
	voiceConnectTimeout = 7 * time.Second
	maxPlayAttempts     = 2
	maxQueueSize        = 100
)

// PlayResult describes what Play did with the URL. Title/Queued/Position
// describe the first track; Playlist (non-nil for playlist URLs) carries batch totals.
type PlayResult struct {
	Title    string
	Queued   bool
	Position int // 1-indexed queue position when Queued; 0 otherwise
	// WhileStopped: bot was disconnected with a saved track; the queued URL won't
	// play until the user resumes.
	WhileStopped bool
	Playlist     *PlaylistInfo
}

// PlaylistInfo describes a playlist that Play loaded and enqueued.
type PlaylistInfo struct {
	Name        string // may be empty if Lavalink didn't supply one
	TotalTracks int
	// QueuedTracks excludes the first track in the fresh-start path (it's playing,
	// not queued); less than TotalTracks if the queue cap was hit mid-batch.
	QueuedTracks int
}

// Player owns the disgolink client and per-guild queue / announce-channel state.
type Player struct {
	link    disgolink.Client
	session *discordgo.Session

	mu               sync.Mutex
	queues           map[snowflake.ID][]lavalink.Track
	announceChannels map[snowflake.ID]string
	playSignals      map[snowflake.ID]chan error
	// pausedTracks: Stop snapshots the active track here. ResumeQueue replays it
	// from position 0 after rejoining. Distinct from a Lavalink-side pause —
	// the bot fully disconnects, so disgolink's player is destroyed.
	pausedTracks map[snowflake.ID]lavalink.Track
}

// New constructs a Player. Register OnTrackEnd and OnTrackException as
// disgolink listeners to wire up auto-advance and failure announcements.
func New(link disgolink.Client, session *discordgo.Session) *Player {
	return &Player{
		link:             link,
		session:          session,
		queues:           map[snowflake.ID][]lavalink.Track{},
		announceChannels: map[snowflake.ID]string{},
		playSignals:      map[snowflake.ID]chan error{},
		pausedTracks:     map[snowflake.ID]lavalink.Track{},
	}
}

// Play loads url and either starts playback or appends to the queue.
// channelID is remembered as the destination for auto-advance "Now playing"
// messages; pass empty to leave it untouched. Voice-connect timeouts retry
// transparently via cycleVoice.
func (p *Player) Play(ctx context.Context, guildID, channelID, userID, url string) (PlayResult, error) {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return PlayResult{}, fmt.Errorf("parse guild id: %w", err)
	}

	node := p.link.BestNode()
	if node == nil {
		return PlayResult{}, errors.New("no lavalink nodes available")
	}

	load, err := loadTracks(ctx, node, url)
	if err != nil {
		return PlayResult{}, err
	}
	firstTrack := load.Tracks[0]
	totalTracks := len(load.Tracks)

	// "In session" includes the stopped state (saved track waiting) — both
	// branches queue rather than starting fresh.
	p.mu.Lock()
	_, isStopped := p.pausedTracks[gID]
	p.mu.Unlock()
	existing := p.link.ExistingPlayer(gID)
	inSession := isStopped || (existing != nil && existing.Track() != nil)

	if inSession {
		p.mu.Lock()
		// Reject if even the first track wouldn't fit — without it we have no
		// usable Position to return.
		if len(p.queues[gID]) >= maxQueueSize {
			p.mu.Unlock()
			return PlayResult{}, ErrQueueFull
		}
		// Playlist tracks past the cap are silently dropped; the count surfaces
		// via PlaylistInfo.QueuedTracks < TotalTracks.
		firstPos := len(p.queues[gID]) + 1
		queuedCount := 0
		for _, t := range load.Tracks {
			if len(p.queues[gID]) >= maxQueueSize {
				break
			}
			p.queues[gID] = append(p.queues[gID], t)
			queuedCount++
		}
		if channelID != "" {
			p.announceChannels[gID] = channelID
		}
		p.mu.Unlock()

		result := PlayResult{
			Title:        firstTrack.Info.Title,
			Queued:       true,
			Position:     firstPos,
			WhileStopped: isStopped,
		}
		if load.IsPlaylist {
			result.Playlist = &PlaylistInfo{
				Name:         load.PlaylistName,
				TotalTracks:  totalTracks,
				QueuedTracks: queuedCount,
			}
		}
		return result, nil
	}

	// Fresh session — caller must be in a voice channel.
	voiceState, err := p.session.State.VoiceState(guildID, userID)
	if err != nil || voiceState == nil || voiceState.ChannelID == "" {
		return PlayResult{}, ErrNotInVoice
	}

	// Set the announce channel up-front so OnTrackException can find it even
	// if the track fails before voice connects.
	if channelID != "" {
		p.mu.Lock()
		p.announceChannels[gID] = channelID
		p.mu.Unlock()
	}

	if err := p.startTrack(ctx, gID, guildID, voiceState.ChannelID, firstTrack); err != nil {
		return PlayResult{}, err
	}

	result := PlayResult{Title: firstTrack.Info.Title}

	// Queue the remaining playlist tracks behind the one that just started.
	if load.IsPlaylist && totalTracks > 1 {
		queuedCount := 0
		p.mu.Lock()
		for _, t := range load.Tracks[1:] {
			if len(p.queues[gID]) >= maxQueueSize {
				break
			}
			p.queues[gID] = append(p.queues[gID], t)
			queuedCount++
		}
		p.mu.Unlock()

		result.Playlist = &PlaylistInfo{
			Name:         load.PlaylistName,
			TotalTracks:  totalTracks,
			QueuedTracks: queuedCount,
		}
	}
	return result, nil
}

// startTrack runs the voice-join + WithTrack + connect-wait retry dance
// for a pre-resolved track. Caller handles the user-in-voice check and
// announce-channel setup. Voice-timeout failures retry (Discord dropped
// events); track-broken / context-cancelled failures bail immediately.
func (p *Player) startTrack(ctx context.Context, gID snowflake.ID, guildID, voiceChannelID string, track lavalink.Track) error {
	var lastErr error
	for attempt := 1; attempt <= maxPlayAttempts; attempt++ {
		if attempt > 1 {
			log.Printf("voice_chat: voice didn't connect on attempt %d for guild %s, retrying with fresh join", attempt-1, guildID)
			if err := p.cycleVoice(ctx, gID, guildID); err != nil {
				return err
			}
		}

		dgPlayer := p.link.Player(gID)

		if err := p.session.ChannelVoiceJoinManual(guildID, voiceChannelID, false, true); err != nil {
			lastErr = fmt.Errorf("voice channel join: %w", err)
			continue
		}

		// Lets OnTrackException short-circuit the connect wait when Lavalink
		// rejects the track outright.
		signal := make(chan error, 1)
		p.mu.Lock()
		p.playSignals[gID] = signal
		p.mu.Unlock()

		if err := dgPlayer.Update(ctx, lavalink.WithTrack(track)); err != nil {
			p.clearPlaySignal(gID)
			_ = p.session.ChannelVoiceJoinManual(guildID, "", false, true)
			return fmt.Errorf("start playback: %w", err)
		}

		err := waitForPlayback(ctx, dgPlayer, signal, voiceConnectTimeout)
		p.clearPlaySignal(gID)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, ErrTrackFailed):
			// Retrying voice won't help; OnTrackException already announced it.
			_ = p.session.ChannelVoiceJoinManual(guildID, "", false, true)
			return err
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return err
		}
		lastErr = ErrVoiceTimeout
	}

	// Final cleanup so the next call sees a clean slate.
	if dgPlayer := p.link.ExistingPlayer(gID); dgPlayer != nil {
		clearCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = dgPlayer.Update(clearCtx, lavalink.WithNullTrack())
		cancel()
	}
	_ = p.session.ChannelVoiceJoinManual(guildID, "", false, true)
	return lastErr
}

// cycleVoice does a leave + wait-for-destroy before the next join attempt.
// The leave is what makes Discord emit fresh VOICE_STATE / VOICE_SERVER —
// otherwise rejoining the same channel ID can fire no events at all.
func (p *Player) cycleVoice(ctx context.Context, gID snowflake.ID, guildID string) error {
	if dgPlayer := p.link.ExistingPlayer(gID); dgPlayer != nil {
		clearCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = dgPlayer.Update(clearCtx, lavalink.WithNullTrack())
		cancel()
	}
	_ = p.session.ChannelVoiceJoinManual(guildID, "", false, true)

	// Wait for disgolink to destroy the player — otherwise a late-arriving
	// leave event can destroy the next attempt's player mid-setup.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if p.link.ExistingPlayer(gID) == nil {
			break
		}
		select {
		case <-time.After(100 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Brief pause so we're not hammering Discord while it's misbehaving.
	select {
	case <-time.After(500 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// waitForPlayback returns:
//   - nil when state.Connected becomes true (track is playing)
//   - ErrTrackFailed when signal receives one (caller should skip the retry)
//   - ErrVoiceTimeout when neither fires before deadline (caller should retry)
func waitForPlayback(ctx context.Context, player disgolink.Player, signal <-chan error, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if player.State().Connected {
			return nil
		}
		select {
		case err := <-signal:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return ErrVoiceTimeout
			}
		}
	}
}

func (p *Player) clearPlaySignal(gID snowflake.ID) {
	p.mu.Lock()
	delete(p.playSignals, gID)
	p.mu.Unlock()
}

// QueueSnapshot is a read-only view of a guild's playback state. Current and
// Paused are mutually exclusive; Upcoming is copied, safe to read without a lock.
type QueueSnapshot struct {
	Current  *lavalink.Track
	Paused   *lavalink.Track
	Upcoming []lavalink.Track
}

func (p *Player) Queue(guildID string) (QueueSnapshot, error) {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return QueueSnapshot{}, fmt.Errorf("parse guild id: %w", err)
	}

	var snap QueueSnapshot
	if existing := p.link.ExistingPlayer(gID); existing != nil {
		snap.Current = existing.Track()
	}

	p.mu.Lock()
	if t, ok := p.pausedTracks[gID]; ok {
		// Copy so callers don't alias the map's value storage.
		cp := t
		snap.Paused = &cp
	}
	if q := p.queues[gID]; len(q) > 0 {
		snap.Upcoming = make([]lavalink.Track, len(q))
		copy(snap.Upcoming, q)
	}
	p.mu.Unlock()

	return snap, nil
}

// Skip advances to the next queued track. If the queue is empty, playback
// stops and the bot leaves voice (caller detects via next == nil).
func (p *Player) Skip(ctx context.Context, guildID string) (skipped, next *lavalink.Track, err error) {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return nil, nil, fmt.Errorf("parse guild id: %w", err)
	}

	dgPlayer := p.link.ExistingPlayer(gID)
	if dgPlayer == nil {
		return nil, nil, ErrNothingPlaying
	}

	skipped = dgPlayer.Track()
	if skipped == nil {
		return nil, nil, ErrNothingPlaying
	}

	p.mu.Lock()
	if len(p.queues[gID]) > 0 {
		n := p.queues[gID][0]
		p.queues[gID] = p.queues[gID][1:]
		next = &n
	}
	p.mu.Unlock()

	if next != nil {
		// Emits TrackEndEvent{Reason=replaced} for the old track; OnTrackEnd
		// ignores that (MayStartNext is false), so the queue isn't double-advanced.
		if err := dgPlayer.Update(ctx, lavalink.WithTrack(*next)); err != nil {
			return skipped, nil, fmt.Errorf("play next: %w", err)
		}
		return skipped, next, nil
	}

	// Queue empty — stop playback and leave voice, matching the natural
	// "queue-empty disconnect" behavior.
	if err := dgPlayer.Update(ctx, lavalink.WithNullTrack()); err != nil {
		return skipped, nil, fmt.Errorf("stop track: %w", err)
	}
	if err := p.session.ChannelVoiceJoinManual(guildID, "", false, true); err != nil {
		return skipped, nil, fmt.Errorf("leave voice: %w", err)
	}
	p.mu.Lock()
	delete(p.announceChannels, gID)
	p.mu.Unlock()
	return skipped, nil, nil
}

// ClearQueue wipes the upcoming queue and any saved/stopped track — a full
// session reset. Does NOT stop a currently-playing track. Returns the
// total number of tracks removed.
func (p *Player) ClearQueue(guildID string) (int, error) {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return 0, fmt.Errorf("parse guild id: %w", err)
	}

	p.mu.Lock()
	count := len(p.queues[gID])
	if _, hasSaved := p.pausedTracks[gID]; hasSaved {
		count++
		delete(p.pausedTracks, gID)
	}
	delete(p.queues, gID)
	p.mu.Unlock()

	return count, nil
}

// Stop snapshots the current track into pausedTracks and disconnects. The
// upcoming queue is left intact — ResumeQueue replays the saved track from
// position 0. Use ClearQueue for a full reset.
// changed=false means the bot was already stopped (idempotent).
// Returns ErrNothingPlaying when nothing is active AND no stopped session exists.
func (p *Player) Stop(ctx context.Context, guildID string) (changed bool, err error) {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return false, fmt.Errorf("parse guild id: %w", err)
	}

	p.mu.Lock()
	_, alreadyStopped := p.pausedTracks[gID]
	p.mu.Unlock()

	dgPlayer := p.link.ExistingPlayer(gID)
	if dgPlayer == nil || dgPlayer.Track() == nil {
		if alreadyStopped {
			return false, nil
		}
		return false, ErrNothingPlaying
	}

	activeTrack := *dgPlayer.Track()
	p.mu.Lock()
	p.pausedTracks[gID] = activeTrack
	p.mu.Unlock()

	// Best-effort; the disconnect below is what actually stops audio.
	if err := dgPlayer.Update(ctx, lavalink.WithNullTrack()); err != nil {
		log.Printf("voice_chat: clear track on stop: %v", err)
	}
	if err := p.session.ChannelVoiceJoinManual(guildID, "", false, true); err != nil {
		return true, fmt.Errorf("leave voice: %w", err)
	}
	return true, nil
}

// ResumeQueue rejoins userID's voice channel and restarts the Stop-saved
// track from position 0; OnTrackEnd then walks the upcoming queue normally.
// channelID becomes the new announce destination.
// Errors: ErrAlreadyPlaying, ErrNothingPlaying, ErrNotInVoice.
func (p *Player) ResumeQueue(ctx context.Context, guildID, channelID, userID string) (lavalink.Track, error) {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return lavalink.Track{}, fmt.Errorf("parse guild id: %w", err)
	}

	if existing := p.link.ExistingPlayer(gID); existing != nil && existing.Track() != nil {
		return lavalink.Track{}, ErrAlreadyPlaying
	}

	p.mu.Lock()
	track, hasSaved := p.pausedTracks[gID]
	p.mu.Unlock()
	if !hasSaved {
		return lavalink.Track{}, ErrNothingPlaying
	}

	voiceState, err := p.session.State.VoiceState(guildID, userID)
	if err != nil || voiceState == nil || voiceState.ChannelID == "" {
		return lavalink.Track{}, ErrNotInVoice
	}

	if channelID != "" {
		p.mu.Lock()
		p.announceChannels[gID] = channelID
		p.mu.Unlock()
	}

	if err := p.startTrack(ctx, gID, guildID, voiceState.ChannelID, track); err != nil {
		// pausedTracks left intact so the user can retry.
		return lavalink.Track{}, err
	}

	p.mu.Lock()
	delete(p.pausedTracks, gID)
	p.mu.Unlock()
	return track, nil
}

// OnTrackException (disgolink listener) signals any in-flight Play so it
// can bail out of the voice-connect wait, and posts a failure message to
// the announce channel.
func (p *Player) OnTrackException(player disgolink.Player, e lavalink.TrackExceptionEvent) {
	defer recoverCallback("OnTrackException", player.GuildID())
	gID := player.GuildID()

	p.mu.Lock()
	announceCh := p.announceChannels[gID]
	signal := p.playSignals[gID]
	p.mu.Unlock()

	if signal != nil {
		select {
		case signal <- ErrTrackFailed:
		default:
		}
	}

	if announceCh == "" {
		return
	}

	title := e.Track.Info.Title
	if title == "" {
		title = "track"
	}
	msg := fmt.Sprintf("⚠️ Couldn't play %s", title)
	if reason := briefExceptionReason(e.Exception.Message); reason != "" {
		msg += ": " + reason
	}
	// Lavalink's AllClientsFailedException can dump several KB; Discord rejects
	// messages >2000 chars.
	if len(msg) > 1900 {
		msg = msg[:1900] + "…"
	}
	if _, err := p.session.ChannelMessageSend(announceCh, msg); err != nil {
		log.Printf("voice_chat: announce track exception in guild %s: %v", gID, err)
	}
}

// briefExceptionReason returns the headline (first line, 200-char cap) of
// what may be a multi-line every-client-tried Lavalink exception.
func briefExceptionReason(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return ""
	}
	if newline := strings.IndexAny(msg, "\r\n"); newline > 0 {
		msg = msg[:newline]
	}
	const maxLen = 200
	if len(msg) > maxLen {
		msg = msg[:maxLen] + "…"
	}
	return msg
}

// OnTrackEnd (disgolink listener) advances the queue and posts "Now playing".
// On an empty queue it disconnects and forgets the announce channel.
func (p *Player) OnTrackEnd(player disgolink.Player, e lavalink.TrackEndEvent) {
	defer recoverCallback("OnTrackEnd", player.GuildID())
	if !e.Reason.MayStartNext() {
		// Stopped / replaced / cleanup — somebody else owns the next step.
		return
	}

	gID := player.GuildID()

	p.mu.Lock()
	queue := p.queues[gID]
	var next *lavalink.Track
	if len(queue) > 0 {
		next = &queue[0]
		p.queues[gID] = queue[1:]
	}
	announceCh := p.announceChannels[gID]
	p.mu.Unlock()

	if next != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := player.Update(ctx, lavalink.WithTrack(*next)); err == nil {
			if announceCh != "" {
				if _, sendErr := p.session.ChannelMessageSend(announceCh, "Now playing: "+next.Info.Title); sendErr != nil {
					log.Printf("voice_chat: announce next track in guild %s: %v", gID, sendErr)
				}
			}
			return
		} else {
			log.Printf("voice_chat: failed to play next queued track in guild %s: %v", gID, err)
			// fall through to disconnect
		}
	}

	if err := p.session.ChannelVoiceJoinManual(gID.String(), "", false, true); err != nil {
		log.Printf("voice_chat: disconnect after track end in guild %s: %v", gID, err)
	}
	p.mu.Lock()
	delete(p.announceChannels, gID)
	p.mu.Unlock()
}

// trackLoad captures what Lavalink resolved from a URL. For single-track and
// search-result URLs Tracks has one element; for playlists, the full list.
type trackLoad struct {
	Tracks       []lavalink.Track
	PlaylistName string
	IsPlaylist   bool
}

func loadTracks(ctx context.Context, node disgolink.Node, url string) (trackLoad, error) {
	var (
		load    trackLoad
		loadErr error
	)
	node.LoadTracksHandler(ctx, url, disgolink.NewResultHandler(
		func(t lavalink.Track) {
			load.Tracks = []lavalink.Track{t}
		},
		func(pl lavalink.Playlist) {
			load.Tracks = pl.Tracks
			load.PlaylistName = pl.Info.Name
			load.IsPlaylist = true
		},
		func(ts []lavalink.Track) {
			// Search result (e.g. ytsearch:foo) — take the first hit.
			if len(ts) > 0 {
				load.Tracks = []lavalink.Track{ts[0]}
			}
		},
		func() {
			loadErr = ErrNoTrackFound
		},
		func(err error) {
			loadErr = fmt.Errorf("load track: %w", err)
		},
	))
	if loadErr != nil {
		return trackLoad{}, loadErr
	}
	if len(load.Tracks) == 0 {
		return trackLoad{}, ErrNoTrackFound
	}
	return load, nil
}

// recoverCallback keeps a panicking listener from killing disgolink's goroutine.
func recoverCallback(name string, gID snowflake.ID) {
	if r := recover(); r != nil {
		log.Printf("voice_chat: panic in %s for guild %s: %v\n%s", name, gID, r, debug.Stack())
	}
}
