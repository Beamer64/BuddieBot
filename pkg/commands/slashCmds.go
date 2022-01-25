package commands

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/api"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/web_scrape"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"strings"
)

var (
	// Commands All commands and options must have a description
	// Commands/options without description will fail the registration
	// of the command.
	Commands = []*discordgo.ApplicationCommand{
		{
			Name:        "version",
			Description: "Gives the Git SHA of the current bot version running",
		},
		{
			Name:        "coin-flip",
			Description: "Flips a coin...",
		},
		{
			Name:        "stop",
			Description: "Stops audio playback",
		},
		{
			Name:        "clear",
			Description: "Clear the audio queue",
		},
		{
			Name:        "skip",
			Description: "Skips current audio in queue",
		},
		{
			Name:        "queue",
			Description: "Gets the current song queue",
		},
		{
			Name:        "play",
			Description: "Play audio from a YouTube link",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "link",
					Description: "Link to YouTube video",
					Required:    true,
				},
			},
		},
		{
			Name:        "horoscope",
			Description: "Gives daily horoscope",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "zodiac",
					Description: "Name of the zodiac sign to return",
					Required:    true,
				},
			},
		},
		{
			Name:        "insult",
			Description: "Tag and insult a user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user",
					Description: "Tag the user to be insulted",
					Required:    true,
				},
			},
		},
	}

	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs){
		"version": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			err := s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("We'we wunnying vewsion `%s` wight nyow", cfg.Version),
					},
				},
			)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
		},
		"coin-flip": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			embed, err := coinFlip(cfg)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
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
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
		},
		"stop": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			err := stopAudioPlayback()
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			err = s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Okay Dad",
					},
				},
			)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
		},
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			audio := ""
			if len(web_scrape.MpFileQueue) > 0 {
				audio = fmt.Sprintf("Skipping %s", web_scrape.MpFileQueue[0])
			} else {
				audio = "Queue is empty, my guy"
			}
			err := s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: audio,
					},
				},
			)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			err = skipPlayback(s, i)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
		},
		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			err := web_scrape.RunMpFileCleanUp()
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			err = s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "This house is clean.",
					},
				},
			)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
		},
		"queue": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			queue := ""
			if len(web_scrape.MpFileQueue) > 0 {
				queue = strings.Join(web_scrape.MpFileQueue, "\n")
			} else {
				queue = "Uh owh, song queue is wempty (>.<)"
			}

			err := s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: queue,
					},
				},
			)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
		},
		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			link := i.ApplicationCommandData().Options[0].StringValue()
			err := s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Playing: %s", link),
					},
				},
			)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			err = playYoutubeLink(s, i, link)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
		},
		"horoscope": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			sign := i.ApplicationCommandData().Options[0].StringValue()
			horoscope, err := web_scrape.ScrapeSign(sign)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			err = s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: horoscope,
					},
				},
			)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
		},
		"insult": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			user := i.ApplicationCommandData().Options[0].StringValue()
			insult, err := api.PostInsult(user, cfg)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			err = s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: insult,
					},
				},
			)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
		},
	}
)
