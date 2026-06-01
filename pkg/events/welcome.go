package events

import (
	"errors"
	"fmt"

	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

// sendWelcomeMessage posts the introductory embed in the guild after the bot
// is first added (or re-added — see WelcomeNeeded). Picks a channel via
// pickWelcomeChannel; returns an error if no channel is writable so the
// caller can log without blocking. Caller is expected to swallow errors —
// failing to send a welcome is annoying but not fatal.
func sendWelcomeMessage(s *discordgo.Session, e *discordgo.GuildCreate) error {
	if s.State == nil || s.State.User == nil {
		return errors.New("welcome: session state not ready")
	}
	botID := s.State.User.ID
	canWrite := func(channelID string) bool {
		perms, err := s.State.UserChannelPermissions(botID, channelID)
		if err != nil {
			return false
		}
		return perms&discordgo.PermissionSendMessages != 0
	}

	channelID := pickWelcomeChannel(e.Guild, canWrite)
	if channelID == "" {
		return errors.New("welcome: no writable channel")
	}

	if _, err := s.ChannelMessageSendEmbed(channelID, welcomeEmbed(e.Guild.Name, s.State.User)); err != nil {
		return fmt.Errorf("welcome: send embed to %s: %w", channelID, err)
	}
	return nil
}

// pickWelcomeChannel picks a channel to post the welcome message in. Prefers
// the guild's designated system channel (Discord's join/leave/boost channel —
// the natural fit); falls back to the first text channel the bot can write
// in. Returns "" if no channel is writable, in which case the caller skips
// posting rather than spamming random channels.
//
// canWrite is parameterized so tests can stub permission checks without
// constructing a real discordgo.Session state.
func pickWelcomeChannel(g *discordgo.Guild, canWrite func(channelID string) bool) string {
	if g == nil {
		return ""
	}
	if g.SystemChannelID != "" && canWrite(g.SystemChannelID) {
		return g.SystemChannelID
	}
	for _, ch := range g.Channels {
		if ch == nil || ch.Type != discordgo.ChannelTypeGuildText {
			continue
		}
		if canWrite(ch.ID) {
			return ch.ID
		}
	}
	return ""
}

// welcomeEmbed builds the introductory message. The layout uses zero-width
// spacer fields to force a 3-rows-of-2 inline grid (Discord auto-fits 3 per
// row otherwise). Then two full-width fields for Getting Started and Privacy.
func welcomeEmbed(guildName string, botUser *discordgo.User) *discordgo.MessageEmbed {
	thumb := ""
	if botUser != nil {
		thumb = botUser.AvatarURL("256")
	}

	return &discordgo.MessageEmbed{
		Title: "👋 Hi, I'm BuddieBot",
		Description: fmt.Sprintf(
			"Thanks for adding me to **%s**. I'm a hobby-started Discord bot focused on "+
				"general use functionality and a grab-bag of fun commands. I am always adding new "+
				"features so lookout for new content. Here's some of what's inside:",
			guildName,
		),
		Color:     helper.RandomDiscordColor(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: thumb},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🎨 Image Effects",
				Value:  "60+ filters, distortions, and meme templates. Stylize avatars, generate text signs, animate things. All under `/image`.",
				Inline: true,
			},
			{
				Name:   "🎲 Games",
				Value:  "Would-you-rather rounds, polls, random picks from a list, and other quick mini-games.",
				Inline: true,
			},
			{
				Name:   "📅 Daily",
				Value:  "Horoscopes, Kanye quotes, affirmations, fun facts, tongue twisters, terrible advice. One per call via `/daily`.",
				Inline: true,
			},
			{
				Name:   "👤 Profiles & Ratings",
				Value:  "`/user profile` shows your stats, balance, command history, and ratings (More in the works). `/rate-this` rates your friends on (made-up) metrics.",
				Inline: true,
			},
			{
				Name:   "😂 Random",
				Value:  "Jokes, yo-momma burns, pickup lines, 8-ball answers, XKCD comics. All under `/get`.",
				Inline: true,
			},
			{
				Name:   "⚙️ For the Admins",
				Value:  "`/admin set-prefix` changes my command prefix from the default `$`. Some prefix commands like `$goodboy` are available alongside slash commands.",
				Inline: true,
			},
			{
				Name:   "🚀 Getting Started",
				Value:  "Run `/tuuck` to browse every command with usage examples. It's the best place to see what each command actually does before you try it.",
				Inline: false,
			},
			{
				Name:   "🔒 Privacy",
				Value:  "I store some minor data like user statistics and ratings. `/user forget-me` permanently deletes all of it.",
				Inline: false,
			},
			{
				Name:   "💬 Feedback",
				Value:  "Found a bug or have a suggestion? Run `/feedback` and pick a category — it routes straight to the maintainer.",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: "BuddieBot"},
	}
}
