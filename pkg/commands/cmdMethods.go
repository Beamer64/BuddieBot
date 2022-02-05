package commands

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/discordBot/pkg/api"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/voice_chat"
	"github.com/beamer64/discordBot/pkg/web"
	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
	"github.com/pkg/errors"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// functions here should mostly be used for the slash commands

//region Utility Commands

func sendTuuckResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	n := reflect.ValueOf(&cfg.Cmd.SlashName).Elem()
	d := reflect.ValueOf(&cfg.Cmd.Desc).Elem()
	e := reflect.ValueOf(&cfg.Cmd.Example).Elem()

	if len(i.ApplicationCommandData().Options) > 0 {
		cmdOption := strings.ToLower(i.ApplicationCommandData().Options[0].StringValue())
		slashCmd := ""
		cmdName := ""
		if strings.Contains(cmdOption, "/") {
			slashCmd = cmdOption
			cmdName = strings.ReplaceAll(slashCmd, "/", "")
		} else {
			slashCmd = fmt.Sprintf("/%s", cmdOption)
			cmdName = cmdOption
		}

		title := ""
		for t := 0; t < n.NumField(); t++ {
			if strings.Contains(fmt.Sprintf("%s", n.Field(t).Interface()), cmdName) {
				title = fmt.Sprintf("%s info", n.Field(t).Interface())
				break
			}
		}

		desc := ""
		for de := 0; de < d.NumField(); de++ {
			cmdDesc := strings.ReplaceAll(cmdName, " ", "")
			lowerDesc := strings.ToLower(d.Type().Field(de).Name)
			if strings.Contains(lowerDesc, cmdDesc) {
				desc = fmt.Sprintf("%s", d.Field(de).Interface())
				break
			}
		}

		example := ""
		for ex := 0; ex < e.NumField(); ex++ {
			cmdExample := strings.ReplaceAll(cmdName, " ", "")
			lowerExample := strings.ToLower(e.Type().Field(ex).Name)
			if strings.Contains(lowerExample, cmdExample) {
				example = fmt.Sprintf("%s", e.Field(ex).Interface())
				break
			}
		}
		usage := fmt.Sprintf("`%s`", slashCmd)

		if title == "" {
			title = fmt.Sprintf("Invalid Command: %s", cmdOption)
		} else if desc == "" {
			desc = "Command not found"
			usage = "N/A"
		} else if example == "" {
			example = "Command not found"
		}

		embed := &discordgo.MessageEmbed{
			Title: title,
			Color: rangeIn(1, 16777215),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Description",
					Value:  desc,
					Inline: false,
				},
				{
					Name:   "Usage",
					Value:  usage,
					Inline: false,
				},
				{
					Name:   "Example",
					Value:  example,
					Inline: false,
				},
			},
		}
		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			},
		)
		if err != nil {
			return err
		}

	} else {
		content := "A list of current Slash commands\n```\n"

		for i := 0; i < n.NumField(); i++ {
			content = content + fmt.Sprintf("%s \n", n.Field(i).Interface())
		}
		content = content + "```\nYou can get more information about a command by using /tuuck <command_name>"

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendVersionResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Version: %s", cfg.Version),
		Color:       62033,
		Description: "You see it up there.",
	}

	err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					embed,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func sendInsultResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	user := i.ApplicationCommandData().Options[0].UserValue(s)

	retVal := ""
	if cfg.Configs.Keys.InsultAPI != "" { // check if insult API is set up
		insultStr, err := GetInsult(cfg.Configs.Keys.InsultAPI)
		if err != nil {
			return err
		}

		retVal = insultStr
	} else {
		retVal = cfg.Cmd.Msg.InsultAPIError
	}

	embed := &discordgo.MessageEmbed{
		Title:       "(ง ͠° ͟ل͜ ͡°)ง",
		Color:       rangeIn(1, 16777215),
		Description: retVal,
	}

	err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("An ode to: <@%s>", user.ID),
				Embeds: []*discordgo.MessageEmbed{
					embed,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func GetInsult(insultURL string) (string, error) {
	res, err := http.Get(insultURL)
	if err != nil {
		return "", err
	}

	var insultObj insult

	err = json.NewDecoder(res.Body).Decode(&insultObj)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	return insultObj.Insult, nil
}

//endregion

//region Game Commands

func sendCoinFlipResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	gifURL, err := api.RequestGifURL("Coin Flip", cfg.Configs.Keys.TenorAPIkey)
	if err != nil {
		return err
	}

	embed := &discordgo.MessageEmbed{
		Title: "Flipping...",
		Color: 16761856,
		Image: &discordgo.MessageEmbedImage{
			URL: gifURL,
		},
	}

	x1 := rand.NewSource(time.Now().UnixNano())
	y1 := rand.New(x1)
	randNum := y1.Intn(200)

	search := ""
	results := ""
	if randNum%2 == 0 {
		search = "Coin Flip Heads"
		results = "Heads"
		gifURL, err = api.RequestGifURL(search, cfg.Configs.Keys.TenorAPIkey)
		if err != nil {
			return err
		}

	} else {
		search = "Coin Flip Tails"
		results = "Tails"
		gifURL, err = api.RequestGifURL(search, cfg.Configs.Keys.TenorAPIkey)
		if err != nil {
			return err
		}
	}

	embed.Description = fmt.Sprintf("It's %s!", results)
	embed.Image = &discordgo.MessageEmbedImage{
		URL: gifURL,
	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					embed,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

//endregion

//region Audio Playback

func sendPlayResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	link := i.ApplicationCommandData().Options[0].StringValue()
	err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Playing: %s", link),
			},
		},
	)
	if err != nil {
		return err
	}

	msg, err := s.ChannelMessageSend(i.ChannelID, "Prepping vidya...")
	if err != nil {
		return err
	}

	//yas
	if i.Member.User.ID == "932843527870742538" {
		link = "https://www.youtube.com/watch?v=kJQP7kiw5Fk"
	}

	link, fileName, err := web.GetYtAudioLink(s, msg, link)
	if err != nil {
		return err
	}

	err = web.DownloadMpFile(i, link, fileName)
	if err != nil {
		return err
	}

	dgv, err := voice_chat.ConnectVoiceChannel(s, i.Member.User.ID, i.GuildID)
	if err != nil {
		return err
	}

	err = web.PlayAudioFile(dgv, fileName, i, s)
	if err != nil {
		return err
	}

	return nil
}

func sendStopResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := stopAudioPlayback()
	if err != nil {
		return err
	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Okay Dad",
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func stopAudioPlayback() error {
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

func sendClearResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := web.RunMpFileCleanUp(fmt.Sprintf("%s/Audio", i.GuildID))
	if err != nil {
		return err
	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This house is clean.",
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func sendSkipResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	audio := ""
	if len(web.MpFileQueue) > 0 {
		audio = fmt.Sprintf("Skipping %s", web.MpFileQueue[0])
	} else {
		audio = "Queue is empty, my guy"
	}
	err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: audio,
			},
		},
	)
	if err != nil {
		return err
	}

	err = skipPlayback(s, i)
	if err != nil {
		return err
	}

	return nil
}

func skipPlayback(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if len(web.MpFileQueue) > 0 {
		err := stopAudioPlayback()
		if err != nil {
			return err
		}

		dgv, err := voice_chat.ConnectVoiceChannel(s, i.Member.User.ID, i.GuildID)
		if err != nil {
			return err
		}

		err = web.PlayAudioFile(dgv, "", i, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendQueueResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	queue := ""
	if len(web.MpFileQueue) > 0 {
		queue = strings.Join(web.MpFileQueue, "\n")
	} else {
		queue = "Uh owh, song queue is wempty (>.<)"
	}

	err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: queue,
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

//endregion

//region Animal Commands

func sendAnimalsResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	switch i.ApplicationCommandData().Options[0].Name {
	case "doggo":
		embed, err := getDoggoEmbed(cfg)
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func getDoggoEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	// a data scientist had to fix this...
	doggoObj, err := callDoggoAPI(cfg)
	if err != nil {
		return nil, err
	}

	if len(doggoObj[0].Breeds) < 1 {
		doggoObj, err = callDoggoAPI(cfg)
		if err != nil {
			return nil, err
		}
	}

	impWeight := checkIfEmpty(doggoObj[0].Breeds[0].Weight.Imperial)
	metWeight := checkIfEmpty(doggoObj[0].Breeds[0].Weight.Metric)
	impHeight := checkIfEmpty(doggoObj[0].Breeds[0].Height.Imperial)
	metHeight := checkIfEmpty(doggoObj[0].Breeds[0].Height.Metric)

	embed := &discordgo.MessageEmbed{
		Title:       doggoObj[0].Breeds[0].Name,
		Color:       rangeIn(1, 16777215),
		Description: doggoObj[0].Breeds[0].Temperament,
		Image: &discordgo.MessageEmbedImage{
			URL: doggoObj[0].URL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Weight",
				Value:  fmt.Sprintf("%s lbs / %s kg", impWeight, metWeight),
				Inline: true,
			},
			{
				Name:   "Height",
				Value:  fmt.Sprintf("%s in / %s cm", impHeight, metHeight),
				Inline: true,
			},
			{
				Name:   "Origin",
				Value:  checkIfEmpty(doggoObj[0].Breeds[0].Origin),
				Inline: true,
			},
			{
				Name:   "Bred For",
				Value:  checkIfEmpty(doggoObj[0].Breeds[0].BredFor),
				Inline: true,
			},
			{
				Name:   "Breed Group",
				Value:  checkIfEmpty(doggoObj[0].Breeds[0].BreedGroup),
				Inline: true,
			},
			{
				Name:   "Life Span",
				Value:  checkIfEmpty(doggoObj[0].Breeds[0].LifeSpan),
				Inline: true,
			},
		},
	}

	return embed, nil
}

func callDoggoAPI(cfg *config.Configs) (doggo, error) {
	res, err := http.Get(cfg.Configs.Keys.DoggoAPI)
	if err != nil {
		return nil, err
	}

	var doggoObj doggo

	err = json.NewDecoder(res.Body).Decode(&doggoObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	return doggoObj, nil
}

//endregion

//region Daily Commands

func sendDailyResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	switch i.ApplicationCommandData().Options[0].Name {
	case "advice":
		embed, err := getDailyAdviceEmbed(cfg)
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "kanye":
		embed, err := getDailyKanyeEmbed(cfg)
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "affirmation":
		embed, err := getDailyAffirmationEmbed(cfg)
		if err != nil {
			return err
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "horoscope":
		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Choose a zodiac sign",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.SelectMenu{
									CustomID:    "horo-select",
									Placeholder: "Zodiac",
									Options: []discordgo.SelectMenuOption{
										{
											Label:   "Aquarius",
											Value:   "aquarius",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🌊"},
										},
										{
											Label:   "Aries",
											Value:   "aries",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🐏"},
										},
										{
											Label:   "Cancer",
											Value:   "cancer",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🦀"},
										},
										{
											Label:   "Capricorn",
											Value:   "capricorn",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🐐"},
										},
										{
											Label:   "Gemini",
											Value:   "gemini",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "♊"},
										},
										{
											Label:   "Leo",
											Value:   "leo",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🦁"},
										},
										{
											Label:   "Libra",
											Value:   "libra",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "⚖️"},
										},
										{
											Label:   "Pisces",
											Value:   "pisces",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🐠"},
										},
										{
											Label:   "Sagittarius",
											Value:   "sagittarius",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🏹"},
										},
										{
											Label:   "Scorpio",
											Value:   "scorpio",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🦂"},
										},
										{
											Label:   "Taurus",
											Value:   "taurus",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "🐃"},
										},
										{
											Label:   "Virgo",
											Value:   "virgo",
											Default: false,
											Emoji:   discordgo.ComponentEmoji{Name: "♍"},
										},
									},
								},
							},
						},
					},
				},
			},
		)
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
			_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
		}
	}

	return nil
}

func getDailyAdviceEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.Keys.AdviceAPI)
	if err != nil {
		return nil, err
	}

	var adviceObj advice

	err = json.NewDecoder(res.Body).Decode(&adviceObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	embed := &discordgo.MessageEmbed{
		Title:       "( ಠ ͜ʖರೃ)",
		Color:       rangeIn(1, 16777215),
		Description: adviceObj.Slip.Advice,
	}

	return embed, nil
}

func getDailyKanyeEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.Keys.KanyeAPI)
	if err != nil {
		return nil, err
	}

	var kanyeObj kanye

	err = json.NewDecoder(res.Body).Decode(&kanyeObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	embed := &discordgo.MessageEmbed{
		Title: "(▀̿Ĺ̯▀̿ ̿)",
		Color: rangeIn(1, 16777215),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  fmt.Sprintf("\"%s\"", kanyeObj.Quote),
				Value: "- Kanye",
			},
		},
	}

	return embed, nil
}

func getDailyAffirmationEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.Keys.AffirmationAPI)
	if err != nil {
		return nil, err
	}

	var affirmObj affirmation

	err = json.NewDecoder(res.Body).Decode(&affirmObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	current := time.Now()
	timeFormat := current.Format("Jan 2, 2006")

	embed := &discordgo.MessageEmbed{
		Title:       timeFormat,
		Color:       rangeIn(1, 16777215),
		Description: affirmObj.Affirmation,
	}

	return embed, nil
}

func getHoroscopeEmbed(sign string) (*discordgo.MessageEmbed, error) {
	horoscope := ""
	found := false
	signNum := ""

	switch strings.ToLower(sign) {
	case "aries":
		signNum = "1"
	case "taurus":
		signNum = "2"
	case "gemini":
		signNum = "3"
	case "cancer":
		signNum = "4"
	case "leo":
		signNum = "5"
	case "virgo":
		signNum = "6"
	case "libra":
		signNum = "7"
	case "scorpio":
		signNum = "8"
	case "sagittarius":
		signNum = "9"
	case "capricorn":
		signNum = "10"
	case "aquarius":
		signNum = "11"
	case "pisces":
		signNum = "12"
	}

	c := colly.NewCollector()

	// Before making a request print "Visiting ..."
	c.OnRequest(
		func(r *colly.Request) {
			fmt.Println("Visiting", r.URL.String())
		},
	)

	// this is ugly, and I'd like to do away with it eventually
	today := time.Now()
	todayFormat := today.Format("Jan 2, 2006")

	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayFormat := yesterday.Format("Jan 2, 2006")

	// On every p element which has style attribute call callback
	c.OnHTML(
		"p", func(e *colly.HTMLElement) {
			if !found {
				if strings.Contains(e.Text, todayFormat) {
					horoscope = e.Text
					found = true

					// website hasn't updated yet
				} else if strings.Contains(e.Text, yesterdayFormat) {
					horoscope = e.Text
					found = true
				}
			}
		},
	)

	// StartServer scraping on https://www.horoscope.com
	err := c.Visit("https://www.horoscope.com/us/horoscopes/general/horoscope-general-daily-today.aspx?sign=" + signNum)
	if err != nil {
		return nil, nil
	}

	embed := &discordgo.MessageEmbed{
		Title:       sign,
		Description: horoscope,
		Color:       rangeIn(1, 16777215),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Via Horoscopes.com",
			IconURL: "https://www.horoscope.com/images-US/horoscope-logo.svg",
		},
	}

	return embed, nil
}

//endregion

//region Pick Commands

func sendPickResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	switch strings.ToLower(i.ApplicationCommandData().Options[0].Name) {
	case "steam":
		choice, err := getSteamGame(cfg)
		if err != nil {
			return err
		}

		embed := &discordgo.MessageEmbed{
			Title: "I have Chosen...",
			Color: rangeIn(1, 16777215),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   choice,
					Value:  "☝(°ロ°)",
					Inline: true,
				},
			},
		}

		err = s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Choosing Steam Game",
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			},
		)
		if err != nil {
			return err
		}

	case "choices":
		content := ""
		for _, v := range i.ApplicationCommandData().Options[0].Options {
			content = content + fmt.Sprintf("[%s] ", v.StringValue())
		}
		content = strings.TrimSpace(content)
		content = fmt.Sprintf("*%s*", content)

		randomIndex := rand.Intn(len(i.ApplicationCommandData().Options))
		choice := i.ApplicationCommandData().Options[0].Options[randomIndex].StringValue()

		embed := &discordgo.MessageEmbed{
			Title: "I have chosen...",
			Color: rangeIn(1, 16777215),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   choice,
					Value:  "☝(°ロ°)",
					Inline: true,
				},
			},
		}

		err := s.InteractionRespond(
			i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Choosing Steam Game",
					Embeds: []*discordgo.MessageEmbed{
						embed,
					},
				},
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func getSteamGame(cfg *config.Configs) (string, error) {
	res, err := http.Get(cfg.Configs.Keys.SteamAPI)
	if err != nil {
		return "", err
	}

	var steamObj steamGames

	err = json.NewDecoder(res.Body).Decode(&steamObj)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)
	if err != nil {
		return "", err
	}

	randomIndex := rand.Intn(len(steamObj.Applist.Apps))
	choice := fmt.Sprintf("%s\nsteam://openurl/https://store.steampowered.com/app/%v", steamObj.Applist.Apps[randomIndex].Name, steamObj.Applist.Apps[randomIndex].Appid)

	return choice, nil
}

//endregion

//region Component Commands

func sendHoroscopeCompResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	sign := i.MessageComponentData().Values[0]
	embed, err := getHoroscopeEmbed(sign)
	if err != nil {
		return err
	}

	msgEdit := discordgo.NewMessageEdit(i.ChannelID, i.Message.ID)
	msgContent := ""
	msgEdit.Content = &msgContent
	msgEdit.Embeds = []*discordgo.MessageEmbed{embed}

	// edit response (i.Interaction) and replace with embed
	_, err = s.ChannelMessageEditComplex(msgEdit)
	if err != nil {
		return err
	}

	// 'This interaction failed' will show if not included
	// todo fix later
	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "",
			},
		},
	)
	if err != nil {
		if !strings.Contains(err.Error(), "Cannot send an empty message") {
			return err
		}
	}

	return nil
}

//endregion
