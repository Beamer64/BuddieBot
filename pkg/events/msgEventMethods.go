package events

import (
	"fmt"
	"net/url"

	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func (r *ReactionHandler) sendLmgtfy(s *discordgo.Session, m *discordgo.Message) error {
	if m.Author.Bot {
		return nil
	}

	strEnc := url.QueryEscape(m.Content)
	lmgtfyURL := fmt.Sprintf("http://lmgtfy2.com/?q=%s", strEnc)

	msgContent := fmt.Sprintf("\"%s\"\n\n%s", m.Content, lmgtfyURL)
	embed := &discordgo.MessageEmbed{
		Color:       helper.RandomDiscordColor(),
		Title:       "Let me Google that for you..",
		Description: msgContent,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "LMGTFY",
		},
	}

	msg, err := s.ChannelMessageSendComplex(
		m.ChannelID, &discordgo.MessageSend{
			Content: fmt.Sprintf("<@!%s>", m.Author.ID),
			Embeds:  []*discordgo.MessageEmbed{embed},
		},
	)

	/*msg, err := s.ChannelMessageSend(m.ChannelID, msgContent)
	if err != nil {
		return err
	}*/

	emojiID := helper.ProdLmgtfyEmojiID
	if helper.IsLaunchedByDebugger() {
		emojiID = helper.TestLmgtfyEmojiID
	}

	err = s.MessageReactionAdd(m.ChannelID, msg.ID, emojiID)
	if err != nil {
		return err
	}

	return nil
}
