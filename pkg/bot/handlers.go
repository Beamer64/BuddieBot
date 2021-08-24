package bot

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/webScrape"
	"github.com/bwmarrin/discordgo"
	_ "github.com/knadh/go-get-youtube/youtube"
	"github.com/pkg/errors"
	"strings"
)

func (d *DiscordBot) messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	if strings.HasPrefix(message.Content, d.cfg.ExternalServicesConfig.BotPrefix) {
		if message.Author.ID == d.botID {
			return
		}

		messageSlices := strings.SplitAfterN(message.Content, " ", 2)
		command := strings.Trim(messageSlices[0], " ")

		param := ""
		if len(messageSlices) > 1 {
			param = messageSlices[1]
		}

		switch strings.ToLower(command) {
		case "$test":
			if d.memberHasRole(session, message, d.cfg.ExternalServicesConfig.BotAdminRole) {
				_, _ = session.ChannelMessageSend(message.ChannelID, "@Beamer64")
			}

		// Sends command list
		case "$tuuck":
			_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandDescriptions.Tuuck+"\n"+d.cfg.CommandDescriptions.McStatus+"\n"+d.cfg.CommandDescriptions.Start+
				"\n"+d.cfg.CommandDescriptions.Stop+"\n"+d.cfg.CommandDescriptions.Horoscope+"\n"+d.cfg.CommandDescriptions.Gif+"\n"+d.cfg.CommandDescriptions.Version+
				"\n"+d.cfg.CommandDescriptions.CoinFlip+"\n"+d.cfg.CommandDescriptions.LMGTFY)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		case "$version":
			_, err := session.ChannelMessageSend(message.ChannelID, "We'we wunnying vewsion `"+d.cfg.Version+"` wight nyow")
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
				_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.NotBotAdmin)
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
				_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.NotBotAdmin)
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
			/*// get the video object (with metdata)
			video, err := youtube.Get("FTl0tl9BGdc")
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}

			// download the video and write to file
			option := &youtube.Option{
				Rename: true,  // rename file using video title
				Resume: true,  // resume cancelled download
				Mp3:    true,  // extract audio to MP3
			}
			video.Download(0, "video.mp4", option)*/

			err := d.playYoutubeLink(session, message, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = session.ChannelMessageSend(message.ChannelID, "No vidya dood.")
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

		case "$insult":
			err := d.postInsult(session, message, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return

		// Sends the "Invalid" command Message
		default:
			_, err := session.ChannelMessageSend(message.ChannelID, d.cfg.CommandMessages.Invalid)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return
		}
	}
}

// GuildCreateHandler This function will be called every time a new guild is joined.
func (d *DiscordBot) guildCreateHandler(session *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	for _, member := range event.Guild.Members {
		if member.User.ID == d.botID {
			return
		}
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			gifURL, err := webScrape.RequestGif("I'm Here", d.cfg.ExternalServicesConfig.TenorAPIkey)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}

			_, err = session.ChannelMessageSend(channel.ID, gifURL)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return
		}
	}
}
