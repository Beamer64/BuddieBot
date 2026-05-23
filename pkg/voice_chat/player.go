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
type PlayResult struct {
	Title    string
	Queued   bool // true when the track was added to the queue rather than started immediately
	Position int  // queue position when Queued (1-indexed); 0 otherwise
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

	track, err := loadTrack(ctx, node, url)
	if err != nil {
		return PlayResult{}, err
	}

	// If something is already playing in this guild, append to the queue
	// instead of trying to take over. Caller doesn't need to be in voice
	// to add to a queue that's already playing somewhere.
	if existing := p.link.ExistingPlayer(gID); existing != nil && existing.Track() != nil {
		p.mu.Lock()
		if len(p.queues[gID]) >= maxQueueSize {
			p.mu.Unlock()
			return PlayResult{}, ErrQueueFull
		}
		p.queues[gID] = append(p.queues[gID], *track)
		position := len(p.queues[gID])
		if channelID != "" {
			p.announceChannels[gID] = channelID
		}
		p.mu.Unlock()
		return PlayResult{Title: track.Info.Title, Queued: true, Position: position}, nil
	}

	// Nothing playing — caller must be in a voice channel to start playback.
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

	var lastErr error
	for attempt := 1; attempt <= maxPlayAttempts; attempt++ {
		if attempt > 1 {
			log.Printf("voice_chat: voice didn't connect on attempt %d for guild %s, retrying with fresh join", attempt-1, guildID)
			if err := p.cycleVoice(ctx, gID, guildID); err != nil {
				return PlayResult{}, err
			}
		}

		dgPlayer := p.link.Player(gID)

		if err := p.session.ChannelVoiceJoinManual(guildID, voiceState.ChannelID, false, true); err != nil {
			lastErr = fmt.Errorf("voice channel join: %w", err)
			continue
		}

		// Register a failure-signal channel so OnTrackException can short-
		// circuit the connect wait if Lavalink rejects the track outright.
		signal := make(chan error, 1)
		p.mu.Lock()
		p.playSignals[gID] = signal
		p.mu.Unlock()

		if err := dgPlayer.Update(ctx, lavalink.WithTrack(*track)); err != nil {
			p.clearPlaySignal(gID)
			_ = p.session.ChannelVoiceJoinManual(guildID, "", false, true)
			return PlayResult{}, fmt.Errorf("start playback: %w", err)
		}

		err := waitForPlayback(ctx, dgPlayer, signal, voiceConnectTimeout)
		p.clearPlaySignal(gID)
		switch {
		case err == nil:
			return PlayResult{Title: track.Info.Title}, nil
		case errors.Is(err, ErrTrackFailed):
			// Track itself is broken — retrying voice won't help. The
			// failure has already been announced via OnTrackException.
			_ = p.session.ChannelVoiceJoinManual(guildID, "", false, true)
			return PlayResult{}, err
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return PlayResult{}, err
		}
		lastErr = ErrVoiceTimeout
	}

	// All attempts failed. Final cleanup so $play can be retried by the user
	// from a clean slate.
	if dgPlayer := p.link.ExistingPlayer(gID); dgPlayer != nil {
		clearCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = dgPlayer.Update(clearCtx, lavalink.WithNullTrack())
		cancel()
	}
	_ = p.session.ChannelVoiceJoinManual(guildID, "", false, true)
	return PlayResult{}, lastErr
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

// Queue returns a snapshot of the current track and upcoming queue for
// the given guild. The upcoming slice is a copy and safe to read from
// the caller without holding any lock.
func (p *Player) Queue(guildID string) (current *lavalink.Track, upcoming []lavalink.Track, err error) {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return nil, nil, fmt.Errorf("parse guild id: %w", err)
	}

	if existing := p.link.ExistingPlayer(gID); existing != nil {
		current = existing.Track()
	}

	p.mu.Lock()
	if q := p.queues[gID]; len(q) > 0 {
		upcoming = make([]lavalink.Track, len(q))
		copy(upcoming, q)
	}
	p.mu.Unlock()

	return current, upcoming, nil
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

// ClearQueue removes all upcoming tracks from the guild's queue without
// affecting the currently-playing track. Use Stop to end playback
// entirely. Returns the number of tracks that were removed.
func (p *Player) ClearQueue(guildID string) (int, error) {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return 0, fmt.Errorf("parse guild id: %w", err)
	}

	p.mu.Lock()
	count := len(p.queues[gID])
	delete(p.queues, gID)
	p.mu.Unlock()

	return count, nil
}

// Stop ends the active track and disconnects from voice. Idempotent.
func (p *Player) Stop(ctx context.Context, guildID string) error {
	gID, err := snowflake.Parse(guildID)
	if err != nil {
		return fmt.Errorf("parse guild id: %w", err)
	}

	p.mu.Lock()
	delete(p.queues, gID)
	delete(p.announceChannels, gID)
	p.mu.Unlock()

	if dgPlayer := p.link.ExistingPlayer(gID); dgPlayer != nil {
		if err := dgPlayer.Update(ctx, lavalink.WithNullTrack()); err != nil {
			log.Printf("voice_chat: clear track on stop: %v", err)
		}
	}

	return p.session.ChannelVoiceJoinManual(guildID, "", false, true)
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
func loadTrack(ctx context.Context, node disgolink.Node, url string) (*lavalink.Track, error) {
	var (
		track   *lavalink.Track
		loadErr error
	)
	node.LoadTracksHandler(ctx, url, disgolink.NewResultHandler(
		func(t lavalink.Track) {
			track = &t
		},
		func(pl lavalink.Playlist) {
			if len(pl.Tracks) > 0 {
				t := pl.Tracks[0]
				track = &t
			}
		},
		func(ts []lavalink.Track) {
			if len(ts) > 0 {
				t := ts[0]
				track = &t
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
		return nil, loadErr
	}
	if track == nil {
		return nil, ErrNoTrackFound
	}
	return track, nil
}

// recoverCallback is deferred at the top of each disgolink listener so a
// panic in our handler doesn't kill the disgolink goroutine silently.
// Logs to stderr — the player has no error-channel config to forward to.
func recoverCallback(name string, gID snowflake.ID) {
	if r := recover(); r != nil {
		log.Printf("voice_chat: panic in %s for guild %s: %v\n%s", name, gID, r, debug.Stack())
	}
}
