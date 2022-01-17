package events

import (
	"fmt"
	"github.com/beamer64/discordBot/pkg/config"
	"github.com/beamer64/discordBot/pkg/voice_chat"
	"github.com/beamer64/discordBot/pkg/web_scrape"
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
					_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
				}
			}
			return

		// Sends command list
		case "$tuuck":
			err := d.sendHelpMessage(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		case "$version":
			if d.memberHasRole(s, m, d.cfg.Configs.Settings.BotAdminRole) {
				_, err := s.ChannelMessageSend(m.ChannelID, "We'we wunnying vewsion `"+d.cfg.Version+"` wight nyow")
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			}
			return

		// Starts the Minecraft Server
		case "$startServer":
			if d.memberHasRole(s, m, d.cfg.Configs.Settings.BotAdminRole) {
				err := d.startServer(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
				}
				return

			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			}

		// Stops the Minecraft Server
		case "$stopServer":
			if d.memberHasRole(s, m, d.cfg.Configs.Settings.BotAdminRole) {
				err := d.stopServer(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
				}
				return

			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.NotBotAdmin)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
				}
			}

		// Stops the Minecraft Server
		case "$mcstatus":
			err := d.sendServerStatusAsMessage(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// Plays youtube link in voice chat
		case "$play":
			err := d.playYoutubeLink(s, m, param)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, "No vidya dood.")
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

			// stops audio playback
		case "$stop":
			err := d.stopYoutubeLink()
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			err = web_scrape.RunMpFileCleanUp()
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			return

			// shows queued songs
		case "$queue":
			err := d.getSongQueue(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

			// clears song queue
		case "$clear":
			err := d.clearSongQueue()
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

			// plays next song in queue
		case "$skip":
			err := d.stopYoutubeLink()
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			/*web_scrape.MpFileQueue = append(web_scrape.MpFileQueue[:0], web_scrape.MpFileQueue[1:]...)*/

			dgv, err := voice_chat.ConnectVoiceChannel(s, m, m.GuildID, d.cfg.Configs.DiscordIDs.ErrorLogChannelID)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			err = web_scrape.PlayAudioFile(dgv, "", m.ChannelID, s)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}

			return

		// Sends daily horoscope
		case "$horoscope":
			err := d.displayHoroscope(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// Sends gif response
		case "$gif":
			err := d.sendGif(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// Flips a coins, sends gif for results
		case "$coinflip":
			err := d.coinFlip(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		case "$insult":
			err := d.postInsult(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// Play Nim game
		case "$nim":
			err := d.playNIM(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSend(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, fmt.Sprintf("%+v", errors.WithStack(err)))
			}
			return

		// Sends the "Invalid" command Message
		default:
			_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.Invalid)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
			}
			return
		}
	}
}
