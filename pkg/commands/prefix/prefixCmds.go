package prefix

import (
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

// Names lists every prefix command the bot dispatches. Used by the startup
// counter in bot.registerCommands. Keep this in sync with the switch in
// ParsePrefixCmds below — drift is caught by TestPrefixNamesMatchSwitch.
var Names = []string{
	"release",
	"weast",
	"palindrome",
	"romans",
}

func ParsePrefixCmds(s *discordgo.Session, m *discordgo.MessageCreate, cfg *config.Configs) {
	// If the message is a command
	if strings.HasPrefix(m.Content, cfg.Settings.BotPrefix) {

		messageSlices := strings.SplitAfterN(m.Content, " ", 2)
		command := strings.TrimPrefix(messageSlices[0], cfg.Settings.BotPrefix) // remove prefix
		command = strings.TrimSpace(command)                                    // remove trailing space from SplitAfterN

		// get command parameter
		param := ""
		if len(messageSlices) > 1 {
			param = messageSlices[1]
		}

		switch strings.ToLower(command) {

		// --------------------------------------------------------DEV--------------------------------------------------------------------

		// {prefix}release
		// sends the release notes to all the available guilds
		case "release":
			if m.GuildID == cfg.DiscordIDs.TestGuildID {
				if helper.MemberHasRole(s, m.Member, m.GuildID, cfg.Settings.BotAdminRole) {
					helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, sendReleaseNotes(s, m))
				} else {
					_, err := s.ChannelMessageSend(m.ChannelID, cfg.Cmd.Msg.NotBotAdmin)
					helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, err)
				}
			}

		// ----------------------------------------------------------Misc-----------------------------------------------------------------

		// {prefix}weast
		// Easter egg
		case "weast":
			helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, sendWeasterEgg(s, m))

		// {prefix}palindrome {text}
		// determines if the text is a palindrome
		case "palindrome":
			helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, checkPalindrome(s, m, param))

		// {prefix}romans {number}
		// converts numbers to the roman numeral equivalent
		case "romans":
			helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, romanNums(s, m, param))

		// Sends the "Invalid" command Message — the message itself is the
		// user feedback, so any send failure only needs to be logged.
		default:
			if _, err := s.ChannelMessageSend(m.ChannelID, cfg.Cmd.Msg.Invalid); err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
			}
		}
	}
}
