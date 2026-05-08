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
	options := i.ApplicationCommandData().Options[0]
	user := fmt.Sprintf("<@!%s>", i.Member.User.ID)

	if len(options.Options) == 1 {
		userName := options.Options[0].UserValue(s)
		user = fmt.Sprintf("<@!%s>", userName.ID)
	}

	embed, err := getRateThisEmbed(options.Name, user)
	if err != nil {
		return helper.ReturnUserError(s, i, "Unable to Rate atm, try again later.", err)
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
	"simp":      {"Rate This Simp", "Simp"},
	"dank":      {"Dank Rating", "Dank"},
	"epicgamer": {"Rate This Epic Gamer", "Epic Gamer"},
	"gay":       {"Gay Rating", "Gay"},
	"stinky":    {"Rate This Stinky", "Stinky"},
	"thot":      {"Rate This Thot", "Thot"},
	"pickme":    {"Rate This Pick-Me", "Pick-Me"},
	"neckbeard": {"Rate This Neck Beard", "Neck Beard"},
	"looks":     {"Rate These Looks", "Looks"},
	"smarts":    {"Rate These Smarts", "Smarts"},
	"nerd":      {"Rate This Nerd", "Nerd"},
	"geek":      {"Rate This Geek", "Geek"},
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
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "dank",
				Description: "Dank Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Dank score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "epicgamer",
				Description: "Epic Gamer Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Epic Gamer score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "gay",
				Description: "Gay Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Gay score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "geek",
				Description: "Geek Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Geek score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "looks",
				Description: "Looks Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Looks score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "neckbeard",
				Description: "Neck Beard Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Neck Beard score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "nerd",
				Description: "Nerd Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Nerd score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "pickme",
				Description: "Pick Me Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Pick Me score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "schmeat",
				Description: "Schmeat Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Schmeat score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "simp",
				Description: "Simp Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User simp score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "smarts",
				Description: "Smarts Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Smarts score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "stinky",
				Description: "Stinky Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Stinky score",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "thot",
				Description: "Thot Rating",
				Required:    false,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "User Thot score",
						Required:    false,
					},
				},
			},
		},
	}
}
