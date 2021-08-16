package bot

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"strings"
)

type DiscordBot struct {
	cfg   *config.Config
	botID string
}

func NewDiscordBot(cfg *config.Config) *DiscordBot {
	return &DiscordBot{
		cfg: cfg,
	}
}

func (d *DiscordBot) Start() error {
	goBot, err := discordgo.New("Bot " + d.cfg.ExternalServicesConfig.Token)
	if err != nil {
		return err
	}

	user, err := goBot.User("@me")
	if err != nil {
		return err
	}
	d.botID = user.ID

	goBot.AddHandler(d.messageHandler)
	err = goBot.Open()
	if err != nil {
		return err
	}

	fmt.Println("DiscordBot is running!")
	return nil
}

func (d *DiscordBot) messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	if strings.HasPrefix(message.Content, d.cfg.ExternalServicesConfig.BotPrefix) {
		if message.Author.ID == d.botID {
			return
		}

		messageSlices := strings.Split(message.Content, " ")
		command := messageSlices[0]

		param := ""
		if len(messageSlices) > 1 {
			param = messageSlices[1]
		}

		// deletes Justins Messages
		/*if message.Author.ID == "282722418093719556" {
			err := session.ChannelMessageDelete(message.ChannelID, message.ID)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				return
			}
			return
		}*/

		switch strings.ToLower(command) {
		// Sends command list
		case "$tuuck":
			err := d.sendMessage(session, message, d.cfg.CommandDescriptions.Tuuck+"\n"+d.cfg.CommandDescriptions.McStatus+"\n"+d.cfg.CommandDescriptions.Start+
				"\n"+d.cfg.CommandDescriptions.Stop+"\n"+d.cfg.CommandDescriptions.Horoscope+"\n"+d.cfg.CommandDescriptions.Gif+"\n"+d.cfg.CommandDescriptions.Version+
				"\n"+d.cfg.CommandDescriptions.CoinFlip+"\n"+d.cfg.CommandDescriptions.LMGTFY)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		case "$version":
			err := d.sendMessage(session, message, "We'we wunnying vewsion `"+d.cfg.Version+"` wight nyow")
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}

		// Starts the Minecraft Server
		case "$start":
			if d.memberHasRole(session, message, d.cfg.ExternalServicesConfig.BotAdminRole) {
				err := d.startServer(session, message)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
				return

			} else {
				err := d.sendMessage(session, message, d.cfg.CommandMessages.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			}

		// Stops the Minecraft Server
		case "$stop":
			if d.memberHasRole(session, message, d.cfg.ExternalServicesConfig.BotAdminRole) {
				err := d.stopServer(session, message)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
				return

			} else {
				err := d.sendMessage(session, message, d.cfg.CommandMessages.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			}

		// Stops the Minecraft Server
		case "$mcstatus":
			err := d.sendServerStatusAsMessage(session, message)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// TODO make this work
		// Plays youtube link in voice chat
		case "$play":
			err := d.playYoutubeLink(session, message, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_ = d.sendMessage(session, message, "No vidya dood.")
			}
			return

		// Sends daily horoscope
		case "$horoscope":
			err := d.displayHoroscope(session, message, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Sends gif response
		case "$gif":
			err := d.sendGif(session, message, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Flips a coins, sends gif for results
		case "$coinflip":
			err := d.coinFlip(session, message)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Creates/sends LMGTFY link for marked msgs
		case "$lmgtfy":
			err := d.sendLmgtfy(session, message)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Sends the "Invalid" command Message
		default:
			err := d.sendMessage(session, message, d.cfg.CommandMessages.Invalid)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return
		}
	}
}
