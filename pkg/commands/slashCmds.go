package commands

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/api"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/web"
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
			Name:        "animals",
			Description: "So CUTE",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "doggo",
					Description: "🐕",
					Required:    false,
				},
			},
		},
		{
			Name:        "daily",
			Description: "Receive daily quotes, horoscopes, affirmations, etc.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "horoscope",
					Description: "Gives daily horoscope",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "affirmation",
					Description: "Gives daily affirmation",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "kanye",
					Description: "Gifts us with a quote from the man himself",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "advice",
					Description: "Words of wisdom",
					Required:    false,
				},
			},
		},
		{
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
					Name:        "steam",
					Description: "Will choose a random Steam game to play.",
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
	}

	// ComponentHandlers for handling components in interactions. Eg. Buttons, Dropdowns, Searchbars Etc.
	ComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs){
		"horo-select": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			sign := i.MessageComponentData().Values[0]
			embed, err := getHoroscopeEmbed(sign)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}

			msgEdit := discordgo.NewMessageEdit(i.ChannelID, i.Message.ID)
			msgContent := ""
			msgEdit.Content = &msgContent
			msgEdit.Embeds = []*discordgo.MessageEmbed{embed}

			// edit response (i.Interaction) and replace with embed
			_, err = s.ChannelMessageEditComplex(msgEdit)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
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
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}
			}
		},
	}

	// CommandHandlers for handling the commands themselves. Main interaction response here.
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
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
		"coin-flip": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			embed, err := getCoinFlipEmbed(cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
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
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
		"stop": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			err := stopAudioPlayback()
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
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
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			err := web.RunMpFileCleanUp(fmt.Sprintf("%s/Audio", i.GuildID))
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
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
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			audio := ""
			if len(web.MpFileQueue) > 0 {
				audio = fmt.Sprintf("Skipping %s", web.MpFileQueue[0])
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
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}

			err = skipPlayback(s, i)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
		"queue": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			queue := ""
			if len(web.MpFileQueue) > 0 {
				queue = strings.Join(web.MpFileQueue, "\n")
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
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
		"animals": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			switch i.ApplicationCommandData().Options[0].Name {
			case "doggo":
				embed, err := getDoggoEmbed(cfg)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
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
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}
			}
		},
		"daily": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			switch i.ApplicationCommandData().Options[0].Name {
			case "advice":
				embed, err := getAdviceEmbed(cfg)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
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
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}

			case "kanye":
				embed, err := getKanyeEmbed(cfg)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
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
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}

			case "affirmation":
				embed, err := getAffirmationEmbed(cfg)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
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
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}

			case "horoscope":
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
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}
			}
		},
		"pick": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			embed, err := getPickEmbed(i.ApplicationCommandData().Options, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}

			content := ""
			switch strings.ToLower(i.ApplicationCommandData().Options[0].Name) {
			case "steam":
				content = "Choosing Steam Game"
			case "choices":
				for _, v := range i.ApplicationCommandData().Options[0].Options {
					content = content + fmt.Sprintf("[%s] ", v.StringValue())
				}
				content = strings.TrimSpace(content)
				content = fmt.Sprintf("*%s*", content)
			}

			err = s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: content,
						Embeds: []*discordgo.MessageEmbed{
							embed,
						},
					},
				},
			)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
		"tuuck": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			cmd := ""
			if len(i.ApplicationCommandData().Options) > 0 {
				cmd = i.ApplicationCommandData().Options[0].StringValue()
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
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
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
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}

			err = playYoutubeLink(s, i, link)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
		"insult": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.ConfigStructs) {
			user := i.ApplicationCommandData().Options[0].UserValue(s)
			randColorInt := rangeIn(1, 16777215)
			embed, err := api.GetInsultEmbed(randColorInt, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
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
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
	}
)
