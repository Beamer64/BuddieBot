package slash

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/BuddieBot/pkg/voice_chat"
	"github.com/bwmarrin/discordgo"
)

func sendAudioResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	gateCtx, gateCancel := context.WithTimeout(context.Background(), 2*time.Second)
	enabled, err := cfg.DB.GuildAudioEnabled(gateCtx, i.GuildID)
	gateCancel()
	if err != nil {
		return helper.ReturnUserError(s, i, "Audio check failed, try again.", fmt.Errorf("guild audio enabled lookup: %w", err))
	}
	if !enabled {
		return helper.ReturnUserError(s, i, "Audio commands aren't enabled in this server. Ask a bot admin for access.", nil)
	}
	if cfg.Player == nil {
		return helper.ReturnUserError(s, i, "Audio is not available right now.", nil)
	}

	sub := i.ApplicationCommandData().Options[0]

	// Defer up-front — play / resume-queue can run past the 3s initial-response deadline.
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /audio %s: %w", sub.Name, err)
	}

	switch sub.Name {
	case "play":
		return audioPlay(s, i, cfg, sub.Options)
	case "stop":
		return audioStop(s, i, cfg)
	case "resume-queue":
		return audioResumeQueue(s, i, cfg)
	case "queue":
		return audioQueue(s, i, cfg)
	case "skip":
		return audioSkip(s, i, cfg)
	case "clear":
		return audioClear(s, i, cfg)
	default:
		return helper.ReturnUserErrorDeferred(s, i, "Unknown audio subcommand.", fmt.Errorf("unknown audio subcommand: %s", sub.Name))
	}
}

// audioPlay dispatches to single- or batch-URL play. Empty/whitespace url
// values are skipped so clearing an option doesn't trip validation.
func audioPlay(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
	provided := map[string]string{}
	for _, opt := range opts {
		provided[opt.Name] = strings.TrimSpace(opt.StringValue())
	}
	var urls []string
	for _, name := range []string{"url-1", "url-2", "url-3"} {
		if v := provided[name]; v != "" {
			urls = append(urls, v)
		}
	}
	if len(urls) == 0 {
		return audioEditMessage(s, i, "`url-1` is required.")
	}

	if len(urls) == 1 {
		return audioPlayOne(s, i, cfg, urls[0])
	}
	return audioPlayBatch(s, i, cfg, urls)
}

// audioPlayOne handles the single-URL case (single track or playlist).
// Per-error friendly translation, unlike the batch summary.
func audioPlayOne(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, url string) error {
	// 2 min covers cold voice-connect retry + a large playlist load.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := cfg.Player.Play(ctx, i.GuildID, i.ChannelID, i.Member.User.ID, url)
	if err != nil {
		// User-facing errors: edit-and-return-nil so wrap() doesn't log as a bug.
		// Other errors: ReturnUserErrorDeferred surfaces the message AND returns
		// the wrapped err for the error channel.
		if voice_chat.IsUserFacingError(err) {
			return audioEditMessage(s, i, voice_chat.FriendlyPlayError(err))
		}
		return helper.ReturnUserErrorDeferred(s, i, voice_chat.FriendlyPlayError(err), fmt.Errorf("audio play: %w", err))
	}
	return audioEditMessage(s, i, voice_chat.FormatPlayResult(result, "/audio resume-queue"))
}

// audioPlayBatch calls Play across multiple URLs sequentially. Fatal
// errors (no voice, voice timeout, queue full) abort the batch; per-URL
// errors (bad link, broken track) are counted and reported in the summary.
// Playlists inside a batch are summarized as one line, not enumerated.
func audioPlayBatch(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, urls []string) error {
	// 2 min covers the worst case: cold voice-connect retry + N playlist resolves.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	type queuedItem struct {
		title    string
		position int
	}
	type playlistEntry struct {
		name     string
		added    int
		total    int
		firstPos int // queue position of the first added track; valid only when !started
		started  bool
	}

	var (
		playingTitle string
		queued       []queuedItem
		playlists    []playlistEntry
		failures     int
		fatalMsg     string
		whileStopped bool
	)

	for _, url := range urls {
		result, err := cfg.Player.Play(ctx, i.GuildID, i.ChannelID, i.Member.User.ID, url)
		if err == nil {
			if result.WhileStopped {
				whileStopped = true
			}
			switch {
			case result.Playlist != nil:
				name := result.Playlist.Name
				if name == "" {
					name = "playlist"
				}
				pe := playlistEntry{
					name:     name,
					added:    result.Playlist.QueuedTracks,
					total:    result.Playlist.TotalTracks,
					firstPos: result.Position,
					started:  !result.Queued,
				}
				if pe.started {
					playingTitle = result.Title
				}
				playlists = append(playlists, pe)
			case result.Queued:
				queued = append(queued, queuedItem{title: result.Title, position: result.Position})
			default:
				playingTitle = result.Title
			}
			continue
		}

		// Fatal errors fail every remaining URL — surface one message.
		if errors.Is(err, voice_chat.ErrNotInVoice) ||
			errors.Is(err, voice_chat.ErrVoiceTimeout) ||
			errors.Is(err, voice_chat.ErrQueueFull) {
			fatalMsg = voice_chat.FriendlyPlayError(err)
			break
		}
		// Per-URL failure — skip and keep going. ErrTrackFailed details
		// were already announced via OnTrackException.
		failures++
	}

	var msg strings.Builder
	if playingTitle != "" {
		fmt.Fprintf(&msg, "Now playing: **%s**\n", playingTitle)
	}
	for _, pl := range playlists {
		if pl.started {
			if pl.added > 0 {
				fmt.Fprintf(&msg, "Queued %d more from **%s**\n", pl.added, pl.name)
			}
		} else {
			fmt.Fprintf(
				&msg, "Queued %d tracks from **%s** (starting at position %d)\n",
				pl.added, pl.name, pl.firstPos,
			)
		}
		played := 0
		if pl.started {
			played = 1
		}
		if missed := pl.total - pl.added - played; missed > 0 {
			fmt.Fprintf(&msg, "  (queue full — %d not added from **%s**)\n", missed, pl.name)
		}
	}
	if len(queued) == 1 {
		fmt.Fprintf(&msg, "Added to queue: **%s** (position %d)\n", queued[0].title, queued[0].position)
	} else if len(queued) > 1 {
		msg.WriteString("Added to queue:\n")
		for idx, item := range queued {
			line := fmt.Sprintf("  %d. %s\n", item.position, item.title)
			// Discord 2000-char cap; leave room for the overflow line.
			if msg.Len()+len(line) > 1900 {
				fmt.Fprintf(&msg, "  …and %d more\n", len(queued)-idx)
				break
			}
			msg.WriteString(line)
		}
	}
	if failures > 0 {
		suffix := "s"
		if failures == 1 {
			suffix = ""
		}
		fmt.Fprintf(&msg, "Couldn't resolve %d URL%s.\n", failures, suffix)
	}
	if fatalMsg != "" {
		msg.WriteString(fatalMsg)
		msg.WriteString("\n")
	}
	if whileStopped {
		msg.WriteString("Use /audio resume-queue to start playback.\n")
	}
	if msg.Len() == 0 {
		msg.WriteString("Nothing happened.")
	}
	return audioEditMessage(s, i, strings.TrimRight(msg.String(), "\n"))
}

func audioStop(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	changed, err := cfg.Player.Stop(ctx, i.GuildID)
	if err != nil {
		if errors.Is(err, voice_chat.ErrNothingPlaying) {
			return audioEditMessage(s, i, "Nothing is playing.")
		}
		return helper.ReturnUserErrorDeferred(s, i, "Failed to stop playback.", fmt.Errorf("audio stop: %w", err))
	}

	msg := "Stopped. Use /audio resume-queue to pick up where you left off."
	if !changed {
		msg = "Already stopped. Use /audio resume-queue to pick up where you left off."
	}
	return audioEditMessage(s, i, msg)
}

func audioResumeQueue(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	// Rejoin runs the full voice-connect retry loop.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	track, err := cfg.Player.ResumeQueue(ctx, i.GuildID, i.ChannelID, i.Member.User.ID)
	if err != nil {
		switch {
		case errors.Is(err, voice_chat.ErrNothingPlaying):
			return audioEditMessage(s, i, "Nothing to resume.")
		case errors.Is(err, voice_chat.ErrAlreadyPlaying):
			return audioEditMessage(s, i, "Already playing.")
		case voice_chat.IsUserFacingError(err):
			return audioEditMessage(s, i, voice_chat.FriendlyPlayError(err))
		}
		return helper.ReturnUserErrorDeferred(s, i, "Failed to resume playback.", fmt.Errorf("audio resume-queue: %w", err))
	}
	return audioEditMessage(s, i, "Resumed: "+track.Info.Title)
}

func audioQueue(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	snap, err := cfg.Player.Queue(i.GuildID)
	if err != nil {
		return helper.ReturnUserErrorDeferred(s, i, "Failed to fetch queue.", fmt.Errorf("audio queue: %w", err))
	}

	if snap.Current == nil && snap.Paused == nil && len(snap.Upcoming) == 0 {
		return audioEditMessage(s, i, "Nothing playing.")
	}

	var msg strings.Builder
	switch {
	case snap.Current != nil:
		msg.WriteString("Now playing: ")
		msg.WriteString(snap.Current.Info.Title)
		msg.WriteString("\n")
	case snap.Paused != nil:
		msg.WriteString("Stopped: ")
		msg.WriteString(snap.Paused.Info.Title)
		msg.WriteString(" (use /audio resume-queue to replay)\n")
	}
	if len(snap.Upcoming) > 0 {
		msg.WriteString("Up next:\n")
		for idx, t := range snap.Upcoming {
			line := fmt.Sprintf("  %d. %s\n", idx+1, t.Info.Title)
			// Discord 2000-char cap; leave room for the "…and N more" line.
			if msg.Len()+len(line) > 1900 {
				fmt.Fprintf(&msg, "  …and %d more\n", len(snap.Upcoming)-idx)
				break
			}
			msg.WriteString(line)
		}
	}
	return audioEditMessage(s, i, strings.TrimRight(msg.String(), "\n"))
}

func audioSkip(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	skipped, next, err := cfg.Player.Skip(ctx, i.GuildID)
	if err != nil {
		if errors.Is(err, voice_chat.ErrNothingPlaying) {
			return audioEditMessage(s, i, "Nothing is playing.")
		}
		return helper.ReturnUserErrorDeferred(s, i, "Failed to skip track.", fmt.Errorf("audio skip: %w", err))
	}

	var msg string
	if next != nil {
		msg = fmt.Sprintf("Skipped: %s.\nNow playing: %s", skipped.Info.Title, next.Info.Title)
	} else {
		msg = fmt.Sprintf("Skipped: %s. Queue is empty — leaving voice.", skipped.Info.Title)
	}
	return audioEditMessage(s, i, msg)
}

func audioClear(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	count, err := cfg.Player.ClearQueue(i.GuildID)
	if err != nil {
		return helper.ReturnUserErrorDeferred(s, i, "Failed to clear queue.", fmt.Errorf("audio clear: %w", err))
	}

	var msg string
	switch {
	case count == 0:
		msg = "Queue is already empty."
	case count == 1:
		msg = "Cleared 1 track."
	default:
		msg = fmt.Sprintf("Cleared %d tracks.", count)
	}
	return audioEditMessage(s, i, msg)
}

func audioEditMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error {
	if _, err := s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		},
	); err != nil {
		return fmt.Errorf("send /audio response: %w", err)
	}
	return nil
}

func audioSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "audio",
		Description: "Audio playback controls",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "play",
				Description: "Play an audio URL (or queue it if something is already playing)",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "url-1",
						Description: "audio URL",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "url-2",
						Description: "audio URL",
						Required:    false,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "url-3",
						Description: "audio URL",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "stop",
				Description: "Disconnect, but save the queue for /audio resume-queue",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "resume-queue",
				Description: "Rejoin voice and restart the saved track from the beginning",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "queue",
				Description: "Show what's playing and what's coming up",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "skip",
				Description: "Skip to the next track in the queue",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "clear",
				Description: "Clear the upcoming queue (and the saved track if stopped)",
			},
		},
	}
}
