package prefix

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/bb_data/buddie"
	"github.com/bwmarrin/discordgo"
)

func sendReleaseNotes(s *discordgo.Session, m *discordgo.MessageCreate) error {
	embed := buildReleaseNotesEmbed()

	embed.Author.Name = m.Author.Username
	embed.Author.IconURL = m.Author.AvatarURL("")

	msg := &discordgo.MessageSend{
		Content: "@everyone",
		Embed:   embed,
	}

	// only send to test channel if just testing
	if helper.IsLaunchedByDebugger() {
		if _, err := s.ChannelMessageSendComplex(m.ChannelID, msg); err != nil {
			return fmt.Errorf("send release notes to channel %s: %w", m.ChannelID, err)
		}
	} else {
		for _, guild := range s.State.Guilds {
			for _, channel := range guild.Channels {
				if channel.Type == discordgo.ChannelTypeGuildText {
					if _, err := s.ChannelMessageSendComplex(channel.ID, msg); err != nil {
						return fmt.Errorf("send release notes to guild %s channel %s: %w", guild.ID, channel.ID, err)
					}
					break
				}
			}
		}
	}
	return nil
}

func sendGoodBoy(s *discordgo.Session, m *discordgo.MessageCreate) error {
	pic := buddie.Random()
	if pic.Filename == "" {
		return errors.New("buddie: no image available (bb_data.Load not called?)")
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Buddie Baxter!",
		Color:       helper.RandomDiscordColor(),
		Description: "For those who don't know, BuddieBot was named after this guy!",
		Image:       &discordgo.MessageEmbedImage{URL: "attachment://" + pic.Filename},
	}
	msg := &discordgo.MessageSend{
		Embed: embed,
		Files: []*discordgo.File{
			{
				Name:        pic.Filename,
				ContentType: contentTypeFor(pic.Filename),
				Reader:      bytes.NewReader(pic.Data),
			},
		},
	}
	if _, sendErr := s.ChannelMessageSendComplex(m.ChannelID, msg); sendErr != nil {
		return fmt.Errorf("send goodboy: %w", sendErr)
	}
	return nil
}

// contentTypeFor maps a filename's extension to a Discord-friendly MIME.
// The buddie/ set is JPEG today (.jpg + .jpeg both present); the switch
// covers .png and .gif too so future additions don't need a code change.
func contentTypeFor(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func sendWeasterEgg(s *discordgo.Session, m *discordgo.MessageCreate) error {
	_, err := s.ChannelMessageSend(
		m.ChannelID,
		"Is mayonnaise an instrument?\n───────────────▄████████▄────────\n──────────────██▒▒▒▒▒▒▒▒██───────\n─────────────██▒▒▒▒▒▒▒▒▒██───────\n────────────██▒▒▒▒▒▒▒▒▒▒██───────\n"+
			"───────────██▒▒▒▒▒▒▒▒▒██▀────────\n"+
			"──────────██▒▒▒▒▒▒▒▒▒▒██─────────\n─────────██▒▒▒▒▒▒▒▒▒▒▒██─────────\n────────██▒████▒████▒▒██─────────\n────────██▒▒▒▒▒▒▒▒▒▒▒▒██─────────\n────────██▒────▒▒────▒██─────────\n────────██▒─██─▒▒─██─▒██─────────\n────────██▒────▒▒────▒██─────────\n────────██▒▒▒▒▒▒▒▒▒▒▒▒██─────────\n───────██▒▒█▀▀▀▀▀▀▀█▒▒▒▒██───────\n─────██▒▒▒▒▒█▄▄▄▄▄█▒▒▒▒▒▒▒██─────\n───██▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒██───\n─██▒▒▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒▒▒██─\n█▒▒▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒▒▒█\n█▒▒▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒▒▒█\n█▒▒████▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒████▒▒█\n▀████▒▒▒▒▒▒▒▒▒▓▓▓▓▒▒▒▒▒▒▒▒▒▒████▀\n──█▌▌▌▌▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▌▌▌███──\n───█▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌█────\n───█▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌█────\n────▀█▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌██▀─────\n─────█▌▌▌▌▌▌████████▌▌▌▌▌██──────\n──────██▒▒██────────██▒▒██───────\n──────▀████▀────────▀████▀───────",
	)
	return err
}

func checkPalindrome(s *discordgo.Session, m *discordgo.MessageCreate, str string) error {
	// Runes, not bytes — multi-byte characters (emoji, accents) must compare whole.
	runes := []rune(str)
	isPalindrome := true

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		if runes[i] != runes[j] {
			isPalindrome = false
			break
		}
	}

	msg := "No is palindrome 👎"
	if isPalindrome {
		msg = "Is palindrome 👍"
	}

	_, err := s.ChannelMessageSend(m.ChannelID, msg)
	return err
}

func romanNums(s *discordgo.Session, m *discordgo.MessageCreate, str string) error {
	if intVal, err := strconv.Atoi(str); err == nil {
		romanLetters := []struct {
			value   int
			letters string
		}{
			{1000, "M"},
			{900, "CM"},
			{500, "D"},
			{400, "CD"},
			{100, "C"},
			{90, "XC"},
			{50, "L"},
			{40, "XL"},
			{10, "X"},
			{9, "IX"},
			{5, "V"},
			{4, "IV"},
			{1, "I"},
		}

		roman := ""
		for _, v := range romanLetters {
			for intVal >= v.value {
				roman += v.letters
				intVal -= v.value
			}
		}

		content := fmt.Sprintf("%s as roman value: %v", str, roman)
		_, err = s.ChannelMessageSend(m.ChannelID, content)
		if err != nil {
			return err
		}

	} else if errors.Is(err, strconv.ErrSyntax) {
		str = strings.ToUpper(str)
		strUp := str
		romanValues := map[rune]int{
			'I': 1,
			'V': 5,
			'X': 10,
			'L': 50,
			'C': 100,
			'D': 500,
			'M': 1000,
		}

		// Expand subtraction pairs (CM/CD/XC/XL/IX/IV) so a sum works.
		replacer := strings.NewReplacer("CM", "CCCCCCCCC", "CD", "CCCC", "XC", "XXXXXXXXX", "XL", "XXXX", "IX", "IIIIIIIII", "IV", "IIII")
		str = replacer.Replace(str)

		total := 0
		for _, v := range str {
			total += romanValues[v]
		}

		content := fmt.Sprintf("%s as numeric value: %v", strUp, total)
		_, err = s.ChannelMessageSend(m.ChannelID, content)
		if err != nil {
			return err
		}

	} else {
		return err
	}

	return nil
}
