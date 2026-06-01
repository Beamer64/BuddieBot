package slash

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

// feedbackLimiter throttles /feedback to one submission per user per hour.
// The destination is a maintainer-visible channel; the cap keeps that channel
// readable when a single user gets feature-request happy.
var feedbackLimiter = helper.NewRateLimiter(time.Hour)

// feedbackMinBodyLen is enforced both client-side (via the spec's MinLength)
// and server-side (here) — Discord's client validation can be bypassed by a
// custom client, so we never trust the wire.
const feedbackMinBodyLen = 30

// feedbackMaxBodyLen caps the details field so a single submission can't
// dominate the destination channel. Well under Discord's 6000-char option
// limit but generous enough for paragraphs of useful detail.
const feedbackMaxBodyLen = 1000

func sendFeedbackResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	submitter := invokerUser(i)
	if submitter == nil {
		return errors.New("feedback: cannot identify invoker")
	}

	// Rate-limit BEFORE any work — pre-defer ephemeral response slot.
	if ok, retry := feedbackLimiter.Allow(submitter.ID); !ok {
		return helper.SendEphemeralMsgPreDeferred(s, i,
			fmt.Sprintf("One feedback per hour, please. Try again in `%s`.", formatRetryAfter(retry)))
	}

	var category, body string
	for _, opt := range i.ApplicationCommandData().Options {
		switch opt.Name {
		case "category":
			category = opt.StringValue()
		case "details":
			body = strings.TrimSpace(opt.StringValue())
		}
	}

	// Server-side validation guards against a misbehaving client. The spec's
	// MinLength/MaxLength enforce the same in the Discord UI for everyone else.
	if len(body) < feedbackMinBodyLen {
		return helper.SendEphemeralMsgPreDeferred(s, i,
			fmt.Sprintf("Your feedback needs at least `%d` characters of detail (got `%d`).", feedbackMinBodyLen, len(body)))
	}

	channelID := feedbackChannelFor(category, cfg.DiscordIDs.BuddieBotHQSuggestionChannelID, cfg.DiscordIDs.BuddieBotHQBugChannelID)
	if channelID == "" {
		return helper.LogSendEphemeralMsgPreDeferred(s, i,
			"Feedback isn't configured for that category yet — try a different one.",
			fmt.Errorf("feedback channel for category %q is not configured", category))
	}

	// Best-effort guild lookup for context — embed handles nil cleanly.
	// State cache is free; we never fall back to a network fetch since the
	// extra round trip isn't worth blocking a feedback submission for.
	var originGuild *discordgo.Guild
	if i.GuildID != "" && s.State != nil {
		if g, _ := s.State.Guild(i.GuildID); g != nil {
			originGuild = g
		}
	}

	if _, err := s.ChannelMessageSendEmbed(channelID, feedbackEmbed(category, body, submitter, originGuild)); err != nil {
		return helper.LogSendEphemeralMsgPreDeferred(s, i,
			"Couldn't deliver your feedback right now. Please try again later.",
			fmt.Errorf("feedback send to channel %s: %w", channelID, err))
	}

	return helper.SendEphemeralMsgPreDeferred(s, i,
		fmt.Sprintf("Thanks! Your %s has been sent through.", categoryLabel(category)))
}

// feedbackChannelFor maps a chosen category to the destination channel ID.
// Bug routes to bugChannelID; everything else (feature, other, plus any
// future category) routes to suggestionChannelID. Takes the IDs directly
// rather than the whole config so the function is trivially testable.
func feedbackChannelFor(category, suggestionChannelID, bugChannelID string) string {
	switch category {
	case "bug":
		return bugChannelID
	default:
		return suggestionChannelID
	}
}

// feedbackTitleAndColor maps a category to its display heading + embed color.
// Colors are picked to be glanceable in the destination channel: green for
// new ideas, red for bugs, neutral gray for "other" / fallback.
func feedbackTitleAndColor(category string) (string, int) {
	switch category {
	case "feature":
		return "💡 Feature Suggestion", 0x4ADE80
	case "bug":
		return "🐛 Bug Report", 0xEF4444
	case "other":
		return "💬 Other Feedback", 0x6B7280
	default:
		return "💬 Feedback", 0x6B7280
	}
}

// categoryLabel returns the user-facing noun used in the success message —
// "Thanks! Your bug report has been sent through."
func categoryLabel(category string) string {
	switch category {
	case "feature":
		return "feature suggestion"
	case "bug":
		return "bug report"
	default:
		return "feedback"
	}
}

// feedbackEmbed builds the message posted to the maintainer channel. The
// submitter's user ID appears both as a mention (clickable) and raw value
// (copy-pasteable for DBeaver / log lookups). originGuild may be nil — DM
// context renders "(direct message)" instead.
func feedbackEmbed(category, body string, submitter *discordgo.User, originGuild *discordgo.Guild) *discordgo.MessageEmbed {
	title, color := feedbackTitleAndColor(category)

	guildLine := "(direct message)"
	if originGuild != nil {
		guildLine = fmt.Sprintf("%s · `%s`", originGuild.Name, originGuild.ID)
	}

	avatar := ""
	if submitter != nil {
		avatar = submitter.AvatarURL("64")
	}
	username := ""
	submitterID := ""
	if submitter != nil {
		username = submitter.Username
		submitterID = submitter.ID
	}

	return &discordgo.MessageEmbed{
		Title: title,
		Color: color,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    username,
			IconURL: avatar,
		},
		Description: body,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "From", Value: fmt.Sprintf("<@%s> · `%s`", submitterID, submitterID), Inline: false},
			{Name: "Guild", Value: guildLine, Inline: false},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// formatRetryAfter renders a duration in a friendly form for the rate-limit
// message. Rounds to whole minutes when retry > 90s so users don't see
// "59m 33s remaining"-style noise.
func formatRetryAfter(d time.Duration) string {
	if d > 90*time.Second {
		return d.Round(time.Minute).String()
	}
	return d.Round(time.Second).String()
}

// invokerUser returns the user who triggered the interaction — i.Member.User
// in guild context, i.User in DM. Returns nil if neither is set (Discord
// shouldn't allow this, but defending against it keeps the handler honest).
func invokerUser(i *discordgo.InteractionCreate) *discordgo.User {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User
	}
	return i.User
}

func feedbackSpec() *discordgo.ApplicationCommand {
	minLen := feedbackMinBodyLen
	return &discordgo.ApplicationCommand{
		Name:        "feedback",
		Description: "Send a feature suggestion, bug report, or other feedback to the maintainer.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "category",
				Description: "What kind of feedback?",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Feature suggestion", Value: "feature"},
					{Name: "Bug report", Value: "bug"},
					{Name: "Other", Value: "other"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "details",
				Description: "Describe your feedback — at least 30 characters of useful detail.",
				Required:    true,
				MinLength:   &minLen,
				MaxLength:   feedbackMaxBodyLen,
			},
		},
	}
}
