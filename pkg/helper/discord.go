package helper

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// maxEmbedFieldValue is below Discord's 1024-char field limit to leave room for
// our code-fence wrapper around stack snippets.
const maxEmbedFieldValue = 1000

// truncateForEmbed keeps the first maxEmbedFieldValue characters of s,
// suffixing "..." when truncated. Best when the start of the string has the
// most informative content (e.g., a human-readable error message).
func truncateForEmbed(s string) string {
	if len(s) <= maxEmbedFieldValue {
		return s
	}
	return s[:maxEmbedFieldValue-3] + "..."
}

// tailForEmbed keeps the last maxEmbedFieldValue characters of s, prefixing
// "..." when truncated. Best for stack traces where the deepest (last) frames
// are typically closest to the actual failure point.
func tailForEmbed(s string) string {
	if len(s) <= maxEmbedFieldValue {
		return s
	}
	return "..." + s[len(s)-(maxEmbedFieldValue-3):]
}

func GetErrorEmbed(err error, s *discordgo.Session, gID string) *discordgo.MessageEmbed {
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

func LogAndReact(s *discordgo.Session, m *discordgo.MessageCreate, errorLogChannelID string, err error) {
	if err == nil {
		return
	}
	LogErrorsToErrorChannel(s, errorLogChannelID, err, m.GuildID)
	if dmErr := SendErrorDMToUser(s, m); dmErr != nil {
		log.Printf("prefix: DM error to user %s failed (%v) — falling back to reaction", m.Author.ID, dmErr)
		if reactErr := s.MessageReactionAdd(m.ChannelID, m.ID, ErrorReaction); reactErr != nil {
			log.Printf("prefix: add error reaction on message %s: %v", m.ID, reactErr)
		}
	}
}

// SendErrorDMToUser opens a DM channel with the message author
func SendErrorDMToUser(s *discordgo.Session, m *discordgo.MessageCreate) error {
	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return fmt.Errorf("open DM with user %s: %w", m.Author.ID, err)
	}
	if _, err := s.ChannelMessageSend(dm.ID, "There was an error with this request. Big Brother is already looking into it."); err != nil {
		return fmt.Errorf("send error DM to user %s: %w", m.Author.ID, err)
	}
	return nil
}

func SendEphemeralError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6, // Ephemeral
				Content: message,
			},
		},
	)

	return err
}

// ReturnUserError sends an ephemeral error message to the invoking user and
// returns the original handler error. If sending the user-facing message
// itself fails, that secondary failure is logged and the original error is
// still returned so callers see the underlying cause.
func ReturnUserError(s *discordgo.Session, i *discordgo.InteractionCreate, userMsg string, err error) error {
	if sendErr := SendEphemeralError(s, i, userMsg); sendErr != nil {
		log.Printf("failed to send error response: %v (original: %v)", sendErr, err)
	}
	return err
}

// EditWithErrorMessage replaces a previously-deferred interaction response
// with a user-facing error message. Use this in handlers that defer the
// interaction up-front — SendEphemeralError would 404 in that flow
// because the initial response was already consumed by the defer.
func EditWithErrorMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	_, err := s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Content: &message,
		},
	)
	return err
}

// ReturnUserErrorDeferred is the deferred-interaction counterpart of
// ReturnUserError: surfaces the user-facing message via InteractionResponseEdit
// and returns the original handler error so callers see the underlying cause.
func ReturnUserErrorDeferred(s *discordgo.Session, i *discordgo.InteractionCreate, userMsg string, err error) error {
	if sendErr := EditWithErrorMessage(s, i, userMsg); sendErr != nil {
		log.Printf("failed to edit deferred response with error: %v (original: %v)", sendErr, err)
	}
	return err
}

// LogErrorsToErrorChannel logs full stack to the console and sends a summary
// embed plus an in-memory text-file attachment with the full stack to the
// configured Discord error channel. The .txt is streamed directly to Discord
// — nothing is written to local disk.
func LogErrorsToErrorChannel(s *discordgo.Session, errorLogChannelID string, err error, guildID string) {
	fullStack := fmt.Sprintf("%+v", errors.WithStack(err))
	log.Print(fullStack)

	msg := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{GetErrorEmbed(err, s, guildID)},
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

// IsAudioGuild reports whether the given guild has audio commands
// enabled. Currently only the master and test guilds; audio commands
// gate on this and return a user-facing message in other guilds.
func IsAudioGuild(guildID, masterGuildID, testGuildID string) bool {
	return guildID == masterGuildID || guildID == testGuildID
}

func MemberHasRole(session *discordgo.Session, m *discordgo.Member, guildID string, roleName string) bool {
	if guildID == "" {
		guildID = m.GuildID
	}
	roleName = strings.ToLower(roleName)

	for _, roleID := range m.Roles {
		role, err := session.State.Role(guildID, roleID)
		if err != nil {
			log.Printf("MemberHasRole: resolve role %s in guild %s: %v", roleID, guildID, err)
			continue
		}

		if strings.ToLower(role.Name) == roleName {
			return true
		}
	}
	return false
}
