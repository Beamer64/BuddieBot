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
			Name:        "horoscope",
			Description: "Gives daily horoscope",
		},
		{
			Name:        "tuuck",
			Description: "I've fallen and can't get up!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "Specify a command for a description",
					Required:    false,
				},
			},
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
			Name:        "insult",
			Description: "Tag and insult a user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Tag the user to be insulted",
					Required:    true,
				},
			},
		},
	}

	ComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs){
		"horo-select": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			sign := i.MessageComponentData().Values[0]
			embed, err := getHoroscopeEmbed(sign)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			msgEdit := discordgo.NewMessageEdit(i.ChannelID, i.Message.ID)
			msgContent := ""
			msgEdit.Content = &msgContent
			msgEdit.Embeds = []*discordgo.MessageEmbed{embed}

			// edit response (i.Interaction) and replace with embed
			_, err = s.ChannelMessageEditComplex(msgEdit)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

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
					_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
				}
			}
		},
	}

	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs){
		"version": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			embed := &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("Version: %s", cfg.Version),
				Color:       62033,
				Description: "You see it up there.",
			}

			err := s.InteractionRespond(
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
		"coin-flip": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			embed, err := getCoinFlipEmbed(cfg)
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
		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			err := web_scrape.RunMpFileCleanUp(fmt.Sprintf("%s/Audio/", i.GuildID), cfg.Configs.DiscordIDs.ErrorLogChannelID, s)
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
		"horoscope": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			err := s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Choose a zodiac sign",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.SelectMenu{
										CustomID:    "horo-select",
										Placeholder: "Zodiac",
										Options: []discordgo.SelectMenuOption{
											{
												Label:   "Aquarius",
												Value:   "aquarius",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "🌊"},
											},
											{
												Label:   "Aries",
												Value:   "aries",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "🐏"},
											},
											{
												Label:   "Cancer",
												Value:   "cancer",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "🦀"},
											},
											{
												Label:   "Capricorn",
												Value:   "capricorn",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "🐐"},
											},
											{
												Label:   "Gemini",
												Value:   "gemini",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "♊"},
											},
											{
												Label:   "Leo",
												Value:   "leo",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "🦁"},
											},
											{
												Label:   "Libra",
												Value:   "libra",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "⚖️"},
											},
											{
												Label:   "Pisces",
												Value:   "pisces",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "🐠"},
											},
											{
												Label:   "Sagittarius",
												Value:   "sagittarius",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "🏹"},
											},
											{
												Label:   "Scorpio",
												Value:   "scorpio",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "🦂"},
											},
											{
												Label:   "Taurus",
												Value:   "taurus",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "🐃"},
											},
											{
												Label:   "Virgo",
												Value:   "virgo",
												Default: false,
												Emoji:   discordgo.ComponentEmoji{Name: "♍"},
											},
										},
									},
								},
							},
						},
					},
				},
			)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
		},
		"tuuck": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			cmd := ""
			if i.ApplicationCommandData().Options != nil {
				cmd = i.ApplicationCommandData().Options[0].StringValue()
			} else {
				cmd = ""
			}
			embed := getTuuckEmbed(cmd, cfg)

			err := s.InteractionRespond(
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
		"insult": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			user := i.ApplicationCommandData().Options[0].UserValue(s)
			randColorInt := rangeIn(1, 16777215)
			embed, err := api.GetInsultEmbed(randColorInt, cfg)
			if err != nil {
				_, _ = s.ChannelMessageSend(cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			err = s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("An ode to: <@%s>", user.ID),
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
	}
)
