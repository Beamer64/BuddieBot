package prefix

import (
	"fmt"
	"github.com/beamer64/buddieBot/pkg/config"
	"github.com/beamer64/buddieBot/pkg/helper"
	"github.com/beamer64/buddieBot/pkg/web"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func ParsePrefixCmds(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs) {
	// If the message is a command
	if strings.HasPrefix(m.Content, cfg.Configs.Settings.BotPrefix) {

		messageSlices := strings.SplitAfterN(m.Content, " ", 2)                                       // split command from param
		command := strings.Trim(messageSlices[0], fmt.Sprintf("%s ", cfg.Configs.Settings.BotPrefix)) // trim spaces and prefix

		// get command parameter
		param := ""
		if len(messageSlices) > 1 {
			param = messageSlices[1]
		}

		switch strings.ToLower(command) {

		// ///////////Dev///////////////////
		case "test":
			err := testMethod(s, m, param)
			if err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
			}

		case "release":
			if m.GuildID == cfg.Configs.DiscordIDs.TestGuildID {
				if helper.MemberHasRole(s, m.Member, m.GuildID, cfg.Configs.Settings.BotAdminRole) {
					err := sendReleaseNotes(s, m)
					if err != nil {
						helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
					}

				} else {
					_, err := s.ChannelMessageSend(m.ChannelID, cfg.Cmd.Msg.NotBotAdmin)
					if err != nil {
						helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
					}
				}
			}

		case "updatedbitems":
			if m.GuildID == cfg.Configs.DiscordIDs.TestGuildID {
				if helper.MemberHasRole(s, m.Member, m.GuildID, cfg.Configs.Settings.BotAdminRole) {

				} else {
					_, err := s.ChannelMessageSend(m.ChannelID, cfg.Cmd.Msg.NotBotAdmin)
					if err != nil {
						helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
					}
				}
			}

		// ///////////Misc///////////////////

		case "cistercian":
			err := sendCistercianNumeral(s, m, cfg, param)
			if err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
			}

		case "weast":
			err := sendWeasterEgg(s, m)
			if err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
			}

		case "palindrome":
			err := checkPalindrome(s, m, param)
			if err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
			}

		case "romans":
			err := romanNums(s, m, param)
			if err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
			}

		// ///////////Misc///////////////////

		// ///////////Games///////////////////

		// ///////////Games///////////////////

		// ///////////Audio///////////////////

		// Play audio
		case "play":
			if m.GuildID == cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == cfg.Configs.DiscordIDs.TestGuildID {
				err := playAudioLink(s, m, param)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "stop":
			if m.GuildID == cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == cfg.Configs.DiscordIDs.TestGuildID {
				err := stopAudioPlayback()
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "queue":
			if m.GuildID == cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == cfg.Configs.DiscordIDs.TestGuildID {
				err := sendQueue(s, m)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "skip":
			if m.GuildID == cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == cfg.Configs.DiscordIDs.TestGuildID {
				err := sendSkipMessage(s, m)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "clear":
			if m.GuildID == cfg.Configs.DiscordIDs.MasterGuildID || m.GuildID == cfg.Configs.DiscordIDs.TestGuildID {
				err := web.MpFileCleanUp(fmt.Sprintf("%s/Audio", m.GuildID))
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}

				_, err = s.ChannelMessageSend(m.ChannelID, "\"This house is clean.\"")
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		// ///////////NSFW///////////////////

		// Sends the "Invalid" command Message
		default:
			_, err := s.ChannelMessageSend(m.ChannelID, cfg.Cmd.Msg.Invalid)
			if err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
			}
		}

		// If the message is not a command
	} else {
		// moderate NSFW content
		err := modNSFWimgs(s, m, cfg)
		if err != nil {
			helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
		}

	}
}
