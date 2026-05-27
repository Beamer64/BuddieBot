package prefix

import (
	"context"
	"log"
	"strings"
	"time"

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
	// Fast path: a warm cache is a plain map read — no context, no DB
	prefix, ok := cfg.DB.CachedGuildPrefix(m.GuildID)
	if !ok {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		var err error
		prefix, err = cfg.DB.GetGuildPrefixOverride(ctx, m.GuildID)
		cancel()
		if err != nil {
			// prefix is still a usable default here. Log locally rather than
			// posting to the channel or error log on every message during a hiccup.
			log.Printf("prefix lookup for guild %s failed, using %q: %v", m.GuildID, prefix, err)
		}
	}

	if !strings.HasPrefix(m.Content, prefix) {
		return
	}

	messageSlices := strings.SplitAfterN(m.Content, " ", 2)
	command := strings.TrimSpace(strings.TrimPrefix(messageSlices[0], prefix))

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
				_, sendErr := s.ChannelMessageSend(m.ChannelID, "You dont have permission to use this command.")
				helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, sendErr)
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
		if _, sendErr := s.ChannelMessageSend(m.ChannelID, "Invalid prefix command."); sendErr != nil {
			helper.LogErrorsToErrorChannel(s, cfg.DiscordIDs.ErrorLogChannelID, sendErr, m.GuildID)
		}
	}
}
