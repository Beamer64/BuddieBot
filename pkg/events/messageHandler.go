package events

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/web"
	"github.com/pkg/errors"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type MessageCreateHandler struct {
	cfg   *config.Configs
	botID string
}

func NewMessageCreateHandler(cfg *config.Configs, u *discordgo.User) *MessageCreateHandler {
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
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

		case "$release":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				if d.memberHasRole(s, m, d.cfg.Configs.Settings.BotAdminRole) {
					err := d.sendReleaseNotes(s, m)
					if err != nil {
						fmt.Printf("%+v", errors.WithStack(err))
						_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
					}
				}
			}
			return

		// Starts the Minecraft Server
		case "$startServer":
			if d.memberHasRole(s, m, d.cfg.Configs.Settings.BotAdminRole) {
				err := d.startServer(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}
				return

			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}
			}

		// Stops the Minecraft Server
		case "$stopServer":
			if d.memberHasRole(s, m, d.cfg.Configs.Settings.BotAdminRole) {
				err := d.stopServer(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}
				return

			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}
			}

		// Stops the Minecraft Server
		case "$serverStatus":
			err := d.sendServerStatusAsMessage(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
			}
			return

			/////////////Games///////////////////

		// Play Nim game
		case "$nim":
			err := d.playNIM(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
			}
			return

			/////////////Audio///////////////////

			// Play audio
		case "$play":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				err := d.playAudio(s, m, param)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

		case "$stop":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				err := d.stopAudioPlayback()
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

		case "$queue":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				err := d.sendQueue(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

		case "$skip":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				err := d.sendSkipMessage(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

		case "$clear":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				err := web.MpFileCleanUp(fmt.Sprintf("%s/Audio", m.GuildID))
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}

				_, err = s.ChannelMessageSend(m.ChannelID, "\"This house is clean.\"")
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

			/////////////NSFW///////////////////

		// Sends the "Invalid" command Message
		default:
			_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.Invalid)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, config.GetErrorEmbed(err, s, m.GuildID))
			}
			return
		}
	}
}
