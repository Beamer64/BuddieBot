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
	"cistercian",
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

		// ------------Dev-------------

		case "test":
			err := testMethod(s, m, cfg, param)
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
				err := playAudioLink(s, m, cfg, param)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "stop":
			if isAudioGuild(m.GuildID, cfg) {
				err := stopAudioPlayback(s, m, cfg)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "queue":
			if isAudioGuild(m.GuildID, cfg) {
				err := sendQueue(s, m, cfg)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "skip":
			if isAudioGuild(m.GuildID, cfg) {
				err := skipPlayback(s, m, cfg)
				if err != nil {
					helper.LogErrorsToErrorChannel(s, cfg.Configs.DiscordIDs.ErrorLogChannelID, err, m.GuildID)
				}
			}

		case "clear":
			if isAudioGuild(m.GuildID, cfg) {
				err := clearQueue(s, m, cfg)
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
