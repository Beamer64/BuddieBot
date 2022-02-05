package commands

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
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
	ComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs){
		"horo-select": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendHoroscopeCompResponse(s, i)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
	}

	// CommandHandlers for handling the commands themselves. Main interaction response here.
	CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs){
		"version": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendVersionResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"coin-flip": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendCoinFlipResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"stop": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendStopResponse(s, i)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"clear": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendClearResponse(s, i)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendSkipResponse(s, i)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"queue": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendQueueResponse(s, i)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"animals": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendAnimalsResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"daily": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendDailyResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"pick": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendPickResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"tuuck": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendTuuckResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"play": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendPlayResponse(s, i)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},

		"insult": func(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) {
			err := sendInsultResponse(s, i, cfg)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
		},
	}
)
