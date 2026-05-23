package slash

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/bb_data/eightball"
	"github.com/Beamer64/bb_data/jokes"
	"github.com/Beamer64/bb_data/pickuplines"
	"github.com/Beamer64/bb_data/roasts"
	"github.com/Beamer64/bb_data/yomomma"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
)

func sendGetResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	// Flat options: `type` (required choice), `user` (optional), `text` (optional).
	optMap := map[string]*discordgo.ApplicationCommandInteractionDataOption{}
	for _, opt := range i.ApplicationCommandData().Options {
		optMap[opt.Name] = opt
	}
	cmdType := optMap["type"].StringValue()
	errRespMsg := "Unable to make call at this moment, please try later :("

	// Defer the interaction response to avoid timeout
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /get type %s: %w", cmdType, err)
	}

	var err error
	var embed *discordgo.MessageEmbed

	pingedUser := fmt.Sprintf("<@!%s>", i.Member.User.ID)
	if userOpt, ok := optMap["user"]; ok {
		if u := userOpt.UserValue(s); u != nil {
			pingedUser = fmt.Sprintf("<@!%s>", u.ID)
		}
	}

	switch cmdType {
	case "rekd":

		embed = &discordgo.MessageEmbed{
			Title: "(ง ͠° ͟ل͜ ͡°)ง",
			Color: helper.RandomDiscordColor(),
			Fields: []*discordgo.MessageEmbedField{
				{Name: roasts.Random(), Value: ""},
			},
		}

	case "joke":
		embed = &discordgo.MessageEmbed{
			Title: "☜(˚▽˚)☞",
			Color: helper.RandomDiscordColor(),
			Fields: []*discordgo.MessageEmbedField{
				{Name: jokes.Random(), Value: ""},
			},
		}

	case "8ball":
		embed = &discordgo.MessageEmbed{
			Title: "༼ ºل͟º ༼ ºل͟º ༼ ºل͟º ༽ ºل͟º ༽ ºل͟º ༽",
			Color: helper.RandomDiscordColor(),
			Fields: []*discordgo.MessageEmbedField{
				{Name: eightball.Random(), Value: ""},
			},
		}

	case "yomomma":

		embed = &discordgo.MessageEmbed{
			Title: "(•.•) ( •.•)>⌐■-■ (⌐■_■)",
			Color: helper.RandomDiscordColor(),
			Fields: []*discordgo.MessageEmbedField{
				{Name: yomomma.Random(), Value: ""},
			},
		}

	case "pickup-line":
		embed = &discordgo.MessageEmbed{
			Title: "ಠ‿↼",
			Color: helper.RandomDiscordColor(),
			Fields: []*discordgo.MessageEmbedField{
				{Name: pickuplines.Random(), Value: ""},
			},
		}

	case "xkcd":
		embed, err = getXkcdEmbed(cfg)

		/*case "captcha":
		data, err := client.WTP()
		if err != nil {
			return err
		}*/

	default:
		return fmt.Errorf("unknown option: %s", cmdType)
	}
	if err != nil {
		return helper.ReturnUserErrorDeferred(s, i, errRespMsg, fmt.Errorf("sendGetResponse %s: %w", cmdType, err))
	}

	webhookEdit := &discordgo.WebhookEdit{
		Content: &pingedUser,
		Embeds:  &[]*discordgo.MessageEmbed{embed},
	}

	// Edit the interaction response with the generated data
	if _, err = s.InteractionResponseEdit(
		i.Interaction, webhookEdit,
	); err != nil {
		return fmt.Errorf("send /get response for type %s: %w", cmdType, err)
	}

	return nil
}

func getXkcdEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	resp, err := http.Get(cfg.Configs.ApiURLs.XkcdAPI)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	imgURL := "https:"
	doc.Find("#comic img").EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			if src, exists := s.Attr("src"); exists {
				imgURL = imgURL + src
				return false // Stop after finding the first image
			}
			return true
		},
	)

	embed := &discordgo.MessageEmbed{
		Title: "People used to read comics.",
		Color: helper.RandomDiscordColor(),
		Image: &discordgo.MessageEmbedImage{
			URL: imgURL,
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Generated with xkcd.com",
			IconURL: imgURL,
		},
	}

	return embed, nil
}

func getSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "get",
		Description: "Get a text based response like a joke or pickup line",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "What can I get you?",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "8ball", Value: "8ball"},
					{Name: "joke", Value: "joke"},
					{Name: "pickup-line", Value: "pickup-line"},
					{Name: "rekd", Value: "rekd"},
					{Name: "xkcd", Value: "xkcd"},
					{Name: "yomomma", Value: "yomomma"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Target someone else",
				Required:    false,
			},
		},
	}
}
