package events

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/subosito/shorturl"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

func getRandomLoadingMessage(possibleMessages []string) string {
	rand.Seed(time.Now().Unix())
	return possibleMessages[rand.Intn(len(possibleMessages))]
}

func GetGuildMembers(session *discordgo.Session, guildID string) ([]*discordgo.Member, error) {
	guild, err := session.State.Guild(guildID)
	if err != nil {
		return nil, err
	}

	return guild.Members, nil
}

func ShortenURL(url string) (string, error) {
	u, err := shorturl.Shorten(url, "tinyurl")
	if err != nil {
		return "", err
	}
	return string(u), nil
}

func CreateLmgtfyURL(s string) string {
	strEnc := url.QueryEscape(s)
	lmgtfyString := "http://lmgtfy.com/?q=" + strEnc
	return lmgtfyString
}

func (d *MessageCreateHandler) memberHasRole(session *discordgo.Session, message *discordgo.MessageCreate, roleName string) bool {
	guildID := message.GuildID
	roleName = strings.ToLower(roleName)

	for _, roleID := range message.Member.Roles {
		role, err := session.State.Role(guildID, roleID)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
		}

		if strings.ToLower(role.Name) == roleName {
			return true
		}
	}
	return false
}
