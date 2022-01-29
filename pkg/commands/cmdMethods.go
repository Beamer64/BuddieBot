package commands

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/discordBot/pkg/api"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/voice_chat"
	"github.com/beamer64/discordBot/pkg/web_scrape"
	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type steamGames struct {
	Applist struct {
		Apps []struct {
			Appid int    `json:"appid"`
			Name  string `json:"name"`
		} `json:"apps"`
	} `json:"applist"`
}

// rangeIn Returns pseudo rand num between low and high.
// For random embed color: rangeIn(1, 16777215)
func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

func getPickEmbed(options []*discordgo.ApplicationCommandInteractionDataOption, cfg *config.ConfigStructs) (*discordgo.MessageEmbed, error) {
	choice := ""
	//content := ""
	if strings.ToLower(options[0].StringValue()) == "steam" {
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
		//choice = steamObj.Applist.Apps[randomIndex].Name
		choice = fmt.Sprintf("%s\nsteam://openurl/https://store.steampowered.com/app/%v", steamObj.Applist.Apps[randomIndex].Name, steamObj.Applist.Apps[randomIndex].Appid)

	} else {
		randomIndex := rand.Intn(len(options))
		choice = options[randomIndex].StringValue()
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
	// On every p element which has style attribute call callback
	c.OnHTML(
		"p", func(e *colly.HTMLElement) {
			if !found {
				if e.Text != "" {
					horoscope = e.Text
					found = true
				}
			}
		},
	)

	// Before making a request print "Visiting ..."
	c.OnRequest(
		func(r *colly.Request) {
			fmt.Println("Visiting", r.URL.String())
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

	link, fileName, err := web_scrape.GetYtAudioLink(s, msg, param)
	if err != nil {
		return err
	}

	err = web_scrape.DownloadMpFile(i, link, fileName)
	if err != nil {
		return err
	}

	dgv, err := voice_chat.ConnectVoiceChannel(s, i.Member.User.ID, i.GuildID)
	if err != nil {
		return err
	}

	err = web_scrape.PlayAudioFile(dgv, fileName, i, s)
	if err != nil {
		return err
	}

	return nil
}

func stopAudioPlayback() error {
	//vc := voice_chat.VoiceConnection{}

	if web_scrape.StopPlaying != nil {
		close(web_scrape.StopPlaying)
		web_scrape.IsPlaying = false

		/*if vc.Dgv != nil {
			vc.Dgv.Close()

		}*/
	}

	return nil
}

func skipPlayback(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if len(web_scrape.MpFileQueue) > 0 {
		err := stopAudioPlayback()
		if err != nil {
			return err
		}

		dgv, err := voice_chat.ConnectVoiceChannel(s, i.Member.User.ID, i.GuildID)
		if err != nil {
			return err
		}

		err = web_scrape.PlayAudioFile(dgv, "", i, s)
		if err != nil {
			return err
		}
	}

	return nil
}
