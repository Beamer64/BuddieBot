package slash

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"

	"github.com/Beamer64/BuddieBot/pkg/api"
	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func sendPlayResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
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
			Color:       helper.RandomDiscordColor(),
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
			Color: helper.RandomDiscordColor(),
		}

		webhookEdit = &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}}

	case "wyr":
		webhookEdit, err = getWYRwebhook(cfg)

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

func getWYRvotesWebhook(cfg *config.Configs, customID string) (*discordgo.WebhookEdit, error) {
	voteA := ""
	voteB := ""

	re := regexp.MustCompile(`L(\d+)-R(\d+)|R(\d+)-L(\d+)`)
	match := re.FindStringSubmatch(customID)
	if len(match) != 0 {
		if match[1] != "" { // L first
			voteA = match[1]
			voteB = match[2]
		} else if match[3] != "" { // R first
			voteB = match[3]
			voteA = match[4]
		}
	}

	webhookEdit := &discordgo.WebhookEdit{
		Components: &[]discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    voteA + "%",
						Emoji:    &discordgo.ComponentEmoji{Name: "👈"},
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("wyr-votes:%v", voteA),
						Disabled: true,
					},
					discordgo.Button{
						Label:    voteB + "%",
						Emoji:    &discordgo.ComponentEmoji{Name: "👉"},
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("wyr-votes:%v", voteB),
						Disabled: true,
					},
					discordgo.Button{
						Label:    "Reroll",
						Style:    discordgo.SecondaryButton,
						CustomID: "wyr-reroll",
					},
				},
			},
		},
	}

	return webhookEdit, nil
}

// wyrVotePercents returns the rounded percentage split for a WYR poll's two
// option vote counts. Extracted for unit testing.
func wyrVotePercents(votesA, votesB int) (percentA, percentB float64) {
	sum := votesA + votesB
	if sum == 0 {
		return 0, 0
	}
	percentA = math.Round((float64(votesA) / float64(sum)) * 100)
	percentB = math.Round((float64(votesB) / float64(sum)) * 100)
	return percentA, percentB
}

func getWYRwebhook(cfg *config.Configs) (*discordgo.WebhookEdit, error) {
	if len(wyrPolls) == 0 {
		return nil, errors.New("WYR polls not loaded")
	}

	randomPoll := wyrPolls[rand.Intn(len(wyrPolls))]
	percentA, percentB := wyrVotePercents(randomPoll.VotesA, randomPoll.VotesB)

	embed := &discordgo.MessageEmbed{
		Title: "Would You Rather?",
		Color: helper.RandomDiscordColor(),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Option 1",
				Value:  randomPoll.OptionA,
				Inline: true,
			},
			{
				Name:   "|",
				Value:  "| \n | \n |",
				Inline: true,
			},
			{
				Name:   "Option 2",
				Value:  randomPoll.OptionB,
				Inline: true,
			},
		},
	}

	webhookEdit := &discordgo.WebhookEdit{
		Components: &[]discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Left",
						Emoji:    &discordgo.ComponentEmoji{Name: "👈"},
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("wyr-votes:L%v-R%v", percentA, percentB),
					},
					discordgo.Button{
						Label:    "Right",
						Emoji:    &discordgo.ComponentEmoji{Name: "👉"},
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("wyr-votes:R%v-L%v", percentB, percentA),
					},
					discordgo.Button{
						Label:    "Reroll",
						Style:    discordgo.SecondaryButton,
						CustomID: "wyr-reroll",
					},
				},
			},
		},
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}

	return webhookEdit, nil
}

func getCoinFlipEmbed(cfg *config.Configs) (*discordgo.MessageEmbed, error) {
	randNum := rand.Intn(200)

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
		Color:       helper.RandomDiscordColor(),
		Image: &discordgo.MessageEmbedImage{
			URL: gifURL,
		},
	}

	return embed, nil
}

func sendWYRvotesResp(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	customID := i.MessageComponentData().CustomID

	webhookEdit, err := getWYRvotesWebhook(cfg, customID)
	if err != nil {
		_ = helper.SendResponseErrorToUser(s, i, "Unable to fetch WYR atm, try again later.")
		return err
	}

	_, err = s.ChannelMessageEditComplex(
		&discordgo.MessageEdit{
			ID:         i.Message.ID, // message ID of the original message with the buttons
			Channel:    i.ChannelID,  // channel where the message is
			Content:    new(string),  // new text or nil if only editing embeds
			Components: webhookEdit.Components,
		},
	)

	// respond with an update to acknowledge interaction
	err = s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		},
	)
	if err != nil {
		_ = helper.SendResponseErrorToUser(s, i, "Unable to fetch WYR atm, try again later.")
		return err
	}

	return nil
}

func sendWYRrerollResp(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	errRespMsg := "Unable to make call at this moment, please try later :("

	// Defer the interaction response to avoid timeout
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /get command WYR: %w", err)
	}

	webhookEdit, err := getWYRwebhook(cfg)
	if err != nil {
		_ = helper.SendResponseErrorToUser(s, i, "Unable to fetch WYR atm, try again later.")
		return err
	}

	if _, err = s.InteractionResponseEdit(
		i.Interaction, webhookEdit,
	); err != nil {
		_ = helper.SendResponseErrorToUser(s, i, errRespMsg)
		return fmt.Errorf("failed to send message for command WYR: %w", err)
	}

	return nil
}

func playSpec() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "play",
		Description: "Play some games! *More coming soon",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "coin-flip",
				Description: "Flips a coin...",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "just-lost",
				Description: "Don't play this..",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "wyr",
				Description: "Would You Rather??",
				Required:    false,
			},
		},
	}
}
