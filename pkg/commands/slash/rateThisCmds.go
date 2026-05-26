package slash

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func sendRateThisResponse(s *discordgo.Session, i *discordgo.InteractionCreate, _ *config.Configs) error {
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction: %w", err)
	}

	optMap := map[string]*discordgo.ApplicationCommandInteractionDataOption{}
	for _, opt := range i.ApplicationCommandData().Options {
		optMap[opt.Name] = opt
	}
	cmdType := optMap["type"].StringValue()

	user := fmt.Sprintf("<@!%s>", i.Member.User.ID)
	if userOpt, ok := optMap["user"]; ok {
		if u := userOpt.UserValue(s); u != nil {
			user = fmt.Sprintf("<@!%s>", u.ID)
		}
	}

	embed, err := getRateThisEmbed(cmdType, user)
	if err != nil {
		return helper.ReturnUserErrorDeferred(s, i, "Unable to Rate atm, try again later.", fmt.Errorf("getRateThisEmbed %s: %w", cmdType, err))
	}

	embeds := []*discordgo.MessageEmbed{embed}
	if _, err = s.InteractionResponseEdit(
		i.Interaction, &discordgo.WebhookEdit{
			Embeds: &embeds,
		},
	); err != nil {
		return fmt.Errorf("send /rate-this %s response: %w", cmdType, err)
	}
	return nil
}

func getRateThisEmbed(ratingName string, user string) (*discordgo.MessageEmbed, error) {
	score := strconv.Itoa(rand.Intn(100))
	rateTitle, rateDesc := getRateTitleAndDesc(ratingName, user, score)

	embed := &discordgo.MessageEmbed{
		Title:       rateTitle,
		Description: rateDesc,
		Color:       helper.RandomDiscordColor(),
	}

	return embed, nil
}

var standardRatings = map[string]struct{ Title, ScoreLabel string }{
	"simp":       {"Rate This Simp", "Simp"},
	"dank":       {"Dank Rating", "Dank"},
	"epic-gamer": {"Rate This Epic Gamer", "Epic Gamer"},
	"gay":        {"Gay Rating", "Gay"},
	"stinky":     {"Rate This Stinky", "Stinky"},
	"thot":       {"Rate This Thot", "Thot"},
	"pickme":     {"Rate This Pick-Me", "Pick-Me"},
	"neckbeard":  {"Rate This Neck Beard", "Neck Beard"},
	"looks":      {"Rate These Looks", "Looks"},
	"smarts":     {"Rate These Smarts", "Smarts"},
	"nerd":       {"Rate This Nerd", "Nerd"},
	"geek":       {"Rate This Geek", "Geek"},
}

func getRateTitleAndDesc(ratingName string, user string, score string) (string, string) {
	if ratingName == "schmeat" {
		size := helper.RangeIn(1, 15)
		schmeat := "C" + strings.Repeat("=", size) + "8"
		return "Schmeat Size", fmt.Sprintf("%s's Thang Thangin' \n%s", user, schmeat)
	}

	r, ok := standardRatings[ratingName]
	if !ok {
		return "", ""
	}
	return r.Title, fmt.Sprintf("%s's %s Score: %s/100", user, r.ScoreLabel, score)
}

func rateThisSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "rate-this",
		Description: "Rate this ...",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "I'll rate the pick too",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "dank", Value: "dank"},
					{Name: "epic-gamer", Value: "epic-gamer"},
					{Name: "gay", Value: "gay"},
					{Name: "geek", Value: "geek"},
					{Name: "looks", Value: "looks"},
					{Name: "neckbeard", Value: "neckbeard"},
					{Name: "nerd", Value: "nerd"},
					{Name: "pickme", Value: "pickme"},
					{Name: "schmeat", Value: "schmeat"},
					{Name: "simp", Value: "simp"},
					{Name: "smarts", Value: "smarts"},
					{Name: "stinky", Value: "stinky"},
					{Name: "thot", Value: "thot"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Rated user",
				Required:    false,
			},
		},
	}
}
