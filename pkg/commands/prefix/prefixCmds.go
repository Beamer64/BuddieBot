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

// NoShowCmds is consumed by the startup counter. Drift vs the ParsePrefixCmds
// switch is caught by TestPrefCmdsListMatchesSwitch; typos here vs
// PrefCmdsList by TestNoShowCmdsAreKnown.
var NoShowCmds = []string{
	"weast",
}
var PrefCmdsList = map[string]string{
	"weast":      "A secret Easter egg command.",
	"palindrome": "Determines if the string is palindrome. Made for a coding challenge.",
	"romans":     "Converts numbers into the roman numeral equivalent.",
	"goodboy":    "The bestest boy I have ever known 🐶",
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

	command, param, ok := splitPrefixCommand(m.Content, prefix)
	if !ok {
		return
	}

	cmdLower := strings.ToLower(command)
	switch cmdLower {
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
	case "test":
		// Test-guild only, admin-gated — sends release notes to every guild.
		if m.GuildID == cfg.DiscordIDs.TestGuildID {
			if helper.MemberHasRole(s, m.Member, m.GuildID, cfg.Settings.BotAdminRole) {
				helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, testFeature(s, m, cfg))
			} else {
				_, sendErr := s.ChannelMessageSend(m.ChannelID, "You dont have permission to use this command.")
				helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, sendErr)
			}
		}

	case "weast":
		helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, sendWeasterEgg(s, m))

	case "goodboy":
		helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, sendGoodBoy(s, m))

	case "palindrome":
		helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, checkPalindrome(s, m, param))

	case "romans":
		helper.LogAndReact(s, m, cfg.DiscordIDs.ErrorLogChannelID, romanNums(s, m, param))

	default:
		// The "Invalid" message IS the user feedback; send failure only needs logging.
		if _, sendErr := s.ChannelMessageSend(m.ChannelID, "Invalid prefix command."); sendErr != nil {
			helper.LogErrorsToErrorChannel(s, cfg.DiscordIDs.ErrorLogChannelID, sendErr, m.GuildID)
		}
		// Unknown command — don't count typos as invocations.
		return
	}

	// Track AFTER dispatch — any case that fell through to here is a recognized
	// command. Adding a new case above gets tracking for free. The "$" key is
	// canonical regardless of the guild's actual prefix; /user profile swaps
	// it for the guild's prefix at render time.
	userID := ""
	if m.Author != nil && !m.Author.Bot {
		userID = m.Author.ID
	}
	trackCtx, trackCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer trackCancel()
	if err := cfg.DB.TrackCommandInvocation(trackCtx, "$"+cmdLower, m.GuildID, userID); err != nil {
		log.Printf("track prefix command %q: %v", cmdLower, err)
	}
}

// splitPrefixCommand parses a raw message body into (command, param) given
// the guild's current prefix. Matching is STRICT — the prefix is consumed
// literally, and any extra whitespace between the prefix and the command
// word means the user didn't intend a prefix invocation. ok=false when the
// content doesn't match, the bot stays silent.
//
//	prefix "$"   + content "$roman 5"      → "roman", "5"   ✓
//	prefix "$"   + content "$ roman 5"     → ok=false       ✗ (stray space)
//	prefix "plz "+ content "plz roman 5"   → "roman", "5"   ✓
//	prefix "plz "+ content "plzroman 5"    → ok=false       ✗ (no space)
//	prefix "plz "+ content "plz  roman 5"  → ok=false       ✗ (double space)
//	either       + content == prefix       → ok=false       ✗ (empty command)
//
// param keeps its original whitespace so commands like palindrome see the
// user's exact input.
func splitPrefixCommand(content, prefix string) (command, param string, ok bool) {
	if !strings.HasPrefix(content, prefix) {
		return "", "", false
	}
	afterPrefix := content[len(prefix):]
	if afterPrefix == "" || afterPrefix[0] == ' ' {
		return "", "", false
	}
	parts := strings.SplitN(afterPrefix, " ", 2)
	command = parts[0]
	if len(parts) > 1 {
		param = parts[1]
	}
	return command, param, true
}
