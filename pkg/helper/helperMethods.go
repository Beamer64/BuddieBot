package helper

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/rand"
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

func GetRandomStringFromSet(set []string) string {
	str := set[rand.Intn(len(set))]
	time.Sleep(1 * time.Millisecond)
	return str
}

func MemberHasRole(session *discordgo.Session, m *discordgo.Member, guildID string, roleName string) bool {
	if guildID == "" {
		guildID = m.GuildID
	}
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

// CheckIfStringEmpty Checks if the value is empty and returns it if not.
// Otherwise, return 'N/A'
func CheckIfStringEmpty(value string) string {
	if value != "" && value != " " {
		return value
	}
	return "N/A"
}

func StringInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func ToConvertedText(text string, convertGroup string) (string, error) {
	letters, err := getLetters()
	if err != nil {
		return "", err
	}

	convertedText := ""
	for _, char := range text {
		randSubs := ""
		subSet := letters[convertGroup][0][string(char)]
		if subSet != nil {
			randSubs = GetRandomStringFromSet(subSet)
		} else {
			randSubs = string(char)
		}
		convertedText += randSubs
	}

	return convertedText, nil

}

func getLetters() (map[string][]map[string][]string, error) {
	letters := make(map[string][]map[string][]string)

	fontsDir := "config_files/text_fonts.json"
	if IsLaunchedByDebugger() {
		fontsDir = "../../config_files/text_fonts.json"
	}

	jsonFile, err := os.Open(fontsDir)
	if err != nil {
		return nil, err
	}

	defer func(jsonFile *os.File) {
		err = jsonFile.Close()
	}(jsonFile)
	if err != nil {
		return nil, err
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &letters)
	if err != nil {
		return nil, err
	}

	return letters, nil
}
