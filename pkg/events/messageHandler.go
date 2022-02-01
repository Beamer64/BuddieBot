package events

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/pkg/errors"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type MessageCreateHandler struct {
	cfg   *config.ConfigStructs
	botID string
}

func NewMessageCreateHandler(cfg *config.ConfigStructs, u *discordgo.User) *MessageCreateHandler {
	return &MessageCreateHandler{
		cfg:   cfg,
		botID: u.ID,
	}
}

func (d *MessageCreateHandler) MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, d.cfg.Configs.Settings.BotPrefix) {
		if m.Author.ID == d.botID {
			return
		}

		messageSlices := strings.SplitAfterN(m.Content, " ", 2)
		command := strings.Trim(messageSlices[0], " ")

		param := ""
		if len(messageSlices) > 1 {
			param = messageSlices[1]
		}

		switch strings.ToLower(command) {
		case "$test":
			if d.memberHasRole(s, m, d.cfg.Configs.Settings.BotAdminRole) {
				err := d.testMethod(s, m, param)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}
			}
			return

		// Starts the Minecraft Server
		case "$startServer":
			if d.memberHasRole(s, m, d.cfg.Configs.Settings.BotAdminRole) {
				err := d.startServer(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}
				return

			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}
			}

		// Stops the Minecraft Server
		case "$stopServer":
			if d.memberHasRole(s, m, d.cfg.Configs.Settings.BotAdminRole) {
				err := d.stopServer(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}
				return

			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
				}
			}

		// Stops the Minecraft Server
		case "$serverStatus":
			err := d.sendServerStatusAsMessage(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
			return

		// Play Nim game
		case "$nim":
			err := d.playNIM(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
			return

		// Sends the "Invalid" command Message
		default:
			_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.Invalid)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, getErrorEmbed(err))
			}
			return
		}
	}
}
