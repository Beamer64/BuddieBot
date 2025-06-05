package slash

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/buddieBot/pkg/api"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"github.com/mitchellh/mapstructure"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func sendPlayResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	client := cfg.Clients.Dagpi
	options := i.ApplicationCommandData().Options[0]

	var embed *discordgo.MessageEmbed
	var data *discordgo.InteractionResponseData
	var err error

	switch options.Name {
	case "coin-flip":
		embed, err = getCoinFlipEmbed(cfg)
		if err != nil {
			go func() {
				err = helper.SendResponseErrorToUser(s, i, "Unable to flip coin atm, try again later.")
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "just-lost":
		embed = &discordgo.MessageEmbed{
			Title:       "You just lost The Game.",
			Color:       helper.RangeIn(1, 16777215),
			Description: "..Told you not to play.",
		}

		channel, err := s.UserChannelCreate(i.Member.User.ID)
		if err != nil {
			return err
		}

		_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
		if err != nil {
			return err
		}

		embed = &discordgo.MessageEmbed{
			Title: "Check your DM's  👀",
			Color: helper.RangeIn(1, 16777215),
		}

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	// todo finish this
	case "nim":
		/*err := games.SendNimEmbed(s, i, cfg)
		if err != nil {
			return err
		}*/

	// todo finish this
	case "typeracer":

	case "gtl":
		clientData, err := client.GTL()
		if err != nil {
			return err
		}

		embed, err = getGTLembed(clientData)
		if err != nil {
			go func() {
				err = helper.SendResponseErrorToUser(s, i, "Unable to fetch game atm, try again later.")
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "wtp":
		clientData, err := client.WTP()
		if err != nil {
			return err
		}

		embed, err = getWTPembed(clientData, false)
		if err != nil {
			go func() {
				err = helper.SendResponseErrorToUser(s, i, "Unable to fetch game atm, try again later.")
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}

	case "wyr":
		embed, err = getWYREmbed(cfg)
		if err != nil {
			go func() {
				err = helper.SendResponseErrorToUser(s, i, "Unable to fetch game atm, try again later.")
			}()
			return err
		}

		data = &discordgo.InteractionResponseData{
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Another One! (▀̿Ĺ̯▀̿ ̿)",
							Style:    1,
							CustomID: "wyr-button",
						},
					},
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		}
	}

	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		},
	)
	if err != nil {
		return fmt.Errorf("error sendind Interaction: %v", err)
	}

	return nil
}

func getGTLembed(data interface{}) (*discordgo.MessageEmbed, error) {
	var gtlObj gtl
	err := mapstructure.Decode(data, &gtlObj)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Clue",
				Value:  gtlObj.Clue,
				Inline: false,
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: gtlObj.Question,
		},
	}

	return embed, nil
}

func getWYREmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	res, err := http.Get(cfg.Configs.ApiURLs.WYRAPI)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call failed with status code %d", res.StatusCode)
	}

	var wyrObj wyr
	err = json.NewDecoder(res.Body).Decode(&wyrObj)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(res.Body)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Would You Rather?",
		Color:       helper.RangeIn(1, 16777215),
		Description: wyrObj.Data,
	}

	return embed, nil
}

func getWTPembed(data interface{}, isAnswer bool) (*discordgo.MessageEmbed, error) {
	var wtpObj wtp
	err := mapstructure.Decode(data, &wtpObj)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{}

	if isAnswer {

	} else {
		embed = &discordgo.MessageEmbed{
			Image: &discordgo.MessageEmbedImage{
				URL: wtpObj.Question,
			},
		}
	}

	return embed, nil
}

func getCoinFlipEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	x1 := rand.NewSource(time.Now().UnixNano())
	y1 := rand.New(x1)
	randNum := y1.Intn(200)

	search, results := "", ""
	if randNum%2 == 0 {
		search = "Coin Flip Heads"
		results = "Heads"

	} else {
		search = "Coin Flip Tails"
		results = "Tails"
	}

	gifURL, err := api.RequestGifURL(search, cfg.Configs.Keys.TenorAPIkey)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Flipping...",
		Description: fmt.Sprintf("It's %s!", results),
		Color:       helper.RangeIn(1, 16777215),
		Image: &discordgo.MessageEmbedImage{
			URL: gifURL,
		},
	}

	return embed, nil
}

func sendWYRCompResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	embed, err := getWYREmbed(cfg)
	if err != nil {
		go func() {
			err = helper.SendResponseErrorToUser(s, i, "Unable to fetch WYR atm, try again later.")
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
