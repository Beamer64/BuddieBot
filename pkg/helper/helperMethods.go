package helper

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var ApprovalWords = []string{
	"enabled",
	"on",
	"true",
	"yes",
	"sure",
}

var DisapprovalWords = []string{
	"disabled",
	"off",
	"false",
	"no",
	"nope",
}

func GetErrorEmbed(err error, s *discordgo.Session, gID string) *discordgo.MessageEmbed {
	var guild *discordgo.Guild
	guildID := "N/A"
	guildName := "N/A"

	if gID != "" {
		guild, _ = s.Guild(gID)
		guildID = gID
		guildName = guild.Name
	}

	embed := &discordgo.MessageEmbed{
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
				Name:   "Stack",
				Value:  fmt.Sprintf("%+v", errors.WithStack(err)),
				Inline: false,
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

func GetRandomLoadingMessage(possibleMessages []string) string {
	rand.Seed(time.Now().Unix())
	return possibleMessages[rand.Intn(len(possibleMessages))]
}

// GetGuildMembers Discordgo and the discord api are broken atm so niether will get member list
func GetGuildMembers(guildID string, cfg *config.Configs) ([]*discordgo.Member, error) {
	token := ""
	if IsLaunchedByDebugger() {
		token = cfg.Configs.Keys.TestBotToken
	} else {
		token = cfg.Configs.Keys.ProdBotToken
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://discord.com/api/guilds/%s/members", guildID), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	req.Header.Add("Authorization", "Bot "+token)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var memberObj []*discordgo.Member

	err = json.NewDecoder(res.Body).Decode(&memberObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	return memberObj, nil
}

func MemberHasRole(session *discordgo.Session, m *discordgo.Member, roleName string) bool {
	guildID := m.GuildID
	roleName = strings.ToLower(roleName)

	for _, roleID := range m.Roles {
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

// RangeIn Returns pseudo rand num between low and high.
// For random embed color: rangeIn(1, 16777215)
func RangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

// CheckIfEmpty Checks if the value is empty and returns it if not.
// Otherwise, return 'N/A'
func CheckIfEmpty(value string) string {
	if value != "" {
		return value
	}
	return "N/A"
}

func StringInSlice(a string, list []string) bool {
	for _, v := range list {
		if v == a {
			return true
		}
	}
	return false
}
