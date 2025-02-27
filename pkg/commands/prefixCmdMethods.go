package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/StephaneBunel/bresenham"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/beamer64/buddieBot/pkg/voice_chat"
	"github.com/beamer64/buddieBot/pkg/web"
	"github.com/bwmarrin/discordgo"
	"github.com/subosito/shorturl"
	"image"
	"image/color"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// functions here should mostly be used for the prefix commands ($)

// region dev commands
func testMethod(s *discordgo.Session, m *discordgo.MessageCreate, param string) error {
	if helper.IsLaunchedByDebugger() {
		err := playAudioLink(s, m, param)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendReleaseNotes(s *discordgo.Session, m *discordgo.MessageCreate) error {
	embed := releaseNotesEmbed

	embed.Author.Name = m.Author.Username
	embed.Author.IconURL = m.Author.AvatarURL("")

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

// endregion dev commands

func sendLmgtfy(s *discordgo.Session, m *discordgo.Message) error {
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
func sendStartUpMessages(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs) error {
	// sleep for 1 minute while saying funny things and to wait for instance to start up
	sm := 0
	for i := 1; i < 5; i++ {
		loadingMessage := helper.GetRandomStringFromSet(cfg.LoadingMessages)
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
func playAudioLink(s *discordgo.Session, m *discordgo.MessageCreate, link string) error {
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

func stopAudioPlayback() error {
	// vc := voice_chat.VoiceConnection{}

	if web.StopPlaying != nil {
		close(web.StopPlaying)
		web.IsPlaying = false

		/*if vc.Dgv != nil {
			vc.Dgv.Close()

		}*/
	}

	return nil
}

func sendQueue(s *discordgo.Session, m *discordgo.MessageCreate) error {
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

func sendSkipMessage(s *discordgo.Session, m *discordgo.MessageCreate) error {
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

	err = skipPlayback(s, m)
	if err != nil {
		return err
	}

	return nil
}

func skipPlayback(s *discordgo.Session, m *discordgo.MessageCreate) error {
	if len(web.MpFileQueue) > 0 {
		err := stopAudioPlayback()
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

// endregion audio commands

// region misc

func sendCistercianNumeral(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs, param string) error {
	posNum, hasPrefix := strings.CutPrefix(param, "-")

	// check if the param is a number
	if intNum, err := strconv.Atoi(posNum); err == nil {
		if intNum >= -9999 && intNum <= 9999 {

			img, err := drawCistLines(hasPrefix, posNum)
			if err != nil {
				return err
			}

			imgPath := "../../res/genFiles/symbol.png"
			err = helper.CreateImgFile(imgPath, img)
			if err != nil {
				return err
			}

			imgURL, err := helper.GetImgbbUploadURL(cfg, imgPath, 10)
			if err != nil {
				return err
			}

			embed := &discordgo.MessageEmbed{
				Title: fmt.Sprintf("Cistercian Numeral for %v", intNum),
				Color: helper.RangeIn(1, 16777215),
				Image: &discordgo.MessageEmbedImage{
					URL: imgURL,
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "https://en.wikipedia.org/wiki/Cistercian_numerals",
				},
			}

			_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
			if err != nil {
				return err
			}

		} else {
			_, err = s.ChannelMessageSend(m.ChannelID, "Please enter a number from -9999 to 9999")
			if err != nil {
				return err
			}
		}
	} else {
		_, err = s.ChannelMessageSend(m.ChannelID, "Please enter a positive or negative number only")
		if err != nil {
			return err
		}
	}

	return nil
}

func drawCistLines(hasPrefix bool, posNum string) (image.Image, error) {
	var imgRect = image.Rect(0, 0, 200, 200)
	var img = image.NewRGBA(imgRect)
	r := helper.RangeIn(0, 255)
	g := helper.RangeIn(0, 255)
	b := helper.RangeIn(0, 255)
	var col = color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}

	if hasPrefix {
		// draw horizontal line
		bresenham.DrawLine(img, 60, 100, 140, 100, col)
	}

	// draw vertical line
	bresenham.DrawLine(img, 100, 20, 100, 180, col)

	var x1 int
	var x2 int
	var y1 int
	var y2 int
	for pos, char := range posNum {
		// fmt.Printf("character %c starts at byte position %d\n", char, pos)
		switch pos {
		case 0: // thous
			switch char {
			case '5':
				bresenham.DrawLine(img, 60, 180, 100, 140, col)
			case '7':
				bresenham.DrawLine(img, 60, 180, 60, 140, col)
			case '8':
				bresenham.DrawLine(img, 60, 140, 60, 180, col)
			case '9':
				bresenham.DrawLine(img, 60, 180, 60, 140, col)
				bresenham.DrawLine(img, 60, 140, 100, 140, col)
			}

			x1 = thous[string(char)].x1
			y1 = thous[string(char)].y1
			x2 = thous[string(char)].x2
			y2 = thous[string(char)].y2
		case 1: // hunds
			switch char {
			case '5':
				bresenham.DrawLine(img, 140, 180, 100, 140, col)
			case '7':
				bresenham.DrawLine(img, 140, 180, 140, 140, col)
			case '8':
				bresenham.DrawLine(img, 140, 140, 140, 180, col)
			case '9':
				bresenham.DrawLine(img, 140, 180, 140, 140, col)
				bresenham.DrawLine(img, 140, 140, 100, 140, col)
			}

			x1 = hunds[string(char)].x1
			y1 = hunds[string(char)].y1
			x2 = hunds[string(char)].x2
			y2 = hunds[string(char)].y2
		case 2: // tens
			switch char {
			case '5':
				bresenham.DrawLine(img, 60, 20, 100, 60, col)
			case '7':
				bresenham.DrawLine(img, 60, 20, 60, 60, col)
			case '8':
				bresenham.DrawLine(img, 60, 60, 60, 20, col)
			case '9':
				bresenham.DrawLine(img, 60, 20, 60, 60, col)
				bresenham.DrawLine(img, 60, 60, 100, 60, col)
			}

			x1 = tens[string(char)].x1
			y1 = tens[string(char)].y1
			x2 = tens[string(char)].x2
			y2 = tens[string(char)].y2
		case 3: // ones
			switch char {
			case '5':
				bresenham.DrawLine(img, 100, 60, 140, 20, col)
			case '7':
				bresenham.DrawLine(img, 140, 20, 140, 60, col)
			case '8':
				bresenham.DrawLine(img, 140, 60, 140, 20, col)
			case '9':
				bresenham.DrawLine(img, 140, 20, 140, 60, col)
				bresenham.DrawLine(img, 140, 60, 100, 60, col)
			}

			x1 = ones[string(char)].x1
			y1 = ones[string(char)].y1
			x2 = ones[string(char)].x2
			y2 = ones[string(char)].y2
		}

		bresenham.DrawLine(img, x1, y1, x2, y2, col)
	}

	return img, nil
}

func sendWeasterEgg(s *discordgo.Session, m *discordgo.MessageCreate) error {
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

func checkPalindrome(s *discordgo.Session, m *discordgo.MessageCreate, str string) error {
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

// endregion misc

// region moderation
func modNSFWimgs(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs) error {
	if cfg.Configs.Settings.EnableNSFWModeration {
		if m.Attachments != nil && len(m.Attachments) > 0 {
			for _, attachment := range m.Attachments {
				/*_, err := s.ChannelMessageSend(m.ChannelID, attachment.URL)
				if err != nil {
					return err
				}*/
				fmt.Println(attachment.URL)
			}

		} else {
			/*_, err := s.ChannelMessageSend(m.ChannelID, "There is no attachment")
			if err != nil {
				return nil
			}*/
		}
	}
	return nil
}
