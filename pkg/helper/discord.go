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

// ReturnUserError sends an ephemeral message and returns err so callers
// see the underlying cause; a send-failure is logged separately.
func ReturnUserError(s *discordgo.Session, i *discordgo.InteractionCreate, userMsg string, err error) error {
	if sendErr := SendEphemeralError(s, i, userMsg); sendErr != nil {
		log.Printf("failed to send error response: %v (original: %v)", sendErr, err)
	}
	return err
}

// EditWithErrorMessage edits a deferred interaction response. Use after
// defer — SendEphemeralError 404s because the initial slot is consumed.
func EditWithErrorMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	_, err := s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Content: &message,
		},
	)
	return err
}

// ReturnUserErrorDeferred — deferred-interaction counterpart of ReturnUserError.
func ReturnUserErrorDeferred(s *discordgo.Session, i *discordgo.InteractionCreate, userMsg string, err error) error {
	if sendErr := EditWithErrorMessage(s, i, userMsg); sendErr != nil {
		log.Printf("failed to edit deferred response with error: %v (original: %v)", sendErr, err)
	}
	return err
}

// LogErrorsToErrorChannel: console log + summary embed + .txt attachment
// streamed in-memory (nothing hits disk).
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
