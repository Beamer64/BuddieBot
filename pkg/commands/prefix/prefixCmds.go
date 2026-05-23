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
	"test",
	"release",
	"weast",
	"palindrome",
	"romans",
	"play",
	"stop",
	"queue",
	"skip",
	"clear",
}

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

		// --------------------------------------------------------DEV--------------------------------------------------------------------

		// {prefix}test
		// used to test specific commands or scenarios
		case "test":
			helper.LogAndReact(s, m, cfg, testMethod(s, m, cfg, param))

		// {prefix}release
		// sends the release notes to all the available guilds
		case "release":
			if m.GuildID == cfg.Configs.DiscordIDs.TestGuildID {
				if helper.MemberHasRole(s, m.Member, m.GuildID, cfg.Configs.Settings.BotAdminRole) {
					helper.LogAndReact(s, m, cfg, sendReleaseNotes(s, m))
				} else {
					_, err := s.ChannelMessageSend(m.ChannelID, cfg.Cmd.Msg.NotBotAdmin)
					helper.LogAndReact(s, m, cfg, err)
				}
			}

		// ----------------------------------------------------------Misc-----------------------------------------------------------------

		// {prefix}weast
		// Easter egg
		case "weast":
			helper.LogAndReact(s, m, cfg, sendWeasterEgg(s, m))

		// {prefix}palindrome {text}
		// determines if the text is a palindrome
		case "palindrome":
			helper.LogAndReact(s, m, cfg, checkPalindrome(s, m, param))

		// {prefix}romans {number}
		// converts numbers to the roman numeral equivalent
		case "romans":
			helper.LogAndReact(s, m, cfg, romanNums(s, m, param))

		// ----------------------------------------------------------Audio-----------------------------------------------------------------

		// {prefix}play {audio url}
		// plays the audio url in the same voice channel as the person who triggered
		case "play":
			if isAudioGuild(m.GuildID, cfg) {
				helper.LogAndReact(s, m, cfg, playAudioLink(s, m, cfg, param))
			}

		// {prefix}stop
		// stops playing the current audio
		case "stop":
			if isAudioGuild(m.GuildID, cfg) {
				helper.LogAndReact(s, m, cfg, stopAudioPlayback(s, m, cfg))
			}

		// {prefix}queue
		// displays the audio queue
		case "queue":
			if isAudioGuild(m.GuildID, cfg) {
				helper.LogAndReact(s, m, cfg, sendQueue(s, m, cfg))
			}

		// {prefix}skip
		// skips to the next audio in the queue
		case "skip":
			if isAudioGuild(m.GuildID, cfg) {
				helper.LogAndReact(s, m, cfg, skipPlayback(s, m, cfg))
			}

		// {prefix}clear
		// clears the audio queue
		case "clear":
			if isAudioGuild(m.GuildID, cfg) {
				helper.LogAndReact(s, m, cfg, clearQueue(s, m, cfg))
			}

		// Sends the "Invalid" command Message — the message itself is the
		// user feedback, so any send failure only needs to be logged.
		default:
			if _, err := s.ChannelMessageSend(m.ChannelID, cfg.Cmd.Msg.Invalid); err != nil {
				helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
			}
		}
	}
}
