// Package voice_chat plays audio in Discord voice channels via a Lavalink
// service. The bot doesn't open its own voice WebSocket; Lavalink handles
// the voice connection (including DAVE/E2EE) while we just forward voice
// state events from discordgo and tell Lavalink which track to play.
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
	// ErrNotInVoice is returned when the requesting user isn't in a voice channel.
	ErrNotInVoice = errors.New("user is not in a voice channel")
	// ErrNoTrackFound is returned when the URL didn't resolve to a playable track.
	ErrNoTrackFound = errors.New("no playable track at that URL")
	// ErrAlreadyPlaying is returned when the guild already has an active track.
	// (No longer surfaced from Play — kept for backward compatibility with any
	// caller that wants to detect "something was playing"; today's Play queues
	// instead of erroring.)
	ErrAlreadyPlaying = errors.New("already playing in this guild")
	// ErrVoiceTimeout is returned when Discord doesn't deliver the voice
	// state/server events within the wait window — Lavalink ends up
	// without voice info and audio never starts. The fix is to retry.
	ErrVoiceTimeout = errors.New("voice connection didn't establish")
	// ErrQueueFull is returned when Play tries to enqueue but the guild's
	// queue is at maxQueueSize.
	ErrQueueFull = errors.New("queue is full")
	// ErrNothingPlaying is returned by Skip / Queue when nothing is playing.
	ErrNothingPlaying = errors.New("nothing is playing")
	// ErrTrackFailed is returned by Play when Lavalink reports a
	// TrackExceptionEvent during the connect window. Distinct from
	// ErrVoiceTimeout because retrying is pointless — the track itself
	// is the problem, not Discord's voice event delivery.
	ErrTrackFailed = errors.New("track failed to play")
)

// FriendlyPlayError maps a Player error to a short, user-facing message.
// ErrTrackFailed cases get a deliberately brief message here because
// detail has already been posted via Player.OnTrackException to the
// announce channel — no point repeating it. Returns a default
// "Failed to start playback." for any error this function doesn't
// recognize; callers should use IsUserFacingError to gate which errors
// they translate vs let bubble up as bugs.
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

// IsUserFacingError reports whether err has a meaningful translation
// via FriendlyPlayError. Lets callers suppress error-channel logs for
// routine user mistakes (wrong voice channel, bad URL, queue full)
// while still surfacing genuine bugs to the error log.
func IsUserFacingError(err error) bool {
	return errors.Is(err, ErrNotInVoice) ||
		errors.Is(err, ErrNoTrackFound) ||
		errors.Is(err, ErrQueueFull) ||
		errors.Is(err, ErrVoiceTimeout) ||
		errors.Is(err, ErrTrackFailed)
}

// FormatPlayResult renders a single PlayResult as a user-facing
// message. Handles single tracks and playlist results uniformly.
// resumeCmd is the bot-specific command suggestion appended when the
// play happened while the bot was stopped (e.g. "$resume-queue" or
// "/audio resume-queue").
func FormatPlayResult(r PlayResult, resumeCmd string) string {
	var b strings.Builder
	if r.Playlist != nil {
		name := r.Playlist.Name
		if name == "" {
			name = "playlist"
		}
		if r.Queued {
			// In-session playlist — every track went into the queue.
			fmt.Fprintf(&b, "Added %d tracks from %s to the queue (starting at position %d)",
				r.Playlist.QueuedTracks, name, r.Position)
		} else {
			// Fresh-start playlist — first track is playing, the rest
			// queued behind it.
			fmt.Fprintf(&b, "Now playing: %s", r.Title)
			if r.Playlist.QueuedTracks > 0 {
				fmt.Fprintf(&b, "\nQueued %d more from %s", r.Playlist.QueuedTracks, name)
			}
		}
		// Note any tracks dropped because the queue hit its cap.
		// In the fresh-start case the first track is playing (not queued),
		// so subtract one from "accounted for" when computing missed.
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

// voiceConnectTimeout is how long each individual voice-connect attempt
// waits for Lavalink to report Connected before declaring that attempt
// failed. maxPlayAttempts is how many times we'll cycle (leave + rejoin)
// before giving up and surfacing ErrVoiceTimeout to the user.
// maxQueueSize caps the per-guild upcoming-track queue.
const (
	voiceConnectTimeout = 7 * time.Second
	maxPlayAttempts     = 2
	maxQueueSize        = 100
)

// PlayResult tells the caller what Play actually did with the URL.
// Title / Queued / Position / WhileStopped describe the *first* track
// of the resolved URL (a single track or the head of a playlist).
// When the URL resolved to a playlist, Playlist is non-nil and carries
// totals for the whole batch — callers can use that to compose a
// playlist-aware message.
type PlayResult struct {
	Title    string
	Queued   bool // true when the track was added to the queue rather than started immediately
	Position int  // queue position when Queued (1-indexed); 0 otherwise
	// WhileStopped is true when Queued and the bot was in a stopped
	// state (Stop was called, no active voice connection, saved track
	// waiting). The queued URL won't actually play until the user runs
	// $resume-queue. Lets the handler tack on a hint to that effect.
	WhileStopped bool
	// Playlist is set when the URL resolved to a Lavalink playlist
	// (typically a YouTube playlist URL). nil otherwise.
	Playlist *PlaylistInfo
}

// PlaylistInfo describes a playlist that Play loaded and enqueued.
type PlaylistInfo struct {
	// Name is the playlist title from Lavalink (e.g. the YouTube
	// playlist's name). May be empty if Lavalink didn't supply one.
	Name string
	// TotalTracks is the count of tracks Lavalink returned for the
	// playlist URL.
	TotalTracks int
	// QueuedTracks is how many of those tracks made it into the queue.
	// Less than TotalTracks if the queue hit its size cap mid-batch.
	// For the "fresh start" path the first track plays immediately and
	// isn't counted here — only the rest that landed in the queue.
	QueuedTracks int
}

// Player is the entry point for playback. One instance per bot — it owns
// the disgolink client and the per-guild queue / announce-channel state.
type Player struct {
	link    disgolink.Client
	session *discordgo.Session

	mu               sync.Mutex
	queues           map[snowflake.ID][]lavalink.Track
	announceChannels map[snowflake.ID]string     // guildID -> text channel ID of the most recent $play
	playSignals      map[snowflake.ID]chan error // guildID -> in-flight Play's failure-signal channel
	// pausedTracks holds the active-track snapshot when Stop is called,
	// so ResumeQueue can replay it from the start after rejoining
	// voice. Distinct from a Lavalink-side pause — the bot fully
	// disconnects, so we have to remember the track ourselves rather
	// than relying on disgolink's destroyed player.
	pausedTracks map[snowflake.ID]lavalink.Track
}

// New constructs a Player. Call OnTrackEnd as a disgolink listener so the
// player can advance queues / disconnect when tracks finish.
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

// Play loads url and either starts playback (if nothing's playing) or
// appends the track to the guild's queue. The caller distinguishes via
// PlayResult.Queued. On a voice-connection timeout (Discord failing to
// deliver a VOICE_STATE / VOICE_SERVER event), Play automatically leaves
// voice, waits for the destroy to propagate, and rejoins — usually fixing
// the issue without the user knowing.
//
// channelID is the text channel where the $play command was issued; it's
// remembered as the destination for auto-advance "Now playing" messages
// when subsequent queued tracks start. Pass an empty string to leave the
// existing announce channel untouched.
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

	// A track is "in session" if Lavalink is currently playing one OR
	// the bot is in a stopped state (disconnected after $stop with a
	// saved track waiting). Either way the new URL is appended to the
	// queue instead of taking over.
	p.mu.Lock()
	_, isStopped := p.pausedTracks[gID]
	p.mu.Unlock()
	existing := p.link.ExistingPlayer(gID)
	inSession := isStopped || (existing != nil && existing.Track() != nil)

	if inSession {
		p.mu.Lock()
		// Reject if even the first track wouldn't fit — without it we
		// can't return a usable Position, so it's cleaner to fail
		// rather than partially succeed.
		if len(p.queues[gID]) >= maxQueueSize {
			p.mu.Unlock()
			return PlayResult{}, ErrQueueFull
		}
		// Append tracks one at a time until the queue caps; remaining
		// playlist tracks (if any) are silently dropped and reflected
		// in PlaylistInfo.QueuedTracks < TotalTracks.
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

	// Nothing playing or stopped — caller must be in a voice channel
	// to start a fresh session.
	voiceState, err := p.session.State.VoiceState(guildID, userID)
	if err != nil || voiceState == nil || voiceState.ChannelID == "" {
		return PlayResult{}, ErrNotInVoice
	}

	// Record the announce channel up-front so OnTrackException can find it
	// even when the track fails before voice has connected.
	if channelID != "" {
		p.mu.Lock()
		p.announceChannels[gID] = channelID
		p.mu.Unlock()
	}

	if err := p.startTrack(ctx, gID, guildID, voiceState.ChannelID, firstTrack); err != nil {
		return PlayResult{}, err
	}

	result := PlayResult{Title: firstTrack.Info.Title}

	// If this URL was a playlist, the rest of its tracks queue up
	// behind the one that just started.
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

// startTrack performs the voice-join + WithTrack + connect-wait retry
// dance for a single track that the caller has already resolved (either
// loaded from a URL via Play or pulled from the saved-track cache via
// ResumeQueue). The caller is also responsible for the user-in-voice
// check and for setting the announce channel.
//
// On success the track is playing and Lavalink + Discord are in sync.
// On failure all voice resources are torn down so the next attempt can
// start from a clean state. voiceTimeout failures retry up to
// maxPlayAttempts — that's Discord dropping VOICE_STATE/VOICE_SERVER
// events, fixable by leaving and rejoining. Track-broken / context-
// cancelled failures bail immediately because retrying won't help.
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

		// Register a failure-signal channel so OnTrackException can short-
		// circuit the connect wait if Lavalink rejects the track outright.
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
			// Track itself is broken — retrying voice won't help. The
			// failure has already been announced via OnTrackException.
			_ = p.session.ChannelVoiceJoinManual(guildID, "", false, true)
			return err
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return err
		}
		lastErr = ErrVoiceTimeout
	}

	// All attempts failed. Final cleanup so the next call sees a clean
	// slate.
	if dgPlayer := p.link.ExistingPlayer(gID); dgPlayer != nil {
		clearCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = dgPlayer.Update(clearCtx, lavalink.WithNullTrack())
		cancel()
	}
	_ = p.session.ChannelVoiceJoinManual(guildID, "", false, true)
	return lastErr
}

// cycleVoice performs a leave + wait-for-cleanup before the next join
// attempt. The leave is what convinces Discord to send fresh VOICE_STATE
// and VOICE_SERVER events on the next join — without it, retrying the
// join with the same channel ID can result in Discord deciding nothing
// changed and not emitting events at all.
func (p *Player) cycleVoice(ctx context.Context, gID snowflake.ID, guildID string) error {
	if dgPlayer := p.link.ExistingPlayer(gID); dgPlayer != nil {
		clearCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = dgPlayer.Update(clearCtx, lavalink.WithNullTrack())
		cancel()
	}
	_ = p.session.ChannelVoiceJoinManual(guildID, "", false, true)

	// Wait for disgolink to fully destroy the player (triggered by the
	// VOICE_STATE_UPDATE(channelID=nil) event the leave will produce).
	// Without this, the next attempt's player can be destroyed mid-setup
	// by the late-arriving leave event.
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

	// Brief pause so we're not hammering Discord while it's already
	// misbehaving.
	select {
	case <-time.After(500 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// waitForPlayback waits for one of three outcomes after we've sent a
// track to Lavalink:
//
//   - state.Connected becomes true: voice connected, track is playing.
//     Returns nil.
//   - signal receives an error: Lavalink emitted TrackExceptionEvent
//     while we were waiting (track is broken). Returns that error
//     (ErrTrackFailed) so the caller can skip the cycleVoice retry.
//   - timeout elapses: neither event fired. Returns ErrVoiceTimeout so
//     the caller can attempt the leave/rejoin retry — this is the
//     "Discord dropped a voice event" case the retry was built for.
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

// clearPlaySignal removes the per-guild failure-signal channel after Play
// has finished waiting. Safe to call when no signal is registered.
func (p *Player) clearPlaySignal(gID snowflake.ID) {
	p.mu.Lock()
	delete(p.playSignals, gID)
	p.mu.Unlock()
}

// QueueSnapshot is a read-only view of a guild's playback state. In
// normal operation Current and Paused are mutually exclusive: the bot
// is either actively playing (Current set), in a stopped session
// (Paused set, will be replayed by ResumeQueue), or fully idle (both
// nil). Upcoming is the queue that plays after the active/paused track,
// in play order — copied at snapshot time, safe to read without a lock.
type QueueSnapshot struct {
	Current  *lavalink.Track
	Paused   *lavalink.Track
	Upcoming []lavalink.Track
}

// Queue returns a snapshot of the guild's playback state, including the
// saved-by-Stop track (Paused) when the bot is in a stopped session.
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
		// Copy so the caller can hold the pointer without aliasing
		// the map's value storage.
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

// Skip ends the current track and starts the next one from the queue.
// If the queue is empty after skipping, playback stops and the bot
// leaves voice — caller can detect this by checking whether next is nil.
// Returns ErrNothingPlaying when nothing is currently playing.
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
		// Update the player with the next track. This emits a
		// TrackEndEvent for the old track with Reason=replaced;
		// OnTrackEnd ignores that case (MayStartNext is false), so it
		// won't double-advance the queue.
		if err := dgPlayer.Update(ctx, lavalink.WithTrack(*next)); err != nil {
			return skipped, nil, fmt.Errorf("play next: %w", err)
		}
		return skipped, next, nil
	}

	// Queue is empty — stop playback and leave voice, matching the
	// "queue-empty disconnect" behavior used when a track ends naturally.
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

// ClearQueue removes all upcoming tracks from the guild's queue. If the
// bot is currently stopped (Stop was called and a saved track is
// waiting), that saved track is wiped too — $clear is a full session
// reset. The currently-playing track in an active session is NOT
// affected. Returns the total number of tracks removed.
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

// Stop snapshots the currently-playing track into pausedTracks, clears
// it in Lavalink, and disconnects from voice. The upcoming queue is
// left intact — ResumeQueue replays the saved track from position 0
// and continues with whatever was queued behind it. Use ClearQueue
// (i.e. $clear) to fully reset including the saved track.
//
// changed reports whether the state actually flipped (true=newly
// stopped, false=was already stopped). Returns ErrNothingPlaying if
// nothing is active AND no stopped session exists.
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
			return false, nil // idempotent — already stopped
		}
		return false, ErrNothingPlaying
	}

	// Snapshot the active track so ResumeQueue can restart it.
	activeTrack := *dgPlayer.Track()
	p.mu.Lock()
	p.pausedTracks[gID] = activeTrack
	p.mu.Unlock()

	// Best-effort track clear on the Lavalink side. The disconnect
	// below is what actually stops audio for the user.
	if err := dgPlayer.Update(ctx, lavalink.WithNullTrack()); err != nil {
		log.Printf("voice_chat: clear track on stop: %v", err)
	}
	if err := p.session.ChannelVoiceJoinManual(guildID, "", false, true); err != nil {
		return true, fmt.Errorf("leave voice: %w", err)
	}
	return true, nil
}

// ResumeQueue rejoins userID's current voice channel and restarts the
// track previously saved by Stop, starting at position 0. After that
// track ends, OnTrackEnd advances through the upcoming queue normally.
// channelID is the text channel where the command was issued — used
// as the new announce-channel destination for auto-advance messages.
//
// Returns the resumed track on success.
//
// Errors: ErrAlreadyPlaying if a track is currently playing (use Skip
// or wait), ErrNothingPlaying if no saved session exists, ErrNotInVoice
// if userID isn't in a voice channel.
func (p *Player) ResumeQueue(ctx context.Context, guildID, channelID, userID string) (lavalink.Track, error) {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return lavalink.Track{}, fmt.Errorf("parse guild id: %w", err)
	}

	// If something is already playing, there's nothing to resume from.
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
		// Leave pausedTracks intact so the user can retry.
		return lavalink.Track{}, err
	}

	p.mu.Lock()
	delete(p.pausedTracks, gID)
	p.mu.Unlock()
	return track, nil
}

// OnTrackException is registered with disgolink. When a track fails to
// play (cipher errors, removed videos, region locks, etc.) it does two
// things: signals any in-flight Play() so it can bail out instead of
// burning the voice-connect timeout, and posts a failure message to the
// announce channel so the user understands why the next track skipped
// or why the bot left voice.
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
	// Discord rejects messages > 2000 characters with HTTP 400. Lavalink's
	// AllClientsFailedException can dump several KB of stack-trace-like
	// detail into Exception.Message — keep our final string well under
	// the cap.
	if len(msg) > 1900 {
		msg = msg[:1900] + "…"
	}
	if _, err := p.session.ChannelMessageSend(announceCh, msg); err != nil {
		log.Printf("voice_chat: announce track exception in guild %s: %v", gID, err)
	}
}

// briefExceptionReason returns the first non-empty line of an exception
// message, capped at 200 characters. Lavalink exceptions can be
// multi-line concatenations of every-client-tried details — we just want
// the headline.
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

// OnTrackEnd is registered with disgolink. It advances the per-guild
// queue if there's a next track and posts an auto-advance "Now playing"
// message to the channel of the most recent $play. If the queue is
// empty it disconnects from voice and forgets the announce channel.
func (p *Player) OnTrackEnd(player disgolink.Player, e lavalink.TrackEndEvent) {
	defer recoverCallback("OnTrackEnd", player.GuildID())
	if !e.Reason.MayStartNext() {
		// Stopped, replaced, or cleanup — Player.Stop already disconnected,
		// or another caller is replacing the track. Don't intervene.
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

// loadTrack drives Lavalink's track resolver for a single URL.
// trackLoad captures everything Lavalink resolved from a single URL.
// Tracks is empty on no-match (ErrNoTrackFound). For single-track and
// search-result URLs Tracks has exactly one element; for playlist URLs
// it has the full track list and PlaylistName is set.
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
			// Search result (e.g. ytsearch:foo) — Lavalink returns
			// many candidates; we only take the first since the user
			// asked for one thing.
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

// recoverCallback is deferred at the top of each disgolink listener so a
// panic in our handler doesn't kill the disgolink goroutine silently.
// Logs to stderr — the player has no error-channel config to forward to.
func recoverCallback(name string, gID snowflake.ID) {
	if r := recover(); r != nil {
		log.Printf("voice_chat: panic in %s for guild %s: %v\n%s", name, gID, r, debug.Stack())
	}
}
