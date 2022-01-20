package commands

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
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
				return
			}
		},
	}
)
