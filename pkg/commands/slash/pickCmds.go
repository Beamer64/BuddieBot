package slash

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/bb_data/emojis"
	"github.com/bwmarrin/discordgo"
)

// pickSteamLimiter: Steam Web API has documented per-key rate limits.
var pickSteamLimiter = helper.NewRateLimiter(5 * time.Second)

func sendPickResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	option := strings.ToLower(i.ApplicationCommandData().Options[0].Name)

	if option == "steam" {
		if ok, retry := pickSteamLimiter.Allow(i.Member.User.ID); !ok {
			msg := fmt.Sprintf("Slow down! Try again in `%.0fs`.", retry.Seconds())
			return helper.ReturnUserError(s, i, msg, nil)
		}
	}

	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction: %w", err)
	}

	var data *discordgo.InteractionResponseData
	var err error

	switch option {
	case "steam":
		data, err = sendSteamPickResponse(cfg)
	case "choices":
		data = sendChoicesPickResponse(i)
	case "poll":
		return sendPollResponse(s, i, cfg)
	default:
		return helper.ReturnUserErrorDeferred(s, i, "Unknown pick option.", fmt.Errorf("unknown option: %s", option))
	}
	if err != nil {
		return helper.ReturnUserErrorDeferred(s, i, "Unable to pick atm, try again later.", err)
	}

	if _, err = s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Content: &data.Content,
			Embeds:  &data.Embeds,
		},
	); err != nil {
		return fmt.Errorf("send /pick %s response: %w", option, err)
	}
	return nil
}

func sendSteamPickResponse(cfg *config.Configs) (*discordgo.InteractionResponseData, error) {
	gameURL, err := getSteamGame(cfg)
	if err != nil {
		return nil, err
	}

	data := &discordgo.InteractionResponseData{
		Content: fmt.Sprintf("I have Chosen...\n %s \n☝(°ロ°)☝", gameURL),
	}

	return data, nil
}

func sendChoicesPickResponse(i *discordgo.InteractionCreate) *discordgo.InteractionResponseData {
	content := ""
	for _, v := range i.ApplicationCommandData().Options[0].Options {
		content = content + fmt.Sprintf("[%s] ", v.StringValue())
	}

	content = strings.TrimSpace(content)
	content = fmt.Sprintf("*%s*", content)

	randomIndex := rand.Intn(len(i.ApplicationCommandData().Options[0].Options))
	choice := i.ApplicationCommandData().Options[0].Options[randomIndex].StringValue()

	embed := &discordgo.MessageEmbed{
		Title: "I have chosen...",
		Color: helper.RandomDiscordColor(),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   choice,
				Value:  "☝(°ロ°)",
				Inline: true,
			},
		},
	}

	data := &discordgo.InteractionResponseData{
		Content: content,
		Embeds:  []*discordgo.MessageEmbed{embed},
	}

	return data
}

// sendPollResponse expects the interaction to already be deferred (by sendPickResponse).
func sendPollResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	question := i.ApplicationCommandData().Options[0].Options[0]
	var fields []*discordgo.MessageEmbedField
	var pollEmojis []string

	for _, v := range i.ApplicationCommandData().Options[0].Options {
		emoji := emojis.Random()
		if v.Name != "request" {
			f := &discordgo.MessageEmbedField{
				Name:   v.StringValue(),
				Value:  emoji,
				Inline: false,
			}
			fields = append(fields, f)
			pollEmojis = append(pollEmojis, emoji)
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:  question.StringValue(),
		Color:  helper.RandomDiscordColor(),
		Fields: fields,
	}

	content := helper.PollMessageContent
	embeds := []*discordgo.MessageEmbed{embed}
	if _, err := s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
			Embeds:  &embeds,
		},
	); err != nil {
		return fmt.Errorf("send /pick poll response: %w", err)
	}

	if err := addPollReactions(pollEmojis, i, s); err != nil {
		return fmt.Errorf("add poll reactions: %w", err)
	}
	return nil
}

func getSteamGame(cfg *config.Configs) (string, error) {
	urlCtx, urlCancel := context.WithTimeout(context.Background(), 2*time.Second)
	url, err := cfg.DB.GetApiURL(urlCtx, "steam")
	urlCancel()
	if err != nil {
		return "", fmt.Errorf("get steam url: %w", err)
	}
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	var steamObj steamGames
	err = json.NewDecoder(res.Body).Decode(&steamObj)
	if err != nil {
		return "", err
	}

	randomIndex := rand.Intn(len(steamObj.Applist.Apps))
	for steamObj.Applist.Apps[randomIndex].Name == "" {
		randomIndex = rand.Intn(len(steamObj.Applist.Apps))
	}
	gameURL := fmt.Sprintf("https://store.steampowered.com/app/%v", steamObj.Applist.Apps[randomIndex].Appid)

	return gameURL, nil
}

func addPollReactions(emojis []string, i *discordgo.InteractionCreate, s *discordgo.Session) error {
	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		return err
	}

	pollMsgID := channel.LastMessageID

	for _, v := range emojis {
		err = s.MessageReactionAdd(channel.ID, pollMsgID, v)
		if err != nil {
			err = fmt.Errorf("Emoji: %s \n %s", v, err)
			return err
		}
	}

	return nil
}

func pickSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "pick",
		Description: "I'll pick stuff for you. I'll also pick a steam game with the 1st choice of 'steam'",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "choices",
				Description: "Will choose between 2 or more things.",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "1st",
						Description: "First choice",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "2nd",
						Description: "Second choice",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "3rd",
						Description: "Third choice",
						Required:    false,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "4th",
						Description: "Fourth choice",
						Required:    false,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "5th",
						Description: "Fifth choice",
						Required:    false,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "6th",
						Description: "Sixth choice",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "poll",
				Description: "Gauge the room!",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "request",
						Description: "Post the Question",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "1st_poll_item",
						Description: "First Choice",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "2nd_poll_item",
						Description: "Second Choice",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "3rd_poll_item",
						Description: "Third Choice",
						Required:    false,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "4th_poll_item",
						Description: "Fourth Choice",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "steam",
				Description: "Will choose a random Steam game to play.",
				Required:    false,
			},
		},
	}
}
