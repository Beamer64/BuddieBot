package webScrape

import (
	"github.com/bwmarrin/discordgo"
	"github.com/subosito/shorturl"
	"net/url"
)

func LmgtfyURL(s string) string {
	strEnc := url.QueryEscape(s)
	lmgtfyString := "http://lmgtfy.com/?q=" + strEnc
	return lmgtfyString
}

func ShortenURL(url string, provider string) (string, error) {
	u, err := shorturl.Shorten(url, provider)
	if err != nil {
		return "", err
	}
	return string(u), nil
}

func FindLMGTFY(session *discordgo.Session, message *discordgo.MessageCreate, botID string) (*discordgo.Message, error) {
	msgs, err := session.ChannelMessages(message.ChannelID, 10, message.ID, "", "")
	if err != nil {
		return nil, err
	}

	for _, m := range msgs {
		for _, r := range m.Reactions {
			if message.Author.ID != botID {
				if r.Emoji.Name == "lmgtfy" {
					return m, nil
				}
			}
		}
	}
	return nil, nil
}
