package prefix

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/BuddieBot/pkg/voice_chat"
	"github.com/StephaneBunel/bresenham"
	"github.com/bwmarrin/discordgo"
)

// functions here should mostly be used for the prefix commands ($)

// region dev commands
func testMethod(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs, param string) error {
	if helper.IsLaunchedByDebugger() {
		err := playAudioLink(s, m, cfg, param)
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

// region audio commands

func playAudioLink(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs, link string) error {
	urls := strings.Fields(link)
	if len(urls) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: $play <YouTube URL> [more URLs…]")
		return err
	}

	if cfg.Player == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Audio is not available right now.")
		return err
	}

	if len(urls) == 1 {
		return playSingle(s, m, cfg, urls[0])
	}
	return playBatch(s, m, cfg, urls)
}

// playSingle handles the original one-URL case. Keeps the historical
// status messages ("Now playing: X" / "Added to queue: X (position N)").
func playSingle(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs, url string) error {
	status, err := s.ChannelMessageSend(m.ChannelID, "Resolving audio…")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := cfg.Player.Play(ctx, m.GuildID, m.ChannelID, m.Author.ID, url)
	if err != nil {
		_, editErr := s.ChannelMessageEdit(m.ChannelID, status.ID, friendlyPlayError(err))
		if editErr != nil {
			return fmt.Errorf("play: %w (also: edit status: %v)", err, editErr)
		}
		if isUserFacingPlayError(err) {
			return nil
		}
		return err
	}

	var finalMsg string
	if result.Queued {
		finalMsg = fmt.Sprintf("Added to queue: %s (position %d)", result.Title, result.Position)
	} else {
		finalMsg = "Now playing: " + result.Title
	}
	_, err = s.ChannelMessageEdit(m.ChannelID, status.ID, finalMsg)
	return err
}

// playBatch handles 2+ URLs in one command. Loops Play sequentially —
// the first call sets up voice if nothing's playing; subsequent calls
// see "already playing" inside Play and get queued. Bails out on errors
// that would affect all remaining URLs (no voice channel, voice timeout,
// queue full); skips past per-URL errors (unresolvable links).
func playBatch(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs, urls []string) error {
	status, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Resolving %d URLs…", len(urls)))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	type queuedItem struct {
		title    string
		position int
	}

	var (
		playingTitle string
		queued       []queuedItem
		failures     int
		fatalMsg     string // set if a bail-out error stopped the batch
	)

	for _, url := range urls {
		result, err := cfg.Player.Play(ctx, m.GuildID, m.ChannelID, m.Author.ID, url)
		if err == nil {
			if result.Queued {
				queued = append(queued, queuedItem{title: result.Title, position: result.Position})
			} else {
				playingTitle = result.Title
			}
			continue
		}

		// Errors that would also fail every remaining URL — stop here
		// and surface a single message rather than spamming.
		if errors.Is(err, voice_chat.ErrNotInVoice) ||
			errors.Is(err, voice_chat.ErrVoiceTimeout) ||
			errors.Is(err, voice_chat.ErrQueueFull) {
			fatalMsg = friendlyPlayError(err)
			break
		}

		// Per-URL failure (most commonly ErrNoTrackFound) — skip it,
		// keep processing the rest.
		failures++
	}

	var msg strings.Builder
	if playingTitle != "" {
		msg.WriteString("Now playing: ")
		msg.WriteString(playingTitle)
		msg.WriteString("\n")
	}
	if len(queued) == 1 {
		fmt.Fprintf(&msg, "Added to queue: %s (position %d)\n", queued[0].title, queued[0].position)
	} else if len(queued) > 1 {
		msg.WriteString("Added to queue:\n")
		for i, item := range queued {
			line := fmt.Sprintf("  %d. %s\n", item.position, item.title)
			// Discord caps messages at 2000 chars; leave room for a "...and N more" line.
			if msg.Len()+len(line) > 1900 {
				fmt.Fprintf(&msg, "  …and %d more\n", len(queued)-i)
				break
			}
			msg.WriteString(line)
		}
	}
	if failures > 0 {
		fmt.Fprintf(&msg, "Couldn't resolve %d URL%s.\n", failures, pluralS(failures))
	}
	if fatalMsg != "" {
		msg.WriteString(fatalMsg)
		msg.WriteString("\n")
	}
	if msg.Len() == 0 {
		msg.WriteString("Nothing happened.")
	}

	_, err = s.ChannelMessageEdit(m.ChannelID, status.ID, strings.TrimRight(msg.String(), "\n"))
	return err
}

// friendlyPlayError maps Player.Play errors to user-facing messages.
func friendlyPlayError(err error) string {
	switch {
	case errors.Is(err, voice_chat.ErrNotInVoice):
		return "Join a voice channel first."
	case errors.Is(err, voice_chat.ErrNoTrackFound):
		return "Couldn't resolve audio (invalid URL or unavailable video)."
	case errors.Is(err, voice_chat.ErrQueueFull):
		return "Queue is full (100 tracks max)."
	case errors.Is(err, voice_chat.ErrVoiceTimeout):
		return "Voice connection didn't establish — try again."
	default:
		return "Failed to start playback."
	}
}

// isUserFacingPlayError reports whether an error has already been shown
// to the user via friendlyPlayError, so the caller can suppress the
// error-channel log to avoid double-reporting normal user mistakes.
func isUserFacingPlayError(err error) bool {
	return errors.Is(err, voice_chat.ErrNotInVoice) ||
		errors.Is(err, voice_chat.ErrNoTrackFound) ||
		errors.Is(err, voice_chat.ErrQueueFull) ||
		errors.Is(err, voice_chat.ErrVoiceTimeout)
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func stopAudioPlayback(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs) error {
	if cfg.Player == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Audio is not available right now.")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := cfg.Player.Stop(ctx, m.GuildID); err != nil {
		return fmt.Errorf("stop playback: %w", err)
	}
	_, err := s.ChannelMessageSend(m.ChannelID, "Stopped.")
	return err
}

func sendQueue(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs) error {
	if cfg.Player == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Audio is not available right now.")
		return err
	}

	current, upcoming, err := cfg.Player.Queue(m.GuildID)
	if err != nil {
		return fmt.Errorf("queue: %w", err)
	}

	if current == nil && len(upcoming) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Nothing playing.")
		return err
	}

	var msg strings.Builder
	if current != nil {
		msg.WriteString("Now playing: ")
		msg.WriteString(current.Info.Title)
		msg.WriteString("\n")
	}
	if len(upcoming) > 0 {
		msg.WriteString("Up next:\n")
		for i, t := range upcoming {
			line := fmt.Sprintf("  %d. %s\n", i+1, t.Info.Title)
			// Discord caps messages at 2000 chars; leave room for a "...and N more" line.
			if msg.Len()+len(line) > 1900 {
				msg.WriteString(fmt.Sprintf("  …and %d more\n", len(upcoming)-i))
				break
			}
			msg.WriteString(line)
		}
	}
	_, err = s.ChannelMessageSend(m.ChannelID, msg.String())
	return err
}

func skipPlayback(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs) error {
	if cfg.Player == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Audio is not available right now.")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	skipped, next, err := cfg.Player.Skip(ctx, m.GuildID)
	if err != nil {
		if errors.Is(err, voice_chat.ErrNothingPlaying) {
			_, sErr := s.ChannelMessageSend(m.ChannelID, "Nothing is playing.")
			return sErr
		}
		return fmt.Errorf("skip: %w", err)
	}

	var msg string
	if next != nil {
		msg = fmt.Sprintf("Skipped: %s.\nNow playing: %s", skipped.Info.Title, next.Info.Title)
	} else {
		msg = fmt.Sprintf("Skipped: %s. Queue is empty — leaving voice.", skipped.Info.Title)
	}
	_, err = s.ChannelMessageSend(m.ChannelID, msg)
	return err
}

func clearQueue(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs) error {
	if cfg.Player == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Audio is not available right now.")
		return err
	}

	count, err := cfg.Player.ClearQueue(m.GuildID)
	if err != nil {
		return fmt.Errorf("clear queue: %w", err)
	}

	var msg string
	switch {
	case count == 0:
		msg = "Queue is already empty."
	case count == 1:
		msg = "Cleared 1 queued track."
	default:
		msg = fmt.Sprintf("Cleared %d queued tracks.", count)
	}
	_, err = s.ChannelMessageSend(m.ChannelID, msg)
	return err
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
				Color: helper.RandomDiscordColor(),
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
	return err
}

func checkPalindrome(s *discordgo.Session, m *discordgo.MessageCreate, str string) error {
	// Convert to runes so multi-byte characters (emoji, accented letters, etc.) are handled correctly
	runes := []rune(str)
	isPalindrome := true

	// Compare from both ends moving inward — only need to check half the string
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

		// convert the subtraction instances into their full value.
		// 900, 400, 90, 40, 9, 4
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

// endregion misc
