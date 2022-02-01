package events

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/subosito/shorturl"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func getErrorEmbed(err error) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "ERROR",
		Description: "(ノಠ益ಠ)ノ彡┻━┻",
		Color:       16726843,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Stack",
				Value:  fmt.Sprintf("%+v", errors.WithStack(err)),
				Inline: true,
			},
		},
	}

	return embed
}

// IsLaunchedByDebugger Determines if application is being run by the debugger.
func IsLaunchedByDebugger() bool {
	// gops executable must be in the path. See https://github.com/google/gops
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	if err == nil && strings.Contains(string(gopsOut), "\\dlv.exe") {
		// our parent process is (probably) the Delve debugger
		return true
	}
	return false
}

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
