package slash

import (
	"fmt"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strconv"
	"strings"
)

func sendRateThisResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options[0]
	user := fmt.Sprintf("<@!%s>", i.Member.User.ID)

	if len(options.Options) == 1 {
		userName := options.Options[0].UserValue(s)
		user = fmt.Sprintf("<@!%s>", userName.ID)
	}

	embed, err := getRateThisEmbed(options.Name, user)
	if err != nil {
		go func() {
			err = helper.SendResponseError(s, i, "Unable to Rate atm, try again later.")
		}()
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

	return nil
}

func getRateThisEmbed(ratingName string, user string) (*discordgo.MessageEmbed, error) {
	score := strconv.Itoa(rand.Intn(100))
	rateTitle, rateDesc := getRateTitleAndDesc(ratingName, user, score)

	embed := &discordgo.MessageEmbed{
		Title:       rateTitle,
		Description: rateDesc,
		Color:       helper.RangeIn(1, 16777215),
	}

	return embed, nil
}

func getRateTitleAndDesc(ratingName string, user string, score string) (string, string) {
	switch ratingName {
	case "simp":
		return "Rate This Simp", fmt.Sprintf("%s's Simp Score: %s/100", user, score)
	case "dank":
		return "Dank Rating", fmt.Sprintf("%s's Dank Score: %s/100", user, score)
	case "epicgamer":
		return "Rate This Epic Gamer", fmt.Sprintf("%s's Epic Gamer Score: %s/100", user, score)
	case "gay":
		return "Gay Rating", fmt.Sprintf("%s's Gay Score: %s/100", user, score)
	case "schmeat":
		size := helper.RangeIn(1, 15)
		schmeat := "C" + strings.Repeat("=", size) + "8"
		return "Schmeat Size", fmt.Sprintf("%s's Thang Thangin' \n%s", user, schmeat)
	case "stinky":
		return "Rate This Stinky", fmt.Sprintf("%s's Stinky Score: %s/100", user, score)
	case "thot":
		return "Rate This Thot", fmt.Sprintf("%s's Thot Score: %s/100", user, score)
	case "pickme":
		return "Rate This Pick-Me", fmt.Sprintf("%s's Pick-Me Score: %s/100", user, score)
	case "neckbeard":
		return "Rate This Neck Beard", fmt.Sprintf("%s's Neck Beard Score: %s/100", user, score)
	case "looks":
		return "Rate These Looks", fmt.Sprintf("%s's Looks Score: %s/100", user, score)
	case "smarts":
		return "Rate These Smarts", fmt.Sprintf("%s's Smarts Score: %s/100", user, score)
	case "nerd":
		return "Rate This Nerd", fmt.Sprintf("%s's Nerd Score: %s/100", user, score)
	case "geek":
		return "Rate This Geek", fmt.Sprintf("%s's Geek Score: %s/100", user, score)
	default:
		return "", ""
	}
}
