package events

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
