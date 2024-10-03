package events

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/subosito/shorturl"
	"net/url"
)

// functions here should be used for the events handled

func (r *ReactionHandler) sendLmgtfy(s *discordgo.Session, m *discordgo.Message) error {
	strEnc := url.QueryEscape(m.Content)
	lmgtfyURL := fmt.Sprintf("http://lmgtfy.com/?q=%s", strEnc)

	lmgtfyShortURL, err := shorturl.Shorten(lmgtfyURL, "tinyurl")
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("\"%s\"\n%s", m.Content, string(lmgtfyShortURL)))
	if err != nil {
		return err
	}

	return nil
}
