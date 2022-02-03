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

func callDoggoAPI(cfg *config.ConfigStructs) (doggo, error) {
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

func getDoggoEmbed(cfg *config.ConfigStructs) (*discordgo.MessageEmbed, error) {
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

func getAdviceEmbed(cfg *config.ConfigStructs) (*discordgo.MessageEmbed, error) {
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

func getKanyeEmbed(cfg *config.ConfigStructs) (*discordgo.MessageEmbed, error) {
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

func getAffirmationEmbed(cfg *config.ConfigStructs) (*discordgo.MessageEmbed, error) {
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

func getPickEmbed(options []*discordgo.ApplicationCommandInteractionDataOption, cfg *config.ConfigStructs) (*discordgo.MessageEmbed, error) {
	choice := ""
	switch strings.ToLower(options[0].Name) {
	case "steam":
		res, err := http.Get(cfg.Configs.Keys.SteamAPI)
		if err != nil {
			return nil, err
		}

		var steamObj steamGames

		err = json.NewDecoder(res.Body).Decode(&steamObj)
		if err != nil {
			return nil, err
		}

		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				return
			}
		}(res.Body)

		randomIndex := rand.Intn(len(steamObj.Applist.Apps))
		choice = fmt.Sprintf("%s\nsteam://openurl/https://store.steampowered.com/app/%v", steamObj.Applist.Apps[randomIndex].Name, steamObj.Applist.Apps[randomIndex].Appid)

	case "choices":
		randomIndex := rand.Intn(len(options))
		choice = options[0].Options[randomIndex].StringValue()
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

	return embed, nil
}

func getTuuckEmbed(cmd string, cfg *config.ConfigStructs) *discordgo.MessageEmbed {
	n := reflect.ValueOf(&cfg.Cmd.Name).Elem()
	d := reflect.ValueOf(&cfg.Cmd.Desc).Elem()

	desc := ""
	title := "A list of current Slash commands:"
	cmd = strings.ToLower(cmd)

	if cmd == "" {
		for i := 0; i < n.NumField(); i++ {
			desc = desc + fmt.Sprintf("%s \n", n.Field(i).Interface())
		}
	} else {
		for i := 0; i < n.NumField(); i++ {
			if strings.Contains(fmt.Sprintf("%s", n.Field(i).Interface()), cmd) {
				title = fmt.Sprintf("%s", n.Field(i).Interface())
				break
			}
		}

		for i := 0; i < d.NumField(); i++ {
			name := strings.ToLower(d.Type().Field(i).Name)
			if strings.Contains(name, cmd) {
				desc = fmt.Sprintf("%s", d.Field(i).Interface())
				break
			}
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Color:       rangeIn(1, 16777215),
		Description: desc,
	}

	return embed
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

func getCoinFlipEmbed(cfg *config.ConfigStructs) (*discordgo.MessageEmbed, error) {
	gifURL, err := api.RequestGifURL("Coin Flip", cfg.Configs.Keys.TenorAPIkey)
	if err != nil {
		return nil, err
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
			return nil, err
		}

	} else {
		search = "Coin Flip Tails"
		results = "Tails"
		gifURL, err = api.RequestGifURL(search, cfg.Configs.Keys.TenorAPIkey)
		if err != nil {
			return nil, err
		}
	}

	embed.Description = fmt.Sprintf("It's %s!", results)
	embed.Image = &discordgo.MessageEmbedImage{
		URL: gifURL,
	}

	return embed, nil
}

func playYoutubeLink(s *discordgo.Session, i *discordgo.InteractionCreate, param string) error {
	msg, err := s.ChannelMessageSend(i.ChannelID, "Prepping vidya...")
	if err != nil {
		return err
	}

	//yas
	if i.Member.User.ID == "932843527870742538" {
		param = "https://www.youtube.com/watch?v=kJQP7kiw5Fk"
	}

	link, fileName, err := web.GetYtAudioLink(s, msg, param)
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
