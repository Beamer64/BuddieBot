package events

import (
	"context"
	"errors"
	"fmt"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/beamer64/buddieBot/pkg/voice_chat"
	"github.com/beamer64/buddieBot/pkg/web"
	"github.com/bwmarrin/discordgo"
	"github.com/subosito/shorturl"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// functions here should mostly be used for the prefix commands ($)

// region dev commands
func (d *MessageCreateHandler) testMethod(s *discordgo.Session, m *discordgo.MessageCreate, param string) error {
	if helper.IsLaunchedByDebugger() {
		err := d.playAudioLink(s, m, param)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *MessageCreateHandler) sendReleaseNotes(s *discordgo.Session, m *discordgo.MessageCreate) error {
	embed := &discordgo.MessageEmbed{
		Title: "Release Notes!",
		URL:   "https://github.com/Beamer64/BuddieBot/blob/master/res/release.md",
		Description: "SUM MOAR BIG BOI CHANGES\n\nDetailed list can be found in the Title link above." +
			"\nCheck it out\n-----------------------------------------------------------------------------\n\n- Command changes:",
		Color: 11091696,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    m.Author.Username,
			IconURL: m.Author.AvatarURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "New Command Group: /ratethis",
				Value:  "Give/Get some new ratings",
				Inline: false,
			},
			{
				Name:   "New Commands: /pick album",
				Value:  "<@!282722418093719556>'s Album recommender api. Recommends a music album based on liked tags.",
				Inline: false,
			},
			{
				Name:   "New Commands: /pick poll",
				Value:  "Poll command...for polling things..",
				Inline: false,
			},
			{
				Name:   "New Commands: ${COMMAND} SpongeBob easter egg",
				Value:  "It's my bot, I can do what I want.",
				Inline: false,
			},
			{
				Name:   "Bug Fix: Youtube mobile links",
				Value:  "(When working..) Audio will play with the mobile link 'm.youtube.com...'",
				Inline: false,
			},
			{
				Name:   "Enhancement: Audio Queue",
				Value:  "The Audio Queue will show a cleaned title from the old 'Name-Title-Sum_Numbers.mp3'.",
				Inline: false,
			},
		},
	}

	msg := &discordgo.MessageSend{
		Content: "@everyone",
		Embed:   embed,
	}

	if helper.IsLaunchedByDebugger() {
		_, err := s.ChannelMessageSendComplex(m.ChannelID, msg)
		if err != nil {
			return err
		}
	} else {
		for _, guild := range s.State.Guilds {
			for _, channel := range guild.Channels {
				if channel.Type == discordgo.ChannelTypeGuildText {
					_, err := s.ChannelMessageSendComplex(channel.ID, msg)
					if err != nil {
						return err
					}
					break
				}
			}
		}
	}
	return nil
}

//endregion

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

// sendStartUpMessages is used when spinning up servers for minecraft for example
func (d *MessageCreateHandler) sendStartUpMessages(s *discordgo.Session, m *discordgo.MessageCreate) error {
	// sleep for 1 minute while saying funny things and to wait for instance to start up
	sm := 0
	for i := 1; i < 5; i++ {
		loadingMessage := helper.GetRandomStringFromSet(d.cfg.LoadingMessages)
		time.Sleep(3 * time.Second)

		_, err := s.ChannelMessageSend(m.ChannelID, loadingMessage)
		if err != nil {
			return err
		}

		sm += i
	}
	time.Sleep(3 * time.Second)
	return nil
}

// region audio commands
func (d *MessageCreateHandler) playAudioLink(s *discordgo.Session, m *discordgo.MessageCreate, link string) error {
	msg, err := s.ChannelMessageSend(m.ChannelID, "Prepping vidya...")
	if err != nil {
		return err
	}

	link, fileName, err := web.GetYtAudioLink(s, msg, link)
	if err != nil {
		// if context timed out because no link found
		if errors.Is(err, context.DeadlineExceeded) {
			_, err = s.ChannelMessageEdit(m.ChannelID, msg.ID, "Audio Unavailable..")
			if err != nil {
				return err
			}
			err = nil
		}
		return err
	}

	err = web.DownloadMpFile(m, link, fileName)
	if err != nil {
		return err
	}

	dgv, err := voice_chat.ConnectVoiceChannel(s, m.Author.ID, m.GuildID)
	if err != nil {
		return err
	}

	err = web.PlayAudioFile(dgv, fileName, m, s)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) stopAudioPlayback() error {
	//vc := voice_chat.VoiceConnection{}

	if web.StopPlaying != nil {
		close(web.StopPlaying)
		web.IsPlaying = false

		/*if vc.Dgv != nil {
			vc.Dgv.Close()

		}*/
	}

	return nil
}

func (d *MessageCreateHandler) sendQueue(s *discordgo.Session, m *discordgo.MessageCreate) error {
	queue := ""
	if len(web.MpFileQueue) > 0 {
		queue = strings.Join(web.MpFileQueue, "\n")
	} else {
		queue = "Uh owh, song queue is wempty (>.<)"
	}

	_, err := s.ChannelMessageSend(m.ChannelID, queue)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) sendSkipMessage(s *discordgo.Session, m *discordgo.MessageCreate) error {
	audio := ""
	if len(web.MpFileQueue) > 0 {
		audio = fmt.Sprintf("Skipping %s", web.MpFileQueue[0])
	} else {
		audio = "Queue is empty, my guy"
	}

	_, err := s.ChannelMessageSend(m.ChannelID, audio)
	if err != nil {
		return err
	}

	err = d.skipPlayback(s, m)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) skipPlayback(s *discordgo.Session, m *discordgo.MessageCreate) error {
	if len(web.MpFileQueue) > 0 {
		err := d.stopAudioPlayback()
		if err != nil {
			return err
		}

		dgv, err := voice_chat.ConnectVoiceChannel(s, m.Author.ID, m.GuildID)
		if err != nil {
			return err
		}

		err = web.PlayAudioFile(dgv, "", m, s)
		if err != nil {
			return err
		}
	}

	return nil
}

//endregion

// region misc
func (d *MessageCreateHandler) sendWeasterEgg(s *discordgo.Session, m *discordgo.MessageCreate) error {
	_, err := s.ChannelMessageSend(
		m.ChannelID,
		"Is mayonnaise an instrument?\n───────────────▄████████▄────────\n──────────────██▒▒▒▒▒▒▒▒██───────\n─────────────██▒▒▒▒▒▒▒▒▒██───────\n────────────██▒▒▒▒▒▒▒▒▒▒██───────\n"+
			"───────────██▒▒▒▒▒▒▒▒▒██▀────────\n"+
			"──────────██▒▒▒▒▒▒▒▒▒▒██─────────\n─────────██▒▒▒▒▒▒▒▒▒▒▒██─────────\n────────██▒████▒████▒▒██─────────\n────────██▒▒▒▒▒▒▒▒▒▒▒▒██─────────\n────────██▒────▒▒────▒██─────────\n────────██▒─██─▒▒─██─▒██─────────\n────────██▒────▒▒────▒██─────────\n────────██▒▒▒▒▒▒▒▒▒▒▒▒██─────────\n───────██▒▒█▀▀▀▀▀▀▀█▒▒▒▒██───────\n─────██▒▒▒▒▒█▄▄▄▄▄█▒▒▒▒▒▒▒██─────\n───██▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒██───\n─██▒▒▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒▒▒██─\n█▒▒▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒▒▒█\n█▒▒▒▒██▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒██▒▒▒▒█\n█▒▒████▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒████▒▒█\n▀████▒▒▒▒▒▒▒▒▒▓▓▓▓▒▒▒▒▒▒▒▒▒▒████▀\n──█▌▌▌▌▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▌▌▌███──\n───█▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌█────\n───█▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌█────\n────▀█▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌▌██▀─────\n─────█▌▌▌▌▌▌████████▌▌▌▌▌██──────\n──────██▒▒██────────██▒▒██───────\n──────▀████▀────────▀████▀───────",
	)
	if err != nil {
		return err
	}

	return nil
}

func (d *MessageCreateHandler) checkPalindrome(s *discordgo.Session, m *discordgo.MessageCreate, str string) error {
	revStr := ""
	isPalindrome := true
	for i := len(str) - 1; i >= 0; i-- {
		revStr += string(str[i])
	}
	for i := range str {
		if str[i] != revStr[i] {
			isPalindrome = false
			break
		}
	}

	if isPalindrome {
		_, err := s.ChannelMessageSend(m.ChannelID, "Is palindrome 👍")
		if err != nil {
			return err
		}
	} else {
		_, err := s.ChannelMessageSend(m.ChannelID, "No is palindrome 👎")
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *MessageCreateHandler) romanNums(s *discordgo.Session, m *discordgo.MessageCreate, str string) error {
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
		romanNums := map[rune]int{
			'I': 1,
			'V': 5,
			'X': 10,
			'L': 50,
			'C': 100,
			'D': 500,
			'M': 1000,
		}

		// convert the subtraction instances into their full value.
		// 900, 400, 90, 40, 9, 4
		replacer := strings.NewReplacer("CM", "CCCCCCCCC", "CD", "CCCC", "XC", "XXXXXXXXX", "XL", "XXXX", "IX", "IIIIIIIII", "IV", "IIII")
		str = replacer.Replace(str)

		total := 0
		for _, v := range str {
			total += romanNums[v]
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

//endregion
