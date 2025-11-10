package slash

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/api"
	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
	"github.com/mitchellh/mapstructure"
)

func sendPlayResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	client := cfg.Clients.Dagpi
	commandName := i.ApplicationCommandData().Options[0].Name
	errRespMsg := "Unable to fetch game atm, try again later."

	// Defer the interaction response to avoid timeout
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /get command %s: %w", commandName, err)
	}

	var webhookEdit *discordgo.WebhookEdit
	var embed *discordgo.MessageEmbed
	var err error

	switch commandName {
	case "coin-flip":
		embed, err = getCoinFlipEmbed(cfg)
		if err == nil {
			webhookEdit = &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}}
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

		webhookEdit = &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}}

	// todo finish this
	case "nim":
		/*err := games.SendNimEmbed(s, i, cfg)
		if err != nil {
			return err
		}*/

	// todo finish this
	case "typeracer":

	case "gtl": // todo: cmd not currently registered in slashCmds.go
		clientData, err := client.GTL()
		if err != nil {
			return err
		}

		embed, err = getGTLembed(clientData)
		if err == nil {
			webhookEdit = &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}}
		}

	case "wtp": // todo: cmd not currently registered in slashCmds.go
		clientData, err := client.WTP()
		if err != nil {
			return err
		}

		embed, err = getWTPembed(clientData, false)
		if err == nil {
			webhookEdit = &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}}
		}

	case "wyr":
		embed, err = getWYREmbed(cfg)
		if err == nil {
			webhookEdit = &discordgo.WebhookEdit{
				Components: &[]discordgo.MessageComponent{
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
				Embeds: &[]*discordgo.MessageEmbed{embed},
			}
		}

	default:
		return fmt.Errorf("unknown option: %s", commandName)
	}

	if err != nil {
		_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
		return fmt.Errorf("error in gameCmds.sendPlayResponse() : %w", err)
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
	/*fontsDir := "datasets/WRY.csv"
	if helper.IsLaunchedByDebugger() {
		fontsDir = "../../datasets/text_fonts.json"
	}*/

	file, err := os.Open(cfg.Configs.ReqFileDirs.Datasets + "WYR.csv")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)

	// Read all records (assuming first row is headers)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) <= 1 {
		return nil, fmt.Errorf("no data rows found in CSV: %w", err)
	}

	// Parse rows into a slice of Poll structs
	var polls []wyrPoll
	for _, row := range records[1:] { // skip header
		if len(row) < 4 {
			continue
		}

		votesA, _ := strconv.Atoi(row[1])
		votesB, _ := strconv.Atoi(row[3])

		polls = append(
			polls, wyrPoll{
				OptionA: row[0],
				VotesA:  votesA,
				OptionB: row[2],
				VotesB:  votesB,
			},
		)
	}

	// Select a random poll
	randomIndex := rand.Intn(len(polls))
	randomPoll := polls[randomIndex]

	embed := &discordgo.MessageEmbed{
		Title:       "Would You Rather?",
		Color:       helper.RangeIn(1, 16777215),
		Description: fmt.Sprintf("%s OR %s", randomPoll.OptionA, randomPoll.OptionB),
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
