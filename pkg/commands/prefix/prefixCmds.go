package prefix

import (
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

// Names is consumed by the startup counter. Drift vs the ParsePrefixCmds
// switch is caught by TestPrefixNamesMatchSwitch.
var Names = []string{
	"release",
	"weast",
	"palindrome",
	"romans",
}

func ParsePrefixCmds(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs) {
	if strings.HasPrefix(m.Content, cfg.Settings.BotPrefix) {
		messageSlices := strings.SplitAfterN(m.Content, " ", 2)
		command := strings.TrimPrefix(messageSlices[0], cfg.Settings.BotPrefix)
		command = strings.TrimSpace(command)

		param := ""
		if len(messageSlices) > 1 {
			param = messageSlices[1]
		}

		switch strings.ToLower(command) {
		case "release":
			// Test-guild only, admin-gated — sends release notes to every guild.
			if m.GuildID == cfg.DiscordIDs.TestGuildID {
				if helper.MemberHasRole(s, m.Member, m.GuildID, cfg.Settings.BotAdminRole) {
					helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, sendReleaseNotes(s, m))
				} else {
					_, err := s.ChannelMessageSend(m.ChannelID, cfg.Cmd.Msg.NotBotAdmin)
					helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, err)
				}
			}

		case "weast":
			helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, sendWeasterEgg(s, m))

		case "palindrome":
			helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, checkPalindrome(s, m, param))

		case "romans":
			helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, romanNums(s, m, param))

		default:
			// The "Invalid" message IS the user feedback; send failure only needs logging.
			if _, err := s.ChannelMessageSend(m.ChannelID, cfg.Cmd.Msg.Invalid); err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
			}
		}
	}
}
