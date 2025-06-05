package slash

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
	"io"
	"net/http"
	"strings"
	"time"
)

func sendDailyResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	client := cfg.Clients.Dagpi

	optionName := i.ApplicationCommandData().Options[0].Name
	var err error
	var data *discordgo.InteractionResponseData

	switch optionName {
	case "advice":
		data, err = getDailyEmbed(cfg, getDailyAdviceEmbed)

	case "kanye":
		data, err = getDailyEmbed(cfg, getDailyKanyeEmbed)

	case "affirmation":
		data, err = getDailyEmbed(cfg, getDailyAffirmationEmbed)

	case "horoscope":
		data = getHoroscopeResponseData()

	case "fact":
		var clientData interface{}
		clientData, err = client.Fact()

		data = &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("%s", clientData),
		}
	default:
		return fmt.Errorf("unknown option: %s", optionName)
	}
	if err != nil {
		go func() {
			if err = helper.SendResponseErrorToUser(s, i, "Unable to fetch data atm, Try again later."); err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, i.GuildID)
			}
		}()
		return err
	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func getDailyEmbed(cfg *config.Configs, embedFunc func(*config.Configs) (*discordgo.MessageEmbed, error)) (*discordgo.InteractionResponseData, error) {
	embed, err := embedFunc(cfg)
	if err != nil {
		return nil, err
	}

	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			embed,
		},
	}, nil
}

func getHoroscopeResponseData() *discordgo.InteractionResponseData {
	return &discordgo.InteractionResponseData{
		Content: "Choose a zodiac sign",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "horo-select",
						Placeholder: "Zodiac",
						Options: []discordgo.SelectMenuOption{
							{Label: "Aquarius", Value: "aquarius", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "🌊"}},
							{Label: "Aries", Value: "aries", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "🐏"}},
							{Label: "Cancer", Value: "cancer", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "🦀"}},
							{Label: "Capricorn", Value: "capricorn", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "🐐"}},
							{Label: "Gemini", Value: "gemini", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "♊"}},
							{Label: "Leo", Value: "leo", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "🦁"}},
							{Label: "Libra", Value: "libra", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "⚖️"}},
							{Label: "Pisces", Value: "pisces", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "🐠"}},
							{Label: "Sagittarius", Value: "sagittarius", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "🏹"}},
							{Label: "Scorpio", Value: "scorpio", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "🦂"}},
							{Label: "Taurus", Value: "taurus", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "🐃"}},
							{Label: "Virgo", Value: "virgo", Default: false, Emoji: &discordgo.ComponentEmoji{Name: "♍"}},
						},
					},
				},
			},
		},
	}
}

func sendHoroscopeCompResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	sign := i.MessageComponentData().Values[0]

	embed, err := getHoroscopeEmbed(sign)
	if err != nil {
		go func() {
			err = helper.SendResponseErrorToUser(s, i, "Unable to fetch Horoscope atm, try again later.")
		}()
		return err
	}

	msgEdit := discordgo.NewMessageEdit(i.ChannelID, i.Message.ID)
	msgContent := ""
	msgEdit.Content = &msgContent
	msgEdit.Embeds = &[]*discordgo.MessageEmbed{embed}

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

func getDailyAdviceEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.ApiURLs.AdviceAPI)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	var adviceObj advice
	err = json.NewDecoder(res.Body).Decode(&adviceObj)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "( ಠ ͜ʖರೃ)",
		Color:       helper.RangeIn(1, 16777215),
		Description: adviceObj.Slip.Advice,
	}

	return embed, nil
}

func getDailyKanyeEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.ApiURLs.KanyeAPI)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	var kanyeObj kanye
	err = json.NewDecoder(res.Body).Decode(&kanyeObj)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title: "(▀̿Ĺ̯▀̿ ̿)",
		Color: helper.RangeIn(1, 16777215),
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
	res, err := http.Get(cfg.Configs.ApiURLs.AffirmationAPI)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	var affirmObj affirmation
	err = json.NewDecoder(res.Body).Decode(&affirmObj)
	if err != nil {
		return nil, err
	}

	current := time.Now()
	timeFormat := current.Format("Jan 2, 2006")

	embed := &discordgo.MessageEmbed{
		Title:       timeFormat,
		Color:       helper.RangeIn(1, 16777215),
		Description: affirmObj.Affirmation,
	}

	return embed, nil
}

func getHoroscopeEmbed(sign string) (*discordgo.MessageEmbed, error) {
	signNum := getSignNumber(sign)
	if signNum == "" {
		return nil, fmt.Errorf("invalid sign: %s", sign)
	}

	horoscope, err := scrapeHoroscope(signNum)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title:       sign,
		Description: horoscope,
		Color:       helper.RangeIn(1, 16777215),
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Via Horoscopes.com",
			IconURL: "https://www.horoscope.com/images-US/horoscope-logo.svg",
		},
	}

	return embed, nil
}

func scrapeHoroscope(signNum string) (string, error) {
	c := colly.NewCollector()

	var horoscope string
	found := false

	todayFormat := time.Now().Format("Jan 2, 2006")
	yesterdayFormat := time.Now().AddDate(0, 0, -1).Format("Jan 2, 2006")

	c.OnHTML(
		"p", func(e *colly.HTMLElement) {
			if !found {
				if strings.Contains(e.Text, todayFormat) || strings.Contains(e.Text, yesterdayFormat) {
					horoscope = e.Text
					found = true
				}
			}
		},
	)

	err := c.Visit("https://www.horoscope.com/us/horoscopes/general/horoscope-general-daily-today.aspx?sign=" + signNum)
	if err != nil {
		return "", fmt.Errorf("failed to scrape horoscope: %v", err)
	}

	return horoscope, nil
}

func getSignNumber(sign string) string {
	sign = strings.ToLower(sign)
	signNum := ""

	switch sign {
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
	default:
		signNum = ""
	}
	return signNum
}
