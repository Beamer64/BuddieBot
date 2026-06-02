package slash

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Beamer64/BuddieBot/pkg/config"
	"github.com/Beamer64/BuddieBot/pkg/database"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/bwmarrin/discordgo"
)

func sendAdminResponse(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	// Admin actions are ephemeral — no need to broadcast config changes.
	if err := s.InteractionRespond(
		i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
		},
	); err != nil {
		return fmt.Errorf("failed to defer interaction for /admin: %w", err)
	}

	sub := i.ApplicationCommandData().Options[0]
	switch sub.Name {
	case "set-prefix":
		return adminSetPrefix(s, i, cfg, sub.Options)
	case "ban":
		return adminBan(s, i, cfg, sub.Options)
	case "unban":
		return adminUnban(s, i, cfg, sub.Options)
	case "banned-list":
		return adminBannedList(s, i, cfg)
	default:
		return helper.LogSendEphemeralFollowUpPostDeferred(s, i, "Unknown admin subcommand.", fmt.Errorf("unknown admin subcommand: %s", sub.Name))
	}
}

func adminSetPrefix(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
	// new-prefix is optional; omitting (or blanking) it resets to the default.
	newPrefix := ""
	trailingSpace := false
	for _, opt := range opts {
		switch opt.Name {
		case "new-prefix":
			newPrefix = strings.TrimLeft(opt.StringValue(), " ")
		case "trailing-space":
			trailingSpace = opt.BoolValue()
		}
	}

	// Discord strips trailing whitespace from String options at the client
	// layer, so a "plz " style prefix can't arrive in newPrefix directly. The
	// trailing-space boolean is the user's opt-in to that mode — append the
	// space here, BEFORE validation, so the 5-char cap counts it.
	if trailingSpace && newPrefix != "" && !strings.HasSuffix(newPrefix, " ") {
		newPrefix += " "
	}

	if newPrefix != "" {
		// 5-char max applies to the full prefix INCLUDING trailing spaces —
		// keeps the rendered prefix short regardless of style choice.
		if len(newPrefix) > 5 {
			return helper.EditMsgPostDeferred(s, i, "Prefix must be 5 characters or fewer.")
		}
		if strings.ContainsAny(newPrefix, "\t\n\r") {
			return helper.EditMsgPostDeferred(s, i, "Prefix can't contain tabs or newlines.")
		}
		// Trailing spaces are allowed — e.g. setting "plz " lets users
		// type "plz roman" naturally. Embedded spaces are not.
		if strings.Contains(strings.TrimRight(newPrefix, " "), " ") {
			return helper.EditMsgPostDeferred(s, i, "Spaces are only allowed at the end of the prefix.")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cfg.DB.SetGuildPrefixOverride(ctx, i.GuildID, newPrefix); err != nil {
		return helper.LogSendEphemeralFollowUpPostDeferred(s, i, "Couldn't update the prefix.", fmt.Errorf("set prefix: %w", err))
	}

	content := fmt.Sprintf("Prefix set to `%s` for this server.", newPrefix)
	if newPrefix == "" {
		content = "Prefix reset to the default `$`."
		newPrefix = "$"
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content}); err != nil {
		return fmt.Errorf("send /admin set-prefix response: %w", err)
	}

	if err := sendPrefixUpdateMsg(s, i, newPrefix); err != nil {
		return fmt.Errorf("send /admin sendPrefixUpdateMsg: %w", err)
	}

	return nil
}

func sendPrefixUpdateMsg(s *discordgo.Session, i *discordgo.InteractionCreate, newPrefix string) error {
	// Discord visually collapses trailing whitespace inside single backticks,
	// so when the prefix ends in a space the rendered `plz ` looks identical
	// to `plz`. Call that out in plain text so members know to type the space.
	extra := ""
	if strings.HasSuffix(newPrefix, " ") {
		extra = " (the trailing space is required)"
	}

	desc := fmt.Sprintf(
		"A server admin has updated the BuddieBot command prefix to `%s`%s.\nExample Usage: `%sgoodboy`",
		newPrefix, extra, newPrefix,
	)

	_, err := s.ChannelMessageSendComplex(
		i.ChannelID, &discordgo.MessageSend{
			Content: "@everyone",
			Embed: &discordgo.MessageEmbed{
				Color:       helper.RandomDiscordColor(),
				Title:       "(°ロ°)👇",
				Description: desc,
			},
		},
	)

	return err
}

// /admin ban — issues (or updates) a global ban. Server-admin gated via the
// parent /admin's DefaultMemberPermissions. The ban applies in every guild
// the bot serves; bans are stored globally in BannedUser.
//
// User data is intentionally NOT touched — profile, ratings, and command
// history all survive so an unban resumes from where the user left off.
func adminBan(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
	var targetID, reason string
	for _, opt := range opts {
		switch opt.Name {
		case "target":
			if u := opt.UserValue(s); u != nil {
				targetID = u.ID
			}
		case "reason":
			reason = strings.TrimSpace(opt.StringValue())
		}
	}
	if targetID == "" {
		return helper.EditMsgPostDeferred(s, i, "No target user specified.")
	}

	// Sanity guards — these are user-facing status messages, not system
	// failures, so they stay public-via-defer (no Log… helper).
	invokerID := ""
	if i.Member != nil && i.Member.User != nil {
		invokerID = i.Member.User.ID
	}
	if targetID == invokerID {
		return helper.EditMsgPostDeferred(s, i, "You can't ban yourself.")
	}
	if s.State != nil && s.State.User != nil && targetID == s.State.User.ID {
		return helper.EditMsgPostDeferred(s, i, "I can't ban myself — that would be problematic.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	created, err := cfg.DB.BanUser(ctx, targetID, reason, invokerID)
	if err != nil {
		return helper.LogSendEphemeralFollowUpPostDeferred(s, i, "Couldn't ban that user.", err)
	}

	verb := "Ban issued"
	if !created {
		verb = "Ban updated"
	}
	msg := fmt.Sprintf("%s for <@%s>.", verb, targetID)
	if reason != "" {
		msg += " Reason: " + reason
	}
	if err := helper.EditMsgPostDeferred(s, i, msg); err != nil {
		return err
	}

	// Audit notification — best-effort, never blocks. Empty channel ID
	// (unconfigured) just skips silently.
	postBanAudit(s, cfg.DiscordIDs.BuddieBotHQBanChannelID, banAuditBan(created), targetID, invokerID, reason)
	return nil
}

// /admin unban — removes the ban row. The user's stored profile, ratings,
// and command counts are untouched, so they pick up where they left off.
func adminUnban(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
	var targetID string
	for _, opt := range opts {
		if opt.Name == "target" {
			if u := opt.UserValue(s); u != nil {
				targetID = u.ID
			}
		}
	}
	if targetID == "" {
		return helper.EditMsgPostDeferred(s, i, "No target user specified.")
	}

	invokerID := ""
	if i.Member != nil && i.Member.User != nil {
		invokerID = i.Member.User.ID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	removed, err := cfg.DB.UnbanUser(ctx, targetID)
	if err != nil {
		return helper.LogSendEphemeralFollowUpPostDeferred(s, i, "Couldn't unban that user.", err)
	}
	if !removed {
		return helper.EditMsgPostDeferred(s, i, fmt.Sprintf("<@%s> isn't currently banned.", targetID))
	}

	if err := helper.EditMsgPostDeferred(s, i, fmt.Sprintf("Unbanned <@%s>. Their stored data is untouched.", targetID)); err != nil {
		return err
	}

	postBanAudit(s, cfg.DiscordIDs.BuddieBotHQBanChannelID, "unban", targetID, invokerID, "")
	return nil
}

// /admin banned-list — server admins see banned users who have invoked the
// bot in THIS guild. Per-guild filter happens at the DB layer via a JOIN
// on User; users banned globally but with no history here don't appear.
func adminBannedList(s *discordgo.Session, i *discordgo.InteractionCreate, cfg *config.Configs) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := cfg.DB.ListBannedUsersInGuild(ctx, i.GuildID)
	if err != nil {
		return helper.LogSendEphemeralFollowUpPostDeferred(s, i, "Couldn't load the banned list.", err)
	}

	if len(rows) == 0 {
		return helper.EditMsgPostDeferred(s, i, "No banned users have used the bot in this server.")
	}

	embed := buildBannedListEmbed(s, rows)
	embeds := []*discordgo.MessageEmbed{embed}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds}); err != nil {
		return fmt.Errorf("send /admin banned-list: %w", err)
	}
	return nil
}

// bannedListMaxFields caps the embed at Discord's 25-field hard limit. If we
// ever exceed it, the renderer notes the overflow rather than truncating
// silently — encourages adding pagination if real-world counts hit it.
const bannedListMaxFields = 25

func buildBannedListEmbed(s *discordgo.Session, rows []*database.BannedUser) *discordgo.MessageEmbed {
	total := len(rows)
	shown := rows
	overflow := 0
	if total > bannedListMaxFields {
		shown = rows[:bannedListMaxFields]
		overflow = total - bannedListMaxFields
	}

	fields := make([]*discordgo.MessageEmbedField, 0, len(shown))
	for _, b := range shown {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   bannedUserDisplayName(s, b.DiscordUserID),
			Value:  bannedListRowValue(b),
			Inline: false,
		})
	}

	title := fmt.Sprintf("Banned users in this server (%d)", total)
	embed := &discordgo.MessageEmbed{
		Title:  title,
		Color:  0xEF4444,
		Fields: fields,
	}
	if overflow > 0 {
		embed.Description = fmt.Sprintf("Showing %d of %d. Add pagination if this becomes routine.", bannedListMaxFields, total)
	}
	return embed
}

// bannedUserDisplayName tries the state cache first (free), falls back to a
// network fetch, and finally to "Unknown User · ID" so the row always has
// something useful in the name slot.
func bannedUserDisplayName(s *discordgo.Session, userID string) string {
	if s != nil {
		if u, err := s.User(userID); err == nil && u != nil {
			return fmt.Sprintf("%s · `%s`", u.Username, userID)
		}
	}
	return fmt.Sprintf("Unknown User · `%s`", userID)
}

func bannedListRowValue(b *database.BannedUser) string {
	reason := "(no reason given)"
	if b.Reason.Valid && b.Reason.String != "" {
		reason = b.Reason.String
	}
	issuer := "—"
	if b.BannedBy.Valid && b.BannedBy.String != "" {
		issuer = fmt.Sprintf("<@%s>", b.BannedBy.String)
	}
	when := b.BannedAt
	if t, err := time.Parse("2006-01-02 15:04:05", b.BannedAt); err == nil {
		when = fmt.Sprintf("<t:%d:R>", t.Unix())
	}
	return fmt.Sprintf("**Reason:** %s\n**Banned by:** %s · %s", reason, issuer, when)
}

// banAuditBan returns the audit-action label for ban events: "ban" for a
// brand-new row, "ban-updated" when an existing ban's metadata was
// refreshed (re-issued with new reason/issuer).
func banAuditBan(created bool) string {
	if created {
		return "ban"
	}
	return "ban-updated"
}

// postBanAudit posts the audit embed to the configured channel. Best-effort:
// any failure is silently swallowed (the ban itself already succeeded, and
// audit-channel issues shouldn't block the primary action).
func postBanAudit(s *discordgo.Session, channelID, action, targetID, issuerID, reason string) {
	if channelID == "" {
		return
	}
	title, color := banAuditTitleAndColor(action)

	userField := bannedUserDisplayName(s, targetID)
	fields := []*discordgo.MessageEmbedField{
		{Name: "User", Value: userField, Inline: false},
	}
	if reason != "" {
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Reason", Value: reason, Inline: false})
	}
	if issuerID != "" {
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Issued by", Value: fmt.Sprintf("<@%s>", issuerID), Inline: false})
	}

	embed := &discordgo.MessageEmbed{
		Title:     title,
		Color:     color,
		Fields:    fields,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	_, _ = s.ChannelMessageSendEmbed(channelID, embed)
}

// banAuditTitleAndColor maps an audit action to its display heading + color.
// Red for bans, amber for ban-updates (metadata refresh), green for unbans —
// glanceable at a scroll-through of the audit channel.
func banAuditTitleAndColor(action string) (string, int) {
	switch action {
	case "ban":
		return "🚫 User Banned", 0xEF4444
	case "ban-updated":
		return "🔁 Ban Updated", 0xF59E0B
	case "unban":
		return "✅ User Unbanned", 0x4ADE80
	default:
		return "Ban Event", 0x6B7280
	}
}

func adminSpec() *discordgo.ApplicationCommand {
	perm := int64(discordgo.PermissionManageGuild)
	return &discordgo.ApplicationCommand{
		Name:                     "admin",
		Description:              "Server-admin configuration commands",
		Contexts:                 helper.GuildOnly,
		DefaultMemberPermissions: &perm,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set-prefix",
				Description: "Set this server's $-command prefix. Call this where everyone can see.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Name:        "trailing-space",
						Description: "Require a space between prefix and command? (e.g. 'plz roman') Off by default.",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "new-prefix",
						Description: "New prefix; leave empty to reset to the default ($)",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "ban",
				Description: "Ban a user from using BuddieBot. Their stored data is preserved.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "target",
						Description: "User to ban",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "reason",
						Description: "Why is this user being banned? (optional)",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "unban",
				Description: "Unban a user. Their stored data resumes from where it left off.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "target",
						Description: "User to unban",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "banned-list",
				Description: "List banned users who have used the bot in this server.",
			},
		},
	}
}
