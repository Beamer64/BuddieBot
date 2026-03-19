package prefix

import (
	"fmt"
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/BuddieBot/pkg/web"
	"github.com/bwmarrin/discordgo"
)

// isAudioGuild checks if the message is from a guild that has audio commands enabled
func isAudioGuild(guildID string, cfg *config.Configs) bool {
	return guildID == cfg.Configs.DiscordIDs.MasterGuildID || guildID == cfg.Configs.DiscordIDs.TestGuildID
}

func ParsePrefixCmds(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs) {
	// If the message is a command
	if strings.HasPrefix(m.Content, cfg.Configs.Settings.BotPrefix) {

		messageSlices := strings.SplitAfterN(m.Content, " ", 2)
		command := strings.TrimPrefix(messageSlices[0], cfg.Configs.Settings.BotPrefix) // remove prefix
		command = strings.TrimSpace(command)                                            // remove trailing space from SplitAfterN

		// get command parameter
		param := ""
		if len(messageSlices) > 1 {
			param = messageSlices[1]
		}

		switch strings.ToLower(command) {

		// ------------Dev-------------

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

		// -------------Misc-------------

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

		// -------------Audio-------------

		case "play":
			if isAudioGuild(m.GuildID, cfg) {
				err := playAudioLink(s, m, param)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "stop":
			if isAudioGuild(m.GuildID, cfg) {
				err := stopAudioPlayback()
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "queue":
			if isAudioGuild(m.GuildID, cfg) {
				err := sendQueue(s, m)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "skip":
			if isAudioGuild(m.GuildID, cfg) {
				err := sendSkipMessage(s, m)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "clear":
			if isAudioGuild(m.GuildID, cfg) {
				err := web.MpFileCleanUp(fmt.Sprintf("%s/Audio", m.GuildID))
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}

				_, err = s.ChannelMessageSend(m.ChannelID, "\"This house is clean.\"")
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		// Sends the "Invalid" command Message
		default:
			_, err := s.ChannelMessageSend(m.ChannelID, cfg.Cmd.Msg.Invalid)
			if err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
			}
		}
	}
}
