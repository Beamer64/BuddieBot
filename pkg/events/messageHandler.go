package events

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/database"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/beamer64/buddieBot/pkg/web"
	"github.com/pkg/errors"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type MessageCreateHandler struct {
	cfg      *config.Configs
	botID    string
	dbClient *dynamodb.DynamoDB
}

func NewMessageCreateHandler(cfg *config.Configs, u *discordgo.User, dbc *dynamodb.DynamoDB) *MessageCreateHandler {
	return &MessageCreateHandler{
		cfg:      cfg,
		botID:    u.ID,
		dbClient: dbc,
	}
}

func (d *MessageCreateHandler) MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, d.cfg.Configs.Settings.BotPrefix) {
		if m.Author.ID == d.botID {
			return
		}

		messageSlices := strings.SplitAfterN(m.Content, " ", 2)                                         //split command from param
		command := strings.Trim(messageSlices[0], fmt.Sprintf("%s ", d.cfg.Configs.Settings.BotPrefix)) //trim spaces and prefix

		//get command parameter
		param := ""
		if len(messageSlices) > 1 {
			param = messageSlices[1]
		}

		switch strings.ToLower(command) {

		/////////////Dev///////////////////
		case "test":
			err := d.testMethod(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
			}
			return

		case "release":
			if m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				if helper.MemberHasRole(s, m.Member, m.GuildID, d.cfg.Configs.Settings.BotAdminRole) {
					err := d.sendReleaseNotes(s, m)
					if err != nil {
						fmt.Printf("%+v", errors.WithStack(err))
						_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
					}
				} else {
					_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.NotBotAdmin)
					if err != nil {
						fmt.Printf("%+v", errors.WithStack(err))
						_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
					}
				}
			}
			return

		case "updatedbitems":
			if m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				if helper.MemberHasRole(s, m.Member, m.GuildID, d.cfg.Configs.Settings.BotAdminRole) {
					err := database.UpdateDBitems(d.dbClient, d.cfg)
					if err != nil {
						fmt.Printf("%+v", errors.WithStack(err))
						_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
					}
				} else {
					_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.NotBotAdmin)
					if err != nil {
						fmt.Printf("%+v", errors.WithStack(err))
						_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
					}
				}
			}

		/////////////Misc///////////////////

		case "weast":
			err := d.sendWeasterEgg(s, m)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
			}

		case "palindrome":
			err := d.checkPalindrome(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
			}

		case "romans":
			err := d.romanNums(s, m, param)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
			}

		/////////////Misc///////////////////

		/////////////Games///////////////////

		/////////////Games///////////////////

		/////////////Audio///////////////////

		// Play audio
		case "play":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				err := d.playAudioLink(s, m, param)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

		case "stop":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				err := d.stopAudioPlayback()
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

		case "queue":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				err := d.sendQueue(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

		case "skip":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				err := d.sendSkipMessage(s, m)
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

		case "clear":
			if m.GuildID == d.cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == d.cfg.Configs.DiscordIDs.TestGuildID {
				err := web.MpFileCleanUp(fmt.Sprintf("%s/Audio", m.GuildID))
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
				}

				_, err = s.ChannelMessageSend(m.ChannelID, "\"This house is clean.\"")
				if err != nil {
					fmt.Printf("%+v", errors.WithStack(err))
					_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
				}
			}
			return

		/////////////NSFW///////////////////

		// Sends the "Invalid" command Message
		default:
			_, err := s.ChannelMessageSend(m.ChannelID, d.cfg.Cmd.Msg.Invalid)
			if err != nil {
				fmt.Printf("%+v", errors.WithStack(err))
				_, _ = s.ChannelMessageSendEmbed(d.cfg.Configs.DiscordIDs.ErrorLogChannelID, helper.GetErrorEmbed(err, s, m.GuildID))
			}
			return
		}
	}
}
