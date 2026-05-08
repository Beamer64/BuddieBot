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

func SendResponseErrorToUser(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
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
	if sendErr := SendResponseErrorToUser(s, i, userMsg); sendErr != nil {
		log.Printf("failed to send error response: %v (original: %v)", sendErr, err)
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
		Files: []*discordgo.File{{
			Name:        fmt.Sprintf("error-%s.txt", time.Now().Format("20060102-150405")),
			ContentType: "text/plain",
			Reader:      strings.NewReader(fullStack),
		}},
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
			log.Printf("%+v", errors.WithStack(err))
		}

		if strings.ToLower(role.Name) == roleName {
			return true
		}
	}
	return false
}
