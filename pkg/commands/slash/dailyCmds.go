package slash

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
)

func sendDailyResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	commandName := i.ApplicationCommandData().Options[0].Name
	errRespMsg := "Unable to make call at this moment, please try later :("

	// Defer the interaction response to avoid timeout
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /get command %s: %w", commandName, err)
	}

	var webhookEdit *discordgo.WebhookEdit
	var err error

	embeds := map[string]func(*config.Configs) (*discordgo.MessageEmbed, error){
		"advice":      getDailyAdviceEmbed,
		"kanye":       getDailyKanyeEmbed,
		"affirmation": getDailyAffirmationEmbed,
		"fact":        getDailyFactEmbed,
	}

	switch commandName {
	case "advice",
		"kanye",
		"affirmation",
		"fact":
		var embed *discordgo.MessageEmbed

		embed, err = embeds[commandName](cfg)
		if err == nil {
			webhookEdit = &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}}
		}

	case "horoscope":
		webhookEdit = getHoroscopeWebHookEdit()

	default:
		return fmt.Errorf("unknown option: %s", commandName)
	}
	if err != nil {
		_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
		return fmt.Errorf("error in dailyCmds.sendDailyResponse() : %w", err)
	}

	// Edit the interaction response with the generated data
	if _, err = s.InteractionResponseEdit(
		i.Interaction, webhookEdit,
	); err != nil {
		_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
		return fmt.Errorf("failed to send message for command %s: %w", commandName, err)
	}

	return nil
}

func getHoroscopeWebHookEdit() *discordgo.WebhookEdit {
	content := "Choose a zodiac sign"
	return &discordgo.WebhookEdit{
		Content: &content,
		Components: &[]discordgo.MessageComponent{
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

	// Defer the response to prevent interaction timeout
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction: %w", err)
	}

	embed, err := getHoroscopeEmbed(sign)
	if err != nil {
		// Respond to user with a fallback message if something goes wrong
		_ = helper.SendResponseErrorToUser(s, i, "Unable to fetch Horoscope atm, try again later.")
		return fmt.Errorf("failed to get horoscope embed: %w", err)
	}

	// Replace the interaction message with new content
	_, err = s.ChannelMessageEditComplex(
		&discordgo.MessageEdit{
			Channel: i.ChannelID,
			ID:      i.Message.ID,
			Content: new(string), // Empty string to clear content
			Embeds:  &[]*discordgo.MessageEmbed{embed},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to edit interaction message: %w", err)
	}

	return nil
}

func getDailyFactEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	client := cfg.Clients.Dagpi
	var clientData interface{}

	clientData, err := client.Fact()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch daily fact: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Fun Fact",
		Color:       helper.RangeIn(1, 16777215),
		Description: fmt.Sprintf("%v", clientData),
	}

	return embed, nil
}

func getDailyAdviceEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	resp, err := http.Get(cfg.Configs.ApiURLs.AdviceAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch advice: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("advice API returned status %d", resp.StatusCode)
	}

	var adviceObj advice
	if err = json.NewDecoder(resp.Body).Decode(&adviceObj); err != nil {
		return nil, fmt.Errorf("failed to decode advice JSON: %w", err)
	}

	adv := "404 Advice Not Found, So here's a tip instead: Plan for bad API calls."
	if adviceObj.Slip.Advice != "" {
		adv = adviceObj.Slip.Advice
	}

	return &discordgo.MessageEmbed{
		Title:       "( ಠ ͜ʖರೃ)",
		Color:       helper.RangeIn(1, 16777215),
		Description: adv,
	}, nil
}

func getDailyKanyeEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	resp, err := http.Get(cfg.Configs.ApiURLs.KanyeAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Kanye quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kanye API returned status code %d", resp.StatusCode)
	}

	var kanyeObj kanye
	if err = json.NewDecoder(resp.Body).Decode(&kanyeObj); err != nil {
		return nil, fmt.Errorf("failed to decode Kanye API response: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		Title: "(▀̿Ĺ̯▀̿ ̿)",
		Color: helper.RangeIn(1, 16777215),
		Fields: []*discordgo.MessageEmbedField{
			{Name: fmt.Sprintf("\"%s\"", kanyeObj.Quote), Value: "\\- Kanye"},
		},
	}

	return embed, nil
}

func getDailyAffirmationEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	resp, err := http.Get(cfg.Configs.ApiURLs.AffirmationAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch affirmation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("affirmation API returned status code %d", resp.StatusCode)
	}

	var affirmObj affirmation
	if err = json.NewDecoder(resp.Body).Decode(&affirmObj); err != nil {
		return nil, fmt.Errorf("failed to decode affirmation response: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		Title:       time.Now().Format("Jan 2, 2006"),
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

	if horoscope == "" {
		horoscope = "404 Horoscope Not Found, So here's a tip instead: Plan for bad API calls."
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
	signMap := map[string]string{
		"aries":       "1",
		"taurus":      "2",
		"gemini":      "3",
		"cancer":      "4",
		"leo":         "5",
		"virgo":       "6",
		"libra":       "7",
		"scorpio":     "8",
		"sagittarius": "9",
		"capricorn":   "10",
		"aquarius":    "11",
		"pisces":      "12",
	}

	return signMap[strings.ToLower(sign)]
}
