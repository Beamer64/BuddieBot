package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/subosito/shorturl"
	"math/rand"
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

func ShortenURL(url string, provider string) (string, error) {
	u, err := shorturl.Shorten(url, provider)
	if err != nil {
		return "", err
	}
	return string(u), nil
}
