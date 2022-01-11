package events

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/pkg/errors"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type MessageHandler struct {
	cfg   *config.Config
	botID string
}

func NewMessageHandler(cfg *config.Config, u *discordgo.User) *MessageHandler {
	return &MessageHandler{
		cfg:   cfg,
		botID: u.ID,
	}
}

func (d *MessageHandler) Handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, d.cfg.ExternalServicesConfig.BotPrefix) {
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
			if d.memberHasRole(s, m, d.cfg.ExternalServicesConfig.BotAdminRole) {
				//d.testMethod(s, m)
			}
			return

		// Sends command list
		case "$tuuck":
			err := d.sendHelpMessage(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		case "$version":
			if d.memberHasRole(s, m, d.cfg.ExternalServicesConfig.BotAdminRole) {
				_, err := s.ChannelMessageSend(m.ChannelID, "We'we wunnying vewsion `"+d.cfg.Version+"` wight nyow")
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.CommandMessages.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			}
			return

		// Starts the Minecraft Server
		case "$startServer":
			if d.memberHasRole(s, m, d.cfg.ExternalServicesConfig.BotAdminRole) {
				err := d.startServer(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
				}
				return

			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.CommandMessages.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			}

		// Stops the Minecraft Server
		case "$stopServer":
			if d.memberHasRole(s, m, d.cfg.ExternalServicesConfig.BotAdminRole) {
				err := d.stopServer(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
				}
				return

			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.CommandMessages.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			}

		// Stops the Minecraft Server
		case "$mcstatus":
			err := d.sendServerStatusAsMessage(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// TODO make this work
		// Plays youtube link in voice chat
		case "$play":
			err := d.playYoutubeLink(s, m, param)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, "No vidya dood.")
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

			// stops youtube link in voice chat
		case "$stop":
			if d.memberHasRole(s, m, d.cfg.ExternalServicesConfig.BotAdminRole) {
				err := d.stopYoutubeLink()
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
				}
			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.CommandMessages.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			}
			return

		// Sends daily horoscope
		case "$horoscope":
			err := d.displayHoroscope(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// Sends gif response
		case "$gif":
			err := d.sendGif(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// Flips a coins, sends gif for results
		case "$coinflip":
			err := d.coinFlip(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// Creates/sends LMGTFY link for marked msgs
		case "$lmgtfy":
			err := d.sendLmgtfy(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		case "$insult":
			err := d.postInsult(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// Play Nim game
		case "$nim":
			err := d.playNIM(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.ExternalServicesConfig.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// Sends the "Invalid" command Message
		default:
			_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.CommandMessages.Invalid)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return
		}
	}
}
