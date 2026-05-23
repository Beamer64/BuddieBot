package slash

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/bb_data/affirmations"
	"github.com/Beamer64/bb_data/facts"
	"github.com/Beamer64/bb_data/kanye"
	"github.com/Beamer64/bb_data/tonguetwister"
	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
)

func sendDailyResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	cmdType := i.ApplicationCommandData().Options[0].StringValue()
	errRespMsg := "Unable to make call at this moment, please try later :("

	// Defer the interaction response to avoid timeout
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /daily type %s: %w", cmdType, err)
	}

	var webhookEdit *discordgo.WebhookEdit
	var err error

	embeds := map[string]func(*config.Configs) (*discordgo.MessageEmbed, error){
		"advice":         getDailyAdviceEmbed,
		"kanye":          getDailyKanyeEmbed,
		"affirmation":    getDailyAffirmationEmbed,
		"fact":           getDailyFactEmbed,
		"tongue-twister": getDailyTongueTwister,
	}

	switch cmdType {
	case "advice",
		"kanye",
		"affirmation",
		"fact",
		"tongue-twister":
		var embed *discordgo.MessageEmbed

		embed, err = embeds[cmdType](cfg)
		if err == nil {
			webhookEdit = &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}}
		}

	case "horoscope":
		webhookEdit = getHoroscopeWebHookEdit()

	default:
		return fmt.Errorf("unknown option: %s", cmdType)
	}
	if err != nil {
		return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fmt.Errorf("sendDailyResponse %s: %w", cmdType, err))
	}

	// Edit the interaction response with the generated data
	if _, err = s.InteractionResponseEdit(
		i.Interaction, webhookEdit,
	); err != nil {
		return fmt.Errorf("send /daily response for type %s: %w", cmdType, err)
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

func sendHoroscopeCompResponse(s *discordgo.Session, i *discordgo.InteractionCreate, _ *config.Configs) error {
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
		return helper.ReturnUserErrorDeferred(s, i, "Unable to fetch Horoscope atm, try again later.", fmt.Errorf("get horoscope embed for %s: %w", sign, err))
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

func getDailyFactEmbed(_ *config.Configs) (*discordgo.MessageEmbed, error) {
	return &discordgo.MessageEmbed{
		Title:       "Fun Fact",
		Color:       helper.RandomDiscordColor(),
		Description: facts.Random(),
	}, nil
}

func getDailyTongueTwister(_ *config.Configs) (*discordgo.MessageEmbed, error) {
	return &discordgo.MessageEmbed{
		Title:       "Try This Tongue Twister",
		Color:       helper.RandomDiscordColor(),
		Description: tonguetwister.Random(),
	}, nil
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
		Color:       helper.RandomDiscordColor(),
		Description: adv,
	}, nil
}

func getDailyKanyeEmbed(_ *config.Configs) (*discordgo.MessageEmbed, error) {
	return &discordgo.MessageEmbed{
		Title: "(▀̿Ĺ̯▀̿ ̿)",
		Color: helper.RandomDiscordColor(),
		Fields: []*discordgo.MessageEmbedField{
			{Name: fmt.Sprintf("\"%s\"", kanye.Random()), Value: "\\- Kanye"},
		},
	}, nil
}

func getDailyAffirmationEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	text, err := fetchAffirmationFromAPI(cfg)
	if err != nil {
		log.Printf("affirmation API unavailable, using local fallback: %v", err)
		text = affirmations.Random()
		if text == "" {
			return nil, fmt.Errorf("affirmation API failed and local fallback pool is empty: %w", err)
		}
	}

	return &discordgo.MessageEmbed{
		Title:       time.Now().Format("Jan 2, 2006"),
		Color:       helper.RandomDiscordColor(),
		Description: text,
	}, nil
}

// fetchAffirmationFromAPI tries the configured affirmation API with a
// short timeout. Returns the affirmation text or an error describing
// the specific failure (network, non-200, decode, empty body).
func fetchAffirmationFromAPI(cfg *config.Configs) (string, error) {
	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(cfg.Configs.ApiURLs.AffirmationAPI)
	if err != nil {
		return "", fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var affirmObj affirmation
	if err := json.NewDecoder(resp.Body).Decode(&affirmObj); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if strings.TrimSpace(affirmObj.Affirmation) == "" {
		return "", fmt.Errorf("empty affirmation field")
	}
	return affirmObj.Affirmation, nil
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
		Color:       helper.RandomDiscordColor(),
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
		return "", fmt.Errorf("failed to scrape horoscope: %w", err)
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

func dailySpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "daily",
		Description: "Receive daily quotes, horoscopes, affirmations, etc.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Which daily response to get",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "advice", Value: "advice"},
					{Name: "affirmation", Value: "affirmation"},
					{Name: "fact", Value: "fact"},
					{Name: "horoscope", Value: "horoscope"},
					{Name: "kanye", Value: "kanye"},
					{Name: "tongue-twister", Value: "tongue-twister"},
				},
			},
		},
	}
}
