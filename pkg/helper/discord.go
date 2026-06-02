package helper

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// maxEmbedFieldValue: below Discord's 1024 cap to leave room for code fences.
const maxEmbedFieldValue = 1000

// truncateForEmbed keeps the head — informative front (error message).
func truncateForEmbed(s string) string {
	if len(s) <= maxEmbedFieldValue {
		return s
	}
	return s[:maxEmbedFieldValue-3] + "..."
}

// tailForEmbed keeps the tail — informative back (deepest stack frames).
func tailForEmbed(s string) string {
	if len(s) <= maxEmbedFieldValue {
		return s
	}
	return "..." + s[len(s)-(maxEmbedFieldValue-3):]
}

// ─────────────────────────────────────────────────────────────────────
// Interaction error helpers
//
// Slash and component handlers report errors back to the invoker
// through one of these. Pick based on where you are in the interaction
// lifecycle and how visible the error should be:
//
//	pre-defer / component click:
//	    SendEphemeralMsgPreDeferred, LogSendEphemeralMsgPreDeferred
//
//	post-defer, error inherits the defer's visibility:
//	    EditMsgPostDeferred
//
//	post-defer, error must be ephemeral regardless of the defer:
//	    SendEphemeralFollowUpPostDeferred, LogSendEphemeralFollowUpPostDeferred
//
// The Return* variants additionally return err so wrap() logs the
// underlying cause to the error channel. The Send* variants don't.
// ─────────────────────────────────────────────────────────────────────

// SendEphemeralMsgPreDeferred sends an ephemeral message into the interaction's
// initial-response slot. Use BEFORE a defer (or for component clicks,
// which never defer) when you have a user-facing message but no underlying
// err to propagate.
func SendEphemeralMsgPreDeferred(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	return s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: message,
			},
		},
	)
}

// LogSendEphemeralMsgPreDeferred = SendEphemeralMsgPreDeferred + return err. Use BEFORE defer when
// validation or a rate-limit check fails AND you want wrap() to log the
// cause to the error channel.
func LogSendEphemeralMsgPreDeferred(s *discordgo.Session, i *discordgo.InteractionCreate, userMsg string, err error) error {
	if sendErr := SendEphemeralMsgPreDeferred(s, i, userMsg); sendErr != nil {
		log.Printf("failed to send error response: %v (original: %v)", sendErr, err)
	}
	return err
}

// EditMsgPostDeferred edits the deferred interaction response with the given
// message. Use AFTER defer when you have a user-facing message but no err
// to propagate (e.g. post-defer validation failure). Visibility inherits
// from the original defer.
func EditMsgPostDeferred(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	_, err := s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{Content: &message},
	)
	return err
}

// LogEditMsgPostDeferred edits the deferred interaction response with
// userMsg and returns err for wrap() to log. Use AFTER defer for any
// post-defer failure. Visibility inherits from the original defer —
// public unless the defer set MessageFlagsEphemeral.
func LogEditMsgPostDeferred(s *discordgo.Session, i *discordgo.InteractionCreate, userMsg string, err error) error {
	if _, sendErr := s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{Content: &userMsg},
	); sendErr != nil {
		log.Printf("failed to edit deferred response with error: %v (original: %v)", sendErr, err)
	}
	return err
}

// SendEphemeralFollowUpPostDeferred sends an ephemeral FOLLOW-UP after a
// public defer. Use when the success path is public but the error should
// stay private to the invoker — the deferred "Bot is thinking…" placeholder
// is deleted first so the user doesn't see an orphan public message next to
// the ephemeral one.
func SendEphemeralFollowUpPostDeferred(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	if delErr := s.InteractionResponseDelete(i.Interaction); delErr != nil {
		log.Printf("delete deferred response before ephemeral followup: %v", delErr)
	}
	_, err := s.FollowupMessageCreate(
		i.Interaction, true, &discordgo.WebhookParams{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	)
	return err
}

// LogSendEphemeralFollowUpPostDeferred = SendEphemeralFollowUpPostDeferred + return err
// for wrap() to log. Use when you want both: a private error reply AND
// wrap() to record the cause in the error channel.
func LogSendEphemeralFollowUpPostDeferred(s *discordgo.Session, i *discordgo.InteractionCreate, userMsg string, err error) error {
	if sendErr := SendEphemeralFollowUpPostDeferred(s, i, userMsg); sendErr != nil {
		log.Printf("failed to send ephemeral error followup: %v (original: %v)", sendErr, err)
	}
	return err
}

// ─────────────────────────────────────────────────────────────────────
// Prefix-command error helpers
// ─────────────────────────────────────────────────────────────────────

// LogAndReact is the prefix-command error finisher. On non-nil err it
// posts the error to the error channel, DMs the invoker a generic notice,
// and falls back to a ⚠️ reaction on the original message if the DM is
// blocked. No-op when err is nil.
func LogAndReact(s *discordgo.Session, m *discordgo.MessageCreate, errorLogChannelID string, err error) {
	if err == nil {
		return
	}
	LogErrorsToErrorChannel(s, errorLogChannelID, err, m.GuildID)
	if dmErr := sendErrorDMToUser(s, m); dmErr != nil {
		log.Printf("prefix: DM error to user %s failed (%v) — falling back to reaction", m.Author.ID, dmErr)
		if reactErr := s.MessageReactionAdd(m.ChannelID, m.ID, ErrorReaction); reactErr != nil {
			log.Printf("prefix: add error reaction on message %s: %v", m.ID, reactErr)
		}
	}
}

// sendErrorDMToUser opens a DM with the message author and posts the
// generic prefix-error notice. Private — only LogAndReact calls it.
func sendErrorDMToUser(s *discordgo.Session, m *discordgo.MessageCreate) error {
	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return fmt.Errorf("open DM with user %s: %w", m.Author.ID, err)
	}
	if _, err := s.ChannelMessageSend(dm.ID, "There was an error with this request. Big Brother is already looking into it."); err != nil {
		return fmt.Errorf("send error DM to user %s: %w", m.Author.ID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────
// Channel logging
// ─────────────────────────────────────────────────────────────────────

// LogErrorsToErrorChannel posts a summary embed + full-stack .txt
// attachment to the bot's error channel. Called by wrap() (handler errors
// and panics), event handlers, and LogAndReact for prefix commands.
// Console logs the full stack regardless of channel-send outcome.
func LogErrorsToErrorChannel(s *discordgo.Session, errorLogChannelID string, err error, guildID string) {
	fullStack := fmt.Sprintf("%+v", errors.WithStack(err))
	log.Print(fullStack)

	msg := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{getErrorEmbed(err, s, guildID)},
		Files: []*discordgo.File{
			{
				Name:        fmt.Sprintf("error-%s.txt", time.Now().Format("20060102-150405")),
				ContentType: "text/plain",
				Reader:      strings.NewReader(fullStack),
			},
		},
	}

	if _, sendErr := s.ChannelMessageSendComplex(errorLogChannelID, msg); sendErr != nil {
		log.Printf("failed to send error report to channel: %v (original: %v)", sendErr, err)
	}
}

// getErrorEmbed builds the summary embed posted to the error channel.
// Private — only LogErrorsToErrorChannel calls it.
func getErrorEmbed(err error, s *discordgo.Session, gID string) *discordgo.MessageEmbed {
	var guild *discordgo.Guild
	guildID := "N/A"
	guildName := "N/A"

	if gID != "" {
		guild, _ = s.Guild(gID)
		guildID = gID
		if guild != nil {
			guildName = guild.Name
		}
	}

	fullStack := fmt.Sprintf("%+v", errors.WithStack(err))

	return &discordgo.MessageEmbed{
		Title:       "ERROR",
		Description: "(ノಠ益ಠ)ノ彡┻━┻",
		Color:       16726843,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Guild ID",
				Value:  guildID,
				Inline: true,
			},
			{
				Name:   "Guild Name",
				Value:  guildName,
				Inline: true,
			},
			{
				Name:   "Error",
				Value:  truncateForEmbed(err.Error()),
				Inline: false,
			},
			{
				Name:   "Stack (truncated)",
				Value:  "```\n" + tailForEmbed(fullStack) + "\n```",
				Inline: false,
			},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────
// Misc
// ─────────────────────────────────────────────────────────────────────

// IsBotOwner reports whether userID is listed as a bot maintainer in the
// supplied owner-ID slice. Empty slice → always false (nobody is owner —
// the safe default if config is unpopulated). Used to gate owner-only
// surfaces ($release, $test).
func IsBotOwner(userID string, ownerIDs []string) bool {
	if userID == "" {
		return false
	}
	for _, id := range ownerIDs {
		if id == userID {
			return true
		}
	}
	return false
}
